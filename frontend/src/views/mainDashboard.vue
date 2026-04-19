<script setup lang="ts">
import Button from 'primevue/button'
import { apiFetch } from '../services/logout/autoLogoutRedirect.ts'
import { requestElevation } from '../services/elevation/elevate.ts'
import router from "../router/index.ts";

function test () {
  apiFetch('/api/protected/test', {})
}

async function logout() {
  await fetch(`${import.meta.env.VITE_API_URL}/api/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  })
  await router.push('/?logout=true')
}



async function testActionElevation() {
  try {
    console.log('Starting action elevation...')
    const ok = await requestElevation('action')
    console.log('Result:', ok)
  } catch (e) {
    console.error('Elevation error:', e)
  }
}

async function testViewElevation() {
  const ok = await requestElevation('view')
  console.log(ok ? 'View elevation granted!' : 'View elevation failed')
}
</script>

<template>
<p>Auth comeplete, dashboard loaded</p>
  <div class="grid grid-cols-1 gap-4 max-w-sm w-full mx-auto px-4 py-3">
  <Button label="Expired Test" @click=test() />
  <Button label="Logout" @click=logout() />
  <Button label="Test Action Elevation" @click="testActionElevation()" />
  <Button label="Test View Elevation" @click="testViewElevation()" />
  </div>
</template>

<style scoped>

</style>