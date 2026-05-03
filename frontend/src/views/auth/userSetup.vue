<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'
import router from '../../router'

const firstName = ref('')
const lastName = ref('')
const phone = ref('')
const email = ref('')
const initials = ref('')
const submitting = ref(false)
const error = ref<string | null>(null)

watch([firstName, lastName], ([first, last]) => {
  const f = first.trim()[0] ?? ''
  const l = last.trim()[0] ?? ''
  if (f || l) initials.value = (f + l).toUpperCase()
})

function onInitialsInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  initials.value = val.toUpperCase().slice(0, 3)
}

function formatPhone(val: string): string {
  const digits = val.replace(/\D/g, '').slice(0, 11)
  if (digits.length <= 1) return digits.startsWith('1') ? '+1' : digits
  const local = digits.startsWith('1') ? digits.slice(1) : digits
  let formatted = '+1'
  if (local.length > 0) formatted += ' ' + local.slice(0, 3)
  if (local.length > 3) formatted += '-' + local.slice(3, 6)
  if (local.length > 6) formatted += '-' + local.slice(6, 10)
  return formatted
}

function onPhoneInput(e: Event) {
  phone.value = formatPhone((e.target as HTMLInputElement).value)
}

const allFilled = computed(() =>
    firstName.value.trim() !== '' &&
    lastName.value.trim() !== '' &&
    phone.value.trim() !== '' &&
    initials.value.trim() !== ''
)

onMounted(async () => {
  const res = await apiFetch('/api/auth/me', {})
  if (!res?.ok) return
  const data = await res.json()
  email.value = data.email ?? ''
})

async function submit() {
  if (!allFilled.value || submitting.value) return
  submitting.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/user/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        first_name: firstName.value.trim(),
        last_name: lastName.value.trim(),
        phone: phone.value.trim(),
        initials: initials.value.trim(),
      }),
    })
    if (!res?.ok) {
      error.value = await res?.text() ?? 'Something went wrong'
      return
    }
    await router.push('/dashboard')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen px-4">
    <div class="w-full max-w-md">

      <div class="mb-8">
        <h1 class="text-2xl font-semibold mb-1">Account Setup</h1>
        <p class="text-sm opacity-60">Complete your profile to continue.</p>
      </div>

      <div class="flex flex-col gap-5">

        <div class="flex flex-col gap-1.5">
          <label class="text-xs font-medium uppercase tracking-widest opacity-60">First Name</label>
          <InputText v-model="firstName" placeholder="Jane" class="w-full" />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-xs font-medium uppercase tracking-widest opacity-60">Last Name</label>
          <InputText v-model="lastName" placeholder="Smith" class="w-full" />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-xs font-medium uppercase tracking-widest opacity-60">Email</label>
          <InputText v-model="email" disabled class="w-full opacity-50 cursor-not-allowed" />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-xs font-medium uppercase tracking-widest opacity-60">Phone Number</label>
          <InputText :value="phone" @input="onPhoneInput" placeholder="+1 555-555-5555" class="w-full" />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-xs font-medium uppercase tracking-widest opacity-60">
            Initials
            <span class="normal-case tracking-normal font-normal ml-1 opacity-60">(max 3 — editable here only)</span>
          </label>
          <InputText
              :value="initials"
              @input="onInitialsInput"
              placeholder="JS"
              maxlength="3"
              class="w-full font-mono tracking-widest"
          />
        </div>

        <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>

        <Button
            label="Complete Setup"
            :disabled="!allFilled || submitting"
            :loading="submitting"
            @click="submit"
            class="w-full mt-2"
            :class="{ 'opacity-40 cursor-not-allowed': !allFilled }"
        />

      </div>
    </div>
  </div>
</template>