<template>
  <div class="register-container">
    <Card class="register-card">
      <template #header>
        <div class="register-header">
          <h2>Benutzerkonto erstellen</h2>
          <p class="register-subtitle">Registrieren Sie sich für ein neues Konto</p>
        </div>
      </template>

      <div v-if="successMessage" class="success-message">
        <i class="pi pi-check-circle"></i>
        {{ successMessage }}
      </div>

      <div v-if="errorMessage" class="error-message">
        <i class="pi pi-exclamation-triangle"></i>
        {{ errorMessage }}
      </div>

      <form @submit.prevent="submitRegister" v-if="!registered">
        <!-- Personal Information -->
        <div class="form-section">
          <h3>Persönliche Informationen</h3>

          <div class="form-group">
            <label for="firstName">Vorname *</label>
            <InputText
              id="firstName"
              v-model="form.firstName"
              placeholder="Vorname"
              @keyup.enter="submitRegister"
            />
          </div>

          <div class="form-group">
            <label for="lastName">Nachname *</label>
            <InputText
              id="lastName"
              v-model="form.lastName"
              placeholder="Nachname"
              @keyup.enter="submitRegister"
            />
          </div>

          <div class="form-group">
            <label for="email">E-Mail (Anmeldename) *</label>
            <InputText
              id="email"
              v-model="form.email"
              type="email"
              placeholder="E-Mail"
              @keyup.enter="submitRegister"
            />
          </div>

          <div class="form-group">
            <label for="password">Passwort (mindestens 8 Zeichen) *</label>
            <Password
              id="password"
              v-model="form.password"
              placeholder="Passwort"
              toggle-mask
              :feedback="false"
              @keyup.enter="submitRegister"
            />
          </div>
        </div>

        <!-- Address Information -->
        <div class="form-section">
          <h3>Adresse</h3>

          <div class="form-group">
            <label for="street">Straße und Hausnummer *</label>
            <InputText
              id="street"
              v-model="form.street"
              placeholder="Straße und Hausnummer"
              @keyup.enter="submitRegister"
            />
          </div>

          <div class="form-row">
            <div class="form-group form-col">
              <label for="zipCode">Postleitzahl *</label>
              <InputText
                id="zipCode"
                v-model="form.zipCode"
                placeholder="PLZ"
                @keyup.enter="submitRegister"
              />
            </div>

            <div class="form-group form-col">
              <label for="city">Stadt *</label>
              <InputText
                id="city"
                v-model="form.city"
                placeholder="Stadt"
                @keyup.enter="submitRegister"
              />
            </div>
          </div>
        </div>

        <!-- Submit Button -->
        <div class="form-actions">
          <Button
            type="submit"
            label="Konto erstellen"
            :loading="isLoading"
            icon="pi pi-check"
          />
          <Button
            type="button"
            label="Zur Anmeldung"
            severity="secondary"
            icon="pi pi-arrow-left"
            @click="goToLogin"
          />
        </div>

        <!-- Login Link -->
        <p class="login-link">
          Haben Sie bereits ein Konto?
          <a href="#" @click.prevent="goToLogin">Hier anmelden</a>
        </p>
      </form>

      <!-- Success Screen -->
      <div v-else class="success-screen">
        <div class="success-icon">
          <i class="pi pi-check-circle"></i>
        </div>
        <h3>Konto erfolgreich erstellt!</h3>
        <p>Ihr Benutzerkonto wurde erfolgreich registriert.</p>
        <p>Sie können sich jetzt mit Ihrer E-Mail und Ihrem Passwort anmelden.</p>
        <Button
          label="Zur Anmeldung"
          icon="pi pi-arrow-right"
          @click="goToLogin"
        />
      </div>
    </Card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'

const router = useRouter()
const { loginWithEmail } = useAuth()

const form = ref({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  street: '',
  zipCode: '',
  city: '',
})

const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const registered = ref(false)

const submitRegister = async () => {
  errorMessage.value = ''
  successMessage.value = ''

  // Basic validation
  if (!form.value.firstName || !form.value.lastName || !form.value.email ||
      !form.value.password || !form.value.street || !form.value.zipCode ||
      !form.value.city) {
    errorMessage.value = 'Bitte füllen Sie alle Felder aus.'
    return
  }

  if (form.value.password.length < 8) {
    errorMessage.value = 'Passwort muss mindestens 8 Zeichen lang sein.'
    return
  }

  isLoading.value = true

  try {
    const apiBase = typeof __API_BASE__ !== 'undefined' ? __API_BASE__ : '/'
    const response = await fetch(`${apiBase}api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        firstName: form.value.firstName,
        lastName: form.value.lastName,
        email: form.value.email,
        password: form.value.password,
        street: form.value.street,
        zipCode: form.value.zipCode,
        city: form.value.city,
      }),
    })

    if (response.status === 201) {
      successMessage.value = 'Registrierung erfolgreich! Redirect wird durchgeführt...'
      registered.value = true
      // Auto-redirect after 2 seconds
      setTimeout(() => goToLogin(), 2000)
    } else if (response.status === 400) {
      const data = await response.json()
      errorMessage.value = data.error || 'Ungültige Eingabedaten.'
    } else if (response.status === 409) {
      errorMessage.value = 'Diese E-Mail-Adresse ist bereits registriert.'
    } else {
      errorMessage.value = 'Registrierung fehlgeschlagen. Bitte versuchen Sie später erneut.'
    }
  } catch (err) {
    console.error('Registration error:', err)
    errorMessage.value = 'Ein Fehler ist aufgetreten. Bitte versuchen Sie später erneut.'
  } finally {
    isLoading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 100%;
  max-width: 600px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  border-radius: 12px;
}

.register-header {
  text-align: center;
  padding: 2rem;
  border-bottom: 1px solid #e9ecef;
}

.register-header h2 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  color: #333;
}

.register-subtitle {
  margin: 0;
  color: #666;
  font-size: 0.95rem;
}

form {
  padding: 2rem;
}

.form-section {
  margin-bottom: 2rem;
}

.form-section h3 {
  margin: 0 0 1rem 0;
  font-size: 1.1rem;
  color: #333;
  font-weight: 500;
  border-bottom: 2px solid #667eea;
  padding-bottom: 0.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-weight: 500;
  font-size: 0.95rem;
  margin-bottom: 0.5rem;
  color: #333;
}

.form-group :deep(.p-inputtext),
.form-group :deep(.p-password) {
  min-height: 42px;
  box-sizing: border-box;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 0.75rem;
  font-size: 1rem;
}

.form-group :deep(.p-inputtext):focus,
.form-group :deep(.p-password) :deep(.p-inputtext):focus {
  border-color: #667eea;
  outline: none;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-col {
  margin-bottom: 0;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.form-actions :deep(.p-button) {
  flex: 1;
  height: 42px;
}

.login-link {
  text-align: center;
  margin-top: 1.5rem;
  color: #666;
  font-size: 0.95rem;
}

.login-link a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.login-link a:hover {
  text-decoration: underline;
}

.success-message,
.error-message {
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.95rem;
}

.success-message {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.success-screen {
  text-align: center;
  padding: 2rem;
}

.success-icon {
  font-size: 4rem;
  color: #28a745;
  margin-bottom: 1rem;
}

.success-screen h3 {
  font-size: 1.5rem;
  margin: 0 0 1rem 0;
  color: #333;
}

.success-screen p {
  color: #666;
  margin: 0 0 1rem 0;
  line-height: 1.5;
}

.success-screen :deep(.p-button) {
  margin-top: 1rem;
}

@media (max-width: 600px) {
  .register-container {
    padding: 1rem;
  }

  .register-card {
    max-width: 100%;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }

  .register-header h2 {
    font-size: 1.5rem;
  }
}
</style>
