<template>
  <section class="card" style="max-width: 520px; margin: 0 auto;">
    <h2>Login</h2>
    <p class="muted">Authentifizierung für Indexierung und Administration.</p>

    <div style="display:grid; gap:0.8rem;">
      <span class="p-float-label">
        <InputText id="user" v-model="username" style="width:100%" />
        <label for="user">Benutzername</label>
      </span>

      <span class="p-float-label">
        <Password id="pass" v-model="password" :feedback="false" toggleMask style="width:100%" />
        <label for="pass">Passwort</label>
      </span>

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

const router = useRouter()
const username = ref('admin')
const password = ref('admin123')
const message = ref('')
const messageType = ref('info')

async function login() {
  message.value = ''
  try {
    const { data } = await api.post('/api/v1/auth/login', {
      username: username.value,
      password: password.value,
    })
    localStorage.setItem('schematic2_token', data.token)
    messageType.value = 'success'
    message.value = 'Login erfolgreich.'
    router.push('/search')
  } catch (err) {
    messageType.value = 'error'
    message.value = err?.response?.data?.error || 'Login fehlgeschlagen.'
  }
}
</script>
