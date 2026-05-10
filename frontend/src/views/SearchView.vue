<template>
  <section class="card">
    <h2>Dokumentsuche</h2>
    <p class="muted">Durchsuche Schaltpläne, Dokumentationen und PDFs über Tags und Volltext.</p>

    <div style="display:grid; gap:0.8rem; margin-bottom:1rem;">
      <InputText v-model="query" placeholder="Suche nach Begriffen" />
      <AutoComplete
        v-model="selectedTags"
        :suggestions="suggestedTags"
        @complete="onTagSuggest"
        placeholder="Tags auswählen"
        multiple
        forceSelection
        style="width:100%"
      />
      <div style="display:flex; gap:0.5rem; flex-wrap:wrap;">
        <Button label="Suchen" icon="pi pi-search" @click="search" />
        <Button v-if="isLoggedIn" label="Upload" icon="pi pi-upload" severity="secondary" @click="showUploadDialog = true" />
      </div>
    </div>

    <DataTable :value="results" stripedRows>
      <Column field="document.id" header="ID" />
      <Column field="document.manufacturer" header="Hersteller" />
      <Column field="document.model" header="Model" />
      <Column field="score" header="Score" />
    </DataTable>

    <UploadDialog v-model="showUploadDialog" @uploaded="search" />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import InputText from 'primevue/inputtext'
import AutoComplete from 'primevue/autocomplete'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import UploadDialog from '../components/UploadDialog.vue'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'

const { isLoggedIn } = useAuth()

const query = ref('')
const selectedTags = ref([])
const suggestedTags = ref([])
const results = ref([])
const showUploadDialog = ref(false)

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

async function search() {
  const params = new URLSearchParams()
  params.set('q', query.value)
  toTags().forEach((tag) => params.append('tag', tag))
  const { data } = await api.get(`/api/v1/documents/search?${params.toString()}`)
  results.value = data.results || []
}
</script>
