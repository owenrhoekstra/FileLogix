<script setup lang="ts">
import { ref, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import { z } from 'zod'
import { passkeyLogin } from '../services/userAuthentication/passkeyLogin'
import { passkeyCreate } from '../services/userAuthentication/passkeyCreate'
import { apiFetch } from '../services/logout/autoLogoutRedirect.ts'

const error = ref<string | null>(null)
const email = ref('')
const hasPasskey = ref(false)

const emailSchema = z.string().email('Invalid email format')
const isEmailValid = computed(() => emailSchema.safeParse(email.value).success)

async function handleSubmit() {
  error.value = null

  const result = emailSchema.safeParse(email.value)
  if (!result.success) {
    error.value = result.error.issues[0].message
    return
  }

  try {
    const checkRes = await apiFetch('/api/auth/check-email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: email.value })
    })

    if (!checkRes || !checkRes.ok) {
      error.value = 'Email check failed'
      return
    }

    const checkData = await checkRes.json().catch(() => null)

    if (!checkData) {
      error.value = 'Server error'
      return
    }

    if (!checkData.allowed) {
      error.value = 'Email not allowed'
      return
    }

    hasPasskey.value = checkData.hasPasskey

    if (hasPasskey.value) {
      await passkeyLogin(email.value)
    } else {
      await passkeyCreate(email.value)
    }

  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Something went wrong'
  }
}
</script>

<template>
  <div class="flex flex-col gap-4 max-w-sm justify-center items-center mx-auto">
    <h1>FileLogix</h1>
    <h2>Please Sign in With Your Email and Passkey</h2>
    <InputText
      v-model="email"
      type="email"
      placeholder="Email Address"
      class="w-full"
    />
    <Button
      label="Continue with Passkey"
      @click="handleSubmit"
      :disabled="!isEmailValid"
      class="w-full"
    />
    <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>
  </div>
</template>
