<script setup>
import { ref, provide, watch } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from './auth'

const currentUser = ref(null)
const isMobileMenuOpen = ref(false)
const { authState } = useAuth()

// Watch for auth state changes to update currentUser
watch(() => authState.user, (newUser) => {
  currentUser.value = newUser
}, { immediate: true })

provide('currentUser', currentUser)

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50">
    <!-- Navigation -->
    <header class="bg-white shadow-md fixed top-0 left-0 w-full z-50">
      <nav class="container mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <!-- Logo and desktop menu -->
          <div class="flex items-center">
            <RouterLink to="/" class="flex items-center space-x-3">
              <img src="/logo.svg" alt="Логотип" class="h-10">
            </RouterLink>
          </div>
          <div class="hidden md:flex items-center space-x-8">
            <RouterLink to="/" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">Вершины</RouterLink>
            <RouterLink to="/map" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">Карта</RouterLink>
            <RouterLink to="/top" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">Топ</RouterLink>
            <RouterLink to="/about" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">О проекте</RouterLink>
          </div>

          <!-- Auth buttons -->
          <div class="hidden md:flex items-center space-x-4">
            <template v-if="currentUser">
              <RouterLink :to="'/user/me'" class="flex items-center space-x-2 text-gray-700 hover:text-gray-900">
                <img v-if="currentUser.avatar" :src="currentUser.avatar" class="h-8 w-8 rounded-full" alt="Avatar">
                <span class="text-sm font-medium">{{ currentUser.name }}</span>
              </RouterLink>
              <a href="/auth/logout" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">Выйти</a>
            </template>
            <a v-else href="/auth/oauth/vk" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium">
              Войти через VK
            </a>
          </div>

          <!-- Mobile menu button -->
          <div class="md:hidden flex items-center">
            <button @click="toggleMobileMenu" class="text-gray-700 hover:text-gray-900 focus:outline-none">
              <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path v-if="!isMobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Mobile menu -->
        <div v-if="isMobileMenuOpen" class="md:hidden">
          <div class="px-2 pt-2 pb-3 space-y-1">
            <RouterLink to="/" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">Вершины</RouterLink>
            <RouterLink to="/map" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">Карта</RouterLink>
            <RouterLink to="/top" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">Топ</RouterLink>
            <RouterLink to="/about" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">О проекте</RouterLink>
            <template v-if="currentUser">
              <RouterLink :to="'/user/me'" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">
                {{ currentUser.name }}
              </RouterLink>
              <a href="/auth/logout" class="block text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-base font-medium">Выйти</a>
            </template>
            <a v-else href="/auth/oauth/vk" class="block bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-base font-medium">
              Войти через VK
            </a>
          </div>
        </div>
      </nav>
    </header>

    <!-- Main content -->
    <main class="flex-1 mt-16">
      <div v-if="$route.name !== 'map'" class="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <RouterView />
      </div>
      <RouterView v-else />
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-gray-200">
      <div class="container mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="text-center text-gray-500 text-sm">
          © 2024 Тысячники Южного Урала
        </div>
      </div>
    </footer>
  </div>
</template>
