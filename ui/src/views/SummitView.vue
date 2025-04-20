<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from '../auth'
import { getImageUrl } from '../utils/images'

const route = useRoute()
const { authState } = useAuth()

const summit = ref(null)
const isLoading = ref(true)
const error = ref(null)
const currentPage = ref(1)
const totalPages = ref(1)

const fetchSummit = async (page = 1) => {
  try {
    isLoading.value = true
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}?page=${page}`)
    if (!response.ok) throw new Error('Failed to fetch summit')
    const data = await response.json()
    summit.value = data
    totalPages.value = Math.ceil(data.climbs.length / 10) // Assuming 10 items per page
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
}

const handlePageChange = (page) => {
  currentPage.value = page
  fetchSummit(page)
}

// Watch for route changes to reset pagination
watch(() => route.params, () => {
  currentPage.value = 1
  fetchSummit(1)
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
    <div class="bg-white rounded-lg shadow-md overflow-hidden">
      <div class="relative h-64">
        <img 
          v-if="summit.images && summit.images.length > 0" 
          :src="getImageUrl(summit.images[0].url)" 
          :alt="summit.name || 'Вершина'"
          class="w-full h-full object-cover"
        >
        <div v-else class="w-full h-full bg-gray-200 flex items-center justify-center">
          <span class="text-gray-400">Нет изображения</span>
        </div>
      </div>

      <div class="p-6">
        <div class="flex justify-between items-start">
          <div>
            <h1 class="text-3xl font-bold text-gray-900">{{ summit.name }}</h1>
            <p v-if="summit.name_alt" class="text-gray-600 mt-1">{{ summit.name_alt }}</p>
            <div class="mt-2 flex items-center space-x-4">
              <span class="text-gray-700">
                <span class="font-semibold">Высота:</span> {{ summit.height }} м
              </span>
              <span class="text-gray-700">
                <span class="font-semibold">Хребет:</span> {{ summit.ridge.name }}
              </span>
            </div>
          </div>
          
          <div v-if="authState.user" class="flex space-x-2">
            <RouterLink 
              v-if="!summit.climbed"
              :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
              class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
            >
              Зарегистрировать восхождение
            </RouterLink>
            <RouterLink 
              v-else
              :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
              class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-md text-sm font-medium"
            >
              Редактировать
            </RouterLink>
          </div>
        </div>

        <div v-if="summit.interpretation" class="mt-4 prose max-w-none" v-html="summit.interpretation"></div>

        <div v-if="summit.description" class="mt-4 prose max-w-none">
          {{ summit.description }}
        </div>

        <div v-if="summit.coordinates" class="mt-4">
          <span class="text-gray-700">
            <span class="font-semibold">Координаты:</span> 
            {{ summit.coordinates[0] }}, {{ summit.coordinates[1] }}
          </span>
        </div>
      </div>
    </div>

    <!-- Climbers List -->
    <div class="bg-white rounded-lg shadow-md overflow-hidden">
      <div class="p-6">
        <h2 class="text-2xl font-bold text-gray-900 mb-4">Восходители</h2>
        
        <div v-if="summit.climbs.length === 0" class="text-center text-gray-500 py-4">
          Пока никто не зарегистрировал восхождение
        </div>

        <div v-else class="space-y-4">
          <div v-for="climb in summit.climbs" :key="climb.user_id" class="flex items-center space-x-4 p-4 hover:bg-gray-50 rounded-lg">
            <img 
              :src="getImageUrl(climb.user_image)" 
              :alt="climb.user_name"
              class="h-12 w-12 rounded-full"
            >
            <div class="flex-1">
              <RouterLink 
                :to="`/user/${climb.user_id}`"
                class="text-lg font-medium text-gray-900 hover:text-blue-600"
              >
                {{ climb.user_name }}
              </RouterLink>
              <div class="text-sm text-gray-500">
                <span v-if="climb.date.Year">{{ climb.date.Year }}-{{ String(climb.date.Month).padStart(2, '0') }}-{{ String(climb.date.Day).padStart(2, '0') }}</span>
                <span v-if="climb.comment" class="ml-2">{{ climb.comment }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="mt-6 flex justify-center">
          <nav class="flex items-center space-x-2">
            <button
              v-for="page in totalPages"
              :key="page"
              @click="handlePageChange(page)"
              :class="[
                'px-3 py-1 rounded-md text-sm font-medium',
                currentPage === page
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              ]"
            >
              {{ page }}
            </button>
          </nav>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
