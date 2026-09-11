<template>
  <section class="card">
    <div style="display:flex; align-items:center; gap:0.5rem; margin-bottom:1rem;" class="title-section">
      <Button 
        icon="pi pi-arrow-left" 
        severity="secondary" 
        text
        v-tooltip.bottom="'Zur Startseite'"
        @click="goHome()"
        class="back-button-mobile"
        style="padding:0.25rem; font-size:1rem;"
      />
      <h2 style="margin:0;">Dokumentsuche</h2>
    </div>
    <!-- Desktop: Always visible help text -->
    <p class="muted help-text-desktop">Durchsuche {{ totalDocuments }} Schaltpläne, Dokumentationen und PDFs über Tags und Volltext.<br/> Suchoperatoren: <code>+Begriff</code> (erforderlich), <code>-Begriff</code> (ausschließen), <code>Begriff*</code> (Prefix-Matching). Mehrere Begriffe (ohne + oder -) werden als ODER verknüpft. <br/>Tags werden immer UND verknüpft.</p>
    
    <!-- Mobile: Collapsible help section -->
    <Fieldset 
      legend="Hilfe zu Suche" 
      :toggleable="true" 
      v-model:collapsed="showHelpPanel"
      style="margin-bottom:1rem; margin-top:0;"
      class="help-fieldset"
    >
      <p class="muted" style="margin:0;">
        Durchsuche {{ totalDocuments }} Schaltpläne, Dokumentationen und PDFs über Tags und Volltext.<br/> 
        Suchoperatoren: <code>+Begriff</code> (erforderlich), <code>-Begriff</code> (ausschließen), <code>Begriff*</code> (Prefix-Matching). Mehrere Begriffe (ohne + oder -) werden als ODER verknüpft. <br/>
        Tags werden immer UND verknüpft.
      </p>
    </Fieldset>
    <div style="display:grid; gap:0.8rem; margin-bottom:1rem;">
      <div class="search-controls" style="display:grid; grid-template-columns:1fr 1fr auto; gap:0.5rem; align-items:center;">
        <InputText 
          v-model="query" 
          placeholder="Suche nach Begriffen"
          @keydown.enter="search"
          class="search-query"
          style="width:100%"
        />
        <AutoComplete
          v-model="selectedTags"
          :suggestions="suggestedTags"
          @complete="onTagSuggest"
          placeholder="Tags auswählen"
          multiple
          forceSelection
          class="search-tags"
          style="width:100%"
        />
        <div class="search-actions" style="display:flex; gap:0.5rem; justify-content:flex-end; align-items:center;">
          <Button v-if="isLoggedIn"
            icon="pi pi-lock" 
            :disabled="!isLoggedIn"
            :severity="privateOnly ? 'warning' : 'secondary'"
            v-tooltip.bottom="privateOnly ? 'Nur private' : 'Private Filter'"
            @click="togglePrivateAndSearch" />
          <Button icon="pi pi-search" v-tooltip.bottom="'Suchen'" @click="search" :loading="isSearching" />
          <Button v-if="isLoggedIn" icon="pi pi-upload" v-tooltip.bottom="'Upload'" severity="success" @click="showUploadDialog = true" />
          <Button v-if="selectedDocument" icon="pi pi-link" v-tooltip.bottom="'Link kopieren'" severity="secondary" @click="copyDocumentLink()" />
          <Button v-if="isLoggedIn && selectedDocument" icon="pi pi-pencil" v-tooltip.bottom="'Bearbeiten'" severity="info" @click="showEditDialog = true" />
          <Button v-if="isLoggedIn && selectedDocument" icon="pi pi-trash" v-tooltip.bottom="'Löschen'" severity="danger" @click="confirmDeleteDocument" />
        </div>
      </div>
      <div style="display:flex; justify-content:flex-end; align-items:center; gap:0.4rem;">
        <label style="font-size:0.9em;">Ergebnisse pro Seite:</label>
        <Dropdown v-model="selectedLimit" :options="limitOptions" @change="onLimitChange" style="width:7rem;" />
      </div>
    </div>

    <div class="results-layout" style="display:flex; gap:1rem; height:calc(100vh - 300px); margin-bottom:1rem;">
      <!-- Treffertabelle (50%, nur wenn nicht versteckt) -->
      <div v-if="!hideSearchResults && !expandDetailPanel" style="flex:2; border:1px solid #e0e0e0; border-radius:4px; overflow:hidden;">
        <DataTable :value="results" stripedRows
          scrollable
          scrollHeight="flex"
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
      <div v-if="(showDetailPanel || hideSearchResults) && !expandDetailPanel" class="splitter-bar" style="width:20px; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:0.3rem; padding:0.25rem; background:#f5f5f5; border-left:1px solid #e0e0e0; border-right:1px solid #e0e0e0;">
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
      <div v-if="showDetailPanel && selectedDocument && !expandDetailPanel" class="detail-panel" :style="{ flex: 1, overflow: 'auto', border: '1px solid #e0e0e0', borderRadius: '4px', background: '#f9f9f9', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '1rem' }">
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
        <div class="detail-files-section" style="flex:1; overflow-y:auto; border-top:1px solid #e0e0e0; padding-top:1rem;">
          <h4 style="margin-top:0;">Dateien</h4>
          <div class="detail-files-table-wrapper" style="max-height:15rem; overflow-y:auto;">
            <DataTable v-if="selectedDocument.files && selectedDocument.files.length > 0" 
              :value="selectedDocument.files" 
              stripedRows
              size="small"
              selectionMode="single"
              v-model:selection="selectedFile"
              @rowSelect="onFileSelect"
              @rowUnselect="onFileUnselect">
              <Column field="type" header="Type" style="width:5rem;">
                <template #body="{ data }">
                  <span v-if="isMobileView" class="doc-type-icon" :title="getDocTypeLabel(data.type) || data.type">
                    <i :class="getDocTypeIcon(data.type)"></i>
                  </span>
                  <span v-else>{{ getDocTypeLabel(data.type) || data.type }}</span>
                </template>
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
      <div v-if="!isMobileView && selectedDocument && selectedFile && !expandDetailPanel" :style="{ flex: hideSearchResults ? 3 : 1, border: '1px solid #e0e0e0', borderRadius: '4px', background: '#f9f9f9', display: 'flex', flexDirection: 'column', overflow: 'hidden' }">
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
      <div v-if="!isMobileView && expandDetailPanel && selectedDocument && selectedFile" class="detail-panel-expanded" style="flex:1; display:flex; gap:1rem; border:1px solid #e0e0e0; border-radius:4px; background:#f9f9f9; overflow:hidden; position:relative;">
        <!-- Close-Button (oben rechts) -->
        <Button 
          icon="pi pi-angle-right"
          severity="secondary"
          text
          v-tooltip.bottom="'Zurück zur Übersicht'"
          @click="expandDetailPanel = false"
          style="position:absolute; top:0.5rem; right:0.5rem; z-index:10; padding:0.5rem;" />

        <!-- Detail Panel (linke Seite) -->
        <div class="detail-panel-expanded-left" style="flex:1; overflow:auto; padding:1rem; border-right:1px solid #e0e0e0; display:flex; flex-direction:column; gap:1rem;">
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
          <div class="detail-files-section-expanded" style="flex:1; overflow-y:auto; border-top:1px solid #e0e0e0; padding-top:1rem;">
            <h4 style="margin-top:0;">Dateien</h4>
            <div class="detail-files-table-wrapper-expanded" style="max-height:calc(100vh - 400px); overflow-y:auto;">
              <DataTable v-if="selectedDocument.files && selectedDocument.files.length > 0" 
                :value="selectedDocument.files" 
                stripedRows
                size="small"
                selectionMode="single"
                v-model:selection="selectedFile"
                @rowSelect="onFileSelect"
                @rowUnselect="onFileUnselect">
                <Column field="type" header="Type" style="width:5rem;">
                    <template #body="{ data }">
                      <span v-if="isMobileView" class="doc-type-icon" :title="getDocTypeLabel(data.type) || data.type">
                        <i :class="getDocTypeIcon(data.type)"></i>
                      </span>
                      <span v-else>{{ getDocTypeLabel(data.type) || data.type }}</span>
                    </template>
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

        <!-- Viewer Panel (rechte Seite) -->
        <div style="flex:2; display:flex; flex-direction:column; overflow:hidden;">
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

      <!-- Großer Fileview - Alte Version (wird durch obiges ersetzt) -->
      <div v-if="false && expandDetailPanel && selectedDocument && selectedFile" style="flex:1; border:1px solid #e0e0e0; border-radius:4px; background:#f9f9f9; display:flex; flex-direction:column; overflow:hidden; position:relative;">
        <!-- Close-Button (oben rechts) -->

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
    <EditDialog v-model="showEditDialog" :document="selectedDocument" @updated="onDocumentUpdated" />

    <div v-if="isMobileView && showMobileFileViewer && selectedFile" class="mobile-file-viewer-overlay">
      <div class="mobile-file-viewer-header">
        <div class="mobile-file-viewer-title">{{ selectedFile.name }}</div>
        <div style="display:flex; gap:0.35rem; align-items:center;">
          <Button v-if="isImageFile(selectedFile)" icon="pi pi-download" severity="secondary" text @click="downloadImage()" />
          <Button icon="pi pi-times" severity="secondary" text @click="showMobileFileViewer = false" />
        </div>
      </div>

      <div class="mobile-file-viewer-content">
        <div v-if="isPdfFile(selectedFile)" style="width:100%; height:100%; display:flex; flex-direction:column;">
          <embed v-if="selectedFile.data" :src="'data:application/pdf;base64,' + selectedFile.data" type="application/pdf" style="flex:1; width:100%; border:none;" />
          <div v-else class="mobile-file-viewer-placeholder">PDF wird geladen...</div>
        </div>

        <div v-else-if="isImageFile(selectedFile)" style="width:100%; height:100%; display:flex; align-items:center; justify-content:center; overflow:auto; background:#fff;">
          <img
            v-if="selectedFile.data"
            :src="'data:' + selectedFile.mimetype + ';base64,' + selectedFile.data"
            :alt="selectedFile.name"
            style="object-fit: contain; width: 100%; height: 100%; max-height: 100%;"
          />
          <div v-else class="mobile-file-viewer-placeholder">Bild wird geladen...</div>
        </div>

        <div v-else class="mobile-file-viewer-placeholder" style="flex-direction:column; gap:0.5rem; text-align:center;">
          <i class="pi pi-file" style="font-size:2rem; color:#888;"></i>
          <div><strong>{{ selectedFile.name }}</strong></div>
          <div>{{ selectedFile.type || selectedFile.mimetype || '-' }}</div>
          <div v-if="selectedFile.page">Seite {{ selectedFile.page }}</div>
          <div>Vorschau für diesen Dateityp nicht verfügbar.</div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import AutoComplete from 'primevue/autocomplete'
import Button from 'primevue/button'
import Dropdown from 'primevue/dropdown'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Fieldset from 'primevue/fieldset'
import UploadDialog from '../components/UploadDialog.vue'
import EditDialog from '../components/EditDialog.vue'
import Image from 'primevue/image'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'
import { getDocTypeLabel, getDocTypeIcon } from '../constants/docTypes'

const router = useRouter()
const { isLoggedIn } = useAuth()
const { info } = useToast()

const query = ref('')
const selectedTags = ref([])
const suggestedTags = ref([])
const results = ref([])
const showHelpPanel = ref(true) // collapsed=true → zugeklappt beim Start
const showUploadDialog = ref(false)
const showEditDialog = ref(false)
const selectedDocument = ref(null)
const selectedFile = ref(null)
const activeShareToken = ref('')
const showMobileFileViewer = ref(false)
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
const totalDocuments = ref(null)
const sortField = ref(null)
const sortOrder = ref(null)
const privateOnly = ref(false)
const isMobileView = ref(false)

function updateViewportState() {
  isMobileView.value = window.innerWidth <= 767
}

function toTags() {
  return selectedTags.value
    .map((tag) => String(tag || '').trim())
    .filter(Boolean)
}

function goHome() {
  router.push('/')
}

async function loadAppInfo() {
  try {
    const { data } = await api.get('/api/v1/info')
    if (data && typeof data.documentCount === 'number') {
      totalDocuments.value = data.documentCount
    }
  } catch (err) {
    console.error('Fehler beim Laden der App-Info:', err)
  }
}

function getDocumentLink() {
  if (!activeShareToken.value) return null
  const baseUrl = window.location.origin + window.location.pathname
  const params = new URLSearchParams()
  params.set('share', activeShareToken.value)
  return `${baseUrl}?${params.toString()}`
}

async function copyDocumentLink() {
  if (!selectedDocument.value) return

  try {
    const { data } = await api.post(`/api/v1/documents/${selectedDocument.value.id}/shares`)
    if (data?.id) {
      activeShareToken.value = data.id
    }

    const link = typeof data?.link === 'string' && data.link.trim() !== ''
      ? data.link
      : getDocumentLink()

    if (!link) {
      info('Share-Link konnte nicht erzeugt werden')
      return
    }

    await navigator.clipboard.writeText(link)
    info('Link kopiert')
  } catch (err) {
    console.error('Fehler beim Erzeugen des Share-Links:', err)
    info('Fehler beim Erzeugen des Share-Links')
  }
}

function getFileApiUrl(fileName) {
  if (!selectedDocument.value) return null
  const encodedFileName = encodeURIComponent(fileName)
  if (activeShareToken.value) {
    return `/api/v1/shares/${encodeURIComponent(activeShareToken.value)}/files/${encodedFileName}`
  }
  return `/api/v1/documents/${selectedDocument.value.id}/files/${encodedFileName}`
}

onMounted(async () => {
  updateViewportState()
  window.addEventListener('resize', updateViewportState)
  loadAppInfo()
  
  // Check if document ID is in URL parameters
  const params = new URLSearchParams(window.location.search)
  const docId = params.get('id')
  const share = params.get('share')
  
  if (share) {
    try {
      activeShareToken.value = share
      const { data } = await api.get(`/api/v1/shares/${encodeURIComponent(share)}`)
      if (data) {
        selectedDocument.value = data
        selectedFile.value = null
        if (data.files && data.files.length > 0) {
          selectedFile.value = data.files[0]
          await loadFileData(selectedFile.value)
        }
        expandDetailPanel.value = true
        hideSearchResults.value = true
        showDetailPanel.value = false
      }
    } catch (err) {
      console.error('Geteiltes Dokument nicht gefunden:', err)
      info('Geteiltes Dokument nicht gefunden')
    }
  } else if (docId) {
    try {
      const { data } = await api.get(`/api/v1/documents/${docId}`)
      if (data) {
        selectedDocument.value = data
        selectedFile.value = null
        showMobileFileViewer.value = false
        if (!isMobileView.value && data.files && data.files.length > 0) {
          selectedFile.value = data.files[0]
          await loadFileData(selectedFile.value)
        }
        expandDetailPanel.value = !isMobileView.value
        hideSearchResults.value = true
        showDetailPanel.value = isMobileView.value
      }
    } catch (err) {
      console.error('Dokument nicht gefunden:', err)
      info('Dokument nicht gefunden')
    }
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateViewportState)
})

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
  search()
}

function onSort(event) {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  search()
}

function prevPage() {
  currentSkip.value = Math.max(0, currentSkip.value - selectedLimit.value)
  search(false)
}

function nextPage() {
  currentSkip.value = currentSkip.value + selectedLimit.value
  search(false)
}

function togglePrivateAndSearch() {
  privateOnly.value = !privateOnly.value
  search()
}

async function search(resetPage = true) {
  try {
    isSearching.value = true
    activeShareToken.value = ''
    
    // Reset pagination when starting a new search
    if (resetPage) {
      currentSkip.value = 0
    }
    
    // Reset detail panel when searching
    selectedDocument.value = null
    selectedFile.value = null
    showMobileFileViewer.value = false
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
  showMobileFileViewer.value = false
  showDetailPanel.value = true
  hideSearchResults.value = isMobileView.value
  expandDetailPanel.value = false
  
  // Erste Datei automatisch selektieren, falls vorhanden
  if (!isMobileView.value && selectedDocument.value.files && selectedDocument.value.files.length > 0) {
    selectedFile.value = selectedDocument.value.files[0]
  }
}

async function onFileSelect(event) {
  const file = event.data
  selectedFile.value = file

  if (!file) {
    return
  }

  if (!file.data) {
    await loadFileData(file)
  }

  if (isMobileView.value) {
    showMobileFileViewer.value = true
  }
}

function onFileUnselect() {
  selectedFile.value = null
  showMobileFileViewer.value = false
}

watch(selectedFile, (file) => {
  if (!file || file.data) {
    return
  }
  loadFileData(file)
})

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
    let url = getFileApiUrl(file.name)
    if (!url) return
    
    // Request PNG conversion for TIFF files (viewer only)
    if (isTiffFile(file)) {
      url += '?format=png'
    }
    
    const { data } = await api.get(url)
    if (data && data.data) {
      file.data = data.data
      // Update mimetype if it changed (e.g., TIFF converted to PNG)
      if (data.mimetype) {
        file.mimetype = data.mimetype
      }
    }
  } catch (err) {
    info('Fehler beim Laden der Datei')
  }
}

function isTiffFile(file) {
  if (!file) return false
  const tiffMimes = ['image/tiff', 'image/x-tiff', 'image/vnd.tiff']
  const tiffExts = ['.tif', '.tiff']
  
  if (tiffMimes.includes(file.mimetype)) return true
  return tiffExts.some(ext => file.name?.toLowerCase().endsWith(ext))
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
  if (!selectedFile.value) return
  
  // For TIFF files, load the original file without conversion
  const downloadData = async () => {
    try {
      let url = getFileApiUrl(selectedFile.value.name)
      if (!url) return
      // Don't use format=png for downloads - get original file
      const { data } = await api.get(url)
      
      const link = document.createElement('a')
      link.href = 'data:' + data.mimetype + ';base64,' + data.data
      link.download = data.name
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    } catch (err) {
      info('Fehler beim Herunterladen der Datei')
    }
  }
  
  downloadData()
}

async function confirmDeleteDocument() {
  if (!selectedDocument.value) return
  
  if (!confirm(`Wirklich löschen: "${selectedDocument.value.manufacturer} ${selectedDocument.value.model}"?`)) {
    return
  }
  
  try {
    await api.delete(`/api/v1/documents/${selectedDocument.value.id}`)
    info('Dokument gelöscht')
    selectedDocument.value = null
    selectedFile.value = null
    showDetailPanel.value = false
    await search()
  } catch (err) {
    info(`Fehler beim Löschen: ${err?.response?.data?.message || err.message}`)
  }
}

async function onDocumentUpdated() {
  info('Dokument gespeichert')
  await search()
  selectedDocument.value = null
  selectedFile.value = null
  showDetailPanel.value = false
}
</script>

<style scoped>
:deep(.selected-row) {
  background-color: #e3f2fd !important;
}

:deep(.selected-row:hover) {
  background-color: #bbdefb !important;
}

/* Help fieldset - only on mobile */
.help-fieldset {
  display: none;
  margin-bottom: 1rem;
}

.title-section {
  margin-bottom: 1rem;
}

.help-text-desktop {
  display: block;
}

.doc-type-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.4rem;
  height: 1.4rem;
  color: #2f4b6e;
}

.detail-panel,
.detail-panel-expanded,
.detail-panel-expanded-left {
  min-height: 0;
}

.detail-files-section,
.detail-files-section-expanded {
  min-height: auto !important;
  flex: 0 0 auto !important;
  overflow: visible !important;
}

.detail-files-table-wrapper,
.detail-files-table-wrapper-expanded {
  max-height: none !important;
  overflow-y: visible !important;
}

/* On mobile, show fieldset and hide desktop text */
@media (max-width: 767px) {
  .search-controls {
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) auto !important;
    gap: 0.5rem !important;
  }

  .search-actions {
    justify-content: flex-start !important;
    gap: 0.5rem !important;
  }

  .doc-type-icon {
    width: 1.2rem;
    height: 1.2rem;
    font-size: 0.9rem;
  }

  .mobile-file-viewer-overlay {
    position: fixed;
    inset: 0;
    z-index: 10050;
    background: #fff;
    display: flex;
    flex-direction: column;
  }

  .mobile-file-viewer-header {
    height: 3rem;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 0.5rem;
    background: #f8f9fa;
  }

  .mobile-file-viewer-title {
    font-size: 0.85rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    padding-right: 0.5rem;
  }

  .mobile-file-viewer-content {
    flex: 1;
    min-height: 0;
    display: flex;
  }

  .mobile-file-viewer-placeholder {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #666;
    padding: 1rem;
  }

  .detail-panel {
    min-height: 0;
    padding: 0.65rem !important;
    gap: 0.65rem !important;
    overflow-y: auto !important;
  }

  .detail-files-section {
    min-height: auto !important;
    flex: 0 0 auto !important;
    overflow: visible !important;
  }

  .detail-files-table-wrapper {
    max-height: none !important;
    overflow-y: visible !important;
  }

  .results-layout {
    gap: 0 !important;
  }

  .splitter-bar {
    width: 12px !important;
    padding: 0 !important;
    gap: 0.15rem !important;
    border-left: none !important;
    border-right: none !important;
  }

  .splitter-bar :deep(.p-button) {
    padding: 0.05rem !important;
    font-size: 0.75rem !important;
  }

  .title-section {
    margin-bottom: 0 !important;
  }
  
  .help-fieldset {
    display: block !important;
    margin: 0.5rem 0 0.5rem 0 !important;
    padding: 0 !important;
  }
  
  .help-text-desktop {
    display: none !important;
  }
  
  /* Reduce fieldset to minimal box */
  .help-fieldset :deep(.p-fieldset) {
    border: none !important;
    border-bottom: 1px solid #ddd !important;
    background: transparent !important;
    padding: 0 !important;
    margin: 0 !important;
  }
  
  /* Minimize legend/header area */
  .help-fieldset :deep(.p-fieldset-legend) {
    padding: 0 !important;
    margin: 0 !important;
    background: transparent !important;
    border: none !important;
    display: flex !important;
    align-items: center !important;
    height: 1.5rem !important;
  }
  
  /* Minimize text in legend */
  .help-fieldset :deep(.p-fieldset-legend-text) {
    padding: 0 0.3rem !important;
    font-size: 0.8rem !important;
    font-weight: 400 !important;
    line-height: 1 !important;
    margin: 0 !important;
    display: inline !important;
  }
  
  /* Make toggle button very compact */
  .help-fieldset :deep(.p-fieldset-toggle-button) {
    width: auto !important;
    height: auto !important;
    min-width: auto !important;
    padding: 0 0.2rem !important;
    margin: 0 !important;
    background: transparent !important;
    border: none !important;
    display: inline-flex !important;
    align-items: center !important;
    justify-content: center !important;
  }
  
  /* Make icon very small */
  .help-fieldset :deep(.p-icon) {
    font-size: 0.65rem !important;
    width: 0.65rem !important;
    height: 0.65rem !important;
  }
  
  /* Reduce content padding */
  .help-fieldset :deep(.p-fieldset-content) {
    padding: 0.5rem 0 !important;
    margin: 0 !important;
  }
}
</style>
