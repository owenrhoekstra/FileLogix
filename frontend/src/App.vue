<script setup lang="ts">
import { useRegisterSW } from 'virtual:pwa-register/vue'
import { RouterView } from 'vue-router'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import { initToast } from './services/utils/toast.ts'

const toast = useToast()
initToast(toast)

useRegisterSW({
  onRegisteredSW(_swUrl, r) {
    if (!r) return
    r.update()
    setInterval(() => r.update(), 300000)
  }
})
</script>

<template>
  <div class="pb-24">
    <router-view />
  </div>
  <Toast
      position="top-right"
      :breakpoints="{
      '960px': { width: '75vw' },
      '640px': { width: '90vw' }
    }"
  />
  <ConfirmDialog />
</template>
