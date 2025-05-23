<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const summits = ref(null)
const sort_col = ref("ridge")
const search = ref('')

const router = useRouter()
const route = useRoute()

async function loadSummits() {
  const res = await fetch("/api/summits")
  summits.value = (await res.json()).summits
}

function resetSearch(e) {
  search.value = ''
}

function updateSort(e) {
  router.push({ name: "summits", query: { sort: e } })
}

function sort_by_ridge(a, b) {
  /*
    Sorts by ridge name and summits location
    (north to south)
  */
  if (a.ridge < b.ridge) {
    return -1
  }
  if (a.ridge > b.ridge) {
    return 1
  }
  return b.lat - a.lat
}

function sort_by_name(a, b) {
  /*
    Sorts by name field. Summits with
    non-null names comes first
  */
  if (a.name === null && b.name !== null) {
    return 1
  }
  if (a.name !== null && b.name === null) {
    return -1
  }
  if (a.name === null && b.name === null) {
    return a.height - b.height
  }
  if (a.name > b.name) {
    return 1
  }
  if (a.name < b.name) {
    return -1
  }
  return 0
}

const filteredSummits = computed(() => {
  if (summits.value === null) {
    return []
  }
  let result = summits.value.filter((s) => {
    let searchStr = search.value.trim().toLowerCase()
    return (s.name !== null && s.name.toLowerCase().includes(searchStr) ||
      s.ridge.toLowerCase().includes(searchStr)
    )
  })
  if (sort_col.value === "ridge") {
    result.sort(sort_by_ridge)
  }
  if (["height", "visitors"].includes(sort_col.value)) {
    result.sort((a, b) => b[sort_col.value] - a[sort_col.value])
  }
  if (sort_col.value === "name") {
    result.sort(sort_by_name)
  }

  return result
})

onMounted(function () {
  loadSummits()
  let sort = route.query.sort
  if (sort) { sort_col.value = sort }
})

watch(
  () => route.query.sort,
  (sort, prevSort) => {
    sort_col.value = sort ? sort : "ridge"
  }
)
</script>

<template>
  <div class="max-w-screen-md mx-auto overflow-hidden">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Все вершины Южного Урала выше тысячи метров</h1>

    <!-- Search -->
    <div class="mb-6">
      <div class="relative">
        <input type="text" v-model="search" placeholder="Поиск по названию или хребту..."
          class="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
        <button v-if="search" @click="resetSearch"
          class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="w-10"></th>
            <th @click="updateSort('name')"
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
              Название
            </th>
            <th @click="updateSort('height')"
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
              Высота
            </th>
            <th @click="updateSort('ridge')"
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
              Хребет
            </th>
            <th @click="updateSort('visitors')"
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
              Восходителей
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="summit in filteredSummits" :key="summit.id" class="hover:bg-gray-50">
            <td class="text-center">
              <svg v-if="summit.climbed" class="w-5 h-5 mx-auto" viewBox="0 0 24 24" fill="black" xmlns="http://www.w3.org/2000/svg">
                <rect x="7" y="4" width="2" height="16" />
                <path d="M9 4 L20 9 L9 14 Z" />
              </svg>
            </td>
            <td class="pr-6 py-4 whitespace-nowrap">
              <router-link
                :to="{ name: 'summit', params: { ridge_id: summit.ridge_id, summit_id: summit.id } }"
                class="text-blue-600 hover:text-blue-800"
                :class="{ 'font-bold': summit.is_main }"
              >
                {{ summit.name ? summit.name : summit.height }}
              </router-link>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="{ 'font-bold': summit.is_main }">{{ summit.height }}</span>
              <span v-if="summit.rank" class="text-xs text-gray-500 ml-1">{{ summit.rank }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-gray-900">
              {{ summit.ridge }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-gray-900">
              {{ summit.visitors }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped></style>
