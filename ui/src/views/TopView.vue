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
                <div class="w-10 h-10 mr-4">
                  <img 
                    v-if="climber.user_image" 
                    :src="getImageUrl(climber.user_image)" 
                    :alt="climber.user_name"
                    class="w-full h-full object-cover rounded-full"
                  >
                  <div 
                    v-else 
                    class="w-full h-full bg-gray-200 rounded-full flex items-center justify-center text-gray-500"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
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

        <!-- Pagination component -->
        <Pagination 
          :current-page="currentPage" 
          :total-pages="totalPages" 
          @page-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { getImageUrl } from '../utils/images'
import Pagination from '../components/Pagination.vue'
import { usePagination } from '../composables/usePagination'

const topClimbers = ref({
  items: [],
  page: 1,
  total_pages: 1
})

// Fetch function for top climbers
async function fetchTopClimbers(page = 1) {
  try {
    const response = await fetch(`/api/top?page=${page}`)
    if (response.ok) {
      const data = await response.json()
      topClimbers.value = data
      
      // Update totalPages from response
      if (data.total_pages) {
        totalPages.value = data.total_pages
      }
    }
  } catch (error) {
    console.error('Error loading top climbers:', error)
  }
}

// Use the pagination composable
const { currentPage, totalPages, handlePageChange } = usePagination(fetchTopClimbers)

// Initial fetch on mount
onMounted(() => {
  fetchTopClimbers(currentPage.value)
})
</script>

<style scoped>
</style> 