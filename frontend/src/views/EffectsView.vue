<template>
  <section class="card">
    <h2>Effektdatenbank</h2>
    <p class="muted">
      Diese Seite ist als eigener Bereich vorbereitet. Die API-Anbindung folgt in einem nächsten Schritt.
    </p>

    <div style="display:grid; gap:0.8rem; margin: 1rem 0;">
      <InputText v-model="query" placeholder="Effektname oder Typ suchen" />
      <small class="muted">Aktuell Demo-Daten. Später werden hier echte Datensätze geladen.</small>
    </div>

    <DataTable :value="filteredEffects" stripedRows>
      <Column field="name" header="Name" />
      <Column field="category" header="Kategorie" />
      <Column field="technology" header="Technologie" />
      <Column field="notes" header="Hinweis" />
    </DataTable>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import InputText from 'primevue/inputtext'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'

const query = ref('')

const effects = [
  {
    name: 'Chorus CE-2 Style',
    category: 'Modulation',
    technology: 'Analog BBD',
    notes: 'Klassischer Gitarrenchorus als Referenzdatensatz.',
  },
  {
    name: 'Tape Echo Preamp',
    category: 'Delay',
    technology: 'Transistor',
    notes: 'Vorbereitung fuer spaetere Echomodul-Varianten.',
  },
  {
    name: 'Opto Compressor',
    category: 'Dynamics',
    technology: 'Optocoupler',
    notes: 'Basisstruktur fuer Kompressor-Schaltungen.',
  },
]

const filteredEffects = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) {
    return effects
  }

  return effects.filter((effect) => {
    return (
      effect.name.toLowerCase().includes(q) ||
      effect.category.toLowerCase().includes(q) ||
      effect.technology.toLowerCase().includes(q)
    )
  })
})
</script>
