<template>
  <Dialog v-model:visible="visible" header="App Info" :modal="true" style="width: 500px">
    <div style="display:grid; gap:1rem;">
      <div style="display:flex; gap:1rem; align-items:center;">
        <div style="font-weight:bold">Version:</div>
        <div>{{ info.version }}</div>
      </div>
      <div style="display:flex; gap:1rem; align-items:center;">
        <div style="font-weight:bold">Status:</div>
        <div style="display:flex; align-items:center; gap:0.3rem;">
          <span
            style="
              width: 10px;
              height: 10px;
              border-radius: 50%;
              background-color: #22c55e;
            "
          ></span>
          {{ info.status }}
        </div>
      </div>

      <div style="padding-top:1rem; border-top:1px solid #e5e7eb; text-align:right;">
        <Button label="Schließen" @click="close" />
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import api from '../services/api'

const props = defineProps({
  modelValue: Boolean,
})

const emit = defineEmits(['update:modelValue'])

const visible = ref(props.modelValue)
const info = ref({
  version: 'Loading...',
  status: 'Loading...',
})

watch(
  () => props.modelValue,
  (newVal) => {
    visible.value = newVal
    if (newVal) {
      fetchInfo()
    }
  }
)

watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

async function fetchInfo() {
  try {
    const { data } = await api.get('/api/v1/info')
    info.value = {
      version: data.version || 'Unknown',
      status: data.status || 'Unknown',
    }
  } catch (err) {
    info.value = {
      version: 'Error',
      status: 'Error',
    }
  }
}

function close() {
  visible.value = false
}
</script>
