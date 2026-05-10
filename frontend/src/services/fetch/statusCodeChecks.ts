import router from "../../router/index.ts"
import { baseFetch } from './baseFetch.ts'
import { requestElevation } from '../elevation/elevate.ts'
import { toast } from '../utils/toast.ts'

export async function apiFetch(url: string, options: RequestInit = {}) {
    const res = await baseFetch(url, options)

    if (res.status === 401) {
        await router.push('/?logout=true')
        return
    }

    if (res.status === 403) {
        const elevationType = res.headers.get('X-Require-Elevation')
        if (elevationType === 'action' || elevationType === 'view') {
            const ok = await requestElevation(elevationType)
            if (ok) return baseFetch(url, options)
        }

        const msg = res.headers.get('X-Toast')
        if (msg) toast?.add({ severity: 'error', summary: msg, life: 4000 })

        return
    }

    return res
}