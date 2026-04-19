import { base64ToUint8Array, uint8ArrayToBase64url } from '../userAuthentication/utilFunctions.ts'
import { apiFetch } from '../logout/autoLogoutRedirect.ts'

type ElevationType = 'action' | 'view'

type ElevationChallengeResponse = {
    options: {
        publicKey: {
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
    }
    type: ElevationType
}

export async function requestElevation(type: ElevationType): Promise<boolean> {
    // Step 1: Get challenge
    const challengeRes = await apiFetch('/api/auth/elevate/challenge', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type }),
    })
    if (!challengeRes?.ok) return false

    const { options }: ElevationChallengeResponse = await challengeRes.json()

    // Step 2: Sign with passkey
    const challengeBytes = base64ToUint8Array(options.publicKey.challenge)

    const pk: PublicKeyCredentialRequestOptions = {
        challenge: challengeBytes.buffer as ArrayBuffer,
        rpId: options.publicKey.rpId,
        userVerification: options.publicKey.userVerification,
        timeout: options.publicKey.timeout,
    }

    if (options.publicKey.allowCredentials?.length) {
        pk.allowCredentials = options.publicKey.allowCredentials.map((cred) => {
            const idBytes = base64ToUint8Array(cred.id)
            return {
                type: cred.type,
                id: idBytes.buffer as ArrayBuffer,
                transports: cred.transports,
            }
        })
    }

    const credential = await navigator.credentials.get({ publicKey: pk }) as PublicKeyCredential | null
    if (!credential) return false

    const response = credential.response as AuthenticatorAssertionResponse

    // Step 3: Verify
    const verifyRes = await apiFetch('/api/auth/elevate/verify', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Elevation-Type': type,
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
                    : null,
            },
        }),
    })

    return verifyRes?.ok ?? false
}