<template>
  <div class="shell">
    <header class="card" style="margin-bottom: 1rem">
      <div style="display:flex; align-items:center; justify-content:space-between; gap: 1rem; flex-wrap: wrap;">
        <div>
          <h1 style="margin-bottom:0.2rem">Schematic2</h1>
          <div class="muted">Nachfolger von WilliesSchematicsWorld</div>
          <div class="muted" style="font-size:0.85rem">Version {{ APP_VERSION }}</div>
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
  </div>
</template>

<script setup>
import { RouterLink, RouterView } from 'vue-router'
import UserMenu from './components/UserMenu.vue'
import Toast from './components/Toast.vue'
import { useAuth } from './composables/useAuth'
import { APP_VERSION } from './config'

const { isLoggedIn } = useAuth()
</script>
