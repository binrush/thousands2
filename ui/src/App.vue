<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from './auth'
import { provide } from 'vue'

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
  <header>
      <nav class="bg-gray-800 text-white py-4 px-8">
        <ul class="flex space-x-4">
            <li class="hover:text-gray-300"><RouterLink to="/">Вершины</RouterLink></li> 
            <li class="hover:text-gray-300"><RouterLink to="/map">Карта</RouterLink></li> 
            <li class="hover:text-gray-300"><RouterLink to="/about">О проекте</RouterLink></li>
            <li v-if="currentUser" class="hover:text-gray-300">
              <RouterLink to="/user/me">{{ currentUser.name }}</RouterLink>
            </li>
            <li v-if="currentUser" class="hover:text-gray-300"><a href="/auth/logout">Выйти</a></li>
            <li v-if="!currentUser" class="hover:text-gray-300"><a href="/auth/oauth/vk">Войти</a></li>
        </ul>
      </nav>
  </header>

  <RouterView />
</div>
</template>

<style scoped>
</style>
