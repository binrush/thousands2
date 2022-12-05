<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
<div class="container m-5">
  <header>
      <nav class="p-1">
        <ul>
            <li class="inline px-1"><RouterLink to="/">Вершины</RouterLink></li> 
            <li class="inline px-1"><RouterLink to="/about">О проекте</RouterLink></li>
        </ul>
      </nav>
  </header>

  <RouterView />
</div>
</template>

<style scoped>
</style>
