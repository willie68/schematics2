<template>
  <section class="card">
    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:1rem;">
      <h2 style="margin:0;">Effektdatenbank</h2>
      <Button 
        icon="pi pi-envelope" 
        label="Vorschlag" 
        severity="secondary"
        @click="sendSuggestion()"
        v-tooltip.bottom="'Effekt vorschlagen'"
      />
    </div>

    <div style="display:grid; gap:0.8rem; margin: 1rem 0;">
      <div style="display:flex; gap:1rem; align-items:center;">
        <InputText 
          v-model="query" 
          placeholder="Nach Effekt suchen..." 
          @keyup.enter="search()"
          style="flex:1;"
        />
        <Button icon="pi pi-search" @click="search()" severity="primary" v-tooltip.bottom="'Suchen'" />
        <Button v-if="isLoggedIn" icon="pi pi-plus" severity="success" @click="showUploadDialog = true" v-tooltip.bottom="'Effekt hinzufügen'" />
      </div>
      
      <div style="display:flex; gap:1rem; align-items:center;">
        <label>Einträge pro Seite:</label>
        <Dropdown 
          v-model="limit" 
          :options="[10, 20, 50]"
          @change="search()"
          style="width:100px;"
        />
        <span class="muted" v-if="total > 0">
          {{ total }} insgesamt
        </span>
      </div>
    </div>

    <DataTable 
      :value="effects" 
      stripedRows
      :loading="loading"
      responsiveLayout="scroll"
      @rowClick="showDetailModal"
      @sort="onSort"
      :sortField="sortField"
      :sortOrder="sortOrder"
      style="cursor:pointer;"
    >
      <Column field="effectType" header="Typ" style="width:15%;" sortable>
        <template #body="slotProps">
          {{ getEffectTypeDisplay(slotProps.data.effectType) }}
        </template>
      </Column>
      <Column field="manufacturer" header="Hersteller" style="width:15%;" sortable />
      <Column field="model" header="Modell" style="width:20%;" sortable />
      <Column field="voltage" header="Spannung" style="width:12%;" sortable />
      <Column field="current" header="Strom" style="width:12%;" sortable />
      <Column header="Anschluss" style="width:12%;">
        <template #body="slotProps">
          <div style="display:flex; align-items:center; gap:0.5rem;">
            <img 
              v-if="slotProps.data.connector && isConnectorWithIcon(slotProps.data.connector)"
              :src="getConnectorImageUrl(slotProps.data.connector)"
              :alt="slotProps.data.connector"
              style="height:24px; width:auto;"
              :title="slotProps.data.connector"
            />
            <i v-else-if="slotProps.data.connector" class="pi pi-link" :title="slotProps.data.connector"></i>
            <i v-else class="pi pi-times"></i>
            <span class="muted">{{ slotProps.data.connector }}</span>
          </div>
        </template>
      </Column>
      <Column header="Bild" style="width:14%; text-align:center;">
        <template #body="slotProps">
          <img 
            v-if="slotProps.data.image"
            :src="getThumbnailUrl(slotProps.data.id)"
            style="max-width:60px; max-height:60px; cursor:pointer;"
            @click="showImageModal(slotProps.data)"
            :alt="slotProps.data.model"
          />
          <span v-else class="muted">-</span>
        </template>
      </Column>
    </DataTable>

    <div style="display:flex; gap:1rem; justify-content:center; margin-top:1rem; align-items:center;" v-if="total > 0">
      <Button 
        icon="pi pi-chevron-left" 
        :disabled="skip === 0"
        @click="previousPage()"
      />
      <span>Seite {{ currentPage }} von {{ totalPages }}</span>
      <Button 
        icon="pi pi-chevron-right" 
        :disabled="skip + limit >= total"
        @click="nextPage()"
      />
    </div>

    <!-- Image Modal -->
    <Dialog v-model:visible="showImage" :header="selectedEffect?.model" modal>
      <img 
        v-if="selectedEffect?.image"
        :src="getImageUrl(selectedEffect.id)"
        style="width:100%; max-height:600px; object-fit:contain;"
        :alt="selectedEffect.model"
      />
    </Dialog>

    <!-- Detail Modal -->
    <Dialog v-model:visible="showDetail" :header="`${selectedEffectDetail?.manufacturer || ''} ${selectedEffectDetail?.model || ''}`" modal style="width:90%; max-width:1000px;">
      <div style="display:grid; grid-template-columns:1fr 1fr; gap:2rem; height:600px;">
        <!-- Left: Details -->
        <div style="overflow-y:auto;">
          <div style="display:flex; flex-direction:column; gap:1.5rem;">
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Typ</label>
              <span>{{ getEffectTypeDisplay(selectedEffectDetail?.effectType) }}</span>
            </div>
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Hersteller</label>
              <span>{{ selectedEffectDetail?.manufacturer }}</span>
            </div>
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Modell</label>
              <span>{{ selectedEffectDetail?.model }}</span>
            </div>
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Spannung</label>
              <span>{{ selectedEffectDetail?.voltage }}</span>
            </div>
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Strom</label>
              <span>{{ selectedEffectDetail?.current }}</span>
            </div>
            <div>
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Anschluss</label>
              <div style="display:flex; align-items:center; gap:1rem;">
                <img 
                  v-if="selectedEffectDetail?.connector && isConnectorWithIcon(selectedEffectDetail.connector)"
                  :src="getConnectorImageUrl(selectedEffectDetail.connector)"
                  :alt="selectedEffectDetail.connector"
                  style="height:80px; width:auto; max-width:200px;"
                />
                <div style="display:flex; flex-direction:column; gap:0.5rem;">
                  <i v-if="!isConnectorWithIcon(selectedEffectDetail?.connector) && selectedEffectDetail?.connector" class="pi pi-link" style="font-size:2rem;"></i>
                  <i v-else-if="!selectedEffectDetail?.connector" class="pi pi-times" style="font-size:2rem;"></i>
                  <span>{{ selectedEffectDetail?.connector }}</span>
                </div>
              </div>
            </div>
            <div v-if="selectedEffectDetail?.tags && selectedEffectDetail.tags.length > 0">
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Tags</label>
              <div style="display:flex; gap:0.5rem; flex-wrap:wrap;">
                <span v-for="tag in selectedEffectDetail.tags" :key="tag" style="background:#e8f5e9; padding:0.25rem 0.75rem; border-radius:4px; font-size:0.875rem;">{{ tag }}</span>
              </div>
            </div>
            <div v-if="selectedEffectDetail?.comment">
              <label style="font-weight:bold; display:block; margin-bottom:0.5rem;">Kommentar</label>
              <span style="display:block; word-break:break-word;">{{ selectedEffectDetail.comment }}</span>
            </div>
          </div>
        </div>

        <!-- Right: Image -->
        <div style="display:flex; align-items:center; justify-content:center; background:#f5f5f5; border-radius:8px;">
          <img 
            v-if="selectedEffectDetail?.image"
            :src="getImageUrl(selectedEffectDetail.id)"
            style="max-width:100%; max-height:100%; object-fit:contain;"
            :alt="selectedEffectDetail.model"
          />
          <span v-else class="muted">Kein Bild vorhanden</span>
        </div>
      </div>

      <template #footer>
        <div style="display:flex; gap:1rem; justify-content:flex-end;">
          <Button 
            label="Schließen" 
            severity="secondary" 
            @click="showDetail = false"
          />
          <Button 
            v-if="isLoggedIn"
            label="Bearbeiten" 
            icon="pi pi-pencil"
            severity="primary"
            @click="editEffect(selectedEffectDetail.id)"
          />
          <Button 
            v-if="isLoggedIn"
            label="Löschen" 
            icon="pi pi-trash" 
            severity="danger" 
            @click="confirmDeleteEffect()"
          />
        </div>
      </template>
    </Dialog>

    <!-- Effect Upload Dialog Component -->
    <EffectUploadDialog 
      :visible="showUploadDialog" 
      @update:visible="showUploadDialog = $event"
      :effect-types="effectTypes"
      @effect-created="onEffectCreated"
    />
  </section>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dropdown from 'primevue/dropdown'
import Dialog from 'primevue/dialog'
import EffectUploadDialog from '../components/EffectUploadDialog.vue'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()

const { isLoggedIn } = useAuth()

const query = ref('')
const effects = ref([])
const loading = ref(false)
const skip = ref(0)
const limit = ref(10)
const total = ref(0)
const sortField = ref('manufacturer')
const sortOrder = ref(1) // 1 for ASC, -1 for DESC

const currentPage = computed(() => Math.floor(skip.value / limit.value) + 1)
const totalPages = computed(() => Math.ceil(total.value / limit.value))

const showImage = ref(false)
const selectedEffect = ref(null)

const showDetail = ref(false)
const selectedEffectDetail = ref(null)

// Upload dialog
const showUploadDialog = ref(false)
const effectTypes = ref([])
const effectTypeMap = ref({}) // TypeName -> German Translation

onMounted(() => {
  search()
  fetchEffectTypes()
})

const fetchEffectTypes = async () => {
  try {
    const response = await api.get('/api/v1/effecttypes')
    if (response?.data && Array.isArray(response.data)) {
      // Create translation map and dropdown options
      const typeMap = {}
      const typeOptions = response.data
        .map(type => {
          if (!type) return null
          const display = (type?.i18n?.de || type?.typeName || '').trim()
          const typeValue = (type?.typeName || '').trim()
          
          if (!typeValue || !display) return null
          
          // Store in map for table lookup
          typeMap[typeValue] = display
          
          return {
            type: typeValue,
            display: display
          }
        })
        .filter(Boolean)
      
      effectTypes.value = typeOptions
      effectTypeMap.value = typeMap
    }
  } catch (error) {
    console.error('Failed to fetch effect types:', error)
    effectTypes.value = []
    effectTypeMap.value = {}
  }
}

const getEffectTypeDisplay = (typeName) => {
  return effectTypeMap.value[typeName] || typeName || '-'
}

const onEffectCreated = () => {
  search()
}

const search = async (resetPage = true) => {
  loading.value = true
  if (resetPage) {
    skip.value = 0
  }
  try {
    const response = await api.get('/api/v1/effects/search', {
      params: {
        q: query.value,
        skip: skip.value,
        limit: limit.value,
        sort: sortField.value,
        order: sortOrder.value === 1 ? 'asc' : 'desc'
      }
    })
    effects.value = response.data.results || []
    total.value = response.data.total || 0
  } catch (error) {
    console.error('Search failed:', error)
    effects.value = []
  } finally {
    loading.value = false
  }
}

const onSort = (event) => {
  sortField.value = event.sortField || 'manufacturer'
  sortOrder.value = event.sortOrder
  search(false)
}

const nextPage = () => {
  skip.value += limit.value
  search(false)
}

const previousPage = () => {
  skip.value = Math.max(0, skip.value - limit.value)
  search(false)
}

const getThumbnailUrl = (effectId) => {
  const base = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
  return `${base}api/v1/effects/${effectId}/image`;
}

const getImageUrl = (effectId) => {
  const base = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
  return `${base}api/v1/effects/${effectId}/image`;
}

const showImageModal = (effect) => {
  selectedEffect.value = effect
  showImage.value = true
}

const showDetailModal = (event) => {
  selectedEffectDetail.value = event.data
  showDetail.value = true
}

const sendSuggestion = () => {
  const subject = encodeURIComponent('Vorschlag: Effekt')
  const body = encodeURIComponent('Hier können Sie einen Effektvorschlag eintragen:\n\nHersteller:\nModell:\nTyp:\nSpannung:\nStrom:\nAnschluss:\nKommentar:\n\n(Optional können Sie auch ein Bild anhängen)\n\nVielen Dank für Ihren Beitrag!')
  window.location.href = `mailto:info@wk-music.de?subject=${subject}&body=${body}`
}

const isConnectorWithIcon = (connector) => {
  if (!connector) return false
  const c = connector.toUpperCase()
  return c === 'HI-A+' || c === 'HI+A-'
}

const getConnectorImageUrl = (connector) => {
  const base = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
  return `${base}api/v1/connectors/${encodeURIComponent(connector)}`
}

const getConnectorIcon = (connector) => {
  if (!connector) return 'pi pi-times'
  
  const c = connector.toUpperCase()
  if (c === 'HI-A+' || c === 'HI+A-') {
    return 'pi pi-sitemap'
  }
  
  return 'pi pi-link'
}

const editEffect = (effectId) => {
  router.push(`/effects/${effectId}/edit`)
  // Schließe das Modal nach kurzer Verzögerung für sanfte Transition
  setTimeout(() => {
    showDetail.value = false
  }, 100)
}

const confirmDeleteEffect = async () => {
  if (!confirm(`Wirklich löschen: "${selectedEffectDetail.value.id} ${selectedEffectDetail.value.manufacturer} ${selectedEffectDetail.value.model}"?`)) {
    return
  }

  try {
    await api.delete(`/api/v1/effects/${selectedEffectDetail.value.id}`)
    toast.success('Effekt gelöscht')
    selectedEffectDetail.value = null
    selectedEffect.value = null
    showDetail.value = false
    await search()
  } catch (err) {
    console.error('Delete error:', err)
    toast.error(`Fehler beim Löschen: ${err?.response?.data?.message || err.message}`)
  }
}

</script>

<style scoped>
.muted {
  color: #999;
  font-size: 0.875rem;
}
</style>
