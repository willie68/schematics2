<template>
  <section class="card">
    <div style="display:flex; align-items:center; gap:1rem; margin-bottom:1rem;">
      <Button 
        icon="pi pi-arrow-left" 
        severity="secondary" 
        @click="goBack()"
        v-tooltip.bottom="'Zurück'"
      />
      <h2 style="margin:0;">Effekt bearbeiten</h2>
    </div>

    <div v-if="loading" style="text-align:center; padding:2rem;">
      <i class="pi pi-spin pi-spinner" style="font-size:3rem;"></i>
    </div>

    <div v-else-if="loadError" style="text-align:center; padding:2rem;">
      <p style="color:red; font-size:1.1rem; margin-bottom:1rem;">⚠️ {{ loadError }}</p>
      <p style="color:#666; margin-bottom:1rem;">Effekt-ID: {{ effectID }}</p>
      <Button label="Zurück" @click="goBack()" />
    </div>

    <div v-else-if="effect">
      <form @submit.prevent="saveEffect" style="display:grid; gap:1.5rem;">
        <!-- Basic Info Grid -->
        <div style="display:grid; grid-template-columns:1fr 1fr; gap:1rem;">
          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Typ *</label>
            <Dropdown 
              v-model="form.effectType" 
              :options="effectTypes"
              optionLabel="display"
              optionValue="type"
              placeholder="Typ wählen"
              :editable="false"
              style="width:100%;"
            />
          </div>

          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Hersteller *</label>
            <AutoComplete 
              v-model="form.manufacturer" 
              :suggestions="manufacturerSuggestions"
              @complete="searchManufacturers"
              placeholder="Hersteller eingeben"
            />
          </div>

          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Modell *</label>
            <InputText 
              v-model="form.model" 
              placeholder="Modell eingeben"
            />
          </div>

          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Anschluss *</label>
            <Dropdown 
              v-model="form.connector" 
              :options="connectorOptions"
              placeholder="Anschluss wählen"
              editable
              style="width:100%;"
            />
          </div>

          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Spannung</label>
            <InputText 
              v-model="form.voltage" 
              placeholder="z.B. 230V"
            />
          </div>

          <div>
            <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Strom</label>
            <InputText 
              v-model="form.current" 
              placeholder="z.B. 1A"
            />
          </div>
        </div>

        <!-- Comment -->
        <div>
          <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Kommentar</label>
          <Textarea 
            v-model="form.comment" 
            placeholder="Kommentar eingeben"
            rows="3"
          />
        </div>

        <!-- Image Section -->
        <div style="border:1px solid #ddd; border-radius:8px; padding:1rem;">
          <label style="font-weight:bold; display:block; margin-bottom:1rem;">Bild</label>

          <!-- Current Image Display -->
          <div style="margin-bottom:1.5rem;">
            <div v-if="form.image" style="text-align:center;">
              <img :src="getImageUrl(effect.id)" style="max-width:200px; max-height:200px; border-radius:4px; object-fit:cover;" />
              <p style="font-size:0.85rem; color:#666; margin-top:0.5rem;">{{ form.image.name }}</p>
            </div>
            <div v-else style="padding:1rem; background:#f9f9f9; border-radius:4px; text-align:center; color:#999;">
              Kein Bild vorhanden
            </div>
          </div>

          <!-- Upload New Image -->
          <div>
            <input 
              ref="fileInput"
              type="file" 
              accept="image/*"
              @change="onImageSelected"
              style="width:100%;"
            />
            <small style="color:#999; display:block; margin-top:0.25rem;">
              {{ newImageFileName || 'keine Datei gewählt' }}
            </small>
          </div>
        </div>

        <!-- Submit Buttons -->
        <div style="display:flex; gap:1rem; justify-content:flex-end;">
          <Button 
            label="Abbrechen" 
            severity="secondary" 
            @click="goBack()"
            :loading="saving"
          />
          <Button 
            label="Speichern" 
            severity="success" 
            icon="pi pi-check"
            @click="saveEffect"
            :loading="saving"
          />
        </div>
      </form>
    </div>

    <div v-else style="text-align:center; padding:2rem;">
      <p style="color:#999;">Effekt nicht gefunden</p>
      <Button label="Zurück" @click="goBack()" />
    </div>
  </section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Dropdown from 'primevue/dropdown'
import Checkbox from 'primevue/checkbox'
import AutoComplete from 'primevue/autocomplete'
import api from '../services/api'
import { useToast } from '../composables/useToast'

const router = useRouter()
const route = useRoute()
const toast = useToast()

const effectID = route.params.id
const effect = ref(null)
const loading = ref(true)
const saving = ref(false)
const loadError = ref(null)
const effectTypes = ref([])
const manufacturerSuggestions = ref([])
const connectorOptions = ref(['HI-A+', 'HI+A-', '3.5mm', 'USB', 'XLR', 'RJ45'])

const newImage = ref(null)
const newImageFileName = ref('')
const fileInput = ref(null)

const form = ref({
  effectType: '',
  manufacturer: '',
  model: '',
  voltage: '',
  current: '',
  connector: '',
  comment: '',
  image: null
})

onMounted(async () => {
  await loadEffect()
  await loadEffectTypes()
})

const loadEffect = async () => {
  try {
    console.log('Loading effect with ID:', effectID)
    const response = await api.get(`/api/v1/effects/${effectID}`)
    console.log('Effect loaded successfully:', response.data)
    effect.value = response.data
    form.value = {
      effectType: response.data.effectType,
      manufacturer: response.data.manufacturer,
      model: response.data.model,
      voltage: response.data.voltage,
      current: response.data.current,
      connector: response.data.connector,
      comment: response.data.comment,
      image: response.data.image || null
    }
    loading.value = false
  } catch (error) {
    console.error('Failed to load effect:', error)
    loadError.value = error.message || 'Effekt konnte nicht geladen werden'
    loading.value = false
    toast.error(loadError.value)
  }
}

const loadEffectTypes = async () => {
  try {
    const response = await api.get('/api/v1/effecttypes')
    if (Array.isArray(response.data)) {
      effectTypes.value = response.data
        .map(type => ({
          type: type?.typeName || '',
          display: (type?.i18n?.de || type?.typeName || '').trim()
        }))
        .filter(t => t.type && t.display)
    }
  } catch (error) {
    console.error('Failed to load effect types:', error)
  }
}

const searchManufacturers = async (event) => {
  if (!event.query || event.query.length < 1) {
    manufacturerSuggestions.value = []
    return
  }

  try {
    const response = await api.get('/api/v1/manufacturers/suggest', {
      params: { q: event.query, limit: 10 }
    })
    manufacturerSuggestions.value = response.data || []
  } catch (error) {
    console.error('Failed to search manufacturers:', error)
    manufacturerSuggestions.value = []
  }
}

const onImageSelected = (event) => {
  const file = event.target?.files?.[0]
  if (file) {
    newImage.value = file
    newImageFileName.value = file.name
  }
}

const removeImage = () => {
  form.value.image = null
}

const getImageUrl = (effectId) => {
  const base = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
  return `${base}api/v1/effects/${effectId}/image`;
}

const saveEffect = async () => {
  // Validate required fields
  if (!form.value.effectType || !form.value.manufacturer || !form.value.model || !form.value.connector) {
    toast.error('Bitte füllen Sie alle erforderlichen Felder aus')
    return
  }

  saving.value = true
  try {
    const formData = new FormData()
    formData.append('effectType', form.value.effectType)
    formData.append('manufacturer', form.value.manufacturer)
    formData.append('model', form.value.model)
    formData.append('voltage', form.value.voltage)
    formData.append('current', form.value.current)
    formData.append('connector', form.value.connector)
    formData.append('comment', form.value.comment)

    // Only append image if a new one was selected
    if (newImage.value) {
      formData.append('image', newImage.value)
    }

    await api.patch(`/api/v1/effects/${effectID}`, formData)

    toast.success('Effekt wurde aktualisiert')
    
    // Reload effect to show updated image
    await loadEffect()
    
    // Reset image selection
    newImage.value = null
    newImageFileName.value = ''
    
    // Go back to effects list
    setTimeout(() => goBack(), 1500)
  } catch (error) {
    console.error('Failed to save effect:', error)
    toast.error('Effekt konnte nicht aktualisiert werden')
  } finally {
    saving.value = false
  }
}

const goBack = () => {
  router.go(-1)
}
</script>

<style scoped>
.card {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
