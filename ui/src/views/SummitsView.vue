<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const summits = ref(null)
const sort_col = ref("ridge")

const router = useRouter()
const route = useRoute()

async function loadSummits() {
    const res = await fetch("http://localhost:5000/api/summits/table")
    summits.value = (await res.json()).summits
}

function update_sort(e) {
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
    if (sort_col.value === "ridge") {
        summits.value.sort(sort_by_ridge)
    }
    if (["height", "visitors"].includes(sort_col.value)) {
        summits.value.sort((a, b) => b[sort_col.value] - a[sort_col.value])
    }
    if (sort_col.value === "name") {
        summits.value.sort(sort_by_name)
    }
   
    return summits.value.map(function (o) {
        // replace name with height if name does not exist
        let summit = JSON.parse(JSON.stringify(o))
        summit.name = o.name !== null ? o.name : o.height.toString()
        return summit
    })
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
  <table class="table-auto">
      <thead>
      <tr>
          <th class="border" @click="update_sort('name')">Название</th>
          <th class="border" @click="update_sort('height')">Высота</th>
          <th class="border" @click="update_sort('ridge')">Хребет</th>
          <th class="border" @click="update_sort('visitors')">Восходителей</th>
      </tr>
      </thead>
      <tr v-for="summit in filteredSummits">
          <td class="border">{{ summit.name }}</td>
          <td class="border">{{ summit.height }}</td>
          <td class="border">{{ summit.ridge }}</td>
          <td class="border">{{ summit.visitors }}</td>
      </tr>
  </table>
</template>

<style>
</style>
