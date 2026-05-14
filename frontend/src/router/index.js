import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import SearchView from '../views/SearchView.vue'
import EffectsView from '../views/EffectsView.vue'
import EditEffectView from '../views/EditEffectView.vue'
import PrivacyView from '../views/PrivacyView.vue'
import ImpressumView from '../views/ImpressumView.vue'
import DisclaimerView from '../views/DisclaimerView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: HomeView },
    { path: '/login', component: LoginView },
    { path: '/register', component: RegisterView },
    { path: '/search', component: SearchView },
    { path: '/effects', component: EffectsView },
    { path: '/effects/:id/edit', component: EditEffectView },
    { path: '/datenschutz', component: PrivacyView },
    { path: '/impressum', component: ImpressumView },
    { path: '/haftungsausschluss', component: DisclaimerView },
  ],
})

export default router
