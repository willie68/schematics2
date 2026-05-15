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
      header="Über Schematics2"
      :modal="true"
      :closable="true"
      style="width: 420px"
    >
      <p><strong>Schematics2</strong></p>
      <p>Es ermöglicht das Indexieren, Suchen und Verwalten von Schaltplan-Dokumenten und Effektbeschreibungen.</p>
      <p style="margin-top:1rem; font-size:0.85rem; color:#888">Versionen: App {{ APP_VERSION }}, Backend {{ info.version }}</p>
      <p>Status: {{ info.status }}</p>
    </Dialog>

    <Dialog
      v-model:visible="accountVisible"
      header="Mein Konto"
      :modal="true"
      :closable="true"
      style="width: 500px"
    >
      <div v-if="currentUser" style="display:grid; gap:1rem;">
        <div>
          <label style="font-weight:bold; color:#666;">E-Mail</label>
          <p>{{ currentUser.email }}</p>
        </div>
        <div>
          <label style="font-weight:bold; color:#666;">Name</label>
          <p>{{ currentUser.firstName }} {{ currentUser.lastName }}</p>
        </div>
        <div style="border-top:1px solid #e0e0e0; padding-top:1rem;">
          <label style="font-weight:bold; color:#666;">Adresse</label>
          <p v-if="currentUser.address">
            {{ currentUser.address.street }}<br />
            {{ currentUser.address.zipCode }} {{ currentUser.address.city }}
          </p>
          <p v-else style="color:#999;">Keine Adresse gespeichert</p>
        </div>
        <div style="border-top:1px solid #e0e0e0; padding-top:1rem; font-size:0.85rem; color:#888;">
          <p>Erstellt: {{ formatDate(currentUser.created) }}</p>
          <p>Zuletzt aktualisiert: {{ formatDate(currentUser.updated) }}</p>
        </div>
      </div>
      <div v-else style="text-align:center; padding:2rem; color:#999;">
        Daten werden geladen...
      </div>
    </Dialog>

    <Dialog
      v-model:visible="changePasswordVisible"
      header="Passwort ändern"
      :modal="true"
      :closable="true"
      style="width: 450px"
    >
      <div style="display:grid; gap:1rem;">
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Aktuelles Passwort</label>
          <Password v-model="passwordForm.oldPassword" :toggleMask="true" style="width:100%;" />
        </div>
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Neues Passwort</label>
          <Password v-model="passwordForm.newPassword" :toggleMask="true" style="width:100%;" />
          <small style="color:#999; display:block; margin-top:0.25rem;">Mindestens 8 Zeichen</small>
        </div>
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Passwort wiederholen</label>
          <Password v-model="passwordForm.confirmPassword" :toggleMask="true" style="width:100%;" />
        </div>
        <div style="color:#e74c3c; font-size:0.9em;" v-if="passwordError">{{ passwordError }}</div>
        <div style="display:flex; gap:0.5rem; justify-content:flex-end; margin-top:1rem;">
          <Button label="Abbrechen" severity="secondary" @click="changePasswordVisible = false; resetPasswordForm()" />
          <Button label="Ändern" icon="pi pi-check" :loading="passwordChanging" @click="submitPasswordChange" />
        </div>
      </div>
    </Dialog>
  </div>
</template> 
<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Avatar from 'primevue/avatar'
import Menu from 'primevue/menu'
import Dialog from 'primevue/dialog'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'
import { APP_VERSION } from '../config'
import api from '../services/api'

const router = useRouter()
const { logout } = useAuth()
const { success, error: showError } = useToast()

const menu = ref(null)
const infoVisible = ref(false)
const accountVisible = ref(false)
const changePasswordVisible = ref(false)
const info = ref({
  version: 'Loading...',
  status: 'Loading...',
})
const currentUser = ref(null)
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const passwordError = ref('')
const passwordChanging = ref(false)

const items = [
  {
    label: 'Mein Konto',
    icon: 'pi pi-user',
    command: () => {
      fetchCurrentUser()
      accountVisible.value = true
    },
  },
  {
    label: 'Passwort ändern',
    icon: 'pi pi-key',
    command: () => {
      resetPasswordForm()
      changePasswordVisible.value = true
    },
  },
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

async function fetchCurrentUser() {
  try {
    const { data } = await api.get('/api/v1/users/me')
    currentUser.value = data
  } catch (err) {
    currentUser.value = null
  }
}

function formatDate(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('de-DE')
}

function resetPasswordForm() {
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  passwordError.value = ''
}

function validatePasswordForm() {
  passwordError.value = ''
  
  if (!passwordForm.value.oldPassword) {
    passwordError.value = 'Aktuelles Passwort erforderlich'
    return false
  }
  if (!passwordForm.value.newPassword) {
    passwordError.value = 'Neues Passwort erforderlich'
    return false
  }
  if (passwordForm.value.newPassword.length < 8) {
    passwordError.value = 'Neues Passwort muss mindestens 8 Zeichen lang sein'
    return false
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordError.value = 'Passwörter stimmen nicht überein'
    return false
  }
  
  return true
}

async function submitPasswordChange() {
  if (!validatePasswordForm()) {
    return
  }
  
  passwordChanging.value = true
  try {
    await api.post('/api/v1/users/change-password', {
      oldPassword: passwordForm.value.oldPassword,
      newPassword: passwordForm.value.newPassword,
    })
    
    success('Passwort erfolgreich geändert')
    changePasswordVisible.value = false
    resetPasswordForm()
  } catch (err) {
    passwordError.value = err.response?.data?.message || 'Fehler beim Ändern des Passworts'
  } finally {
    passwordChanging.value = false
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

.user-menu-wrapper {
  display: flex;
  align-items: center;
}
</style>
