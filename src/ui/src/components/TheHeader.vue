<template>
  <nav class="relative bg-white dark:bg-gray-800 fixed top-0 left-0 w-full z-50">
    <div class="container px-6 py-2 mx-auto">
      <div class="lg:flex lg:items-center lg:justify-between">
        <div class="flex items-center justify-between">
          <RouterLink to="/" class="flex items-center">
            <img src="/logo.svg" alt="Логотип" class="w-auto h-12 sm:h-14">
          </RouterLink>

          <!-- Mobile menu button -->
          <div class="flex lg:hidden">
            <button @click="toggleMobileMenu" type="button" class="text-gray-500 dark:text-gray-200 hover:text-gray-600 dark:hover:text-gray-400 focus:outline-none focus:text-gray-600 dark:focus:text-gray-400" aria-label="toggle menu">
              <svg v-if="!isMobileMenuOpen" xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 8h16M4 16h16" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Mobile Menu open: "block", Menu closed: "hidden" -->
        <div :class="[isMobileMenuOpen ? 'translate-x-0 opacity-100' : 'opacity-0 -translate-x-full']" class="absolute inset-x-0 z-20 w-full px-6 py-4 transition-all duration-300 ease-in-out bg-white dark:bg-gray-800 lg:mt-0 lg:p-0 lg:top-0 lg:relative lg:bg-transparent lg:w-auto lg:opacity-100 lg:translate-x-0 lg:flex lg:items-center">
          <div class="flex flex-col -mx-6 lg:flex-row lg:items-center lg:mx-8">
            <RouterLink to="/" class="px-3 py-2 mx-3 mt-2 text-gray-700 transition-colors duration-300 transform rounded-md lg:mt-0 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700">Вершины</RouterLink>
            <RouterLink to="/map" class="px-3 py-2 mx-3 mt-2 text-gray-700 transition-colors duration-300 transform rounded-md lg:mt-0 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700">Карта</RouterLink>
            <RouterLink to="/top" class="px-3 py-2 mx-3 mt-2 text-gray-700 transition-colors duration-300 transform rounded-md lg:mt-0 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700">Восходители</RouterLink>
            <RouterLink to="/about" class="px-3 py-2 mx-3 mt-2 text-gray-700 transition-colors duration-300 transform rounded-md lg:mt-0 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700">О проекте</RouterLink>
          </div>

          <div class="flex items-center mt-4 lg:mt-0">
            <template v-if="currentUser">
              <div class="flex items-center">
                <RouterLink :to="'/user/me'" class="flex items-center focus:outline-none" aria-label="toggle profile dropdown">
                  <div class="w-8 h-8 overflow-hidden border-2 border-gray-400 rounded-full">
                    <img :src="getImageUrl(currentUser.image_s)" class="object-cover w-full h-full" alt="Avatar">
                  </div>
                </RouterLink>
                <a href="/auth/logout" class="px-3 py-2 mx-3 text-gray-700 transition-colors duration-300 transform rounded-md dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700">Выйти</a>
              </div>
            </template>
            <a v-else href="/auth/oauth/vk" class="px-6 py-2 mx-3 font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-600 rounded-lg hover:bg-blue-500 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-80">
              Войти через VK
            </a>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { ref, inject } from 'vue'
import { RouterLink } from 'vue-router'
import { getImageUrl } from '../utils/images'

const currentUser = inject('currentUser')
const isMobileMenuOpen = ref(false)

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}
</script> 