<template>
  <Dialog 
    v-model:visible="isVisible" 
    header="Effekt hinzufügen" 
    modal 
    :closable="true"
    @hide="onDialogHide"
    style="width:90%; max-width:600px;"
  >
    <div style="display:grid; gap:1.5rem;">
      <div>
        <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Typ *</label>
        <Dropdown 
          v-model="form.effectType" 
          :options="effectTypes"
          optionLabel="display"
          optionValue="type"
          placeholder="Wählen Sie einen Typ"
          style="width:100%;"
        />
      </div>

      <div>
        <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Hersteller *</label>
        <AutoComplete 
          v-model="form.manufacturer" 
          :suggestions="manufacturerSuggestions"
          @complete="onManufacturerSearch"
          :force-selection="false"
          :min-length="1"
          placeholder="Hersteller eingeben"
          style="width:100%;"
        />
      </div>

      <div>
        <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Modell *</label>
        <InputText v-model="form.model" placeholder="Modellname" style="width:100%;" />
      </div>

      <div style="display:grid; grid-template-columns:1fr 1fr; gap:1rem;">
        <div>
          <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Spannung</label>
          <InputText v-model="form.voltage" placeholder="z.B. 9V" style="width:100%;" />
        </div>
        <div>
          <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Strom</label>
          <InputText v-model="form.current" placeholder="z.B. 100mA" style="width:100%;" />
        </div>
      </div>

      <div>
        <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Anschluss *</label>
        <Dropdown 
          v-model="form.connector" 
          :options="connectorOptions"
          placeholder="Wählen Sie einen Anschluss"
          style="width:100%;"
        />
      </div>

      <div>
        <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Bild</label>
        <input 
          ref="fileInput"
          type="file" 
          accept="image/*"
          @change="onFileSelected"
          style="width:100%;"
        />
        <small style="color:#999; display:block; margin-top:0.25rem;" v-if="form.imageFileName">
          Datei: {{ form.imageFileName }}
        </small>
      </div>

      <div style="color:#e74c3c; font-size:0.9em;" v-if="errorMessage">{{ errorMessage }}</div>

      <div style="display:flex; gap:0.5rem; justify-content:flex-end;">
        <Button label="Abbrechen" severity="secondary" @click="onCancel" v-tooltip.bottom="'Abbrechen'" />
        <Button label="Hochladen" icon="pi pi-check" severity="success" :loading="uploading" @click="submitEffect" v-tooltip.bottom="'Hochladen'" />
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dropdown from 'primevue/dropdown'
import AutoComplete from 'primevue/autocomplete'
import { useToast } from '../composables/useToast'
import api from '../services/api'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  effectTypes: {
    type: Array,
    required: false,
    default: () => []
  }
})

const emit = defineEmits(['update:visible', 'effect-created'])

const { success, error: showError } = useToast()

const fileInput = ref(null)
const uploading = ref(false)
const errorMessage = ref('')
const manufacturerSuggestions = ref([])

const connectorOptions = [
  'HI-A+',
  'HI+A-',
  'proprietär',
  'DC-Buchse',
  'Miniklinke',
  'USB',
  'MikroUSB',
  'USB-C',
  'Netz'
]

const form = ref({
  effectType: null,
  manufacturer: '',
  model: '',
  voltage: '',
  current: '',
  connector: null,
  imageFile: null,
  imageFileName: ''
})

const isVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const onManufacturerSearch = async (event) => {
  if (!event?.query || event.query.length === 0) {
    manufacturerSuggestions.value = []
    return
  }
  try {
    const response = await api.get('/api/v1/manufacturers/suggest', {
      params: { q: event.query }
    })
    // API returns { manufacturers: [...] }
    if (response.data?.manufacturers && Array.isArray(response.data.manufacturers)) {
      manufacturerSuggestions.value = response.data.manufacturers
    } else if (Array.isArray(response.data)) {
      // Fallback: if response.data is directly an array
      manufacturerSuggestions.value = response.data
    } else {
      manufacturerSuggestions.value = []
    }
  } catch (error) {
    console.error('Failed to fetch manufacturers:', error)
    manufacturerSuggestions.value = []
  }
}

const onFileSelected = (event) => {
  const file = event.target?.files?.[0]
  if (file) {
    form.value.imageFile = file
    form.value.imageFileName = file.name
  }
}

const onCancel = () => {
  resetForm()
  isVisible.value = false
}

const onDialogHide = () => {
  resetForm()
}

const resetForm = () => {
  form.value = {
    effectType: null,
    manufacturer: '',
    model: '',
    voltage: '',
    current: '',
    connector: null,
    imageFile: null,
    imageFileName: ''
  }
  errorMessage.value = ''
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const submitEffect = async () => {
  errorMessage.value = ''

  // Robust null/undefined checks
  if (!form.value?.effectType || form.value.effectType === null) {
    errorMessage.value = 'Typ ist erforderlich'
    return
  }
  if (!form.value?.manufacturer?.trim?.()) {
    errorMessage.value = 'Hersteller ist erforderlich'
    return
  }
  if (!form.value?.model?.trim?.()) {
    errorMessage.value = 'Modell ist erforderlich'
    return
  }
  if (!form.value?.connector || form.value.connector === null) {
    errorMessage.value = 'Anschluss ist erforderlich'
    return
  }

  uploading.value = true
  try {
    // Create FormData for multipart upload
    const formData = new FormData()
    formData.append('effectType', form.value.effectType)
    formData.append('manufacturer', form.value.manufacturer)
    formData.append('model', form.value.model)
    formData.append('voltage', form.value.voltage)
    formData.append('current', form.value.current)
    formData.append('connector', form.value.connector)
    
    if (form.value.imageFile) {
      formData.append('image', form.value.imageFile)
    }

    await api.post('/api/v1/effects', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    success('Effekt erfolgreich hinzugefügt')
    emit('effect-created')
    resetForm()
    isVisible.value = false
  } catch (error) {
    errorMessage.value = error.response?.data?.error || 'Fehler beim Hochladen'
    showError(errorMessage.value)
  } finally {
    uploading.value = false
  }
}
</script>
