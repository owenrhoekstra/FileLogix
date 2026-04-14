<script setup lang="ts">
import Button from 'primevue/button'
import { apiFetch } from '../services/logout/autoLogoutRedirect.ts'
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
</script>

<template>
<p>Auth comeplete, dashboard loaded</p>
  <Button label="Expired Test" @click=test() />
  <Button label="Logout" @click=logout() />
</template>

<style scoped>

</style>