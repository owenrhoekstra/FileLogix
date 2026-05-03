<script setup lang="ts">
import DataView from 'primevue/dataview'
import SelectButton from 'primevue/selectbutton'
import Button from 'primevue/button'
import { ref, watch } from 'vue'
import 'primeicons/primeicons.css'
import mainMenuBar from '../components/mainMenuBar.vue'
import footerBar from '../components/footerBar.vue'
import { useDocuments } from '../services/composables/useDocuments'

const filterOptions = [
  { label: 'Added', value: 'added' },
  { label: 'Modified', value: 'modified' },
  { label: 'Viewed', value: 'viewed' },
  { label: 'Deleted', value: 'deleted' },
]

const LIMIT = 20
const activeFilter = ref('added')
const { documents, loading, fetchDocuments } = useDocuments()

watch(activeFilter, (val) => fetchDocuments({ sortBy: val, limit: LIMIT }), { immediate: true })

const getIcon = (type: string) => {
  return type === 'PDF' ? 'pi pi-file-pdf' : 'pi pi-file-word'
}

const getDate = (doc: any) => {
  switch (activeFilter.value) {
    case 'added': return doc.added
    case 'modified': return doc.modified
    case 'viewed': return doc.viewed
    case 'deleted': return doc.deleted
  }
}

const getDateLabel = () => {
  switch (activeFilter.value) {
    case 'added': return 'Added'
    case 'modified': return 'Modified'
    case 'viewed': return 'Viewed'
    case 'deleted': return 'Deleted'
  }
}
</script>

<template>
  <mainMenuBar />
  <div class="flex flex-col gap-4 p-4">

    <!-- Sort By -->
    <div class="flex flex-col sm:flex-row justify-center items-center gap-2">
      <span class="text-lg font-semibold">Sort By:</span>
      <SelectButton
          v-model="activeFilter"
          :options="filterOptions"
          option-label="label"
          option-value="value"
      />
    </div>

    <!-- DataView -->
    <DataView :value="documents" :loading="loading" :paginator="true" :rows="LIMIT">
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
                  v-if="activeFilter !== 'deleted'"
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

  </div>
  <footerBar />
</template>