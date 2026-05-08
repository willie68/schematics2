<template>
  <section class="card">
    <h2>Einstieg</h2>
    <p class="muted">Wähle einen Bereich: Dokumentensuche oder Effektdatenbank.</p>

    <div class="hub-grid">
      <article class="hub-card">
        <h3>Suche</h3>
        <p class="muted">Suche in Schaltplänen, Dokumentationen und PDFs über Volltext und Tags.</p>
        <RouterLink class="hub-link" to="/search">Zur Suche</RouterLink>
      </article>

      <article class="hub-card">
        <h3>Effektdatenbank</h3>
        <p class="muted">Bereich für Effekte aus WilliesSchematicsWorld, vorbereitet für spätere Datenanbindung.</p>
        <RouterLink class="hub-link" to="/effects">Zur Effektdatenbank</RouterLink>
      </article>
    </div>
  </section>
</template>

<script setup>
import { RouterLink } from 'vue-router'
</script>

<style scoped>
.hub-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.hub-card {
  border: 1px solid #d9e2ec;
  border-radius: 12px;
  padding: 1rem;
  background: linear-gradient(180deg, #ffffff 0%, #f7fbff 100%);
}

.hub-link {
  display: inline-block;
  margin-top: 0.7rem;
  color: #0077b6;
  font-weight: 600;
  text-decoration: none;
}

.hub-link:hover {
  text-decoration: underline;
}
</style>
