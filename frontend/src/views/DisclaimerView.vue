<template>
  <div class="disclaimer-container">
    <div class="disclaimer-header">
      <h1>Haftungsausschluss</h1>
      <p class="last-updated">Stand: {{ currentDate }}</p>
    </div>

    <div class="disclaimer-content">
      <!-- Allgemeiner Haftungsausschluss -->
      <section class="disclaimer-section">
        <h2>Allgemeiner Haftungsausschluss</h2>
        <p>
          Die Inhalte dieser Website werden bereitgestellt, wie sie sind ("as is"). 
          Die Betreiber dieser Website geben keine Zusicherungen oder Gewährleistungen, ausdrücklich oder stillschweigend, 
          bezüglich der Richtigkeit, Vollständigkeit, Zuverlässigkeit oder Eignung der Inhalte für einen bestimmten Zweck.
        </p>
      </section>

      <!-- Technische Verfügbarkeit -->
      <section class="disclaimer-section">
        <h2>Verfügbarkeit der Website</h2>
        <p>
          Wir bemühen uns, die Website ständig verfügbar zu halten. 
          Wir können jedoch nicht garantieren, dass die Website fehler- oder unterbrechungsfrei läuft. 
          Die Website kann jederzeit ohne vorherige Ankündigung gewartet, repariert oder verbessert werden.
        </p>
      </section>

      <!-- Haftungsbeschränkung -->
      <section class="disclaimer-section">
        <h2>Haftungsbeschränkung</h2>
        <p>
          Soweit gesetzlich zulässig, haften die Betreiber dieser Website unter keinen Umständen für direkte, 
          indirekte, zufällige, spezielle oder Folgeschäden, die sich aus der Nutzung oder Unfähigkeit zur Nutzung 
          dieser Website oder deren Inhalte ergeben, auch wenn auf die Möglichkeit solcher Schäden hingewiesen wurde.
        </p>
        <p>
          Dies umfasst insbesondere, aber nicht ausschließlich:
        </p>
        <ul>
          <li>Datenverlust</li>
          <li>Betriebsunterbrechung</li>
          <li>Verlust von Geschäftsmöglichkeiten</li>
          <li>Körper- oder Sachschäden</li>
          <li>Entgangene Gewinne</li>
          <li>Geschäftsunterbrechung</li>
          <li>Reputationsschäden</li>
        </ul>
      </section>

      <!-- Schadensersatz -->
      <section class="disclaimer-section">
        <h2>Beschränkung der Schadensersatzpflicht</h2>
        <p>
          Unter keinen Umständen übersteigt die Gesamthaftung der Betreiber dieser Website 
          den Betrag von einhundert Euro (100,00 €).
        </p>
      </section>

      <!-- Externe Links -->
      <section class="disclaimer-section">
        <h2>Externe Links und Verweise</h2>
        <p>
          Diese Website enthält Links zu Websites Dritter. Wir sind nicht verantwortlich für die Inhalte, 
          Genauigkeit, Vollständigkeit oder Qualität dieser verlinkten Websites. 
          Die Verlinkung auf externe Websites bedeutet keine Billigung oder Empfehlung dieser Websites oder ihrer Inhalte.
        </p>
        <p>
          Sie klicken auf externe Links auf eigenes Risiko. Wir empfehlen Ihnen, 
          die Datenschutz- und Nutzungsbedingungen jeder verlinkten Website zu überprüfen, 
          bevor Sie persönliche Daten preisgeben.
        </p>
      </section>

      <!-- Benutzerinhalte -->
      <section class="disclaimer-section">
        <h2>Benutzerinhalte und Benutzerbeiträge</h2>
        <p>
          Benutzer, die Inhalte auf dieser Website hochladen oder einreichen, garantieren, 
          dass diese Inhalte nicht gegen Gesetze verstoßen und keine Rechte Dritter verletzen.
        </p>
        <p>
          Durch das Hochladen von Inhalten stimmen Benutzer zu, dass die Betreiber dieser Website 
          solche Inhalte verwenden, modifizieren, reproduzieren und verbreiten dürfen.
        </p>
      </section>

      <!-- Verletzungen und Urheberrecht -->
      <section class="disclaimer-section">
        <h2>Urheberrechte und Marken</h2>
        <p>
          Alle Inhalte dieser Website, einschließlich Text, Grafiken, Logos, Bilder und Software, 
          sind Eigentum der Betreiber oder der Inhaltsanbieter und unterliegen dem Schutz durch 
          Urheberrechtsgesetze und internationale Verträge.
        </p>
        <p>
          Ohne vorherige schriftliche Genehmigung der Betreiber darf keine der Inhalte 
          vervielfältigt, vertrieben, übermittelt, zweckentfremdet oder in irgendeiner Weise 
          verwendet werden.
        </p>
      </section>

      <!-- Datenschutz -->
      <section class="disclaimer-section">
        <h2>Datenschutz und Sicherheit</h2>
        <p>
          Während wir Ihre Privatsphäre schützen, können wir die absolute Sicherheit 
          der über das Internet übermittelten Daten nicht garantieren. 
          Sie verwenden diese Website auf eigenes Risiko.
        </p>
        <p>
          Weitere Informationen zur Datenverarbeitung finden Sie in unserer 
          <RouterLink to="/datenschutz">Datenschutzerklärung</RouterLink>.
        </p>
      </section>

      <!-- Geltende Gesetze -->
      <section class="disclaimer-section">
        <h2>Anwendbares Recht</h2>
        <p>
          Dieser Haftungsausschluss unterliegt den Gesetzen der Bundesrepublik Deutschland. 
          Im Falle von Rechtsstreitigkeiten sind die Gerichte am Sitz der Betreiber zuständig.
        </p>
      </section>

      <!-- Änderungen -->
      <section class="disclaimer-section">
        <h2>Änderungen des Haftungsausschlusses</h2>
        <p>
          Die Betreiber dieser Website behalten sich vor, diesen Haftungsausschluss 
          jederzeit ohne vorherige Ankündigung zu ändern, zu ergänzen oder zu aktualisieren. 
          Die fortgesetzte Nutzung der Website nach solchen Änderungen gilt als Zustimmung zu den 
          geänderten Bedingungen.
        </p>
      </section>

      <!-- Kontakt -->
      <section class="disclaimer-section">
        <h2>Kontakt</h2>
        <p>
          Wenn Sie Bedenken oder Fragen zu diesem Haftungsausschluss haben, 
          kontaktieren Sie uns bitte über die im <RouterLink to="/impressum">Impressum</RouterLink> 
          angegebenen Kontaktinformationen.
        </p>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const currentDate = computed(() => {
  const now = new Date()
  const day = String(now.getDate()).padStart(2, '0')
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const year = now.getFullYear()
  return `${day}.${month}.${year}`
})
</script>

<style scoped>
.disclaimer-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem 1rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
  color: #333;
}

.disclaimer-header {
  text-align: center;
  margin-bottom: 2rem;
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 1.5rem;
}

.disclaimer-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2.5rem;
  color: #1f2937;
}

.disclaimer-header .last-updated {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.disclaimer-content {
  line-height: 1.8;
}

.disclaimer-section {
  margin-bottom: 2rem;
}

.disclaimer-section h2 {
  font-size: 1.5rem;
  color: #1f2937;
  margin: 1.5rem 0 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.disclaimer-section p {
  margin: 0.75rem 0;
  text-align: justify;
}

.disclaimer-section ul {
  margin: 0.75rem 0;
  padding-left: 2rem;
}

.disclaimer-section ul li {
  margin: 0.5rem 0;
}

.disclaimer-section a {
  color: #0066cc;
  text-decoration: none;
  transition: color 0.2s;
}

.disclaimer-section a:hover {
  color: #0052a3;
  text-decoration: underline;
}

/* Dark Mode Support */
@media (prefers-color-scheme: dark) {
  .disclaimer-container {
    color: #e0e0e0;
  }

  .disclaimer-header {
    border-bottom-color: #444;
  }

  .disclaimer-header h1 {
    color: #e0e0e0;
  }

  .disclaimer-header .last-updated {
    color: #aaa;
  }

  .disclaimer-section h2 {
    color: #e0e0e0;
    border-bottom-color: #444;
  }

  .disclaimer-section a {
    color: #64b5f6;
  }

  .disclaimer-section a:hover {
    color: #90caf9;
  }
}

/* Responsive Design */
@media (max-width: 640px) {
  .disclaimer-container {
    padding: 1rem 0.75rem;
  }

  .disclaimer-header h1 {
    font-size: 1.75rem;
  }

  .disclaimer-section h2 {
    font-size: 1.25rem;
  }

  .disclaimer-section p {
    text-align: left;
  }
}
</style>
