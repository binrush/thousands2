<script setup>
import { inject } from 'vue'

const currentUser = inject('currentUser')

</script>

<template>
  <div class="w-full max-w-screen-md mx-auto">
    <div v-if="currentUser">
        <h1>{{ currentUser.name }}</h1>
    </div>
    <div v-else>Not authenticated</div>
  </div>
</template>

<style scoped>
</style>
