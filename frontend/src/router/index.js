import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import SearchView from '../views/SearchView.vue'
import EffectsView from '../views/EffectsView.vue'
import PrivacyView from '../views/PrivacyView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: HomeView },
    { path: '/login', component: LoginView },
    { path: '/register', component: RegisterView },
    { path: '/search', component: SearchView },
    { path: '/effects', component: EffectsView },
    { path: '/datenschutz', component: PrivacyView },
  ],
})

export default router
