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
    router.push({ name: "summits", query: {sort: e} })
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
  <h1 class="text-xl font-bold">Все вершины Южного Урала выше тысячи метров</h1>
  <form>
    <input class="border my-1" type="text" v-model="search" placeholder="Поиск">
    <input type="button" value="X" @click="resetSearch">
  </form>
  <table class="table-auto">
      <thead>
      <tr>
          <th class="border" @click="updateSort('name')">Название</th>
          <th class="border" @click="updateSort('height')">Высота</th>
          <th class="border" @click="updateSort('ridge')">Хребет</th>
          <th class="border" @click="updateSort('visitors')">Восходителей</th>
      </tr>
      </thead>
      <tr v-for="summit in filteredSummits">
          <td class="border" :class="{ 'font-bold': summit.is_main }">
              <router-link :to="{ name: 'summit', params: { ridge_id: summit.ridge_id, summit_id: summit.id}}">
                {{ summit.name ? summit.name : summit.height}}
              </router-link>
          </td>
          <td class="border">
              <span :class="{ 'font-bold': summit.is_main }">{{ summit.height }}</span>
              <span class="text-xs ml-1">{{ summit.rank }}</span>
          </td>
          <td class="border">{{ summit.ridge }}</td>
          <td class="border">{{ summit.visitors }}</td>
      </tr>
  </table>
</template>

<style>
</style>
