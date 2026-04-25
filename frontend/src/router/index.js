import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import SearchView from '../views/SearchView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/search' },
    { path: '/login', component: LoginView },
    { path: '/search', component: SearchView },
  ],
})

export default router
