import { createApp } from 'vue'
import './style.css'
import './assets/theme.css'
import App from './App.vue'
import { registerSW } from 'virtual:pwa-register'
import router from './router'
import PrimeVue from 'primevue/config';
import Aura from '@primeuix/themes/aura';
import ToastService from 'primevue/toastservice'
import Form from '@primevue/forms/form'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Button from 'primevue/button'
import Toast from "primevue/toast";

if (import.meta.env.PROD) {
    registerSW({ immediate: true })
}


createApp(App)
    .use(router)
    .use(PrimeVue, {
        theme: {
            preset: Aura,
            options: {
                darkModeSelector: true
            }
        }
    })
    .use(ToastService)
    .component('Form', Form)
    .component('InputText', InputText)
    .component('Message', Message)
    .component('Button', Button)
    .component('Toast', Toast)
    .mount('#app')
