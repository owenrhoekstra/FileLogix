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
  <div class="fixed bottom-6 left-1/2 -translate-x-1/2 w-full max-w-[1126px] flex items-center px-4 py-2">
    <!-- Left -->
    <div class="flex-1 flex justify-start gap-2">
      <Button :label="isMobile ? undefined : 'Back'" icon="pi pi-arrow-circle-left" @click="router.back()"/>
      <Button :label="isMobile ? undefined : 'Reload'" icon="pi pi-refresh" @click="reloadPage"/>
    </div>
    <!-- Centre -->
    <div class="flex-1 flex justify-center">
      <Button disabled :label="isMobile ? undefined : 'View All Records'" icon="pi pi-eye" @click="router.push('/records')"/>
    </div>
    <!-- Right -->
    <div class="flex-1 flex justify-end">
      <Button :label="isMobile ? undefined : 'Forward'" icon="pi pi-arrow-circle-right" @click="router.forward()"/>
    </div>
  </div>
</template>

<style scoped>
</style>