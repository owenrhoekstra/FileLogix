<script setup lang="ts">
import DataView from 'primevue/dataview'
import Button from 'primevue/button'
import 'primeicons/primeicons.css'
import type { Document } from '../types/documents.ts'

// ── Props / Emits ─────────────────────────────────────────────────────────────

const props = defineProps<{
  documents: Document[]
  loading: boolean
  loadingMore: boolean
  hasMore: boolean
  typeMap: Record<string, string>
}>()

const emit = defineEmits<{
  loadMore: []
  select: [id: string]
}>()

// ── Helpers ───────────────────────────────────────────────────────────────────

function resolveLabel(value: string): string {
  return props.typeMap[value] ?? value
}

function formatDate(raw: string, hasTime = false): string {
  if (!raw) return '—'
  const d = hasTime ? new Date(raw) : new Date(raw + 'T00:00:00')
  if (isNaN(d.getTime())) return raw
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const yyyy = d.getFullYear()
  return `${mm}-${dd}-${yyyy}`
}
</script>

<template>
  <DataView :value="documents" :loading="loading">
    <template #list="{ items }">
      <div class="flex flex-col gap-2">
        <div
            v-for="doc in items"
            :key="doc.id"
            class="flex items-center gap-4 px-4 py-3 rounded-lg border border-surface-200 cursor-pointer hover:bg-surface-50 transition-colors"
            @click="emit('select', doc.id)"
        >
          <!-- Thumbnail -->
          <div class="shrink-0 w-12 h-16 rounded overflow-hidden bg-surface-100 flex items-center justify-center">
            <img
                :src="doc.thumbnail"
                :alt="doc.name"
                class="w-full h-full object-cover"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
          </div>

          <!-- Metadata -->
          <div class="flex flex-col min-w-0 flex-1 gap-1">

            <!-- Name + sensitive lock -->
            <div class="flex items-center gap-2">
              <span class="text-base font-semibold text-surface-700 truncate">{{ doc.name }}</span>
              <i
                  v-if="doc.sensitive"
                  class="pi pi-lock text-xs text-red-400 shrink-0"
                  title="Sensitive"
              />
              <i
                  v-else
                  class="pi pi-lock-open text-xs text-surface-300 shrink-0"
                  title="Not sensitive"
              />
            </div>

            <!-- Types separated by dot -->
            <div class="flex items-center gap-1.5 flex-wrap">
              <template v-for="(type, index) in doc.types" :key="type">
                <span class="text-sm text-surface-500">{{ resolveLabel(type) }}</span>
                <i
                    v-if="Number(index) < doc.types.length - 1"
                    class="pi pi-circle-fill text-surface-300"
                    style="font-size: 0.35rem"
                />
              </template>
            </div>

            <!-- Dates + Location -->
            <div class="flex flex-wrap gap-x-3 gap-y-0.5 mt-0.5">
              <span class="text-xs text-surface-400">
                Document Date: <span class="text-surface-600 font-medium">{{ formatDate(doc.dateOfDoc) }}</span>
              </span>
              <span class="text-xs text-surface-400">
                Filed: <span class="text-surface-600 font-medium">{{ formatDate(doc.dateFiled, true) }}</span>
              </span>
              <span v-if="doc.location" class="text-xs text-surface-400">
                <span class="text-surface-600 font-medium">{{ doc.location }}</span>
              </span>
            </div>

          </div>

          <!-- Chevron -->
          <i class="pi pi-chevron-right text-surface-400 shrink-0" />
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
        @click="emit('loadMore')"
    />
  </div>

  <div v-if="!hasMore && documents.length > 0" class="flex justify-center py-4 text-surface-400 text-sm">
    All documents loaded.
  </div>
</template>