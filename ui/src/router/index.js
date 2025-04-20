import { createRouter, createWebHistory } from 'vue-router'
import AboutView from '../views/AboutView.vue'
import SummitsView from '../views/SummitsView.vue'
import SummitView from '../views/SummitView.vue'
import MapView from '../views/MapView.vue'
import UserView from '../views/UserView.vue'
import ClimbFormView from '../views/ClimbFormView.vue'
import TopView from '../views/TopView.vue'
import { useAuth } from '../auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'summits',
      component: SummitsView,
      meta: { title: 'Вершины' }
    },
    {
      path: '/map',
      name: 'map',
      component: MapView,
      meta: { title: 'Карта' }
    },
    {
      path: '/top',
      name: 'top',
      component: TopView,
      meta: { title: 'Топ альпинистов' }
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView,
      meta: { title: 'О проекте' }
    },
    {
      path: '/:ridge_id/:summit_id',
      name: 'summit',
      component: SummitView,
      meta: { title: 'Вершина' }
    },
    {
      path: '/:ridge_id/:summit_id/climb',
      name: 'edit_climb',
      component: ClimbFormView,
      meta: { 
        title: 'Регистрация восхождения',
        requiresAuth: true 
      }
    },
    {
      path: '/user/me',
      name: 'user_profile',
      component: UserView,
      props: { user_id: 'me' },
      meta: { 
        title: 'Мой профиль',
        requiresAuth: true 
      }
    },
    {
      path: '/user/:user_id',
      name: 'user',
      component: UserView,
      meta: { title: 'Профиль пользователя' }
    }
  ]
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const { fetchAuthStatus, authState } = useAuth()
  
  // Update document title
  document.title = to.meta.title ? `${to.meta.title} | 1000+` : '1000+'
  
  // Ensure auth state is initialized
  if (!authState.isInitialized) {
    await fetchAuthStatus()
  }
  
  // Check if route requires authentication
  if (to.meta.requiresAuth && !authState.user) {
    // Redirect to login
    window.location.href = '/auth/oauth/vk'
    return
  }
  
  next()
})

export default router
