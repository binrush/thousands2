import { createRouter, createWebHistory } from 'vue-router'
import AboutView from '../views/AboutView.vue'
import SummitsView from '../views/SummitsView.vue'
import SummitView from '../views/SummitView.vue'
import MapView from '../views/MapView.vue'
import UserView from '../views/UserView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'summits',
      component: SummitsView
    },
    {
      path: '/map',
      name: 'map',
      component: MapView
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView
    },
    {
      path: '/:ridge_id/:summit_id',
      name: 'summit',
      component: SummitView
    },
    {
      path: '/user/me',
      name: 'user',
      component: UserView 
    }

  ]
})

export default router
