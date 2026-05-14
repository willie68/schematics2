<template>
  <div class="impressum-container">
    <div class="impressum-header">
      <h1>Impressum</h1>
      <p class="last-updated">Stand: {{ currentDate }}</p>
    </div>

    <div class="impressum-content">
      <!-- Verantwortlich -->
      <section class="impressum-section">
        <h2>Verantwortlich für den Inhalt</h2>
        <div class="contact-info">
          <p><strong>Name/Unternehmen:</strong>Wilfried Klaas</p>
          <p><strong>Adresse:</strong> Welperstraße 65<br>45525 Hattingen<br>Deutschland</p>
          <p><strong>Email:</strong> <a href="mailto:w.klaas@gmx.de">w.klaas@gmx.de</a></p>
        </div>
      </section>

      <!-- Haftung für Inhalte -->
      <section class="impressum-section">
        <h2>Haftung für Inhalte</h2>
        <p>
          Die Inhalte unserer Seiten wurden mit größter Sorgfalt erstellt. 
          Für die Richtigkeit, Vollständigkeit und Aktualität der Inhalte können wir jedoch keine Gewähr übernehmen. 
          Als Diensteanbieter sind wir gemäß § 7 Abs.1 TMG für eigene Inhalte auf diesen Seiten nach den allgemeinen Gesetzen verantwortlich. 
          Nach §§ 8 bis 10 des TMG sind wir als Diensteanbieter jedoch nicht verpflichtet, übermittelte oder gespeicherte fremde Informationen zu überwachen oder nach Umständen zu forschen, die auf eine rechtswidrige Tätigkeit hinweisen.
        </p>
        <p>
          Verpflichtungen zur Entfernung oder Sperrung der Nutzung von Informationen bleiben davon unberührt und gelten entsprechend den einschlägigen Gesetzen. 
          Eine diesbezügliche Haftung ist jedoch erst ab dem Zeitpunkt der Kenntnis einer konkreten Rechtsverletzung möglich.
        </p>
      </section>

      <!-- Haftung für Links -->
      <section class="impressum-section">
        <h2>Haftung für Links</h2>
        <p>
          Unser Angebot enthält Links zu externen Webseiten Dritter, auf deren Inhalte wir keinen Einfluss haben. 
          Deshalb können wir für diese fremden Inhalte auch keine Gewähr übernehmen. 
          Für die Inhalte der verlinkten Seiten ist stets der jeweilige Anbieter oder Betreiber der Seiten verantwortlich. 
          Die verlinkten Seiten wurden zum Zeitpunkt der Verlinkung auf mögliche Rechtsverstöße überprüft. 
          Rechtswidrige Inhalte waren zum Zeitpunkt der Verlinkung nicht erkennbar.
        </p>
        <p>
          Eine permanente inhaltliche Kontrolle der verlinkten Seiten ist jedoch ohne konkrete Anhaltspunkte einer Rechtsverletzung nicht zumutbar. 
          Bei Bekanntwerden von Rechtsverletzungen werden wir derartige Links umgehend entfernen.
        </p>
      </section>

      <!-- Urheberrecht -->
      <section class="impressum-section">
        <h2>Urheberrecht</h2>
        <p>
          Die durch die Seitenbetreiber erstellten Inhalte und Werke auf diesen Seiten unterliegen dem deutschen Urheberrecht. 
          Die Vervielfältigung, Bearbeitung, Verbreitung und jede Art der Verwertung außerhalb der Grenzen des Urheberrechtes bedürfen der schriftlichen Zustimmung des Autors oder Urhebers bzw. Seitenbetreibers. 
          Downloads und Kopien dieser Seite sind nur für den privaten, nicht kommerziellen Gebrauch gestattet.
        </p>
        <p>
          Soweit die Inhalte auf dieser Seite nicht von uns erstellt wurden, werden die Urheberrechte Dritter beachtet. 
          Insbesondere werden Inhalte Dritter als solche gekennzeichnet. 
          Sollten Sie trotzdem auf eine Urheberrechtsverletzung aufmerksam werden, bitten wir um einen entsprechenden Hinweis. 
          Bei Bekanntwerden von Rechtsverletzungen werden wir derartige Inhalte umgehend entfernen.
        </p>
      </section>

      <!-- Datenschutz -->
      <section class="impressum-section">
        <h2>Datenschutz</h2>
        <p>
          Informationen zur Verarbeitung personenbezogener Daten finden Sie in unserer <RouterLink to="/datenschutz">Datenschutzerklärung</RouterLink>.
        </p>
      </section>

      <!-- Streitbeilegung -->
      <section class="impressum-section">
        <h2>Streitbeilegung</h2>
        <p>
          Die Europäische Kommission stellt eine Plattform zur Online-Streitbeilegung (OS) bereit, die Sie unter 
          <a href="https://ec.europa.eu/consumers/odr" target="_blank" rel="noopener noreferrer">
            https://ec.europa.eu/consumers/odr
          </a>
          finden. Zur Teilnahme an einem Streitbeilegungsverfahren vor einer Verbraucherschlichtungsstelle sind wir bereit.
        </p>
      </section>

      <!-- Disclaimer -->
      <section class="impressum-section">
        <h2>Disclaimer</h2>
        <p>
          Die Inhalte dieser Website werden bereitgestellt, wie sie sind. 
          Die Betreiber dieser Website geben keine Zusicherungen oder Gewährleistungen, ausdrücklich oder stillschweigend, bezüglich der Richtigkeit, Vollständigkeit oder Eignung der hier enthaltenen Informationen für einen bestimmten Zweck. 
          Die Betreiber haften nicht für Schäden, die sich aus der Verwendung dieser Website oder der hier bereitgestellten Inhalte ergeben.
        </p>
      </section>

      <!-- Kontakt -->
      <section class="impressum-section">
        <h2>Kontakt</h2>
        <p>
          Für Fragen, Kritik oder sonstige Anliegen können Sie uns kontaktieren unter:<br>
          <strong>Email:</strong> <a href="mailto:w.klaas@gmx.de">w.klaas@gmx.de</a><br>
          <strong>Adresse:</strong> Welperstraße 65<br>45525 Hattingen<br>Deutschland
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
.impressum-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem 1rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
  color: #333;
}

.impressum-header {
  text-align: center;
  margin-bottom: 2rem;
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 1.5rem;
}

.impressum-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2.5rem;
  color: #1f2937;
}

.impressum-header .last-updated {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.impressum-content {
  line-height: 1.8;
}

.impressum-section {
  margin-bottom: 2rem;
}

.impressum-section h2 {
  font-size: 1.5rem;
  color: #1f2937;
  margin: 1.5rem 0 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.impressum-section p {
  margin: 0.75rem 0;
  text-align: justify;
}

.impressum-section a {
  color: #0066cc;
  text-decoration: none;
  transition: color 0.2s;
}

.impressum-section a:hover {
  color: #0052a3;
  text-decoration: underline;
}

.contact-info p {
  margin: 0.5rem 0;
  padding-left: 1rem;
}

.contact-info strong {
  display: inline-block;
  min-width: 120px;
  color: #1f2937;
}

/* Dark Mode Support */
@media (prefers-color-scheme: dark) {
  .impressum-container {
    color: #e0e0e0;
  }

  .impressum-header {
    border-bottom-color: #444;
  }

  .impressum-header h1 {
    color: #e0e0e0;
  }

  .impressum-header .last-updated {
    color: #aaa;
  }

  .impressum-section h2 {
    color: #e0e0e0;
    border-bottom-color: #444;
  }

  .contact-info strong {
    color: #e0e0e0;
  }

  .impressum-section a {
    color: #64b5f6;
  }

  .impressum-section a:hover {
    color: #90caf9;
  }
}

/* Responsive Design */
@media (max-width: 640px) {
  .impressum-container {
    padding: 1rem 0.75rem;
  }

  .impressum-header h1 {
    font-size: 1.75rem;
  }

  .impressum-section h2 {
    font-size: 1.25rem;
  }

  .impressum-section p {
    text-align: left;
  }
}
</style>
