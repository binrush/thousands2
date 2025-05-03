<template>
  <footer class="bg-white dark:bg-gray-900 border-t border-gray-200">
    <div class="container flex flex-col items-center justify-between p-6 mx-auto space-y-4 sm:space-y-0 sm:flex-row">
      <RouterLink to="/">
        <img src="/logo.svg" alt="Логотип" class="w-auto h-8">
      </RouterLink>

      <p class="text-sm text-gray-600 dark:text-gray-300">© 2024 Тысячники Южного Урала</p>

      <div class="flex -mx-2">
        <a href="https://vk.com/" class="mx-2 text-gray-600 transition-colors duration-300 dark:text-gray-300 hover:text-blue-500 dark:hover:text-blue-400" aria-label="VK">
          <img src="/vklogo.svg" class="w-5 h-5" alt="VK">
        </a>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { RouterLink } from 'vue-router'
</script> 