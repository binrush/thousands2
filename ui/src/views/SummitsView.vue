<script setup>
import { ref, onMounted, computed } from 'vue'
const summits = ref(null)

async function loadSummits() {
    const res = await fetch("http://localhost:5000/api/summits/table")
    summits.value = (await res.json()).summits
}

const filteredSummits = computed(() => {
    if (summits.value === null) {
        return []
    } else {
        return summits.value
    }
})

onMounted(loadSummits)
</script>

<template>
  <h1 class="text-xl font-bold">Все вершины Южного Урала выше тысячи метров</h1>
  <table class="table-auto">
      <thead>
      <tr>
          <th class="border">Название</th>
          <th class="border">Высота</th>
          <th class="border">Хребет</th>
      </tr>
      </thead>
      <tr v-for="summit in filteredSummits">
          <td class="border">{{ summit.name }}</td>
          <td class="border">{{ summit.height }}</td>
          <td class="border">{{ summit.ridge }}</td>
      </tr>
  </table>
</template>

<style>
</style>
