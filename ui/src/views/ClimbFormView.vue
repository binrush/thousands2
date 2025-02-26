<script setup>
import { ref } from 'vue'

const climbData = ref({
  name: '',
  email: '',
})

function submitClimb(e) {

}
</script>

<template>
<div class="w-full max-w-screen-md mx-auto">
    <h1>Зарегистрировать восхождение</h1>
    
    <form @submit.prevent="submitClimb" class="grid grid-cols-2 gap-4 items-center justify-center">
      <label for="date" class="text-right">Дата:</label>
      <input id="date" type="text" v-model="climbData.date" class="p-1 rounded border">

      <label for="description" class="text-right">Комментарий:</label>
      <textarea id="description" v-model="climbData.description" class="p-1 rounded border"></textarea>

      <div></div>
      <button type="submit" class="inline-block bg-blue-500 text-white py-1 px-4 rounded">Сохранить</button>
    </form>
</div>
</template>

<style scoped>
</style>
