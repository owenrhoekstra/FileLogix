import router from "../../router/index.ts";

const BASE_URL = import.meta.env.VITE_API_URL

if (!BASE_URL) {
    throw new Error('VITE_API_URL environment variable is not set')
}

// Validate it's HTTPS
if (!BASE_URL.startsWith('https://')) {
    throw new Error('VITE_API_URL must be HTTPS')
}

export async function apiFetch(url: string, options: RequestInit = {}) {
    const fullUrl = new URL(url, BASE_URL).toString()

    const res = await fetch(fullUrl, {
        ...options,
        credentials: 'include',
    })

    if (res.status === 401) {
        await router.push('/?logout=true')
        return
    }

    return res
}