<script setup>

import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const route = useRoute()
const router = useRouter()

const summit = ref(null)

async function loadSummit() {
    const res = await fetch(
        `/api/summit/${route.params.ridge_id}/${route.params.summit_id}`)
    if (res.status === 404) {
        return
    }
    summit.value = await res.json()
}

onMounted(function () {
    loadSummit()
})

function submitClimb(e) {
    router.push({name: "edit_climb", ridge_id: route.params.ridge_id, summit_id: route.params.summit_id})
}

</script>

<template>
  <div class="w-full max-w-screen-md mx-auto">
    <div v-if="summit">
        <h1 class="text-xl font-bold">{{ summit.name ? summit.name : summit.height }}</h1>
        <dl>
            <dt class="font-bold" v-if="summit.name_alt">Варианты названий</dt>
            <dd v-if="summit.name_alt">{{ summit.name_alt }}</dd>
        </dl>
        <dl>
            <dt class="font-bold">Высота</dt>
            <dd>{{ summit.height }}</dd>
        </dl>
        <dl>
            <dt class="font-bold">Хребет</dt>
            <dd>{{ summit.ridge.name }}</dd>
        </dl>
        <dl>
            <dt class="font-bold">Координаты</dt>
            <dd>{{summit.coordinates[0]}} {{summit.coordinates[1]}}</dd>
        </dl>
        <dl>
            <dt class="font-bold" v-if="summit.interpretation">Расшифровка названия</dt>
            <dd v-if="summit.interpretation">{{summit.interpretation}}</dd>
        </dl>
        <dl>
            <dt class="font-bold" v-if="summit.description">Дополнительная информация</dt>
            <dd v-if="summit.description">{{summit.description}}</dd>
        </dl>
        <button @click="submitClimb" class="inline-block px-4 py-2 bg-blue-500 text-white rounded">
            Взошли на эту вершину?
        </button>
    </div>
    <div v-else>Загрузка...</div>
  </div>
</template>

<style scoped>
</style>
