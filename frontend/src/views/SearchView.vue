<template>
  <section class="card">
    <h2>Dokumentsuche</h2>
    <p class="muted">Durchsuche Schaltpläne, Dokumentationen und PDFs über Tags und Volltext.</p>

    <div style="display:grid; gap:0.8rem; margin-bottom:1rem;">
      <div style="display:grid; grid-template-columns:1fr 1fr auto auto auto; gap:0.5rem; align-items:center;">
        <InputText 
          v-model="query" 
          placeholder="Suche nach Begriffen"
          @keydown.enter="search"
          style="width:100%"
        />
        <AutoComplete
          v-model="selectedTags"
          :suggestions="suggestedTags"
          @complete="onTagSuggest"
          placeholder="Tags auswählen"
          multiple
          forceSelection
          style="width:100%"
        />
        <Button v-if="isLoggedIn"
          icon="pi pi-lock" 
          :disabled="!isLoggedIn"
          :severity="privateOnly ? 'warning' : 'secondary'"
          v-tooltip.bottom="privateOnly ? 'Nur private' : 'Private Filter'"
          @click="togglePrivateAndSearch" />
        <Button icon="pi pi-search" v-tooltip.bottom="'Suchen'" @click="search" :loading="isSearching" />
        <Button v-if="isLoggedIn" icon="pi pi-upload" v-tooltip.bottom="'Upload'" severity="success" @click="showUploadDialog = true" />
      </div>
      <div style="display:flex; justify-content:flex-end; align-items:center; gap:0.4rem;">
        <label style="font-size:0.9em;">Ergebnisse pro Seite:</label>
        <Dropdown v-model="selectedLimit" :options="limitOptions" @change="onLimitChange" style="width:7rem;" />
      </div>
    </div>

    <div style="display:flex; gap:1rem; height:calc(100vh - 300px); margin-bottom:1rem;">
      <!-- Treffertabelle (50%, nur wenn nicht versteckt) -->
      <div v-if="!hideSearchResults && !expandDetailPanel" style="flex:2; overflow-y:auto; border:1px solid #e0e0e0; border-radius:4px;">
        <DataTable :value="results" stripedRows
          :sortField="sortField" :sortOrder="sortOrder"
          @sort="onSort"
          removableSort
          @rowClick="selectDocument"
          :rowClass="(data) => selectedDocument?.id === data.document.id ? 'selected-row' : ''">
          <Column field="manufacturer" header="Hersteller" sortable>
            <template #body="{ data }">{{ data.document.manufacturer }}</template>
          </Column>
          <Column field="model" header="Model" sortable>
            <template #body="{ data }">{{ data.document.model }}</template>
          </Column>
          <Column field="subtitle" header="Untertitel" sortable>
            <template #body="{ data }">{{ data.document.subtitle }}</template>
          </Column>
          <Column header="Tags">
            <template #body="{ data }">
              <span v-for="tag in data.document.tags" :key="tag" style="display:inline-block; background:#e0e0e0; border-radius:3px; padding:1px 6px; margin:1px 2px; font-size:0.85em;">{{ tag }}</span>
            </template>
          </Column>
          <Column header="Privat" style="width:5rem; text-align:center;">
            <template #body="{ data }">
              <i v-if="data.document.privateFile" class="pi pi-lock" style="color:#888;" />
            </template>
          </Column>
          <Column field="owner" header="Eigentümer" sortable>
            <template #body="{ data }">{{ data.document.owner }}</template>
          </Column>
        </DataTable>
      </div>

      <!-- Toggle-Leiste 1: zwischen Treffertabelle und Detail -->
      <div v-if="(showDetailPanel || hideSearchResults) && !expandDetailPanel" style="width:20px; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:0.3rem; padding:0.25rem; background:#f5f5f5; border-left:1px solid #e0e0e0; border-right:1px solid #e0e0e0;">
        <Button 
          v-if="!hideSearchResults"
          icon="pi pi-angle-left" 
          severity="secondary" 
          text 
          v-tooltip.right="'Treffer ausblenden'"
          @click="hideSearchResults = true"
          style="padding:0.25rem; font-size:0.9rem;" />
        <Button 
          v-if="!hideSearchResults"
          icon="pi pi-angle-right" 
          severity="secondary" 
          text 
          v-tooltip.right="'Detail ausblenden'"
          @click="showDetailPanel = false; selectedDocument = null; selectedFile = null"
          style="padding:0.25rem; font-size:0.9rem;" />
        <Button 
          v-if="hideSearchResults"
          icon="pi pi-angle-right" 
          severity="secondary" 
          text
          v-tooltip.right="'Treffer anzeigen'"
          @click="hideSearchResults = false"
          style="padding:0.25rem; font-size:0.9rem;" />
      </div>

      <!-- Detail Panel (25% wenn showDetailPanel=true, 50% wenn false) -->
      <div v-if="showDetailPanel && selectedDocument && !expandDetailPanel" :style="{ flex: 1, overflow: 'auto', border: '1px solid #e0e0e0', borderRadius: '4px', background: '#f9f9f9', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '1rem' }">
        <h3 style="margin-top:0;">Details</h3>
        <div style="display:grid; gap:0.5rem; font-size:0.9em;">
          <div><strong>Hersteller:</strong> {{ selectedDocument.manufacturer }}</div>
          <div><strong>Model:</strong> {{ selectedDocument.model }}</div>
          <div><strong>Untertitel:</strong> {{ selectedDocument.subtitle }}</div>
          <div><strong>Beschreibung:</strong> {{ selectedDocument.description || '-' }}</div>
          <div><strong>Eigentümer:</strong> {{ selectedDocument.owner }}</div>
          <div><strong>Privat:</strong> {{ selectedDocument.privateFile ? 'Ja' : 'Nein' }}</div>
          <div><strong>Tags:</strong> {{ (selectedDocument.tags || []).join(', ') || '-' }}</div>
        </div>

        <!-- Datei-Tabelle -->
        <div style="flex:1; overflow-y:auto; border-top:1px solid #e0e0e0; padding-top:1rem;">
          <h4 style="margin-top:0;">Dateien</h4>
          <div style="max-height:15rem; overflow-y:auto;">
            <DataTable v-if="selectedDocument.files && selectedDocument.files.length > 0" 
              :value="selectedDocument.files" 
              stripedRows
              size="small"
              selectionMode="single"
              v-model:selection="selectedFile"
              @rowSelect="onFileSelect">
              <Column field="type" header="Type" style="width:5rem;">
                <template #body="{ data }">{{ data.type }}</template>
              </Column>
              <Column field="name" header="Name">
                <template #body="{ data }">{{ data.name }}</template>
              </Column>
              <Column field="page" header="Page" style="width:4rem; text-align:center;">
                <template #body="{ data }">{{ data.page || '-' }}</template>
              </Column>
            </DataTable>
            <div v-else style="padding:1rem; text-align:center; color:#999;">
              Keine Dateien
            </div>
          </div>
        </div>
      </div>

      <!-- Fileviewer Panel (25% in Normalansicht, 75% in Vollansicht) -->
      <div v-if="selectedDocument && selectedFile && !expandDetailPanel" :style="{ flex: hideSearchResults ? 3 : 1, border: '1px solid #e0e0e0', borderRadius: '4px', background: '#f9f9f9', display: 'flex', flexDirection: 'column', overflow: 'hidden' }">
        <!-- PDF Viewer -->
        <div v-if="isPdfFile(selectedFile)" style="width:100%; height:100%; display:flex; flex-direction:column;">
          <div style="flex-shrink:0; padding:0.5rem; background:#f0f0f0; border-bottom:1px solid #e0e0e0; text-align:center; font-size:0.9em;">
            {{ selectedFile.name }}
          </div>
          <embed v-if="selectedFile.data" :src="'data:application/pdf;base64,' + selectedFile.data" type="application/pdf" style="flex:1; width:100%; border:none;" />
          <div v-else style="flex:1; display:flex; align-items:center; justify-content:center; color:#999;">
            PDF wird geladen...
          </div>
        </div>

        <!-- Image Viewer (mit Zoom, Pan, Rotate, Download) -->
        <div v-else-if="isImageFile(selectedFile)" style="width:100%; height:100%; display:flex; flex-direction:column;">
          <div style="flex-shrink:0; padding:0.5rem; background:#f0f0f0; border-bottom:1px solid #e0e0e0; display:flex; justify-content:flex-end; align-items:center; font-size:0.9em;">
            <div style="display:flex; gap:0.3rem;">
              <Button icon="pi pi-download" severity="secondary" text @click="downloadImage()" v-tooltip.bottom="'Download'" style="padding:0.25rem;" />
            </div>
          </div>
          <div style="flex:1; display:flex; align-items:center; justify-content:center; overflow:auto; background:#fff;">
            <Image 
              v-if="selectedFile.data"
              ref="imageRef"
              :src="'data:' + selectedFile.mimetype + ';base64,' + selectedFile.data"
              :alt="selectedFile.name"
              preview
              imageStyle="object-fit: contain; width: 100%; height: 100%; max-height: 100%;"
              style="width: 100%; height: 100%;"
            />
            <div v-else style="color:#999;">Bild wird geladen...</div>
          </div>
        </div>

        <!-- File Info (für andere Typen) -->
        <div v-else style="width:100%; height:100%; display:flex; flex-direction:column; align-items:center; justify-content:center; padding:2rem; text-align:center;">
          <i class="pi pi-file" style="font-size:3rem; color:#ccc; margin-bottom:1rem;"></i>
          <div style="font-size:0.9em; color:#999;">
            <div><strong>{{ selectedFile.name }}</strong></div>
            <div>{{ selectedFile.type }}</div>
            <div v-if="selectedFile.page">Page {{ selectedFile.page }}</div>
            <div style="margin-top:1rem; font-size:0.85em;">
              Vorschau nicht verfügbar
            </div>
          </div>
        </div>
      </div>

      <!-- Großer Fileview (in Vollbildmodus) -->
      <div v-if="expandDetailPanel && selectedDocument && selectedFile" style="flex:1; border:1px solid #e0e0e0; border-radius:4px; background:#f9f9f9; display:flex; flex-direction:column; overflow:hidden; position:relative;">
        <!-- Close-Button (oben rechts) -->
        <Button 
          icon="pi pi-angle-right"
          severity="secondary"
          text
          v-tooltip.bottom="'Zurück zur Übersicht'"
          @click="expandDetailPanel = false"
          style="position:absolute; top:0.5rem; right:0.5rem; z-index:10; padding:0.5rem;" />

        <!-- PDF Viewer -->
        <div v-if="isPdfFile(selectedFile)" style="width:100%; height:100%; display:flex; flex-direction:column;">
          <div style="flex-shrink:0; padding:0.5rem; background:#f0f0f0; border-bottom:1px solid #e0e0e0; text-align:center; font-size:0.9em;">
            {{ selectedFile.name }}
          </div>
          <embed v-if="selectedFile.data" :src="'data:application/pdf;base64,' + selectedFile.data" type="application/pdf" style="flex:1; width:100%; border:none;" />
          <div v-else style="flex:1; display:flex; align-items:center; justify-content:center; color:#999;">
            PDF wird geladen...
          </div>
        </div>

        <!-- Image Viewer (mit Zoom, Pan, Rotate, Download) -->
        <div v-else-if="isImageFile(selectedFile)" style="width:100%; height:100%; display:flex; flex-direction:column;">
          <div style="flex-shrink:0; padding:0.5rem; background:#f0f0f0; border-bottom:1px solid #e0e0e0; display:flex; justify-content:flex-end; align-items:center; font-size:0.9em;">
            <div style="display:flex; gap:0.3rem;">
              <Button icon="pi pi-download" severity="secondary" text @click="downloadImage()" v-tooltip.bottom="'Download'" style="padding:0.25rem;" />
            </div>
          </div>
          <div style="flex:1; display:flex; align-items:center; justify-content:center; overflow:auto; background:#fff;">
            <Image 
              v-if="selectedFile.data"
              ref="imageRefExpanded"
              :src="'data:' + selectedFile.mimetype + ';base64,' + selectedFile.data"
              :alt="selectedFile.name"
              preview
              imageStyle="object-fit: contain; width: 100%; height: 100%; max-height: 100%;"
              style="width: 100%; height: 100%;"
            />
            <div v-else style="color:#999;">Bild wird geladen...</div>
          </div>
        </div>

        <!-- File Info (für andere Typen) -->
        <div v-else style="width:100%; height:100%; display:flex; flex-direction:column; align-items:center; justify-content:center; padding:2rem; text-align:center;">
          <i class="pi pi-file" style="font-size:3rem; color:#ccc; margin-bottom:1rem;"></i>
          <div style="font-size:0.9em; color:#999;">
            <div><strong>{{ selectedFile.name }}</strong></div>
            <div>{{ selectedFile.type }}</div>
            <div v-if="selectedFile.page">Page {{ selectedFile.page }}</div>
            <div style="margin-top:1rem; font-size:0.85em;">
              Vorschau nicht verfügbar
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="totalResults > 0" style="display:flex; align-items:center; gap:0.8rem; margin-top:1rem; flex-wrap:wrap;">
      <Button icon="pi pi-angle-left" :disabled="currentSkip === 0" @click="prevPage" text />
      <span style="font-size:0.9em;">
        {{ currentSkip + 1 }}–{{ Math.min(currentSkip + selectedLimit, totalResults) }} von {{ totalResults }}
      </span>
      <Button icon="pi pi-angle-right" :disabled="currentSkip + selectedLimit >= totalResults" @click="nextPage" text />
    </div>

    <UploadDialog v-model="showUploadDialog" @uploaded="search" />
  </section>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import InputText from 'primevue/inputtext'
import AutoComplete from 'primevue/autocomplete'
import Button from 'primevue/button'
import Dropdown from 'primevue/dropdown'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import UploadDialog from '../components/UploadDialog.vue'
import Image from 'primevue/image'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'

const { isLoggedIn } = useAuth()
const { info } = useToast()

const query = ref('')
const selectedTags = ref([])
const suggestedTags = ref([])
const results = ref([])
const showUploadDialog = ref(false)
const selectedDocument = ref(null)
const selectedFile = ref(null)
const showDetailPanel = ref(false)
const expandDetailPanel = ref(false)
const hideSearchResults = ref(false)
const isSearching = ref(false)

// Image Viewer - PrimeVue Image
const imageRef = ref(null)
const imageRefExpanded = ref(null)

const limitOptions = [10, 20, 50, 100]
const selectedLimit = ref(20)
const currentSkip = ref(0)
const totalResults = ref(0)
const sortField = ref(null)
const sortOrder = ref(null)
const privateOnly = ref(false)

function toTags() {
  return selectedTags.value
    .map((tag) => String(tag || '').trim())
    .filter(Boolean)
}

async function onTagSuggest(event) {
  const queryText = (event.query || '').trim()
  if (!queryText) {
    suggestedTags.value = []
    return
  }

  try {
    const { data } = await api.get('/api/v1/tags/suggest', {
      params: { q: queryText, limit: 10 },
    })
    suggestedTags.value = (data.tags || [])
      .map((tag) => (typeof tag === 'string' ? tag : tag?.name))
      .map((tag) => String(tag || '').trim())
      .filter(Boolean)
  } catch (_err) {
    suggestedTags.value = []
  }
}

function onLimitChange() {
  currentSkip.value = 0
  search()
}

function onSort(event) {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  currentSkip.value = 0
  search()
}

function prevPage() {
  currentSkip.value = Math.max(0, currentSkip.value - selectedLimit.value)
  search()
}

function nextPage() {
  currentSkip.value = currentSkip.value + selectedLimit.value
  search()
}

function togglePrivateAndSearch() {
  privateOnly.value = !privateOnly.value
  currentSkip.value = 0
  search()
}

async function search() {
  try {
    isSearching.value = true
    
    // Reset detail panel when searching
    selectedDocument.value = null
    selectedFile.value = null
    showDetailPanel.value = false
    expandDetailPanel.value = false

    // Guests cannot search private documents
    if (!isLoggedIn.value) {
      privateOnly.value = false
    }
    
    const params = new URLSearchParams()
    params.set('q', query.value)
    toTags().forEach((tag) => params.append('tag', tag))
    params.set('skip', String(currentSkip.value))
    params.set('limit', String(selectedLimit.value))
    if (sortField.value) {
      params.set('sortField', sortField.value)
      params.set('sortOrder', String(sortOrder.value ?? 1))
    }
    if (isLoggedIn.value) {
      params.set('privateOnly', privateOnly.value ? 'true' : 'false')
    }
    const { data } = await api.get(`/api/v1/documents/search?${params.toString()}`)
    results.value = data.results || []
    totalResults.value = data.total ?? (data.results || []).length
    
    const count = totalResults.value
    const countText = count === 1 ? '1 Dokument' : `${count} Dokumente`
    info(`${countText} gefunden`)
  } catch (err) {
    info(`Fehler bei der Suche`)
  } finally {
    isSearching.value = false
  }
}

function selectDocument(event) {
  selectedDocument.value = event.data.document
  selectedFile.value = null
  showDetailPanel.value = true
  hideSearchResults.value = false
  
  // Erste Datei automatisch selektieren, falls vorhanden
  if (selectedDocument.value.files && selectedDocument.value.files.length > 0) {
    selectedFile.value = selectedDocument.value.files[0]
    // Lade die Datei automatisch
    if (!selectedFile.value.data) {
      loadFileData(selectedFile.value)
    }
  }
}

function onFileSelect(event) {
  selectedFile.value = event.data
  
  // Lade die Datei, falls nicht bereits vorhanden
  if (!selectedFile.value.data) {
    loadFileData(selectedFile.value)
  }
}

function isPdfFile(file) {
  if (!file) return false
  return file.mimetype === 'application/pdf' || file.type === 'pdf' || file.name?.endsWith('.pdf')
}

function isImageFile(file) {
  if (!file) return false
  const imageTypes = ['image/jpeg', 'image/png', 'image/bmp', 'image/tiff', 'image/x-tiff', 'image/gif']
  const imageMimes = ['image/jpeg', 'image/png', 'image/bmp', 'image/tiff', 'image/x-tiff', 'image/vnd.tiff', 'image/gif']
  const imageExts = ['.png', '.jpg', '.jpeg', '.bmp', '.tif', '.tiff', '.gif']
  
  if (imageMimes.includes(file.mimetype)) return true
  if (imageTypes.includes(file.type)) return true
  return imageExts.some(ext => file.name?.toLowerCase().endsWith(ext))
}

async function loadFileData(file) {
  try {
    const { data } = await api.get(`/api/v1/documents/${selectedDocument.value.id}/files/${encodeURIComponent(file.name)}`)
    if (data && data.data) {
      file.data = data.data
    }
  } catch (err) {
    info('Fehler beim Laden der Datei')
  }
}

function getFilePreviewUrl(file) {
  if (!file) return ''
  // Wenn die Datei bereits geladen ist (base64 in data)
  if (file.data) {
    if (isPdfFile(file)) {
      return 'data:application/pdf;base64,' + file.data
    }
    return 'data:' + (file.mimetype || 'image/*') + ';base64,' + file.data
  }
  return ''
}

// Image Viewer Funktionen
function initImageZoom() {
  // PrimeVue Image hat native Zoom/Rotate - nichts zu initialisieren
}

function downloadImage() {
  if (!selectedFile.value || !selectedFile.value.data) return
  
  const link = document.createElement('a')
  link.href = 'data:' + selectedFile.value.mimetype + ';base64,' + selectedFile.value.data
  link.download = selectedFile.value.name
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<style scoped>
:deep(.selected-row) {
  background-color: #e3f2fd !important;
}

:deep(.selected-row:hover) {
  background-color: #bbdefb !important;
}
</style>
