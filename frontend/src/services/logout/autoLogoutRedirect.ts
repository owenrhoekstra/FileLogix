const BASE_URL = import.meta.env.VITE_API_URL

export async function apiFetch(url: string, options: RequestInit = {}) {
    const res = await fetch(`${BASE_URL}${url}`, {
        ...options,
        credentials: 'include',
    })

    if (res.status === 401) {
        window.location.href = '/?logout=true'
        return
    }

    return res
}