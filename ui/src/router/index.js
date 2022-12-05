import { createRouter, createWebHistory } from 'vue-router'
import AboutView from '../views/AboutView.vue'
import SummitsView from '../views/SummitsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'summits',
      component: SummitsView
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView
    }
  ]
})

export default router
