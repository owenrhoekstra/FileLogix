<script setup lang="ts">
import { ref, watch } from 'vue'
import DataView from 'primevue/dataview'
import SelectButton from 'primevue/selectbutton'
import Button from 'primevue/button'
import Drawer from 'primevue/drawer'
import DatePicker from 'primevue/datepicker'
import MultiSelect from 'primevue/multiselect'
import 'primeicons/primeicons.css'
import mainMenuBar from '../../components/mainMenuBar.vue'
import footerBar from '../../components/footerBar.vue'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'

// ── Types ────────────────────────────────────────────────────────────────────

interface Document {
  id: string
  name: string
  type: string
  added: string
  modified: string
  viewed: string
  deleted: string | null
}

interface Filters {
  docDateFrom: Date | null
  docDateTo: Date | null
  filedDateFrom: Date | null
  filedDateTo: Date | null
  types: string[]
}

interface DocumentType {
  documentLabel: string
  documentLabelValue: string
}

// ── Constants ────────────────────────────────────────────────────────────────

const LIMIT = 20

const sortOptions = [
  { label: 'Added',    value: 'added'    },
  { label: 'Modified', value: 'modified' },
  { label: 'Viewed',   value: 'viewed'   },
  { label: 'Deleted',  value: 'deleted'  },
]

// ── State ────────────────────────────────────────────────────────────────────

const documents   = ref<Document[]>([])
const loading     = ref(false)
const loadingMore = ref(false)
const hasMore     = ref(true)
const offset      = ref(0)

const activeSort = ref('added')

const drawerOpen = ref(false)

// Working copies inside the drawer (only committed on Apply)
const draftFilters = ref<Filters>({
  docDateFrom:  null,
  docDateTo:    null,
  filedDateFrom: null,
  filedDateTo:   null,
  types:        [],
})

// Applied filters (what the last fetch used)
const appliedFilters = ref<Filters>({ ...draftFilters.value })

// Active filter count badge
const activeFilterCount = ref(0)

// Available document types from backend
const availableTypes = ref<DocumentType[]>([])

// ── Helpers ──────────────────────────────────────────────────────────────────

const getIcon = (type: string) => {
  return type === 'PDF' ? 'pi pi-file-pdf' : 'pi pi-file-word'
}

const getDate = (doc: Document) => {
  switch (activeSort.value) {
    case 'added':    return doc.added
    case 'modified': return doc.modified
    case 'viewed':   return doc.viewed
    case 'deleted':  return doc.deleted ?? '—'
  }
}

const getDateLabel = () => {
  switch (activeSort.value) {
    case 'added':    return 'Added'
    case 'modified': return 'Modified'
    case 'viewed':   return 'Viewed'
    case 'deleted':  return 'Deleted'
  }
}

const countActiveFilters = (f: Filters) => {
  let count = 0
  if (f.docDateFrom || f.docDateTo)   count++
  if (f.filedDateFrom || f.filedDateTo) count++
  if (f.types.length)                 count++
  return count
}

const buildQuery = (currentOffset: number, filters: Filters) => {
  const params = new URLSearchParams({
    sortBy: activeSort.value,
    limit:  String(LIMIT),
    offset: String(currentOffset),
  })
  if (filters.docDateFrom)   params.set('docDateFrom',   filters.docDateFrom.toISOString())
  if (filters.docDateTo)     params.set('docDateTo',     filters.docDateTo.toISOString())
  if (filters.filedDateFrom) params.set('filedDateFrom', filters.filedDateFrom.toISOString())
  if (filters.filedDateTo)   params.set('filedDateTo',   filters.filedDateTo.toISOString())
  if (filters.types.length)  params.set('types',         filters.types.join(','))
  return `/api/protected/documents?${params.toString()}`
}

// ── Fetch ────────────────────────────────────────────────────────────────────

const fetchDocuments = async (reset = false) => {
  if (reset) {
    documents.value = []
    offset.value    = 0
    hasMore.value   = true
  }

  if (!hasMore.value) return

  reset ? (loading.value = true) : (loadingMore.value = true)

  try {
    const res = await apiFetch(buildQuery(offset.value, appliedFilters.value))
    if (!res) return

    const data: Document[] = await res.json()

    documents.value = reset ? data : [...documents.value, ...data]
    offset.value   += data.length
    hasMore.value   = data.length === LIMIT
  } finally {
    loading.value     = false
    loadingMore.value = false
  }
}

const loadMore = () => fetchDocuments(false)

// ── Document types ────────────────────────────────────────────────────────────

const fetchTypes = async () => {
  const res = await apiFetch('/api/protected/form-metadata')
  if (!res) return
  const data = await res.json()
  availableTypes.value = data.documentTypes
}

// ── Filter drawer ─────────────────────────────────────────────────────────────

const openDrawer = () => {
  // Populate draft from currently applied filters so user sees their last state
  draftFilters.value = { ...appliedFilters.value }
  drawerOpen.value   = true
}

const applyFilters = () => {
  appliedFilters.value  = { ...draftFilters.value }
  activeFilterCount.value = countActiveFilters(appliedFilters.value)
  drawerOpen.value      = false
  fetchDocuments(true)
}

const clearFilters = () => {
  const empty: Filters = {
    docDateFrom:   null,
    docDateTo:     null,
    filedDateFrom: null,
    filedDateTo:   null,
    types:         [],
  }
  draftFilters.value    = empty
  appliedFilters.value  = empty
  activeFilterCount.value = 0
  drawerOpen.value      = false
  fetchDocuments(true)
}

// ── Init ──────────────────────────────────────────────────────────────────────

fetchTypes()
fetchDocuments(true)

// Re-fetch on sort change
watch(activeSort, () => fetchDocuments(true))
</script>

<template>
  <mainMenuBar />

  <div class="flex flex-col gap-4 p-4">

    <!-- Sort + Filter Row -->
    <div class="flex flex-col sm:flex-row justify-between items-center gap-3">
      <div class="flex flex-col items-center gap-2">
        <span class="text-lg font-semibold">Sort By:</span>
        <SelectButton
            v-model="activeSort"
            :options="sortOptions"
            option-label="label"
            option-value="value"
        />
      </div>

      <!-- Filter button with active count badge -->
      <div class="relative">
        <Button
            icon="pi pi-filter"
            label="Filter"
            outlined
            @click="openDrawer"
        />
        <span
            v-if="activeFilterCount > 0"
            class="absolute -top-2 -right-2 bg-primary text-primary-contrast text-xs rounded-full h-5 w-5 flex items-center justify-center font-bold"
        >
          {{ activeFilterCount }}
        </span>
      </div>
    </div>

    <!-- Document List -->
    <DataView :value="documents" :loading="loading">
      <template #list="{ items }">
        <div class="flex flex-col gap-2">
          <div
              v-for="doc in items"
              :key="doc.id"
              class="flex items-center justify-between px-3 py-3 rounded-lg border border-surface-border"
          >
            <!-- Icon + Name -->
            <div class="flex items-center gap-3 min-w-0">
              <i :class="getIcon(doc.type)" class="text-2xl shrink-0" />
              <div class="flex flex-col min-w-0">
                <span class="font-medium truncate">{{ doc.name }}</span>
                <span class="text-sm text-surface-400">{{ doc.type }}</span>
              </div>
            </div>

            <!-- Date + Actions -->
            <div class="flex items-center gap-2 shrink-0">
              <div class="hidden sm:flex flex-col items-end">
                <span class="text-sm text-surface-400">{{ getDateLabel() }}</span>
                <span class="text-sm">{{ getDate(doc) }}</span>
              </div>
              <Button
                  v-if="activeSort !== 'deleted'"
                  icon="pi pi-ellipsis-v"
                  text
                  rounded
              />
              <Button
                  v-else
                  icon="pi pi-replay"
                  text
                  rounded
                  v-tooltip="'Restore'"
              />
            </div>
          </div>
        </div>
      </template>

      <template #empty>
        <div class="flex justify-center py-8 text-surface-400">
          No documents found.
        </div>
      </template>
    </DataView>

    <!-- Load More -->
    <div v-if="hasMore && documents.length > 0" class="flex justify-center pt-2 pb-4">
      <Button
          label="Load More"
          outlined
          :loading="loadingMore"
          @click="loadMore"
      />
    </div>

    <div v-if="!hasMore && documents.length > 0" class="flex justify-center py-4 text-surface-400 text-sm">
      All documents loaded.
    </div>

  </div>

  <!-- Filter Drawer -->
  <Drawer v-model:visible="drawerOpen" position="left" header="Filter Documents">
    <div class="flex flex-col gap-6 py-2">

      <!-- Document Date Range -->
      <div class="flex flex-col gap-2">
        <span class="font-semibold text-sm">Document Date</span>
        <div class="flex gap-2">
          <DatePicker
              v-model="draftFilters.docDateFrom"
              placeholder="From"
              date-format="M dd yy"
              class="flex-1"
          />
          <DatePicker
              v-model="draftFilters.docDateTo"
              placeholder="To"
              date-format="M dd yy"
              class="flex-1"
          />
        </div>
      </div>

      <!-- Date Filed Range -->
      <div class="flex flex-col gap-2">
        <span class="font-semibold text-sm">Date Filed</span>
        <div class="flex gap-2">
          <DatePicker
              v-model="draftFilters.filedDateFrom"
              placeholder="From"
              date-format="M dd yy"
              class="flex-1"
          />
          <DatePicker
              v-model="draftFilters.filedDateTo"
              placeholder="To"
              date-format="M dd yy"
              class="flex-1"
          />
        </div>
      </div>

      <!-- Document Type -->
      <div class="flex flex-col gap-2">
        <span class="font-semibold text-sm">Document Type</span>
        <MultiSelect
            v-model="draftFilters.types"
            :options="availableTypes"
            option-label="documentLabel"
            option-value="documentLabelValue"
            placeholder="Select types"
            display="chip"
        />
      </div>

    </div>

    <!-- Footer Actions -->
    <template #footer>
      <div class="flex gap-2 w-full">
        <Button
            label="Clear"
            outlined
            severity="secondary"
            class="flex-1"
            @click="clearFilters"
        />
        <Button
            label="Apply"
            class="flex-1"
            @click="applyFilters"
        />
      </div>
    </template>
  </Drawer>

  <footerBar />
</template>