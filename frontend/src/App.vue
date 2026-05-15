<template>
  <div class="shell">
    <header class="card" style="margin-bottom: 1rem">
      <div style="display:flex; align-items:center; justify-content:space-between; gap: 1rem; flex-wrap: wrap;">
        <div>
          <h1 style="margin-bottom:0.2rem">Schematic2</h1>
        </div>
        <nav style="display:flex; align-items:center; gap:0.6rem;">
          <Avatar
            v-tooltip.bottom="'Startseite'"
            icon="pi pi-home"
            shape="circle"
            class="home-avatar"
            @click="router.push('/')"
            style="cursor: pointer;"
            aria-label="Zur Startseite"
          />
          <Avatar
            v-tooltip.bottom="'Suche'"
            icon="pi pi-search"
            shape="circle"
            class="nav-avatar"
            @click="router.push('/search')"
            style="cursor: pointer;"
            aria-label="Suche"
          />
          <Avatar
            v-tooltip.bottom="'Effektdatenbank'"
            icon="pi pi-star"
            shape="circle"
            class="nav-avatar"
            @click="router.push('/effects')"
            style="cursor: pointer;"
            aria-label="Effektdatenbank"
          />
          <Avatar v-if="!isLoggedIn"
            v-tooltip.bottom="'Login'"
            icon="pi pi-user"
            shape="circle"
            class="nav-avatar"
            @click="router.push('/login')"
            style="cursor: pointer;"
            aria-label="Login"
          />
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
import Avatar from 'primevue/avatar'
import Tooltip from 'primevue/tooltip'
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

// Directive for tooltips
const vTooltip = Tooltip

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

<style scoped>
.home-avatar {
  background-color: #999;
  color: #fff;
  width: 2.4rem;
  height: 2.4rem;
  font-size: 1rem;
  flex-shrink: 0;
  transition: opacity 0.2s;
}

.nav-avatar {
  background-color: #999;
  color: #fff;
  width: 2.4rem;
  height: 2.4rem;
  font-size: 1rem;
  flex-shrink: 0;
  transition: opacity 0.2s;
}

.home-avatar:hover,
.nav-avatar:hover {
  opacity: 0.85;
}
</style>
