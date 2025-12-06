<template>
  <div class="max-w-screen-md mx-auto overflow-hidden">
      <PageHeading>Рейтинг восходителей</PageHeading>
      
      <!-- Tabs Section -->
      <div class="border-b border-gray-200 mb-6">
        <nav class="-mb-px flex space-x-8">
          <button
            @click="switchTab('all-time')"
            :class="[
              activeTab === 'all-time'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm'
            ]"
          >
            За все время
          </button>
          <button
            @click="switchTab('year')"
            :class="[
              activeTab === 'year'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm'
            ]"
          >
            За текущий год
          </button>
        </nav>
      </div>

      <div v-if="!topClimbers.items.length" class="text-center text-gray-500 py-8">
        Нет данных для отображения
      </div>
      <ul v-else>
        <li v-for="climber in topClimbers.items" :key="climber.user_id"
          class="flex items-center py-3 border-b last:border-b-0">
          <div class="w-12 h-8 flex items-center justify-center text-blue-600 rounded-full mr-4 font-bold text-xl">
            {{ climber.place }}
          </div>
          <div class="mr-4">
            <UserAvatar 
              :image-url="climber.user_image"
              :alt-text="climber.user_name"
              :to="`/user/${climber.user_id}`"
            />
          </div>
          <div class="flex-grow flex items-center">
            <RouterLink :to="`/user/${climber.user_id}`"
              class="text-lg font-semibold text-gray-900 hover:text-blue-600">
              {{ climber.user_name }}
            </RouterLink>
          </div>
          <div class="text-lg whitespace-nowrap ml-4">
            {{ climber.climbs_num }}/{{ topClimbers.total_summits }}
          </div>
        </li>
      </ul>
      <!-- Pagination component -->
      <Pagination :current-page="currentPage" :total-pages="totalPages" @page-change="handlePageChange" class="mt-6" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Pagination from '../components/Pagination.vue'
import UserAvatar from '../components/UserAvatar.vue'
import PageHeading from '../components/PageHeading.vue'
import { usePagination } from '../composables/usePagination'

const route = useRoute()
const router = useRouter()

const activeTab = ref(route.query.tab || 'all-time')

const topClimbers = ref({
  items: [],
  page: 1,
  total_pages: 1,
  total_summits: 0
})

// Watch for tab query parameter changes
watch(() => route.query.tab, (newTab) => {
  if (newTab && (newTab === 'all-time' || newTab === 'year')) {
    activeTab.value = newTab
  } else if (!newTab) {
    activeTab.value = 'all-time'
  }
})

// Function to switch tabs and update URL
function switchTab(tab) {
  activeTab.value = tab
  currentPage.value = 1
  router.replace({ 
    query: { ...route.query, tab: tab === 'all-time' ? undefined : tab }
  })
  fetchTopClimbers(1)
}

// Fetch function for top climbers
async function fetchTopClimbers(page = 1) {
  try {
    const endpoint = activeTab.value === 'year' ? '/api/top/year' : '/api/top'
    const response = await fetch(`${endpoint}?page=${page}`)
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

<style scoped></style>