<script setup lang="ts">
import Button from 'primevue/button'
import { useRouter } from 'vue-router'
import { ref, onMounted, onUnmounted } from 'vue';

const router = useRouter();

const reloadPage = () => {
  window.location.reload();
}

const isMobile = ref(window.innerWidth < 768);
const onResize = () => isMobile.value = window.innerWidth < 768;
onMounted(() => window.addEventListener('resize', onResize));
onUnmounted(() => window.removeEventListener('resize', onResize));
</script>

<template>
  <div class="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[1110px]">
    <!-- Background rectangle -->
    <div class="absolute inset-0" style="background-color: var(--app-bg);" />

    <!-- Buttons -->
    <div class="relative flex items-center px-4 py-2 pb-6">
      <div class="flex-1 flex justify-start gap-2">
        <Button :label="isMobile ? undefined : 'Back'" icon="pi pi-arrow-circle-left" @click="router.back()"/>
        <Button :label="isMobile ? undefined : 'Reload'" icon="pi pi-refresh" @click="reloadPage"/>
      </div>
      <div class="flex-1 flex justify-center">
        <Button :label="isMobile ? undefined : 'File Records'" icon="pi pi-qrcode" @click="router.push('/code-scan')"/>
      </div>
      <div class="flex-1 flex justify-end">
        <Button :label="isMobile ? undefined : 'Forward'" icon="pi pi-arrow-circle-right" @click="router.forward()"/>
      </div>
    </div>
  </div>
  </template>

<style scoped>
</style>