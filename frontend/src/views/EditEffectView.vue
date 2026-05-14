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

        <!-- Tags and Comment -->
        <div>
          <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Tags (kommagetrennt)</label>
          <Textarea 
            v-model="tagsInput" 
            placeholder="Tags eingeben, durch Komma getrennt"
            rows="2"
          />
        </div>

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
          <label style="font-weight:bold; display:block; margin-bottom:1rem;">Bilder</label>

          <!-- Current Images -->
          <div v-if="form.images && form.images.length > 0" style="margin-bottom:1rem;">
            <h3 style="margin-top:0; font-size:0.9rem; color:#666;">Aktuelle Bilder:</h3>
            <div style="display:grid; grid-template-columns:repeat(auto-fill, minmax(150px, 1fr)); gap:1rem;">
              <div v-for="(img, idx) in form.images" :key="idx" style="position:relative; border:1px solid #ddd; border-radius:4px; overflow:hidden;">
                <img :src="getImageUrl(effect.id, idx)" style="width:100%; height:150px; object-fit:cover;" />
                <Button 
                  icon="pi pi-trash" 
                  severity="danger" 
                  text 
                  @click="removeImage(idx)"
                  style="position:absolute; top:0; right:0;"
                />
              </div>
            </div>
            <Checkbox 
              v-model="replaceImage" 
              label="Neues Bild ersetzt alte Bilder"
              style="margin-top:1rem;"
            />
          </div>

          <!-- Upload New Image -->
          <div>
            <FileUpload 
              name="image"
              @select="onImageSelected"
              accept="image/*"
              :maxFileSize="32000000"
              :auto="false"
              chooseLabel="Bild wählen"
              cancelLabel="Abbrechen"
            />
            <div v-if="newImage" style="margin-top:1rem;">
              <img :src="newImagePreview" style="max-width:200px; max-height:200px; border-radius:4px;" />
              <p style="font-size:0.85rem; color:#666; margin-top:0.5rem;">{{ newImage.name }}</p>
            </div>
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
import FileUpload from 'primevue/fileupload'
import { useToast } from 'primevue/usetoast'
import api from '../services/api'

const router = useRouter()
const route = useRoute()
const toast = useToast()

const effectID = route.params.id
const effect = ref(null)
const loading = ref(true)
const saving = ref(false)
const effectTypes = ref([])
const manufacturerSuggestions = ref([])
const connectorOptions = ref(['HI-A+', 'HI+A-', '3.5mm', 'USB', 'XLR', 'RJ45'])

const replaceImage = ref(false)
const newImage = ref(null)
const newImagePreview = ref(null)

const form = ref({
  effectType: '',
  manufacturer: '',
  model: '',
  voltage: '',
  current: '',
  connector: '',
  tags: [],
  comment: '',
  images: []
})

const tagsInput = computed({
  get: () => form.value.tags.join(', '),
  set: (val) => {
    form.value.tags = val
      .split(',')
      .map(tag => tag.trim())
      .filter(tag => tag.length > 0)
  }
})

onMounted(async () => {
  await loadEffect()
  await loadEffectTypes()
})

const loadEffect = async () => {
  try {
    const response = await api.get(`/api/v1/effects/${effectID}`)
    effect.value = response.data
    form.value = {
      effectType: response.data.effectType,
      manufacturer: response.data.manufacturer,
      model: response.data.model,
      voltage: response.data.voltage,
      current: response.data.current,
      connector: response.data.connector,
      tags: response.data.tags || [],
      comment: response.data.comment,
      images: response.data.images || []
    }
  } catch (error) {
    console.error('Failed to load effect:', error)
    toast.add({ severity: 'error', summary: 'Fehler', detail: 'Effekt konnte nicht geladen werden' })
    setTimeout(() => goBack(), 2000)
  } finally {
    loading.value = false
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
  const file = event.files[0]
  if (file) {
    newImage.value = file
    const reader = new FileReader()
    reader.onload = (e) => {
      newImagePreview.value = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

const removeImage = (index) => {
  form.value.images.splice(index, 1)
}

const getImageUrl = (effectId, index) => {
  const base = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
  return `${base}api/v1/effects/${effectId}/image`
}

const saveEffect = async () => {
  // Validate required fields
  if (!form.value.effectType || !form.value.manufacturer || !form.value.model || !form.value.connector) {
    toast.add({ severity: 'error', summary: 'Validierungsfehler', detail: 'Bitte füllen Sie alle erforderlichen Felder aus' })
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
    formData.append('tags', form.value.tags.join(','))
    formData.append('comment', form.value.comment)
    formData.append('replaceImage', replaceImage.value)

    if (newImage.value) {
      formData.append('image', newImage.value)
    }

    await api.patch(`/api/v1/effects/${effectID}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    toast.add({ severity: 'success', summary: 'Erfolg', detail: 'Effekt wurde aktualisiert' })
    setTimeout(() => goBack(), 1500)
  } catch (error) {
    console.error('Failed to save effect:', error)
    toast.add({ severity: 'error', summary: 'Fehler', detail: 'Effekt konnte nicht aktualisiert werden' })
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
