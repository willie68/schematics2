<template>
  <footer class="app-footer">
    <div class="footer-content">
      <p>&copy; 2026 Schematics2. Alle Rechte vorbehalten.</p>
      <p class="version-info">Versionen: App {{ APP_VERSION }}, Backend {{ BACKEND_VERSION }}</p>
      <div class="footer-links">
        <RouterLink to="/datenschutz" class="footer-link">
          Datenschutz
        </RouterLink>
        <span class="separator">|</span>
        <RouterLink to="/impressum" class="footer-link">
          Impressum
        </RouterLink>
        <span class="separator">|</span>
        <RouterLink to="/haftungsausschluss" class="footer-link">
          Haftungsausschluss
        </RouterLink>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import { ref, onMounted } from 'vue'
import api from '../services/api'
import { APP_VERSION } from '../config'

const BACKEND_VERSION = ref('Loading...')

onMounted(async () => {
  try {
    const { data } = await api.get('/api/v1/info')
    BACKEND_VERSION.value = data.version || 'Unknown'
  } catch (error) {
    BACKEND_VERSION.value = 'Error'
    console.error('Failed to fetch backend version:', error)
  }
})
</script>

<style scoped>
.app-footer {
  background: #f8f9fa;
  border-top: 1px solid #dee2e6;
  padding: 2rem 1rem;
  margin-top: 3rem;
  text-align: center;
  font-size: 0.9rem;
  color: #666;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
}

.footer-content p {
  margin: 0 0 0.75rem 0;
  color: #666;
}

.version-info {
  font-size: 0.85rem;
  color: #999;
  margin: 0.5rem 0 0.75rem 0;
}

.footer-links {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.footer-links a {
  color: #0066cc;
  text-decoration: none;
  transition: color 0.2s;
}

.footer-links a:hover {
  color: #0052a3;
  text-decoration: underline;
}

.footer-link {
  color: #0066cc;
  text-decoration: none;
  transition: color 0.2s;
}

.footer-link:hover {
  color: #0052a3;
  text-decoration: underline;
}

.separator {
  color: #ccc;
}

@media (max-width: 768px) {
  .app-footer {
    padding: 1.5rem 0.75rem;
    margin-top: 2rem;
  }

  .footer-content p {
    font-size: 0.85rem;
  }

  .footer-links {
    gap: 0.75rem;
  }
}
</style>
