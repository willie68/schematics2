<template>
  <div v-if="!cookiesAccepted" class="cookie-banner">
    <div class="cookie-content">
      <div class="cookie-message">
        <p style="margin:0;">
          <strong>Datenschutz</strong>: Diese Anwendung verwendet LocalStorage zur Speicherung Ihres 
          Authentifizierungs-Tokens. Dies ist technisch erforderlich, um Sie angemeldet zu halten.
        </p>
      </div>
      <div class="cookie-actions">
        <Button 
          label="Verstanden" 
          icon="pi pi-check" 
          severity="success" 
          @click="acceptCookies" 
          size="small"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Button from 'primevue/button'

const COOKIE_CONSENT_KEY = 'schematic2_cookie_consent'
const cookiesAccepted = ref(true) // Standard: akzeptiert

onMounted(() => {
  // Prüfe, ob Benutzer bereits zugestimmt hat
  const consent = localStorage.getItem(COOKIE_CONSENT_KEY)
  if (consent !== 'true') {
    cookiesAccepted.value = false
  }
})

function acceptCookies() {
  localStorage.setItem(COOKIE_CONSENT_KEY, 'true')
  cookiesAccepted.value = true
}
</script>

<style scoped>
.cookie-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(33, 37, 41, 0.95);
  color: white;
  padding: 1rem;
  z-index: 9999;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.3);
}

.cookie-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.cookie-message {
  flex: 1;
  font-size: 0.95rem;
  line-height: 1.5;
}

.cookie-message p {
  margin: 0;
}

.cookie-actions {
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .cookie-content {
    flex-direction: column;
    align-items: stretch;
    gap: 1rem;
  }

  .cookie-actions {
    display: flex;
    gap: 0.5rem;
  }

  .cookie-banner {
    padding: 1rem 0.75rem;
  }
}
</style>
