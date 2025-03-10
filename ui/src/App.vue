<script setup>
import { ref, onMounted, provide } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from './auth'

const currentUser = ref(null)
const { fetchAuthStatus, authState } = useAuth()

onMounted(async () => {
  await fetchAuthStatus()
  currentUser.value = authState.user
})

provide('currentUser', currentUser)
</script>

<template>
  <div class="h-screen flex flex-col">
    <!-- Навигация -->
    <header class="bg-white shadow-md fixed top-0 left-0 w-full z-50">
      <nav class="container mx-auto px-6 py-4 flex justify-between items-center">
        <!-- Логотип -->
        <RouterLink to="/" class="flex items-center space-x-3">
          <img src="/logo.svg" alt="Логотип" class="h-10">
        </RouterLink>

        <!-- Меню -->
        <ul class="hidden md:flex space-x-6 text-lg text-gray-700">
          <li><RouterLink to="/" class="hover:text-gray-500">Вершины</RouterLink></li> 
          <li><RouterLink to="/map" class="hover:text-gray-500">Карта</RouterLink></li> 
          <li><RouterLink to="/about" class="hover:text-gray-500">О проекте</RouterLink></li>
          <li v-if="currentUser"><RouterLink to="/user/me" class="hover:text-gray-500">{{ currentUser.name }}</RouterLink></li>
          <li v-if="currentUser"><a href="/auth/logout" class="hover:text-gray-500">Выйти</a></li>
          <li v-if="!currentUser"><a href="/auth/oauth/vk" class="hover:text-gray-500">Войти</a></li>
        </ul>
      </nav>
    </header>

    <!-- Контент -->
    <main class="mt-20 container mx-auto px-6">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
</style>
