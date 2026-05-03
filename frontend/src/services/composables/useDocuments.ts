import { ref } from 'vue'
import { apiFetch } from '../fetch/statusCodeChecks'

export interface DocumentQuery {
    sortBy: string
    limit?: number
    type?: string
    name?: string
}

export interface Document {
    id: number
    name: string
    type: string
    added: string
    modified: string
    viewed: string
    deleted: string
}

export function useDocuments() {
    const documents = ref<Document[]>([])
    const loading = ref(false)

    async function fetchDocuments(query: DocumentQuery) {
        loading.value = true
        try {
            const params = new URLSearchParams()
            params.set('sortBy', query.sortBy)
            if (query.limit) params.set('limit', String(query.limit))
            if (query.type) params.set('type', query.type)
            if (query.name) params.set('name', query.name)

            const res = await apiFetch(`/api/protected/records?${params}`)
            if (!res || !res.ok) throw new Error('Failed to fetch')
            documents.value = await res.json()
        } catch (err) {
            console.error(err)
            documents.value = []
        } finally {
            loading.value = false
        }
    }

    return { documents, loading, fetchDocuments }
}