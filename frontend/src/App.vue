<script setup lang="ts">
import { useRegisterSW } from 'virtual:pwa-register/vue'
import { RouterView } from 'vue-router'
import Toast from 'primevue/toast'

const { updateServiceWorker } = useRegisterSW({
  onNeedRefresh() {
    updateServiceWorker()
  },
  onRegisteredSW(_swUrl, r) {
    r && setInterval(async () => {
      await r.update()
    }, 300000) // 5 minutes
  }
})
</script>

<template>
  <router-view />
  <Toast
      position="top-right"
      :breakpoints="{
      '960px': { width: '75vw' },
      '640px': { width: '90vw' }
    }"
  />
  <ConfirmDialog />
</template>
