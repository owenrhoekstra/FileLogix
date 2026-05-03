<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const toast = useToast()

type PrintState = 'loading' | 'ready' | 'error'

const state = ref<PrintState>('loading')
const errorMessage = ref<string>('')
let objectUrl: string | null = null

onBeforeUnmount(() => {
  if (objectUrl) URL.revokeObjectURL(objectUrl)
})

const id = Array.isArray(route.params.id) ? route.params.id[0] : route.params.id

async function fetchLabel(): Promise<void> {
  state.value = 'loading'
  try {
    const res = await fetch(`/api/protected/print/${id}`)
    if (!res.ok) throw new Error(`Server returned ${res.status}`)
    const blob = await res.blob()
    objectUrl = URL.createObjectURL(blob)
    state.value = 'ready'
  } catch (err) {
    state.value = 'error'
    errorMessage.value = err instanceof Error ? err.message : 'An unexpected error occurred.'
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'An unexpected error occurred while loading the label. Please try again later.',
      life: 3000,
    })
  }
}

function triggerPrint(): void {
  if (!objectUrl) return
  const win = window.open(objectUrl, '_blank')
  if (!win) return

  const check = setInterval(() => {
    if (win.document.readyState === 'complete') {
      clearInterval(check)
      win.print()
    }
  }, 100)
}

function finish(): void {
  router.push('/code-scan')
}

fetchLabel()
</script>

<template>
  <mainMenuBar />

  <div class="min-h-[calc(100vh-180px)] flex justify-center items-center p-4">
    <div class="flex flex-col items-center gap-6 text-center max-w-sm w-full">

      <!-- Loading -->
      <template v-if="state === 'loading'">
        <i class="pi pi-spin pi-spinner text-4xl text-muted-color" />
        <div class="flex flex-col gap-1">
          <span class="text-lg font-medium">Preparing Label</span>
          <span class="text-sm text-muted-color">Generating data matrix for record</span>
          <code class="text-xs text-muted-color mt-1">{{ id }}</code>
        </div>
      </template>

      <!-- Ready -->
      <template v-else-if="state === 'ready'">
        <i class="pi pi-print text-4xl" />
        <div class="flex flex-col gap-1">
          <span class="text-lg font-medium">Label Ready</span>
          <span class="text-sm text-muted-color">Make sure your document is loaded in the printer before printing.</span>
        </div>
        <div class="flex gap-2 w-full">
          <Button
              label="Print"
              icon="pi pi-print"
              severity="secondary"
              fluid
              @click="triggerPrint"
          />
          <Button
              label="Done"
              icon="pi pi-check"
              fluid
              @click="finish"
          />
        </div>
      </template>

      <!-- Error -->
      <template v-else-if="state === 'error'">
        <i class="pi pi-times-circle text-4xl text-red-500" />
        <div class="flex flex-col gap-1">
          <span class="text-lg font-medium">Failed to Load Label</span>
          <span class="text-sm text-muted-color">{{ errorMessage }}</span>
        </div>
        <Button label="Go to Dashboard" icon="pi pi-home" severity="secondary" @click="router.push('/dashboard')" />
      </template>

    </div>
  </div>

  <footerBar />
</template>