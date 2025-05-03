<template>
  <div class="max-w-screen-md mx-auto">
    <div class="overflow-hidden">
      <div class="p-6">
        <h1 class="text-2xl font-bold text-gray-900 mb-6">Рейтинг восходителей</h1>
        <div v-if="!topClimbers.items.length" class="text-center text-gray-500 py-8">
          Нет данных для отображения
        </div>
        <ul v-else>
          <li
            v-for="climber in topClimbers.items"
            :key="climber.user_id"
            class="flex items-center py-3 border-b last:border-b-0"
          >
            <div class="w-8 h-8 flex items-center justify-center text-blue-600 rounded-full mr-4 font-bold text-xl">
              {{ climber.place }}
            </div>
            <div class="w-10 h-10 mr-4">
              <RouterLink :to="`/user/${climber.user_id}`">
                <img
                  v-if="climber.user_image"
                  :src="getImageUrl(climber.user_image)"
                  :alt="climber.user_name"
                  class="w-full h-full object-cover rounded-full"
                >
                <img
                  v-else
                  src="/climber_no_photo.svg"
                  :alt="climber.user_name"
                  class="w-full h-full object-cover rounded-full"
                >
              </RouterLink>
            </div>
            <div class="flex-grow flex items-center">
              <RouterLink
                :to="`/user/${climber.user_id}`"
                class="text-lg font-semibold text-gray-900 hover:text-blue-600"
              >
                {{ climber.user_name }}
              </RouterLink>
            </div>
            <div class="text-lg whitespace-nowrap ml-4">
              {{ climber.climbs_num }}/255
            </div>
          </li>
        </ul>
        <!-- Pagination component -->
        <Pagination
          :current-page="currentPage"
          :total-pages="totalPages"
          @page-change="handlePageChange"
          class="mt-6"
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