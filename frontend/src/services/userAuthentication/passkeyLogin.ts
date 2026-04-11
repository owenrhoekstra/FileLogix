import { base64ToUint8Array, uint8ArrayToBase64url } from './utilFunctions'
import { apiFetch } from '../logout/autoLogoutRedirect.ts'

type PublicKeyCredentialRequestOptionsJSON = {
    challenge: string
    rpId?: string
    allowCredentials?: Array<{
        id: string
        type: PublicKeyCredentialType
        transports?: AuthenticatorTransport[]
    }>
    userVerification?: UserVerificationRequirement
    timeout?: number
    [key: string]: any
}

type PasskeyLoginOptions = {
    options: {
        publicKey: PublicKeyCredentialRequestOptionsJSON
    }
    sessionId: string
}

export async function passkeyLogin(email: string): Promise<void> {
    const res = await apiFetch('/api/auth/passkey/login-challenge', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
    })

    if (!res) throw new Error('Backend not available')

    if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Login failed: ${errorText || res.statusText}`)
    }

    const { options, sessionId }: PasskeyLoginOptions = await res.json()

    // 🔧 normalize binary ONCE
    const challengeBytes = base64ToUint8Array(options.publicKey.challenge)

    const pk: PublicKeyCredentialRequestOptions = {
        challenge: challengeBytes.buffer as ArrayBuffer,
        rpId: options.publicKey.rpId,
        userVerification: options.publicKey.userVerification,
        timeout: options.publicKey.timeout
    }

    if (options.publicKey.allowCredentials?.length) {
        pk.allowCredentials = options.publicKey.allowCredentials.map((cred) => {
            const idBytes = base64ToUint8Array(cred.id)

            return {
                type: cred.type,
                id: idBytes.buffer as ArrayBuffer,
                transports: cred.transports
            }
        })
    }

    const credential = await navigator.credentials.get({
        publicKey: pk
    }) as PublicKeyCredential | null

    if (!credential) throw new Error('No credential returned')

    const response = credential.response as AuthenticatorAssertionResponse

    const authRes = await apiFetch('/api/auth/passkey/login-verify', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Email': email,
            'X-Session-Id': sessionId
        },
        body: JSON.stringify({
            id: uint8ArrayToBase64url(new Uint8Array(credential.rawId)),
            rawId: uint8ArrayToBase64url(new Uint8Array(credential.rawId)),
            type: credential.type,
            response: {
                authenticatorData: uint8ArrayToBase64url(new Uint8Array(response.authenticatorData)),
                clientDataJSON: uint8ArrayToBase64url(new Uint8Array(response.clientDataJSON)),
                signature: uint8ArrayToBase64url(new Uint8Array(response.signature)),
                userHandle: response.userHandle
                    ? uint8ArrayToBase64url(new Uint8Array(response.userHandle))
                    : null
            }
        })
    })

    if (!authRes) throw new Error('Backend not available')

    if (!authRes.ok) {
        const errorText = await authRes.text()
        throw new Error(`Authentication verification failed: ${errorText || authRes.statusText}`)
    }

    window.location.href = '/dashboard'
}