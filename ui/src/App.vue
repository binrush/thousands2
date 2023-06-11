<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
<div class="h-screen flex flex-col">
  <header>
      <nav class="bg-gray-800 text-white py-4 px-8">
        <ul class="flex space-x-4">
            <li class="hover:text-gray-300"><RouterLink to="/">Вершины</RouterLink></li> 
            <li class="hover:text-gray-300"><RouterLink to="/map">Карта</RouterLink></li> 
            <li class="hover:text-gray-300"><RouterLink to="/about">О проекте</RouterLink></li>
            <li class="hover:text-gray-300"><a href="/auth/oauth/vk">Вход</a></li>
        </ul>
      </nav>
  </header>

  <RouterView />
</div>
</template>

<style scoped>
</style>
