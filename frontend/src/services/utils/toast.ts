// services/utils/toast.ts
import type { ToastServiceMethods } from 'primevue/toastservice'

export let toast: ToastServiceMethods | null = null

export function initToast(t: ToastServiceMethods) {
    toast = t
}