<script setup>
import { computed } from 'vue'

const props = defineProps({
  summit: {
    type: Object,
    required: true
  }
})

function shouldMarkSummit(summit) {
  return (summit.prominence !== null && summit.prominence < 50) || summit.height < 1000
}

const shouldShow = computed(() => shouldMarkSummit(props.summit))
</script>

<template>
  <VDropdown v-if="shouldShow" :distance="8">
    <button
      type="button"
      class="font-bold cursor-pointer focus:outline-none"
      aria-label="Информация о вершине"
      @click.stop
    >
      *
    </button>
    <template #popper>
      <div class="p-3 max-w-xs text-sm text-gray-700 normal-case font-normal tracking-normal">
        Вершина не отвечает критериям тысячника, но по историческим причинам включена в каталог проекта. <router-link to="/about" class="text-blue-600 hover:underline ml-1">Подробнее</router-link>
      </div>
    </template>
  </VDropdown>
</template>

