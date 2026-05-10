export interface Document {
    id: string
    name: string
    types: string[]
    dateOfDoc: string
    dateFiled: string
    location: string
    sensitive: boolean
    thumbnail: string
}

export interface DocumentDetail extends Document {
    pages: string[]
}

export interface DocumentType {
    documentLabel: string
    documentLabelValue: string
}

export interface Filters {
    docDateFrom: Date | null
    docDateTo: Date | null
    filedDateFrom: Date | null
    filedDateTo: Date | null
    types: string[]
}