<script setup lang="ts">
import { ref } from 'vue'
import { z } from 'zod'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import mainMenuBar from '../../components/mainMenuBar.vue'
import footerBar from '../../components/footerBar.vue'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'

const documentLabel = ref('')
const error = ref('')
const success = ref(false)
const loading = ref(false)

const schema = z.string()
    .min(1, 'Document type is required')
    .max(50, 'Maximum 50 characters')
    .regex(/^[a-zA-Z0-9 ]+$/, 'Letters, numbers, and spaces only')

async function submit() {
  error.value = ''
  success.value = false

  const result = schema.safeParse(documentLabel.value)
  if (!result.success) {
    error.value = result.error.errors[0].message
    return
  }

  loading.value = true
  const resp = await apiFetch('/api/protected/settings/add-document-type', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ documentLabel: documentLabel.value.trim() }),
  })
  loading.value = false

  if (!resp) return // 401 redirect handled by apiFetch
  if (!resp.ok) {
    error.value = 'Failed to add document type.'
    return
  }

  documentLabel.value = ''
  success.value = true
}
</script>

<template>
  <mainMenuBar />

  <div class="p-6 max-w-md mx-auto flex flex-col gap-4">
    <h2 class="text-xl font-semibold">Add New Document Type</h2>

    <div class="flex flex-col gap-1">
      <InputText
          v-model="documentLabel"
          placeholder="Document type label"
          :maxlength="50"
          class="w-full"
      />
      <small v-if="error" class="text-red-500">{{ error }}</small>
      <small v-if="success" class="text-green-500">Document type added.</small>
    </div>

    <Button
        label="Submit"
        :loading="loading"
        @click="submit"
    />
  </div>

  <footerBar />
</template>