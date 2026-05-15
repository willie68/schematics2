<template>
  <section class="card" style="max-width: 520px; margin: 0 auto;">
    <h2>Anmeldung</h2>
    <div style="display:grid; gap:1.5rem;">
      <div style="box-sizing:border-box; width:100%;">
        <label for="user" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Benutzername</label>
        <InputText id="user" v-model="username" @keyup.enter="login" style="width:100%; min-height:42px; box-sizing:border-box;" />
      </div>

      <div style="box-sizing:border-box; width:100%;">
        <label for="pass" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Passwort</label>
        <Password id="pass" v-model="password" :feedback="false" toggleMask @keyup.enter="login" style="width:100%; min-height:42px; box-sizing:border-box;" />
      </div>

      <Button label="Einloggen" icon="pi pi-sign-in" @click="login" />
    </div>

    <Message v-if="message" :severity="messageType" style="margin-top:1rem">{{ message }}</Message>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { setLoggedIn } = useAuth()
const username = ref('')
const password = ref('')
const message = ref('')
const messageType = ref('info')

async function login() {
  message.value = ''
  try {
    const { data } = await api.post('/api/v1/auth/login', {
      username: username.value,
      password: password.value,
    })
    localStorage.setItem('schematics2_token', data.token)
    setLoggedIn(true)
    messageType.value = 'success'
    message.value = 'Login erfolgreich.'
    router.push('/')
  } catch (err) {
    messageType.value = 'error'
    message.value = err?.response?.data?.error || 'Login fehlgeschlagen.'
  }
}
</script>

<style scoped>
:deep(.p-inputtext),
:deep(.p-password) {
  width: 100% !important;
  box-sizing: border-box !important;
}

:deep(.p-password .p-password-input) {
  width: 100% !important;
  box-sizing: border-box !important;
}

:deep(.p-password) {
  display: flex !important;
  width: 100% !important;
}
</style>

