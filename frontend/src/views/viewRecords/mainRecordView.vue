<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import SelectButton from 'primevue/selectbutton'
import Button from 'primevue/button'
import Drawer from 'primevue/drawer'
import DatePicker from 'primevue/datepicker'
import MultiSelect from 'primevue/multiselect'
import 'primeicons/primeicons.css'
import mainMenuBar from '../../components/mainMenuBar.vue'
import footerBar from '../../components/footerBar.vue'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'
import documentList from '../../components/documentList.vue'
import documentDetails from '../../components/documentDetails.vue'
import type { Document, DocumentType, Filters } from '../../types/documents.ts'

// ── Constants ────────────────────────────────────────────────────────────────

const LIMIT = 5

const sortOptions = [
  { label: 'Added',    value: 'added'    },
  { label: 'Modified', value: 'modified' },
  { label: 'Deleted',  value: 'deleted'  },
]

const typeMap = computed(() =>
    Object.fromEntries(availableTypes.value.map(t => [t.documentLabelValue, t.documentLabel]))
)

// ── State ────────────────────────────────────────────────────────────────────

const documents   = ref<Document[]>([])
const loading     = ref(false)
const loadingMore = ref(false)
const hasMore     = ref(true)
const offset      = ref(0)

const activeSort = ref('added')

// Filter drawer
const drawerOpen        = ref(false)
const activeFilterCount = ref(0)
const availableTypes    = ref<DocumentType[]>([])

const draftFilters = ref<Filters>({
  docDateFrom:   null,
  docDateTo:     null,
  filedDateFrom: null,
  filedDateTo:   null,
  types:         [],
})

const appliedFilters = ref<Filters>({ ...draftFilters.value })

// Detail drawer
const drawerVisible = ref(false)
const selectedId    = ref<string | null>(null)

// ── Helpers ──────────────────────────────────────────────────────────────────

const countActiveFilters = (f: Filters) => {
  let count = 0
  if (f.docDateFrom || f.docDateTo)     count++
  if (f.filedDateFrom || f.filedDateTo) count++
  if (f.types.length)                   count++
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
  return `/api/protected/fetch-records?${params.toString()}`
}

const onEdited = (id: string, fields: Pick<Document, 'name' | 'sensitive' | 'types' | 'dateOfDoc'>) => {
  const doc = documents.value.find(d => d.id === id)
  if (!doc) return
  doc.name      = fields.name
  doc.sensitive = fields.sensitive
  doc.types     = fields.types
  doc.dateOfDoc = fields.dateOfDoc
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

const fetchTypes = async () => {
  const res = await apiFetch('/api/protected/form-metadata')
  if (!res) return
  const data = await res.json()
  availableTypes.value = data.documentTypes
}

// ── Filter drawer ─────────────────────────────────────────────────────────────

const openFilterDrawer = () => {
  draftFilters.value = { ...appliedFilters.value }
  drawerOpen.value   = true
}

const applyFilters = () => {
  appliedFilters.value    = { ...draftFilters.value }
  activeFilterCount.value = countActiveFilters(appliedFilters.value)
  drawerOpen.value        = false
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
  draftFilters.value      = empty
  appliedFilters.value    = empty
  activeFilterCount.value = 0
  drawerOpen.value        = false
  fetchDocuments(true)
}

// ── Detail drawer ─────────────────────────────────────────────────────────────

const openDetail = (id: string) => {
  selectedId.value    = id
  drawerVisible.value = true
}

const onDeleted = (id: string) => {
  documents.value = documents.value.filter(d => d.id !== id)
}

const onRestored = (id: string) => {
  documents.value = documents.value.filter(d => d.id !== id)
}

// ── Init ──────────────────────────────────────────────────────────────────────

fetchTypes()
fetchDocuments(true)

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
            @click="openFilterDrawer"
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
    <documentList
        :documents="documents"
        :loading="loading"
        :loading-more="loadingMore"
        :has-more="hasMore"
        :type-map="typeMap"
        @load-more="fetchDocuments(false)"
        @select="openDetail"
    />

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

  <!-- Detail Drawer -->
  <documentDetails
      v-model:visible="drawerVisible"
      :document-id="selectedId"
      @deleted="onDeleted"
      @restored="onRestored"
      @edited="onEdited"
  />

  <footerBar />
</template>