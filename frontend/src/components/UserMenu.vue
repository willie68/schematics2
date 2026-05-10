<template>
  <div class="user-menu-wrapper">
    <Avatar
      icon="pi pi-user"
      shape="circle"
      class="user-avatar"
      @click="toggleMenu"
      aria-haspopup="true"
      aria-controls="user-overlay-menu"
    />

    <Menu ref="menu" id="user-overlay-menu" :model="items" :popup="true" />

    <Dialog
      v-model:visible="infoVisible"
      header="Über Schematic2"
      :modal="true"
      :closable="true"
      style="width: 420px"
    >
      <p><strong>Schematic2</strong> ist der Nachfolger von WilliesSchematicsWorld.</p>
      <p>Es ermöglicht das Indexieren, Suchen und Verwalten von Schaltplan-Dokumenten und Effektbeschreibungen.</p>
      <p style="margin-top:1rem; font-size:0.85rem; color:#888">App Version: {{ APP_VERSION }}, Backend: {{ info.version }}</p>
      <p>Status: {{ info.status }}</p>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Avatar from 'primevue/avatar'
import Menu from 'primevue/menu'
import Dialog from 'primevue/dialog'
import { useAuth } from '../composables/useAuth'
import { APP_VERSION } from '../config'
import api from '../services/api'

const router = useRouter()
const { logout } = useAuth()

const menu = ref(null)
const infoVisible = ref(false)
const info = ref({
  version: 'Loading...',
  status: 'Loading...',
})

const items = [
  {
    label: 'Info',
    icon: 'pi pi-info-circle',
    command: () => { infoVisible.value = true },
  },
  {
    label: 'Logout',
    icon: 'pi pi-sign-out',
    command: () => {
      logout()
      router.push('/')
    },
  },
]

function toggleMenu(event) {
  menu.value.toggle(event)
}

async function fetchBackendInfo() {
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

watch(infoVisible, (newVal) => {
  if (newVal) {
    fetchBackendInfo()
  }
})
</script>

<style scoped>
.user-avatar {
  cursor: pointer;
  background-color: var(--primary-color, #3b82f6);
  color: #fff;
  width: 2.4rem;
  height: 2.4rem;
  font-size: 1rem;
  flex-shrink: 0;
  transition: opacity 0.2s;
}
.user-avatar:hover {
  opacity: 0.85;
}
</style>
