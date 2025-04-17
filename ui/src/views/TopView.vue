<template>
  <div class="max-w-screen-md mx-auto px-4 py-8">
    <div class="bg-white rounded-lg shadow-md overflow-hidden">
      <div class="p-6">
        <h1 class="text-2xl font-bold text-gray-900 mb-6">Топ альпинистов</h1>
        
        <div class="space-y-4">
          <div 
            v-for="climber in topClimbers.items" 
            :key="climber.user_id" 
            class="bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors duration-200"
          >
            <div class="p-4">
              <div class="flex items-center">
                <div class="w-10 h-10 flex items-center justify-center bg-blue-100 text-blue-600 rounded-full mr-4 font-semibold">
                  {{ climber.place }}
                </div>
                <div class="flex-grow">
                  <RouterLink 
                    :to="`/user/${climber.user_id}`"
                    class="text-lg font-medium text-gray-900 hover:text-blue-600"
                  >
                    {{ climber.user_name }}
                  </RouterLink>
                  <p class="text-sm text-gray-600">Количество восхождений: {{ climber.climbs_num }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="topClimbers.total_pages > 1" class="mt-6 flex justify-center">
          <nav class="flex items-center space-x-2">
            <!-- Previous page button -->
            <button
              @click="loadTopClimbers(topClimbers.page - 1)"
              :disabled="topClimbers.page === 1"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                topClimbers.page === 1
                  ? 'text-gray-400 cursor-not-allowed'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              ←
            </button>

            <!-- First page -->
            <button
              @click="loadTopClimbers(1)"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                topClimbers.page === 1
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              1
            </button>

            <!-- Left ellipsis -->
            <span v-if="leftEllipsis" class="px-2 text-gray-400">...</span>

            <!-- Page numbers -->
            <button
              v-for="page in visiblePages"
              :key="page"
              @click="loadTopClimbers(page)"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                topClimbers.page === page
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              {{ page }}
            </button>

            <!-- Right ellipsis -->
            <span v-if="rightEllipsis" class="px-2 text-gray-400">...</span>

            <!-- Last page -->
            <button
              v-if="topClimbers.total_pages > 1"
              @click="loadTopClimbers(topClimbers.total_pages)"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                topClimbers.page === topClimbers.total_pages
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              {{ topClimbers.total_pages }}
            </button>

            <!-- Next page button -->
            <button
              @click="loadTopClimbers(topClimbers.page + 1)"
              :disabled="topClimbers.page === topClimbers.total_pages"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                topClimbers.page === topClimbers.total_pages
                  ? 'text-gray-400 cursor-not-allowed'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              →
            </button>
          </nav>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const topClimbers = ref({
  items: [],
  page: 1,
  total_pages: 1
})

const VISIBLE_PAGES_COUNT = 5 // Number of pages to show between ellipsis

const visiblePages = computed(() => {
  const current = topClimbers.value.page
  const total = topClimbers.value.total_pages
  const pages = []
  
  // Calculate range of visible pages
  let start = Math.max(2, current - Math.floor(VISIBLE_PAGES_COUNT / 2))
  let end = Math.min(total - 1, start + VISIBLE_PAGES_COUNT - 1)
  
  // Adjust start if we're near the end
  if (end === total - 1) {
    start = Math.max(2, end - VISIBLE_PAGES_COUNT + 1)
  }
  
  // Add visible page numbers
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

const leftEllipsis = computed(() => {
  return visiblePages.value.length > 0 && visiblePages.value[0] > 2
})

const rightEllipsis = computed(() => {
  return visiblePages.value.length > 0 && 
         visiblePages.value[visiblePages.value.length - 1] < topClimbers.value.total_pages - 1
})

async function loadTopClimbers(page = 1) {
  try {
    const response = await fetch(`/api/top?page=${page}`)
    if (response.ok) {
      topClimbers.value = await response.json()
    }
  } catch (error) {
    console.error('Error loading top climbers:', error)
  }
}

onMounted(() => {
  loadTopClimbers()
})
</script>

<style scoped>
</style> 