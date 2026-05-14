<template>
  <section class="card" style="max-width: 900px; margin: 2rem auto;">
    <h2>Datenschutzerklärung</h2>
    
    <div class="privacy-content">
      <h3>1. Verantwortlicher</h3>
      <p>
        Verantwortlich für die Datenverarbeitung auf dieser Webseite ist:
      </p>
      <p>
        <strong>Willie's Schematic World</strong><br>
        <a href="https://wkla.no-ip.biz/willie/" target="_blank" rel="noopener noreferrer">https://wkla.no-ip.biz/willie/</a>
      </p>

      <h3>2. Erhobene Daten</h3>
      <p>
        Diese Webseite erhebt nur die Daten, die für die technische Funktionalität notwendig sind:
      </p>
      <ul>
        <li><strong>Authentifizierungs-Token (JWT)</strong> – Gespeichert im Browser LocalStorage zur Aufrechterhaltung der Benutzer-Sitzung</li>
        <li><strong>Cookie-Zustimmung</strong> – LocalStorage-Eintrag zur Speicherung Ihrer Zustimmung zu dieser Datenschutzerklärung</li>
        <li><strong>HTTP-Request-Daten</strong> – Standard-Protokolldaten (IP-Adresse, Zeitstempel) auf dem Server für technische Zwecke</li>
      </ul>

      <h3>3. Verwendung der Daten</h3>
      <p>
        Die erhobenen Daten werden ausschließlich für folgende Zwecke verwendet:
      </p>
      <ul>
        <li>Authentifizierung und Autorisierung von Benutzern</li>
        <li>Bereitstellung und Wartung der Webseite</li>
        <li>Fehleranalyse und Sicherheit</li>
      </ul>
      <p>
        <strong>Es werden keine Daten zu Tracking-, Marketing- oder Analysezwecken (z. B. Google Analytics) erhoben oder weitergegeben.</strong>
      </p>

      <h3>4. LocalStorage und Browser-Speicher</h3>
      <p>
        Diese Webseite verwendet Browser LocalStorage ausschließlich für technisch notwendige Daten:
      </p>
      <ul>
        <li><code>schematic2_token</code> – Ihr Authentifizierungs-Token (JWT) zur Aufrechterhaltung der Sitzung</li>
        <li><code>schematic2_cookie_consent</code> – Ihre Zustimmung zu dieser Datenschutzerklärung</li>
      </ul>
      <p>
        Diese Daten werden <strong>nur lokal in Ihrem Browser</strong> gespeichert und nicht an externe Server übertragen (außer dem notwendigen API-Token für Authentifizierung).
      </p>

      <h3>5. Datensicherheit</h3>
      <p>
        Ihre Daten werden durch folgende Maßnahmen geschützt:
      </p>
      <ul>
        <li>HTTPS-Verschlüsselung für alle Verbindungen</li>
        <li>JWT-basierte Authentifizierung mit sicheren Tokens</li>
        <li>Minimale Datenerhebung (nur notwendige Daten)</li>
      </ul>

      <h3>6. Ihre Rechte</h3>
      <p>
        Sie haben folgende Rechte bezüglich Ihrer Daten:
      </p>
      <ul>
        <li><strong>Zugriff</strong> – Sie können jederzeit einsehen, welche Daten über Sie gespeichert sind</li>
        <li><strong>Löschung</strong> – Sie können Ihr Konto und Ihre Daten jederzeit löschen</li>
        <li><strong>Speicherung verwalten</strong> – Sie können LocalStorage in Ihrem Browser jederzeit löschen (führt zu Logout)</li>
      </ul>

      <h3>7. Cookies</h3>
      <p>
        Diese Webseite setzt <strong>keine klassischen HTTP-Cookies</strong>. 
        Der Browser-Speicher (LocalStorage) wird verwendet, da dieser direkter und transparenter ist.
      </p>
      <p>
        Sie können Ihr LocalStorage jederzeit in den Browser-Einstellungen löschen:
      </p>
      <ul>
        <li><strong>Chrome/Chromium:</strong> Developer Tools (F12) → Application → Local Storage</li>
        <li><strong>Firefox:</strong> Developer Tools (F12) → Storage → Local Storage</li>
        <li><strong>Safari:</strong> Preferences → Privacy → Manage Website Data</li>
      </ul>

      <h3>8. Dritte</h3>
      <p>
        Ihre Daten werden <strong>nicht an Dritte weitergegeben</strong>, mit Ausnahme von:
      </p>
      <ul>
        <li>Erforderlich durch Gesetz (z. B. Behördenanfragen)</li>
        <li>Notwendig für die Sicherheit der Webseite</li>
      </ul>

      <h3>9. Änderungen dieser Datenschutzerklärung</h3>
      <p>
        Wir behalten uns das Recht vor, diese Datenschutzerklärung jederzeit zu ändern. 
        Änderungen werden auf dieser Seite veröffentlicht.
      </p>

      <h3>10. Kontakt</h3>
      <p>
        Bei Fragen zur Datenschutzerklärung kontaktieren Sie bitte:
      </p>
      <p>
        <a href="https://wkla.no-ip.biz/willie/pages/impressum.php" target="_blank" rel="noopener noreferrer">
          Impressum / Kontaktdaten
        </a>
      </p>

      <div style="margin-top:2rem; padding:1rem; background:#f0f0f0; border-radius:4px;">
        <p style="margin:0; color:#666; font-size:0.9rem;">
          <strong>Letzte Aktualisierung:</strong> 14. Mai 2026
        </p>
      </div>
    </div>
  </section>
</template>

<script setup>
</script>

<style scoped>
.privacy-content {
  line-height: 1.8;
  color: #333;
}

.privacy-content h3 {
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  color: #000;
  border-bottom: 2px solid #0066cc;
  padding-bottom: 0.5rem;
}

.privacy-content h3:first-child {
  margin-top: 0;
}

.privacy-content p {
  margin-bottom: 1rem;
}

.privacy-content ul {
  margin-left: 2rem;
  margin-bottom: 1rem;
}

.privacy-content ul li {
  margin-bottom: 0.5rem;
}

.privacy-content code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

.privacy-content a {
  color: #0066cc;
  text-decoration: none;
}

.privacy-content a:hover {
  text-decoration: underline;
}
</style>
