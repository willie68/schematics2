<template>
  <div class="shell">
    <header class="card" style="margin-bottom: 1rem">
      <div style="display:flex; align-items:center; justify-content:space-between; gap: 1rem; flex-wrap: wrap;">
        <div>
          <h1 style="margin-bottom:0.2rem">Schematic2</h1>
          <div class="muted">Nachfolger von WilliesSchematicsWorld</div>
        </div>
        <nav style="display:flex; gap:0.6rem;">
          <RouterLink to="/">Start</RouterLink>
          <RouterLink to="/search">Suche</RouterLink>
          <RouterLink to="/effects">Effektdatenbank</RouterLink>
          <RouterLink to="/login">Login</RouterLink>
        </nav>
      </div>
    </header>

    <RouterView />
  </div>
</template>

<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>
