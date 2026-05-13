<template>
  <div class="shell">
    <header class="card" style="margin-bottom: 1rem">
      <div style="display:flex; align-items:center; justify-content:space-between; gap: 1rem; flex-wrap: wrap;">
        <div>
          <h1 style="margin-bottom:0.2rem">Schematic2</h1>
        </div>
        <nav style="display:flex; align-items:center; gap:0.6rem;">
          <RouterLink to="/">Start</RouterLink>
          <RouterLink to="/search">Suche</RouterLink>
          <RouterLink to="/effects">Effektdatenbank</RouterLink>
          <RouterLink v-if="!isLoggedIn" to="/login">Login</RouterLink>
          <UserMenu v-if="isLoggedIn" />
        </nav>
      </div>
    </header>

    <RouterView />
    <Toast />
    <CookieBanner />
    <AppFooter />
  </div>
</template>

<script setup>
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { onMounted } from 'vue'
import UserMenu from './components/UserMenu.vue'
import Toast from './components/Toast.vue'
import CookieBanner from './components/CookieBanner.vue'
import AppFooter from './components/AppFooter.vue'
import { useAuth } from './composables/useAuth'
import { useToast } from './composables/useToast'
import { setApiErrorHandler } from './services/api'

const router = useRouter()
const { isLoggedIn, logout } = useAuth()
const { error: showError } = useToast()

onMounted(() => {
  // Register global error handler for unauthorized responses
  setApiErrorHandler({
    onUnauthorized: () => {
      showError('Sitzung abgelaufen. Bitte melden Sie sich erneut an.')
      logout()
      router.push('/login')
    },
  })
})
</script>
