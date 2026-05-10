<template>
  <Dialog v-model:visible="visible" header="Dokument hochladen" :modal="true" style="width: 600px">
    <div style="display:grid; gap:1rem;">
      <!-- Manufacturer & Model -->
      <div style="display:grid; grid-template-columns: 1fr 1fr; gap:0.8rem;">
        <div>
          <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Hersteller *</label>
          <AutoComplete
            v-model="form.manufacturer"
            :suggestions="suggestedManufacturers"
            @complete="onManufacturerSuggest"
            placeholder="Hersteller eingeben..."
            :typeahead="false"
            style="width:100%"
          />
        </div>
        <div>
          <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Modell *</label>
          <InputText v-model="form.model" placeholder="Modell eingeben..." style="width:100%" />
        </div>
      </div>

      <!-- Subtitle -->
      <span class="p-float-label">
        <InputText id="subtitle" v-model="form.subtitle" style="width:100%" />
        <label for="subtitle">Untertitel</label>
      </span>

      <!-- Tags with Autocomplete -->
      <div>
        <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Tags</label>
        <AutoComplete
          v-model="currentTagQuery"
          :suggestions="suggestedTags"
          @complete="onTagSuggest"
          @item-select="onTagSelect"
          @keydown.enter.prevent="onTagEnter"
          placeholder="Tags eingeben..."
          :typeahead="false"
          style="width:100%"
        />
        <div v-if="selectedTags.length" style="display:flex; flex-wrap:wrap; gap:0.4rem; margin-top:0.5rem;">
          <Chip
            v-for="tag in selectedTags"
            :key="tag"
            :label="tag"
            removable
            @remove="removeTag(tag)"
          />
        </div>
        <small class="p-text-secondary">Beim Tippen werden Vorschläge angezeigt. Mit Enter wird der aktuelle Text als Tag hinzugefügt.</small>
      </div>

      <!-- Description -->
      <span class="p-float-label">
        <Textarea
          id="description"
          v-model="form.description"
          style="width:100%; min-height:80px"
          auto-resize
        />
        <label for="description">Beschreibung</label>
      </span>

      <!-- Private File & Owner -->
      <div style="display:grid; grid-template-columns: auto 1fr; gap:0.8rem; align-items:center;">
        <div>
          <InputSwitch v-model="form.privateFile" />
        </div>
        <label>Privates Dokument</label>
      </div>

      <span v-if="form.privateFile" class="p-float-label">
        <InputText id="owner" v-model="form.owner" style="width:100%" />
        <label for="owner">Besitzer *</label>
      </span>

      <!-- File Input -->
      <div>
        <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Datei *</label>
        <FileUpload
          name="file"
          @select="onFileSelect"
          :show-upload-button="false"
          :show-cancel-button="false"
          accept=".pdf,.jpg,.jpeg,.png,.gif"
          style="width:100%"
        />
      </div>

      <!-- Document Type -->
      <div v-if="form.files.length > 0">
        <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Dokumenttyp *</label>
        <Dropdown v-model="form.files[0].type" :options="docTypes" option-label="label" option-value="value" style="width:100%" />
      </div>

      <!-- Error Message -->
      <Message v-if="errorMessage" severity="error">{{ errorMessage }}</Message>

      <!-- Buttons -->
      <div style="display:flex; gap:0.5rem; justify-content:flex-end; margin-top:1rem;">
        <Button label="Abbrechen" severity="secondary" @click="close" />
        <Button label="Hochladen" @click="submit" :loading="isSubmitting" />
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import InputSwitch from 'primevue/inputswitch'
import FileUpload from 'primevue/fileupload'
import AutoComplete from 'primevue/autocomplete'
import Chip from 'primevue/chip'
import Dropdown from 'primevue/dropdown'
import Button from 'primevue/button'
import Message from 'primevue/message'
import api from '../services/api'

const props = defineProps({
  modelValue: Boolean,
})

const emit = defineEmits(['update:modelValue', 'uploaded'])

const visible = ref(props.modelValue)
const isSubmitting = ref(false)
const errorMessage = ref('')
const suggestedTags = ref([])
const selectedTags = ref([])
const currentTagQuery = ref('')
const suggestedManufacturers = ref([])

const docTypes = [
  { label: 'Schaltplan', value: 'schematic' },
  { label: 'Bedienungsanleitung', value: 'manual' },
  { label: 'Service-Dokumentation', value: 'service' },
  { label: 'Zertifikat', value: 'certificate' },
]

const form = ref({
  manufacturer: '',
  model: '',
  subtitle: '',
  description: '',
  privateFile: false,
  owner: '',
  files: [],
})

watch(
  () => props.modelValue,
  (newVal) => {
    visible.value = newVal
  }
)

watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

function close() {
  visible.value = false
  resetForm()
}

function resetForm() {
  form.value = {
    manufacturer: '',
    model: '',
    subtitle: '',
    description: '',
    privateFile: false,
    owner: '',
    files: [],
  }
  selectedTags.value = []
  errorMessage.value = ''
}

async function onTagSuggest(event) {
  currentTagQuery.value = event.query || ''

  if (!event.query) {
    suggestedTags.value = []
    return
  }

  try {
    const { data } = await api.get('/api/v1/tags/suggest', {
      params: { q: event.query, limit: 10 },
    })
    suggestedTags.value = (data.tags || [])
      .map((tag) => (typeof tag === 'string' ? tag : tag?.name))
      .map((tag) => String(tag || '').trim())
      .filter(Boolean)
  } catch (err) {
    suggestedTags.value = []
  }
}

async function onManufacturerSuggest(event) {
  const query = event.query || ''

  if (!query) {
    suggestedManufacturers.value = []
    return
  }

  try {
    const { data } = await api.get('/api/v1/manufacturers/suggest', {
      params: { q: query, limit: 10 },
    })
    suggestedManufacturers.value = (data.manufacturers || [])
      .map((m) => String(m || '').trim())
      .filter(Boolean)
  } catch (err) {
    suggestedManufacturers.value = []
  }
}

function normalizeTag(tag) {
  return String(tag || '').trim()
}

function hasTag(tag) {
  const needle = normalizeTag(tag).toLowerCase()
  if (!needle) {
    return false
  }
  return selectedTags.value.some((existing) => normalizeTag(existing).toLowerCase() === needle)
}

function onTagEnter() {
  const tag = normalizeTag(currentTagQuery.value)
  if (!tag || hasTag(tag)) {
    return
  }

  selectedTags.value = [...selectedTags.value, tag]
  currentTagQuery.value = ''
}

function onTagSelect(event) {
  const value = event?.value
  const tag = normalizeTag(typeof value === 'string' ? value : value?.name)
  if (!tag || hasTag(tag)) {
    currentTagQuery.value = ''
    return
  }

  selectedTags.value = [...selectedTags.value, tag]
  currentTagQuery.value = ''
}

function removeTag(tag) {
  const needle = normalizeTag(tag).toLowerCase()
  selectedTags.value = selectedTags.value.filter((entry) => normalizeTag(entry).toLowerCase() !== needle)
}

async function onFileSelect(event) {
  const file = event.files[0]
  if (file) {
    const data = await fileToBase64(file)
    form.value.files = [
      {
        name: file.name,
        page: 1,
        mimetype: file.type,
        type: 'schematic',
        data,
      },
    ]
  }
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      const marker = 'base64,'
      const idx = result.indexOf(marker)
      if (idx < 0) {
        reject(new Error('could not read base64 payload'))
        return
      }
      resolve(result.substring(idx + marker.length))
    }
    reader.onerror = () => reject(reader.error || new Error('file read failed'))
    reader.readAsDataURL(file)
  })
}

async function submit() {
  errorMessage.value = ''

  if (!form.value.manufacturer.trim()) {
    errorMessage.value = 'Hersteller ist erforderlich'
    return
  }

  if (!form.value.model.trim()) {
    errorMessage.value = 'Modell ist erforderlich'
    return
  }

  if (form.value.files.length === 0) {
    errorMessage.value = 'Bitte wählen Sie eine Datei aus'
    return
  }

  if (!form.value.files[0].type) {
    errorMessage.value = 'Bitte wählen Sie einen Dokumenttyp aus'
    return
  }

  if (form.value.privateFile && !form.value.owner.trim()) {
    errorMessage.value = 'Besitzer ist erforderlich für private Dokumente'
    return
  }

  isSubmitting.value = true

  try {
    const doc = {
      id: `doc-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      createdAt: new Date().toISOString(),
      lastModifiedAt: new Date().toISOString(),
      manufacturer: form.value.manufacturer,
      model: form.value.model,
      subtitle: form.value.subtitle,
      tags: selectedTags.value.map((t) => normalizeTag(t)).filter(Boolean),
      description: form.value.description,
      privateFile: form.value.privateFile,
      owner: form.value.owner || 'admin',
      files: form.value.files,
    }

    await api.post('/api/v1/documents/index', doc)
    close()
    emit('uploaded')
  } catch (err) {
    errorMessage.value = err?.response?.data?.error || 'Upload fehlgeschlagen'
  } finally {
    isSubmitting.value = false
  }
}
</script>
