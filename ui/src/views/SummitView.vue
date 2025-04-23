<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from '../auth'
import { getImageUrl } from '../utils/images'
import { formatRussianDate } from '../utils/dates'

const route = useRoute()
const { authState } = useAuth()

const summit = ref(null)
const climbs = ref([])
const isLoading = ref(true)
const isLoadingClimbs = ref(false)
const error = ref(null)
const currentPage = ref(1)
const totalPages = ref(1)
const totalClimbs = ref(0)

// Determine which pages to show in pagination
const paginationItems = computed(() => {
  const items = []
  if (totalPages.value <= 6) {
    // If 6 or fewer pages, show all
    for (let i = 1; i <= totalPages.value; i++) {
      items.push({ type: 'page', value: i })
    }
  } else {
    // Always show first 2 pages
    items.push({ type: 'page', value: 1 })
    items.push({ type: 'page', value: 2 })
    
    // Determine if we need ellipsis after page 2
    if (currentPage.value > 4) {
      items.push({ type: 'ellipsis', value: '...' })
    }
    
    // Current page (if not in first 2 or last 2)
    if (currentPage.value > 2 && currentPage.value < totalPages.value - 1) {
      items.push({ type: 'page', value: currentPage.value })
    }
    
    // Determine if we need ellipsis before last 2 pages
    if (currentPage.value < totalPages.value - 3) {
      items.push({ type: 'ellipsis', value: '...' })
    }
    
    // Always show last 2 pages
    if (totalPages.value > 1) {
      items.push({ type: 'page', value: totalPages.value - 1 })
      items.push({ type: 'page', value: totalPages.value })
    }
  }
  
  return items
})

const goToPreviousPage = () => {
  if (currentPage.value > 1) {
    handlePageChange(currentPage.value - 1)
  }
}

const goToNextPage = () => {
  if (currentPage.value < totalPages.value) {
    handlePageChange(currentPage.value + 1)
  }
}

// Initial data fetch to get summit details
const fetchSummitDetails = async () => {
  try {
    isLoading.value = true
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}?page=1`)
    if (!response.ok) throw new Error('Failed to fetch summit')
    const data = await response.json()
    
    // Set summit data but not climbs
    summit.value = { ...data }
    delete summit.value.climbs // Remove climbs from summit data
    
    // Set climbs separately
    climbs.value = data.climbs
    totalClimbs.value = data.total_climbs
    
    // Use total_climbs for pagination
    const itemsPerPage = 10 // Items per page from API
    totalPages.value = Math.ceil(data.total_climbs / itemsPerPage)
    
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
}

// Fetch only the climbs for pagination
const fetchClimbs = async (page = 1) => {
  try {
    isLoadingClimbs.value = true
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}?page=${page}`)
    if (!response.ok) throw new Error('Failed to fetch climbs')
    const data = await response.json()
    
    // Update only the climbs data
    climbs.value = data.climbs
  } catch (err) {
    console.error('Error fetching climbs:', err)
  } finally {
    isLoadingClimbs.value = false
  }
}

const handlePageChange = (page) => {
  // Prevent default anchor behavior
  event?.preventDefault()
  
  if (currentPage.value === page) return
  
  currentPage.value = page
  fetchClimbs(page)
}

// Watch for route changes to reset pagination and fetch data
watch(() => route.params, () => {
  currentPage.value = 1
  fetchSummitDetails()
}, { immediate: true })
</script>

<template>
  <div v-if="error" class="text-red-600 text-center py-4">
    {{ error }}
  </div>

  <div v-else-if="isLoading" class="flex justify-center py-8">
    <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
  </div>

  <div v-else class="space-y-8">
    <!-- Summit Information -->
    <div class="bg-white overflow-hidden">
      <div class="lg:flex">
        <div class="flex w-full lg:w-1/2">
          <div class="max-w-xl">
            <h1 class="text-3xl font-bold text-gray-900">
              {{ summit.name || summit.height }}
              <span v-if="summit.name_alt" class="text-gray-500 text-sm">({{ summit.name_alt }})</span>
            </h1>
            
            <div class="mt-4 flex items-center space-x-4">
              <span class="text-gray-700">
                <span class="font-semibold">Высота:</span> {{ summit.height }} м
              </span>
              <span class="text-gray-700">
                <span class="font-semibold">Хребет:</span> {{ summit.ridge.name }}
              </span>
            </div>

            <div v-if="summit.coordinates" class="mt-2">
              <span class="text-gray-700">
                <span class="font-semibold">Координаты:</span> 
                {{ summit.coordinates[0] }}, {{ summit.coordinates[1] }}
              </span>
            </div>

            <div v-if="summit.interpretation" class="mt-4 prose max-w-none" v-html="summit.interpretation"></div>

            <div v-if="summit.description" class="mt-4 prose max-w-none" v-html="summit.description"></div>

            <div v-if="authState.user" class="flex space-x-2 mt-6">
              <RouterLink 
                v-if="!summit.climbed"
                :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
                class="bg-blue-600 hover:bg-blue-700 text-white px-5 py-2 rounded-md text-sm font-medium tracking-wider"
              >
                Зарегистрировать восхождение
              </RouterLink>
              <RouterLink 
                v-else
                :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
                class="bg-gray-600 hover:bg-gray-700 text-white px-5 py-2 rounded-md text-sm font-medium tracking-wider"
              >
                Редактировать
              </RouterLink>
            </div>
          </div>
        </div>

        <div class="w-full h-64 lg:w-1/2 lg:h-auto">
          <img 
            v-if="summit.images && summit.images.length > 0 && summit.images[0].url"
            :src="getImageUrl(summit.images[0].url)" 
            :alt="summit.name || 'Вершина'"
            class="w-full h-full object-cover"
          />
          <div v-else class="w-full h-full bg-gray-200 flex items-center justify-center">
            <span class="text-gray-400 text-lg font-medium">Нет изображения</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Climbers List -->
    <div class="bg-white overflow-hidden">
      <div class="py-6">
        <h2 class="text-2xl font-bold text-gray-900 mb-4">
          Восходители 
          <span v-if="totalClimbs > 0" class="text-sm font-normal text-gray-500">({{ totalClimbs }})</span>
        </h2>
        
        <!-- Fixed height container to prevent layout shifts -->
        <div class="min-h-[400px]">
          <div v-if="isLoadingClimbs" class="flex justify-center py-8">
            <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
          </div>
          
          <div v-else-if="climbs.length === 0" class="text-center text-gray-500 py-4">
            Пока никто не зарегистрировал восхождение
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            <div 
              v-for="climb in climbs" 
              :key="climb.user_id" 
              class="flex items-start space-x-3 p-3 h-[100px]"
            >
              <!-- Avatar -->
              <div class="flex-shrink-0">
                <img 
                  v-if="climb.user_image"
                  :src="getImageUrl(climb.user_image)" 
                  :alt="climb.user_name"
                  class="h-12 w-12 rounded-full object-cover"
                >
                <img 
                  v-else
                  src="/climber_no_photo.svg" 
                  :alt="climb.user_name"
                  class="h-12 w-12 rounded-full"
                >
              </div>
              
              <!-- User Info -->
              <div class="flex-1 min-w-0 overflow-hidden">
                <RouterLink 
                  :to="`/user/${climb.user_id}`"
                  class="text-lg font-medium text-gray-900 hover:text-blue-600 block truncate"
                >
                  {{ climb.user_name }}
                </RouterLink>
                
                <div class="text-sm mt-1">
                  <div v-if="climb.date" class="text-gray-500">{{ formatRussianDate(climb.date) }}</div>
                  <div v-if="climb.comment" class="mt-1 line-clamp-2">{{ climb.comment }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination - position absolute to prevent jumping -->
        <div v-if="totalPages > 1" class="mt-12 mb-4 relative">
          <div class="flex justify-center">
            <div class="flex">
              <!-- Previous button -->
              <button 
                @click="goToPreviousPage" 
                :class="[
                  'flex items-center justify-center px-4 py-2 mx-1 text-gray-700 transition-colors duration-300 transform bg-white rounded-md rtl:-scale-x-100',
                  currentPage === 1 ? 'cursor-not-allowed text-gray-500' : 'hover:bg-blue-500 hover:text-white'
                ]"
                :disabled="currentPage === 1"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </button>

              <!-- Page numbers and ellipsis -->
              <template v-for="(item, index) in paginationItems" :key="index">
                <button
                  v-if="item.type === 'page'"
                  @click="handlePageChange(item.value)"
                  class="hidden px-4 py-2 mx-1 transition-colors duration-300 transform bg-white rounded-md sm:inline hover:bg-blue-500 hover:text-white"
                  :class="currentPage === item.value ? 'bg-blue-500 text-white' : 'text-gray-700'"
                >
                  {{ item.value }}
                </button>
                
                <span 
                  v-else 
                  class="hidden px-4 py-2 mx-1 text-gray-700 bg-white sm:inline"
                >
                  {{ item.value }}
                </span>
              </template>

              <!-- Next button -->
              <button 
                @click="goToNextPage" 
                :class="[
                  'flex items-center justify-center px-4 py-2 mx-1 text-gray-700 transition-colors duration-300 transform bg-white rounded-md rtl:-scale-x-100',
                  currentPage === totalPages ? 'cursor-not-allowed text-gray-500' : 'hover:bg-blue-500 hover:text-white'
                ]"
                :disabled="currentPage === totalPages"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
