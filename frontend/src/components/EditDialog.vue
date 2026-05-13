<template>
  <Dialog v-model:visible="visible" header="Dokument bearbeiten" :modal="true" style="width: 1000px">
    <div style="display:grid; gap:1rem;">
      <!-- Manufacturer & Model (read-only) -->
      <div style="display:grid; grid-template-columns: 1fr 1fr; gap:0.8rem;">
        <div>
          <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Hersteller</label>
          <InputText v-model="form.manufacturer" disabled style="width:100%" />
        </div>
        <div>
          <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Modell</label>
          <InputText v-model="form.model" disabled style="width:100%" />
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

      <!-- Private File & Owner (read-only) -->
      <div style="display:grid; grid-template-columns: auto 1fr; gap:0.8rem; align-items:center;">
        <div>
          <InputSwitch v-model="form.privateFile" :disabled="true" />
        </div>
        <label>Privates Dokument</label>
      </div>

      <span v-if="form.privateFile" class="p-float-label">
        <InputText id="owner" v-model="form.owner" disabled style="width:100%" />
        <label for="owner">Besitzer</label>
      </span>

      <!-- File Input -->
      <div>
        <label style="display:block; margin-bottom:0.3rem; font-size:0.85rem">Dateien</label>
        <FileUpload
          name="file"
          @select="onFileSelect"
          :show-upload-button="false"
          :show-cancel-button="false"
          accept=".pdf,.jpg,.jpeg,.png,.gif"
          multiple
          style="width:100%"
        />
        <small class="p-text-secondary">Neue Dateien können hinzugefügt werden. Bestehende Dateien können gelöscht werden.</small>
      </div>

      <!-- File List with Types -->
      <div v-if="form.files.length > 0">
        <label style="display:block; margin-bottom:0.5rem; font-size:0.85rem">Dateien</label>
        <div style="display:grid; gap:0.8rem;">
          <div v-for="(file, index) in form.files" :key="index" style="display:grid; grid-template-columns: 1fr 200px auto; gap:0.5rem; align-items:center; padding:0.8rem; border:1px solid #e0e0e0; border-radius:4px;" :style="{ opacity: file.deleted ? 0.5 : 1, textDecoration: file.deleted ? 'line-through' : 'none' }">
            <div style="display:flex; flex-direction:column; gap:0.2rem;">
              <span style="font-weight:500; font-size:0.95rem;">{{ file.name }}</span>
              <span style="font-size:0.85rem; color:#666;">{{ file.mimetype }}</span>
              <span v-if="file.isNew" style="font-size:0.75rem; color:#2196F3; font-weight:500;">✓ Neu</span>
              <span v-if="file.deleted" style="font-size:0.75rem; color:#f44336;">⊗ Zum Löschen markiert</span>
            </div>
            <Dropdown 
              v-model="form.files[index].type" 
              :options="docTypes" 
              option-label="label" 
              option-value="value" 
              placeholder="Typ wählen..."
              :disabled="file.isNew && !file.type"
            />
            <Button 
              icon="pi pi-trash" 
              severity="danger" 
              text 
              rounded 
              size="small" 
              @click="toggleDeleteFile(index)"
              :label="file.deleted ? 'Wiederherstellen' : ''"
            />
          </div>
        </div>
      </div>

      <!-- Error Message -->
      <Message v-if="errorMessage" severity="error">{{ errorMessage }}</Message>

      <!-- Buttons -->
      <div style="display:flex; gap:0.5rem; justify-content:flex-end; margin-top:1rem;">
        <Button label="Abbrechen" severity="secondary" @click="close" />
        <Button label="Speichern" @click="submit" :loading="isSubmitting" />
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
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
  document: Object,
})

const emit = defineEmits(['update:modelValue', 'updated'])

const visible = ref(props.modelValue)
const isSubmitting = ref(false)
const errorMessage = ref('')
const suggestedTags = ref([])
const selectedTags = ref([])
const currentTagQuery = ref('')

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
    if (newVal && props.document) {
      loadDocument(props.document)
    }
  }
)

watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

function loadDocument(doc) {
  form.value = {
    manufacturer: doc.manufacturer || '',
    model: doc.model || '',
    subtitle: doc.subtitle || '',
    description: doc.description || '',
    privateFile: doc.privateFile || false,
    owner: doc.owner || '',
    files: (doc.files || []).map((f) => ({
      ...f,
      isNew: false,
      deleted: false,
    })),
  }
  selectedTags.value = [...(doc.tags || [])]
  errorMessage.value = ''
}

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
  const newFiles = event.files || []
  
  // Get list of file names already added
  const existingNames = new Set(form.value.files.map(f => f.name))
  
  // Only add files that are not already in the list
  for (const file of newFiles) {
    if (!existingNames.has(file.name)) {
      const data = await fileToBase64(file)
      form.value.files.push({
        name: file.name,
        page: 1,
        mimetype: file.type,
        type: '', // User must select
        data,
        isNew: true,
        deleted: false,
      })
      existingNames.add(file.name)
    }
  }
}

function toggleDeleteFile(index) {
  form.value.files[index].deleted = !form.value.files[index].deleted
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

  if (form.value.files.length === 0) {
    errorMessage.value = 'Bitte behalten Sie mindestens eine Datei'
    return
  }

  // Check that all non-deleted files have a type selected
  for (let i = 0; i < form.value.files.length; i++) {
    if (!form.value.files[i].deleted && !form.value.files[i].type) {
      errorMessage.value = `Bitte wählen Sie einen Dokumenttyp für "${form.value.files[i].name}" aus`
      return
    }
  }

  isSubmitting.value = true

  try {
    // Only include new files with data, exclude deleted files
    const newFiles = form.value.files
      .filter(f => f.isNew && !f.deleted)
      .map(f => ({
        name: f.name,
        type: f.type,
        mimetype: f.mimetype,
        page: f.page || 1,
        data: f.data,
      }))

    // Files to delete (existing files marked as deleted)
    const deletedFiles = form.value.files
      .filter(f => !f.isNew && f.deleted)
      .map(f => ({
        name: f.name,
      }))

    // Keep all non-deleted files (existing + new)
    const allFiles = form.value.files
      .filter(f => !f.deleted)
      .map(f => ({
        name: f.name,
        type: f.type,
        mimetype: f.mimetype,
        page: f.page || 1,
        ...(f.isNew ? { data: f.data } : {}),
      }))

    const payload = {
      subtitle: form.value.subtitle,
      tags: selectedTags.value.map((t) => normalizeTag(t)).filter(Boolean),
      description: form.value.description,
      newFiles,
      deletedFiles,
    }

    await api.patch(`/api/v1/documents/${props.document.id}`, payload)
    close()
    emit('updated')
  } catch (err) {
    errorMessage.value = err?.response?.data?.error || 'Bearbeitung fehlgeschlagen'
  } finally {
    isSubmitting.value = false
  }
}
</script>
