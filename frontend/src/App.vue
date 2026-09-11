<template>
  <div class="shell">
    <header class="card app-header" :class="{ 'hide-on-secondary-mobile': hideHeaderOnMobile }" style="margin-bottom: 1rem">
      <div class="app-header-content" style="display:flex; align-items:center; justify-content:space-between; gap: 1rem; flex-wrap: wrap;">
        <div class="brand">
          <img class="brand-logo" :src="logoUrl" width="40" height="40" alt="" />
          <h1 style="margin-bottom:0.2rem">Schematics2</h1>
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
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, computed } from 'vue'
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
const route = useRoute()
const { isLoggedIn, logout } = useAuth()
const { error: showError } = useToast()
const logoUrl = `${import.meta.env.BASE_URL}logo.svg`

// Hide header on search and effects pages on mobile devices
const hideHeaderOnMobile = computed(() => {
  const currentPath = route.path || ''
  return currentPath === '/search' || currentPath.startsWith('/effects')
})

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
.app-header {
  margin-bottom: 1rem;
}

.app-header-content {
  gap: 1rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.brand-logo {
  width: 2.4rem;
  height: 2.4rem;
  flex-shrink: 0;
  border-radius: 0.55rem;
  display: block;
}

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

/* Mobile responsive - portrait (< 576px) */
@media (max-width: 575px) {
  .app-header.hide-on-secondary-mobile {
    display: none;
  }

  .app-header-content {
    gap: 0.5rem;
  }

  .home-avatar,
  .nav-avatar {
    width: 2rem;
    height: 2rem;
    font-size: 0.85rem;
  }

  .brand-logo {
    width: 2rem;
    height: 2rem;
  }
}

/* Mobile responsive - landscape (576px - 767px) */
@media (min-width: 576px) and (max-width: 767px) {
  .app-header {
    margin-bottom: 0.875rem;
  }

  .app-header-content {
    gap: 0.75rem;
  }

  .home-avatar,
  .nav-avatar {
    width: 2.2rem;
    height: 2.2rem;
    font-size: 0.9rem;
  }

  .brand-logo {
    width: 2.2rem;
    height: 2.2rem;
  }
}
</style>
