<template>
  <footer class="app-footer">
    <div class="footer-content">
      <p>&copy; 2026 Schematic2. Alle Rechte vorbehalten.</p>
      <div class="footer-links">
        <a href="https://wkla.no-ip.biz/willie/pages/impressum.php" target="_blank" rel="noopener noreferrer">
          Impressum
        </a>
        <span class="separator">|</span>
        <a href="https://www.disclaimer.de/disclaimer.htm?farbe=FFFFFF/000000/000000/000000" target="_blank" rel="noopener noreferrer">
          Haftungsausschluss
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup>
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
