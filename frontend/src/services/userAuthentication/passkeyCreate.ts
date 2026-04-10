import { base64ToUint8Array, uint8ArrayToBase64url } from './utilFunctions'

type PublicKeyCredentialCreationOptionsJSON = {
    rp: PublicKeyCredentialRpEntity
    user: {
        id: string
        name: string
        displayName: string
    }
    challenge: string
    pubKeyCredParams: PublicKeyCredentialParameters[]
    excludeCredentials?: Array<{
        id: string
        type: PublicKeyCredentialType
        transports?: AuthenticatorTransport[]
    }>
    [key: string]: any
}

type PasskeyCreateOptions = {
    options: {
        publicKey: PublicKeyCredentialCreationOptionsJSON
    }
    sessionId: string
}

export async function passkeyCreate(email: string): Promise<void> {
    const res = await fetch('/api/auth/passkey/register-challenge', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
    })

    if (!res.ok) throw new Error('Backend not available')

    const { options, sessionId }: PasskeyCreateOptions = await res.json()

    // ✅ normalize binary properly
    const challenge = base64ToUint8Array(options.publicKey.challenge)
    const userId = base64ToUint8Array(options.publicKey.user.id)

    const pk: PublicKeyCredentialCreationOptions = {
        rp: options.publicKey.rp,
        user: {
            ...options.publicKey.user,
            id: userId.buffer as ArrayBuffer
        },
        challenge: challenge.buffer as ArrayBuffer,
        pubKeyCredParams: options.publicKey.pubKeyCredParams,
    }

    if (options.publicKey.excludeCredentials) {
        pk.excludeCredentials = options.publicKey.excludeCredentials.map(
            (cred) => {
                const idBytes = base64ToUint8Array(cred.id)
                return {
                    ...cred,
                    id: idBytes.buffer as ArrayBuffer
                }
            }
        )
    }

    const credential = await navigator.credentials.create({
        publicKey: pk
    }) as PublicKeyCredential | null

    if (!credential) throw new Error('No credential created')

    const response = credential.response as AuthenticatorAttestationResponse

    const regRes = await fetch('/api/auth/passkey/register-verify', {
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
                clientDataJSON: uint8ArrayToBase64url(new Uint8Array(response.clientDataJSON)),
                attestationObject: uint8ArrayToBase64url(new Uint8Array(response.attestationObject))
            }
        })
    })

    if (!regRes.ok) throw new Error('Registration failed')

    window.location.href = '/dashboard'
}