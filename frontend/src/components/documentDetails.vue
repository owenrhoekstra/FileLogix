<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import Drawer from 'primevue/drawer'
import Galleria from 'primevue/galleria'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import ToggleSwitch from 'primevue/toggleswitch'
import MultiSelect from 'primevue/multiselect'
import DatePicker from 'primevue/datepicker'
import 'primeicons/primeicons.css'
import { apiFetch } from '../services/fetch/statusCodeChecks.ts'
import { useConfirm } from 'primevue/useconfirm'
import { useSwipe, useDebounceFn } from '@vueuse/core'
import type { Document } from '../types/documents.ts'

const confirm = useConfirm()
const router = useRouter()

// ── Types ────────────────────────────────────────────────────────────────────

interface DocumentType {
  documentLabel: string
  documentLabelValue: string
}

interface DocumentDetail {
  id: string
  name: string
  types: string[]
  dateOfDoc: string
  dateFiled: string
  location: string
  description: string
  sensitive: boolean
  deleted: boolean
  pages: string[]
}

interface EditDraft {
  name: string
  sensitive: boolean
  types: string[]
  dateOfDoc: Date | null
}

// ── Props / Emits ─────────────────────────────────────────────────────────────

const props = defineProps<{
  visible: boolean
  documentId: string | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  deleted: [id: string]
  restored: [id: string]
  edited: [id: string, fields: Pick<Document, 'name' | 'sensitive' | 'types' | 'dateOfDoc'>]
}>()

const galleryRef = ref<HTMLElement | null>(null)

// ── State ────────────────────────────────────────────────────────────────────

const detail          = ref<DocumentDetail | null>(null)
const loading         = ref(false)
const saving          = ref(false)
const activePageIndex = ref(0)
const typeMap         = ref<Record<string, string>>({})
const availableTypes  = ref<DocumentType[]>([])

const isEditing  = ref(false)
const editDraft  = ref<EditDraft>({ name: '', sensitive: false, types: [], dateOfDoc: null })

// ── Computed ─────────────────────────────────────────────────────────────────

const isDeleted = computed(() => detail.value?.deleted ?? false)

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatDate(raw: string): string {
  if (!raw) return '—'
  const d = new Date(raw)
  if (isNaN(d.getTime())) return raw
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const yyyy = d.getFullYear()
  return `${mm}-${dd}-${yyyy}`
}

function resolveLabel(value: string): string {
  return typeMap.value[value] ?? value
}

useSwipe(galleryRef, {
  passive: false,
  onSwipe(e) { e.preventDefault() },
  onSwipeEnd(_e, direction) {
    if (!detail.value) return
    const max = detail.value.pages.length - 1
    if (direction === 'left' && activePageIndex.value < max) activePageIndex.value++
    else if (direction === 'right' && activePageIndex.value > 0) activePageIndex.value--
  }
})

// ── Fetch ────────────────────────────────────────────────────────────────────

const fetchTypeMap = async () => {
  const res = await apiFetch('/api/protected/form-metadata')
  if (!res) return
  const data = await res.json()
  const types: DocumentType[] = data.documentTypes ?? []
  availableTypes.value = types
  typeMap.value = Object.fromEntries(types.map(t => [t.documentLabelValue, t.documentLabel]))
}

const fetchDetail = async (id: string) => {
  loading.value  = true
  isEditing.value = false
  activePageIndex.value = 0

  try {
    const res = await apiFetch(`/api/protected/documents/${id}`)
    if (!res) return
    detail.value = await res.json()
  } finally {
    loading.value = false
  }
}

const debouncedFetchDetail = useDebounceFn(fetchDetail, 500)

watch(
    () => props.documentId,
    (id) => { if (id) debouncedFetchDetail(id) }
)

fetchTypeMap()

// ── Edit ──────────────────────────────────────────────────────────────────────

const startEdit = () => {
  if (!detail.value) return
  editDraft.value = {
    name:      detail.value.name,
    sensitive: detail.value.sensitive,
    types:     [...detail.value.types],
    dateOfDoc: detail.value.dateOfDoc ? new Date(detail.value.dateOfDoc) : null,
  }
  isEditing.value = true
}

const cancelEdit = () => {
  isEditing.value = false
}

const saveEdit = async () => {
  if (!detail.value) return
  saving.value = true
  try {
    const res = await apiFetch(`/api/protected/documents/${detail.value.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name:      editDraft.value.name,
        sensitive: editDraft.value.sensitive,
        types:     editDraft.value.types,
        dateOfDoc: editDraft.value.dateOfDoc?.toISOString().split('T')[0] ?? null,
      }),
    })
    if (!res) return
    await fetchDetail(detail.value.id)
    emit('edited', detail.value.id, {
      name:      editDraft.value.name,
      sensitive: editDraft.value.sensitive,
      types:     editDraft.value.types,
      dateOfDoc: editDraft.value.dateOfDoc?.toISOString().split('T')[0] ?? '',
    })
    isEditing.value = false
  } finally {
    saving.value = false
  }
}

// ── Actions ──────────────────────────────────────────────────────────────────

const printLabel = () => {
  if (!detail.value) return
  const win = window.open(`/print/${detail.value.id}`, '_blank')
  if (!win) return
  win.onload = () => win.print()
}

const refile = () => {
  router.push('/code-scan')
}

const deleteDocument = () => {
  if (!detail.value) return
  confirm.require({
    message: 'Are you sure you want to delete this document?',
    header: 'Delete Document',
    icon: 'pi pi-trash',
    rejectProps: { label: 'Cancel', severity: 'secondary', outlined: true },
    acceptProps: { label: 'Delete', severity: 'danger' },
    accept: async () => {
      const res = await apiFetch(`/api/protected/documents/${detail.value!.id}`, { method: 'DELETE' })
      if (!res) return
      emit('deleted', detail.value!.id)
      emit('update:visible', false)
    },
  })
}

const restoreDocument = () => {
  if (!detail.value) return
  confirm.require({
    message: 'Restore this document?',
    header: 'Restore Document',
    icon: 'pi pi-undo',
    rejectProps: { label: 'Cancel', severity: 'secondary', outlined: true },
    acceptProps: { label: 'Restore', severity: 'success' },
    accept: async () => {
      const res = await apiFetch(`/api/protected/documents/${detail.value!.id}/restore`, { method: 'PATCH' })
      if (!res) return
      emit('restored', detail.value!.id)
      emit('update:visible', false)
    },
  })
}
</script>

<template>
  <Drawer
      :visible="visible"
      position="right"
      :header="isEditing ? 'Edit Document' : (detail?.name ?? 'Document')"
      style="width: min(480px, 100vw)"
      @update:visible="emit('update:visible', $event)"
  >
    <div class="relative">
      <!-- Loading overlay -->
      <div
          v-if="loading"
          class="absolute inset-0 z-10 flex justify-center items-center bg-surface-0/60 rounded-lg"
      >
        <i class="pi pi-spin pi-spinner text-2xl text-surface-400" />
      </div>

      <div v-if="detail" class="flex flex-col gap-6">

        <!-- ── View Mode ── -->
        <template v-if="!isEditing">
          <div class="flex flex-col gap-4">

            <div class="flex items-center gap-1.5">
              <i
                  :class="detail.sensitive ? 'pi pi-lock text-red-400' : 'pi pi-lock-open text-blue-400'"
                  class="text-sm"
              />
              <span
                  :class="detail.sensitive ? 'text-red-400' : 'text-blue-400'"
                  class="text-sm font-medium"
              >
                {{ detail.sensitive ? 'Sensitive Document' : 'Not Sensitive Document' }}
              </span>
            </div>

            <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
              <template v-for="(type, index) in detail.types" :key="type">
                <span class="text-base font-semibold text-surface-700">{{ resolveLabel(type) }}</span>
                <i
                    v-if="index < detail.types.length - 1"
                    class="pi pi-circle-fill text-surface-300"
                    style="font-size: 0.4rem"
                />
              </template>
            </div>

            <div class="grid grid-cols-2 gap-y-4 text-sm">
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-surface-400">Document Date</span>
                <span class="font-medium">{{ formatDate(detail.dateOfDoc) }}</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-surface-400">Date Filed</span>
                <span class="font-medium">{{ formatDate(detail.dateFiled) }}</span>
              </div>
              <div class="flex flex-col gap-0.5 col-span-2">
                <span class="text-xs text-surface-400">Filing Location</span>
                <span class="font-medium">{{ detail.location || '—' }}</span>
              </div>
              <div v-if="detail.description" class="flex flex-col gap-0.5 col-span-2">
                <span class="text-xs text-surface-400">Directions</span>
                <span class="text-surface-600">{{ detail.description }}</span>
              </div>
            </div>

          </div>

          <div class="border-t border-surface-200" />

          <div ref="galleryRef">
            <Galleria
                v-if="detail.pages.length > 0"
                :value="detail.pages"
                v-model:activeIndex="activePageIndex"
                :show-thumbnails="false"
                :show-indicators="detail.pages.length > 1"
                :show-item-navigators="detail.pages.length > 1"
                container-class="w-full rounded-lg overflow-hidden border border-surface-200"
            >
              <template #item="{ item }">
                <img
                    :key="item"
                    :src="item"
                    :alt="`Page ${activePageIndex + 1}`"
                    class="w-full object-contain max-h-96"
                />
              </template>
            </Galleria>
          </div>

          <div v-if="detail.pages.length > 1" class="text-center text-xs text-surface-400 -mt-4">
            {{ activePageIndex + 1 }} / {{ detail.pages.length }}
          </div>
        </template>

        <!-- ── Edit Mode ── -->
        <template v-else>
          <div class="flex flex-col gap-5">

            <div class="flex flex-col gap-1.5">
              <label class="text-xs text-surface-400">Name</label>
              <InputText v-model="editDraft.name" class="w-full" />
            </div>

            <div class="flex flex-col gap-1.5">
              <label class="text-xs text-surface-400">Document Date</label>
              <DatePicker v-model="editDraft.dateOfDoc" date-format="M dd yy" class="w-full" />
            </div>

            <div class="flex flex-col gap-1.5">
              <label class="text-xs text-surface-400">Types</label>
              <MultiSelect
                  v-model="editDraft.types"
                  :options="availableTypes"
                  option-label="documentLabel"
                  option-value="documentLabelValue"
                  placeholder="Select types"
                  display="chip"
                  class="w-full"
              />
            </div>

            <div class="flex items-center justify-between">
              <span class="text-xs text-surface-400">Sensitive Document</span>
              <ToggleSwitch v-model="editDraft.sensitive" />
            </div>

          </div>
        </template>

      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <div v-if="detail" class="flex flex-col gap-2 w-full">

        <!-- View mode footer -->
        <div v-if="!isEditing" class="grid grid-cols-2 gap-2">
          <Button label="Refile"      icon="pi pi-folder-open" outlined class="w-full" @click="refile" />
          <Button label="Print Label" icon="pi pi-print"       outlined class="w-full" @click="printLabel" />
          <Button label="Edit"        icon="pi pi-pencil"      outlined class="w-full" @click="startEdit" />
          <Button
              v-if="!isDeleted"
              label="Delete"
              icon="pi pi-trash"
              outlined
              severity="danger"
              class="w-full"
              @click="deleteDocument"
          />
          <Button
              v-else
              label="Restore"
              icon="pi pi-undo"
              outlined
              severity="success"
              class="w-full"
              @click="restoreDocument"
          />
        </div>

        <!-- Edit mode footer -->
        <div v-else class="flex gap-2">
          <Button label="Cancel" outlined severity="secondary" class="flex-1" @click="cancelEdit" />
          <Button label="Save"   icon="pi pi-check" class="flex-1" :loading="saving" @click="saveEdit" />
        </div>

      </div>
    </template>
  </Drawer>
</template>

<style scoped>
img {
  animation: fadeIn 0.7s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}
</style>