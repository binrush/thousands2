<script setup>
defineProps({
  size: {
    type: String,
    default: 'medium' // 'small' or 'medium'
  }
})
</script>

<template>
  <VDropdown :distance="8">
    <button
      type="button"
      :class="[
        'inline-flex items-center justify-center rounded-full bg-gray-200 text-gray-700 text-xs hover:bg-gray-300 active:bg-gray-300 focus:outline-none',
        size === 'small' ? 'w-4 h-4 ml-1' : 'w-5 h-5 ml-1 focus:ring-2 focus:ring-blue-500'
      ]"
      aria-label="Что такое превышение?"
      @click.stop
    >
      ?
    </button>
    <template #popper>
      <div class="p-3 max-w-xs text-sm text-gray-700 normal-case font-normal tracking-normal">
        <div class="font-medium mb-1">Превышение (относительная высота)</div>
        <div>Высота вершины над высшей седловиной, соединяющей её с более высокой вершиной. Помогает оценить «самостоятельность» вершины относительно окружающего рельефа.</div>
      </div>
    </template>
  </VDropdown>
</template>

