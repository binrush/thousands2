<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from '../auth'
import { getImageUrl } from '../utils/images'
import { formatRussianDate } from '../utils/dates'
import Pagination from '../components/Pagination.vue'
import { usePagination } from '../composables/usePagination'

const route = useRoute()
const { authState } = useAuth()

const summit = ref(null)
const climbs = ref([])
const isLoading = ref(true)
const isLoadingClimbs = ref(false)
const error = ref(null)
const totalClimbs = ref(0)
const climbersSection = ref(null)
const showCommentModal = ref(false)
const selectedComment = ref(null)

// Watch for summit data changes to update page title
watch(summit, (newSummit) => {
  if (newSummit) {
    const title = newSummit.name || newSummit.height
    document.title = `${title}, хр. ${newSummit.ridge.name} | Тысячники Южного Урала`
  }
}, { immediate: true })

// Declare fetchClimbs first before using it in usePagination
async function fetchClimbs(page = 1) {
  try {
    isLoadingClimbs.value = true

    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}/climbs?page=${page}`)
    if (!response.ok) throw new Error('Failed to fetch climbs')
    const data = await response.json()

    // Update the climbs data
    climbs.value = data.climbs
    totalClimbs.value = data.total_climbs

    // Calculate total pages
    const itemsPerPage = 20
    totalPages.value = Math.ceil(data.total_climbs / itemsPerPage)

  } catch (err) {
    console.error('Error fetching climbs:', err)
  } finally {
    isLoadingClimbs.value = false
  }
}

// Now use our pagination composable
const { currentPage, totalPages, handlePageChange } = usePagination(fetchClimbs)

// Initial data fetch to get summit details
const fetchSummitDetails = async () => {
  try {
    isLoading.value = true
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`)
    if (!response.ok) throw new Error('Failed to fetch summit')
    summit.value = await response.json()
  } catch (err) {
    error.value = err.message
  } finally {
    isLoading.value = false
  }
}

// Initial fetch on mount
onMounted(() => {
  fetchSummitDetails()
  fetchClimbs(currentPage.value)
})

const openCommentModal = (climb) => {
  selectedComment.value = climb
  showCommentModal.value = true
}

const closeCommentModal = () => {
  showCommentModal.value = false
  selectedComment.value = null
}
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
    <div class="overflow-hidden">
      <div class="lg:flex">
        <div class="flex w-full lg:w-1/2">
          <div class="max-w-xl">
            <h1 class="text-3xl font-bold text-gray-900">
              {{ summit.name || summit.height }}
              <span v-if="summit.name_alt" class="text-gray-500 text-sm">({{ summit.name_alt }})</span>
            </h1>

            <div class="mt-2 flex items-center">
              <span class="text-gray-700">
                {{ summit.height }}м, хребет {{ summit.ridge.name }}
              </span>
            </div>

            <div class="mt-2">
              <RouterLink
                v-if="summit.coordinates"
                :to="{
                  name: 'map',
                  query: {
                    lat: summit.coordinates[0],
                    lng: summit.coordinates[1],
                    zoom: 12
                  }
                }"
                class="text-blue-600 hover:text-blue-800 underline"
                title="Показать на карте"
              >
                {{ summit.coordinates[0] }}, {{ summit.coordinates[1] }}
              </RouterLink>
            </div>

            <div v-if="summit.interpretation" class="mt-4 prose max-w-none" v-html="summit.interpretation"></div>

            <div v-if="summit.description" class="mt-4 prose max-w-none" v-html="summit.description"></div>

            <div v-if="authState.user" class="flex space-x-2 mt-6">
              <RouterLink v-if="!summit.climb_data" :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
                class="px-6 py-2 font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-600 rounded-lg hover:bg-blue-500 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-80">
                Зарегистрировать восхождение
              </RouterLink>
              <div v-else>
                <div class="text-gray-700 mb-2">
                  <div class="text-lg font-medium text-gray-900">Вы взошли на эту вершину</div>
                  <div v-if="summit.climb_data.date" class="text-sm text-gray-500">
                    {{ formatRussianDate(summit.climb_data.date) }}
                  </div>
                  <div v-if="summit.climb_data.comment" class="text-sm">
                    {{ summit.climb_data.comment }}
                  </div>
                </div>
                <RouterLink :to="`/${route.params.ridge_id}/${route.params.summit_id}/climb`"
                  class="inline-block px-4 py-2 text-sm font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-600 rounded-lg hover:bg-blue-500 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-80">
                  Редактировать
                </RouterLink>
              </div>
            </div>
          </div>
        </div>

        <div class="w-full h-64 lg:w-1/2 lg:h-auto">
          <img v-if="summit.images && summit.images.length > 0 && summit.images[0].url"
            :src="getImageUrl(summit.images[0].url)" :alt="summit.name || 'Вершина'"
            class="w-full h-full object-cover rounded-lg" />
          <div v-else class="w-full h-full bg-gray-200 flex items-center justify-center">
            <span class="text-gray-400 text-lg font-medium">Нет изображения</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Climbers List -->
    <div class="bg-white overflow-hidden" ref="climbersSection">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">
        Восходители
        <span v-if="totalClimbs > 0" class="text-sm font-normal text-gray-500">({{ totalClimbs }})</span>
      </h2>

      <!-- Fixed height container to prevent layout shifts -->
      <div class="min-h-[400px]">
        <div v-if="climbs.length === 0" class="text-center text-gray-500 py-4">
          Пока никто не зарегистрировал восхождение
        </div>

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          <div v-for="climb in climbs" :key="climb.user_id" class="flex items-start space-x-3 p-3 h-[100px]">
            <!-- Avatar -->
            <div class="flex-shrink-0">
              <img v-if="climb.user_image" :src="getImageUrl(climb.user_image)" :alt="climb.user_name"
                class="h-12 w-12 rounded-full object-cover">
              <img v-else src="/climber_no_photo.svg" :alt="climb.user_name" class="h-12 w-12 rounded-full">
            </div>

            <!-- User Info -->
            <div class="flex-1 min-w-0 overflow-hidden">
              <RouterLink :to="`/user/${climb.user_id}`"
                class="text-lg font-medium text-gray-900 hover:text-blue-600 block truncate">
                {{ climb.user_name }}
              </RouterLink>

              <div class="text-sm mt-1">
                <div v-if="climb.date" 
                     class="text-gray-500 cursor-pointer hover:text-blue-600"
                     @click="openCommentModal(climb)">
                  {{ formatRussianDate(climb.date) }}
                </div>
                <div v-if="climb.comment" 
                     class="mt-1 line-clamp-2 cursor-pointer hover:text-blue-600"
                     @click="openCommentModal(climb)">
                  {{ climb.comment }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Replace the pagination section with our component -->
      <div class="mt-12 mb-4">
        <Pagination :current-page="currentPage" :total-pages="totalPages" @page-change="handlePageChange" />
      </div>
    </div>

    <!-- Comment Modal -->
    <teleport to="body">
      <div v-if="showCommentModal"
           class="fixed top-0 left-0 w-screen h-screen bg-black bg-opacity-50 flex items-center justify-center z-50"
           @click="closeCommentModal">
        <div class="bg-white rounded-lg p-6 max-w-lg w-full mx-4" @click.stop>
          <div class="flex justify-between items-start mb-4">
            <div class="flex items-center space-x-3">
              <img v-if="selectedComment?.user_image" 
                   :src="getImageUrl(selectedComment.user_image)" 
                   :alt="selectedComment?.user_name"
                   class="h-12 w-12 rounded-full object-cover">
              <img v-else 
                   src="/climber_no_photo.svg" 
                   :alt="selectedComment?.user_name" 
                   class="h-12 w-12 rounded-full">
              <div>
                <h3 class="text-lg font-medium text-gray-900">
                  {{ selectedComment?.user_name }}
                </h3>
                <div v-if="selectedComment?.date" class="text-sm text-gray-500">
                  {{ formatRussianDate(selectedComment.date) }}
                </div>
              </div>
            </div>
            <button @click="closeCommentModal" class="text-gray-400 hover:text-gray-500">
              <span class="sr-only">Закрыть</span>
              <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div v-if="selectedComment?.comment" class="text-gray-700 whitespace-pre-wrap">
            {{ selectedComment.comment }}
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<style scoped></style>
