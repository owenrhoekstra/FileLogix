<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import { BrowserMultiFormatReader } from '@zxing/browser'
import { NotFoundException } from '@zxing/library'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'
import mainMenuBar from '../../components/mainMenuBar.vue'
import footerBar from '../../components/footerBar.vue'

const router = useRouter()
const toast = useToast()

type ScanStep = 'document' | 'cabinet' | 'submitting' | 'done'

const step = ref<ScanStep>('document')
const videoRef = ref<HTMLVideoElement | null>(null)
const scanning = ref(false)
const errorMessage = ref('')

const documentId = ref<string | null>(null)
const cabinetId = ref<string | null>(null)
const cabinetMeta = ref<{ name: string; location: string } | null>(null)

let reader: BrowserMultiFormatReader | null = null
let streamRef: MediaStream | null = null

async function startScanner() {
  stopScanner()
  scanning.value = true
  errorMessage.value = ''

  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: { ideal: 'environment' } }
    })
    streamRef = stream
    if (videoRef.value) {
      videoRef.value.srcObject = null
      videoRef.value.srcObject = stream
    }

    reader = new BrowserMultiFormatReader()
    await reader.decodeFromStream(stream, videoRef.value!, (result, err, controls) => {
      if (result) {
        controls.stop()
        scanning.value = false
        handleScan(result.getText())
      }
      if (err && !(err instanceof NotFoundException)) {
        controls.stop()
        scanning.value = false
        errorMessage.value = 'Scanner error. Try again.'
      }
    })
  } catch (err) {
    scanning.value = false
    errorMessage.value = err instanceof Error ? err.message : 'Failed to start camera.'
  }
}

function stopScanner() {
  if (streamRef) {
    streamRef.getTracks().forEach(t => t.stop())
    streamRef = null
  }
  reader = null
  scanning.value = false
}

let handling = false

async function handleScan(value: string) {
  if (handling) return
  handling = true
  try {
    if (step.value === 'document') {
      documentId.value = value
      step.value = 'cabinet'
      stopScanner()
      await new Promise(r => setTimeout(r, 500))
      await startScanner()
    } else if (step.value === 'cabinet') {
      cabinetId.value = value
      await fetchCabinetMeta(value)
    }
  } finally {
    handling = false
  }
}

async function fetchCabinetMeta(id: string) {
  if (id === documentId.value) {
    toast.add({ severity: 'error', summary: 'Wrong Code', detail: 'That was the document. Scan the cabinet instead.', life: 3000 })
    errorMessage.value = 'Please scan the cabinet QR code.'
    cabinetId.value = null
    stopScanner()
    await new Promise(r => setTimeout(r, 500))
    await startScanner()
    return
  }

  const res = await apiFetch(`/api/protected/cabinets/${id}`)
  if (!res || !res.ok) {
    toast.add({ severity: 'error', summary: 'Invalid Cabinet', detail: 'Could not resolve cabinet. Try again.', life: 3000 })
    errorMessage.value = 'Could not resolve cabinet. Try scanning again.'
    cabinetId.value = null
    stopScanner()
    await new Promise(r => setTimeout(r, 500))
    await startScanner()
    return
  }
  cabinetMeta.value = await res.json()
}

async function submit() {
  if (!documentId.value || !cabinetId.value) return
  step.value = 'submitting'

  const res = await apiFetch('/api/protected/records/location', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ documentId: documentId.value, cabinetId: cabinetId.value }),
  })

  if (!res || !res.ok) {
    step.value = 'cabinet'
    toast.add({ severity: 'error', summary: 'Failed', detail: 'Could not save location. Try again.', life: 3000 })
    return
  }

  toast.add({ severity: 'success', summary: 'Filed', detail: 'Document location saved.', life: 3000 })
  router.push('/dashboard')
}

function rescan() {
  cabinetId.value = null
  cabinetMeta.value = null
  errorMessage.value = ''
  step.value = 'cabinet'

  toast.add({ severity: 'info', summary: 'Rescanning', detail: 'Scan the cabinet again.', life: 2000 })

  stopScanner()
  setTimeout(() => {
    startScanner()
  }, 500)
}

onBeforeUnmount(() => stopScanner())

// kick off first scan immediately
startScanner()
</script>

<template>
  <mainMenuBar />

  <div class="min-h-[calc(100vh-180px)] flex justify-center items-center p-4">
    <div class="flex flex-col items-center gap-6 text-center max-w-sm w-full">

      <!-- Step: Scan Document -->
      <template v-if="step === 'document'">
        <i class="pi pi-qrcode text-4xl" />
        <div class="flex flex-col gap-1">
          <span class="text-lg font-medium">Scan Document</span>
          <span class="text-sm text-muted-color">Point your camera at the Data Matrix label on the document.</span>
        </div>
        <div class="w-full rounded-xl overflow-hidden border border-surface-border aspect-square bg-black">
          <video ref="videoRef" class="w-full h-full object-cover" autoplay muted playsinline />
        </div>
        <span v-if="errorMessage" class="text-sm text-red-400">{{ errorMessage }}</span>
      </template>

      <!-- Step: Scan Cabinet -->
      <template v-else-if="step === 'cabinet'">
        <!-- Document confirmed -->
        <div class="w-full flex items-center gap-3 px-3 py-3 rounded-lg border border-surface-border text-left">
          <i class="pi pi-file text-xl shrink-0" />
          <div class="flex flex-col min-w-0">
            <span class="text-xs text-muted-color">Document ID</span>
            <code class="text-sm truncate">{{ documentId }}</code>
          </div>
          <i class="pi pi-check-circle text-green-500 ml-auto shrink-0" />
        </div>

        <template v-if="!cabinetMeta">
          <i class="pi pi-box text-4xl" />
          <div class="flex flex-col gap-1">
            <span class="text-lg font-medium">Scan Cabinet</span>
            <span class="text-sm text-muted-color">Point your camera at the QR code on the filing cabinet.</span>
          </div>
          <div class="w-full rounded-xl overflow-hidden border border-surface-border aspect-square bg-black">
            <video ref="videoRef" class="w-full h-full object-cover" autoplay muted playsinline />
          </div>
          <span v-if="errorMessage" class="text-sm text-red-400">{{ errorMessage }}</span>
        </template>

        <!-- Cabinet resolved -->
        <template v-else>
          <div class="w-full flex items-center gap-3 px-3 py-3 rounded-lg border border-surface-border text-left">
            <i class="pi pi-box text-xl shrink-0" />
            <div class="flex flex-col min-w-0">
              <span class="text-xs text-muted-color">Cabinet</span>
              <span class="text-sm font-medium">{{ cabinetMeta.name }}</span>
              <span class="text-xs text-muted-color">{{ cabinetMeta.location }}</span>
            </div>
            <i class="pi pi-check-circle text-green-500 ml-auto shrink-0" />
          </div>

          <div class="flex gap-2 w-full">
            <Button label="Rescan" icon="pi pi-refresh" severity="secondary" fluid @click="rescan" />
            <Button label="Confirm" icon="pi pi-check" fluid @click="submit" />
          </div>
        </template>
      </template>

      <!-- Submitting -->
      <template v-else-if="step === 'submitting'">
        <i class="pi pi-spin pi-spinner text-4xl text-muted-color" />
        <div class="flex flex-col gap-1">
          <span class="text-lg font-medium">Saving Location</span>
          <span class="text-sm text-muted-color">Filing document to cabinet...</span>
        </div>
      </template>

    </div>
  </div>

  <footerBar />
</template>