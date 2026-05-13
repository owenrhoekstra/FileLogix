<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import MultiSelect from 'primevue/multiselect'
import DatePicker from 'primevue/datepicker'
import SelectButton from 'primevue/selectbutton'
import mainMenuBar from '../../components/mainMenuBar.vue'
import footerBar from '../../components/footerBar.vue'
import { apiFetch } from '../../services/fetch/statusCodeChecks.ts'
import { useRouter } from 'vue-router'
import type { FormSubmitEvent } from '@primevue/forms'
import { convertAllToWebP } from '../../services/addNewRecord/imageConvert.ts'
import { Cropper } from 'vue-advanced-cropper'
import type { CropperResult } from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'
import Dialog from 'primevue/dialog'

// ---- Router ----
const router = useRouter()

// ---- Types ----
type DocumentType = {
  documentLabel: string
  documentLabelValue: string
}

type FormValues = {
  documentName: string
  documentType: string[]
  documentDate: Date | null
  documentSensitivity: boolean | null
}

// ---- State ----
const toast = useToast()
const types = ref<DocumentType[]>([])
const selectedFiles = ref<File[]>([])
const previewUrls = ref<string[]>([])
const fileInput = ref<HTMLInputElement | null>(null)
const photoError = ref<string | null>(null)
const submitting = ref(false)
const cropQueue = ref<File[]>([])
const cropModalVisible = ref(false)
const currentCropFile = ref<File | null>(null)
const currentCropUrl = ref<string | null>(null)
const cropperRef = ref<{ getResult(): CropperResult } | null>(null)

const initialValues = reactive<FormValues>({
  documentName: '',
  documentType: [],
  documentDate: null,
  documentSensitivity: null,
})

// ---- Lifecycle ----
onMounted(async () => {
  const res = await apiFetch('/api/protected/form-metadata')
  if (!res) throw new Error('No response')
  const data = await res.json()
  types.value = data.documentTypes
})

// ---- Validation ----
const resolver = (e: any) => {
  const values = e.values as FormValues
  const errors: Record<string, any> = {}

  if (!values.documentName)
    errors.documentName = [{ message: 'Document Name is required.' }]

  if (!values.documentType?.length)
    errors.documentType = [{ message: 'Select at least one document type.' }]

  if (!values.documentDate)
    errors.documentDate = [{ message: 'Select date of document.' }]

  if (values.documentSensitivity === null)
    errors.documentSensitivity = [{ message: 'Select document sensitivity.' }]

  photoError.value = selectedFiles.value.length === 0 ? 'At least one photo required' : null

  return { values, errors }
}

// ---- Handlers ----
const onFormSubmit = async (event: FormSubmitEvent) => {
  const { valid, values } = event
  if (!valid) return
  if (selectedFiles.value.length === 0) {
    photoError.value = 'At least one photo required'
    return
  }

  submitting.value = true

  const formData = new FormData()
  formData.append('documentName', values.documentName)
  const date = new Date(values.documentDate)
  formData.append('documentDate', date.toISOString().split('T')[0])
  formData.append('documentSensitivity', String(values.documentSensitivity))
  values.documentType.forEach((type: string) => formData.append('documentType', type))
  selectedFiles.value.forEach(file => formData.append('photos', file))

  try {
    const res = await apiFetch('/api/protected/records', {
      method: 'POST',
      body: formData
    })

    if (!res || !res.ok) throw new Error('request failed')

    toast.add({ severity: 'success', summary: 'Success', detail: 'Record created', life: 2000 })
    const data = await res.json()
    router.push(`/print/${data.id}`)
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create record', life: 3000 })
  } finally {
    submitting.value = false
  }
}

const onFilesSelected = async (e: Event) => {
  const input = e.target as HTMLInputElement
  if (!input.files) return
  console.log('files selected:', input.files.length)
  try {
    const converted = await convertAllToWebP(input.files)
    console.log('converted:', converted.length)
    cropQueue.value.push(...converted)
    console.log('queue after push:', cropQueue.value.length)
    processNextCrop()
  } catch (err) {
    console.error('onFilesSelected error:', err)
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to process images', life: 3000 })
  } finally {
    if (fileInput.value) fileInput.value.value = ''
  }
}

const processNextCrop = () => {
  console.log('processNextCrop called, queue length:', cropQueue.value.length)
  if (cropQueue.value.length === 0) return
  currentCropFile.value = cropQueue.value[0]
  currentCropUrl.value = URL.createObjectURL(currentCropFile.value)
  cropModalVisible.value = true
  console.log('cropModalVisible set to true')
}

const onCropSave = () => {
  const result = cropperRef.value?.getResult()
  if (!result?.canvas) return
  result.canvas.toBlob((blob: Blob | null) => {
    if (!blob) return
    const file = new File([blob], currentCropFile.value!.name, { type: 'image/webp' })
    selectedFiles.value.push(file)
    previewUrls.value.push(URL.createObjectURL(file))
    cropQueue.value.shift()
    URL.revokeObjectURL(currentCropUrl.value!)
    cropModalVisible.value = false
    processNextCrop()
  }, 'image/webp', 0.9)
}

const triggerFile = () => fileInput.value?.click()

const removeFile = (index: number) => {
  URL.revokeObjectURL(previewUrls.value[index])
  selectedFiles.value.splice(index, 1)
  previewUrls.value.splice(index, 1)
}
</script>

<template>
  <mainMenuBar />

  <div class="min-h-[calc(100vh-180px)] flex justify-center items-center p-4">
    <Form
        v-slot="$form"
        :initialValues="initialValues"
        :resolver="resolver"
        @submit="onFormSubmit"
        class="flex flex-col gap-4 w-full"
    >
      <div class="flex flex-col gap-4">

        <InputText name="documentName" placeholder="Document Name" fluid />
        <Message v-if="$form.documentName?.invalid" severity="error" size="small" variant="simple">
          {{ $form.documentName.error?.message }}
        </Message>

        <MultiSelect
            name="documentType"
            :options="types"
            optionLabel="documentLabel"
            optionValue="documentLabelValue"
            filter
            placeholder="Select Document Type"
            :maxSelectedLabels="3"
        />
        <Message v-if="$form.documentType?.invalid" severity="error" size="small" variant="simple">
          {{ $form.documentType.error?.message }}
        </Message>

        <DatePicker name="documentDate" placeholder="Date of Document" />
        <Message v-if="$form.documentDate?.invalid" severity="error" size="small" variant="simple">
          {{ $form.documentDate.error?.message }}
        </Message>

        <SelectButton
            fluid
            class="justify-center w-full"
            name="documentSensitivity"
            :options="[
            { label: 'Sensitive', sensitive: true },
            { label: 'Not Sensitive', sensitive: false }
          ]"
            optionLabel="label"
            optionValue="sensitive"
            :allowEmpty="true"
        />
        <Message v-if="$form.documentSensitivity?.invalid" severity="error" size="small" variant="simple">
          {{ $form.documentSensitivity.error?.message }}
        </Message>

        <Button type="button" label="Add photo" @click="triggerFile" />
        <Message v-if="photoError" severity="error" size="small" variant="simple">
          {{ photoError }}
        </Message>

        <input
            ref="fileInput"
            type="file"
            accept="image/*"
            capture="environment"
            multiple
            style="display:none"
            @change="onFilesSelected"
        />
      </div>

      <div v-for="(f, index) in selectedFiles" :key="index" class="flex items-center gap-3">
        <img :src="previewUrls[index]" class="w-16 h-16 object-cover rounded" />
        <span class="text-sm">{{ f.name }}</span>
        <Button type="button" icon="pi pi-times" severity="danger" text @click="removeFile(index)" />
      </div>

      <Button
          type="submit"
          severity="secondary"
          label="Submit"
          :loading="submitting"
          :disabled="submitting"
      />
    </Form>
    <Dialog v-model:visible="cropModalVisible" modal header="Crop Photo" :closable="false" style="width: 90vw; max-width: 600px">
      <Cropper
          ref="cropperRef"
          :src="currentCropUrl"
          :stencil-props="{ aspectRatio: null }"
          style="max-height: 70vh"
      />
      <template #footer>
        <Button label="Save" @click="onCropSave" />
      </template>
    </Dialog>
  </div>

  <footerBar />
</template>