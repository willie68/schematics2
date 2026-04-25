<template>
  <section class="card">
    <h2>Dokumentsuche</h2>
    <p class="muted">Durchsuche Schaltpläne, Dokumentationen und PDFs über Tags und Volltext.</p>

    <div style="display:grid; gap:0.8rem; margin-bottom:1rem;">
      <InputText v-model="query" placeholder="Suche nach Begriffen" />
      <InputText v-model="tagInput" placeholder="Tags, z. B. netzteil, smps, reparatur" />
      <div style="display:flex; gap:0.5rem; flex-wrap:wrap;">
        <Button label="Suchen" icon="pi pi-search" @click="search" />
        <Button label="Beispiel indexieren" icon="pi pi-plus" severity="secondary" @click="seed" />
      </div>
    </div>

    <DataTable :value="results" stripedRows>
      <Column field="document.id" header="ID" />
      <Column field="document.title" header="Titel" />
      <Column field="document.path" header="Datei" />
      <Column field="score" header="Score" />
    </DataTable>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import api from '../services/api'

const query = ref('')
const tagInput = ref('')
const results = ref([])

function toTags() {
  return tagInput.value
    .split(',')
    .map((x) => x.trim())
    .filter(Boolean)
}

async function seed() {
  const token = localStorage.getItem('schematic2_token')
  if (!token) {
    alert('Bitte zuerst einloggen.')
    return
  }

  await api.post('/api/v1/documents/index', {
    id: `doc-${Date.now()}`,
    title: 'SMPS Netzteil Schaltplan',
    path: 'docs/netzteil/smps-v2.pdf',
    tags: ['netzteil', 'smps', 'reparatur'],
    text: 'Schaltplan und Service-Dokumentation für Schaltnetzteil Revision 2.',
  })

  await search()
}

async function search() {
  const params = new URLSearchParams()
  params.set('q', query.value)
  toTags().forEach((tag) => params.append('tag', tag))
  const { data } = await api.get(`/api/v1/documents/search?${params.toString()}`)
  results.value = data.results || []
}
</script>
