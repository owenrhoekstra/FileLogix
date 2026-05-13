import { createApp } from 'vue'
import './style.css'
import './assets/theme.css'
import App from './App.vue'
import { registerSW } from 'virtual:pwa-register'
import router from './router'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'
import ToastService from 'primevue/toastservice'
import Form from '@primevue/forms/form'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import ConfirmationService from 'primevue/confirmationservice'
import ConfirmDialog from 'primevue/confirmdialog'
import SelectButton from 'primevue/selectbutton'

// main.ts or a dedicated sw-register.ts
if (import.meta.env.PROD) {
    registerSW({
        immediate: true,
        onRegisteredSW(_swUrl, registration) {
            if (!registration) return
            setInterval(() => registration.update(), 5 * 60 * 1000)
        },
        onOfflineReady() {}
    })
}

const MyPreset = definePreset(Aura, {
    semantic: {
        primary: {
            50: '#eff6ff',
            100: '#dbeafe',
            200: '#bfdbfe',
            300: '#93c5fd',
            400: '#60a5fa',
            500: '#3b82f6',
            600: '#2563eb',
            700: '#1d4ed8',
            800: '#1e40af',
            900: '#1e3a8a',
            950: '#172554',
        }
    }
})

createApp(App)
    .use(ConfirmationService)
    .use(router)
    .use(PrimeVue, {
        theme: {
            preset: MyPreset,
            options: {
                darkModeSelector: 'system',
            }
        }
    })
    .use(ToastService)
    .component('SelectButton', SelectButton)
    .component('ConfirmDialog', ConfirmDialog)
    .component('Form', Form)
    .component('InputText', InputText)
    .component('Message', Message)
    .component('Button', Button)
    .component('Toast', Toast)
    .mount('#app')