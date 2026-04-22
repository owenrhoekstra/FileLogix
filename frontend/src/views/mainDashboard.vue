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

async function ocrCall() {
  // Create a hidden file input that opens the camera
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.capture = 'environment' // rear camera; use 'user' for front

  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return

    // Decode the captured image
    const bitmap = await createImageBitmap(file)

    // Re-encode to high-quality WebP via canvas
    const canvas = document.createElement('canvas')
    canvas.width = bitmap.width
    canvas.height = bitmap.height
    const ctx = canvas.getContext('2d')!
    ctx.drawImage(bitmap, 0, 0)

    const blob = await new Promise<Blob>((resolve, reject) =>
        canvas.toBlob(
            b => b ? resolve(b) : reject(new Error('toBlob failed')),
            'image/webp',
            0.92 // quality 0–1
        )
    )

    // Send to backend
    const form = new FormData()
    form.append('image', blob, 'capture.webp')

    await apiFetch('/api/protected/ocr', {
      method: 'POST',
      body: form,
      // Do NOT set Content-Type — browser sets it with the boundary
    })
  }

  input.click()
}
</script>

<template>
<p>Auth comeplete, dashboard loaded</p>
  <div class="grid grid-cols-1 gap-4 max-w-sm w-full mx-auto px-4 py-3">
  <Button label="Expired Test" @click=test() />
  <Button label="Logout" @click=logout() />
  <Button label="Test Action Elevation" @click="testActionElevation()" />
  <Button label="Test View Elevation" @click="testViewElevation()" />
    <Button label="OCR Call" @click="ocrCall" />
  </div>
</template>

<style scoped>

</style>