<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RouterLink } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isSubmitting = ref(false)
const summit = ref(null)
const isLoading = ref(true)

const climbData = ref({
  date: new Date().toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  }).replace(/\./g, '.'),
  description: '',
  photos: []
})

const MAX_DESCRIPTION_LENGTH = 1000

// Function to validate and format date
function formatDate(dateStr) {
  // Remove any non-digit characters except dots
  const cleaned = dateStr.replace(/[^\d.]/g, '')
  
  // Split by dots and filter out empty strings
  const parts = cleaned.split('.').filter(Boolean)
  
  // If we have only year (e.g. "2012")
  if (parts.length === 1 && parts[0].length === 4) {
    return parts[0]
  }
  
  // If we have month and year (e.g. "2.2012")
  if (parts.length === 2 && parts[0].length <= 2 && parts[1].length === 4) {
    return `${parts[0]}.${parts[1]}`
  }
  
  // If we have full date (e.g. "12.02.2012")
  if (parts.length === 3 && parts[0].length <= 2 && parts[1].length <= 2 && parts[2].length === 4) {
    return `${parts[0].padStart(2, '0')}.${parts[1].padStart(2, '0')}.${parts[2]}`
  }
  
  return dateStr
}

async function fetchSummit() {
  try {
    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`)
    if (!response.ok) throw new Error('Failed to fetch summit')
    summit.value = await response.json()
  } catch (error) {
    console.error('Error fetching summit:', error)
  } finally {
    isLoading.value = false
  }
}

async function submitClimb() {
  if (isSubmitting.value) return
  
  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('date', climbData.value.date)
    formData.append('comment', climbData.value.comment)

    const response = await fetch(`/api/summit/${route.params.ridge_id}/${route.params.summit_id}`, {
      method: 'PUT',
      body: formData
    })

    if (response.ok) {
      router.push({ name: 'summit', params: route.params })
    } else {
      throw new Error('Failed to submit climb')
    }
  } catch (error) {
    console.error('Error submitting climb:', error)
    alert('Произошла ошибка при сохранении восхождения')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchSummit()
})
</script>

<template>
  <div class="climb-form-view">
    <div class="max-w-screen-md mx-auto">
      <h1 class="text-2xl font-bold mb-8">Регистрация восхождения</h1>
      
      <!-- Summit Information -->
      <div v-if="isLoading" class="flex justify-center py-4">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
      
      <div v-else-if="summit" class="mb-8 p-6 bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="flex items-center space-x-2">
          <RouterLink 
            :to="{ name: 'summit', params: { ridge_id: route.params.ridge_id, summit_id: route.params.summit_id }}"
            class="text-lg font-medium text-blue-600 hover:text-blue-800"
          >
            {{ summit.name ? `${summit.name}, ` : '' }}{{ summit.height }}м
          </RouterLink>
          <span class="text-gray-500">|</span>
          <span class="text-gray-600">{{ summit.ridge.name }}</span>
        </div>
      </div>
      
      <form @submit.prevent="submitClimb" class="space-y-6">
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div class="space-y-6">
            <div class="form-group">
              <label for="date" class="block text-sm font-medium text-gray-700 mb-1">Дата восхождения</label>
              <input
                id="date"
                type="text"
                v-model="climbData.date"
                @input="climbData.date = formatDate($event.target.value)"
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                placeholder="дд.мм.гггг"
              >
              <p class="mt-2 text-sm text-gray-500">
                Необязательное поле. Если точная дата неизвестна, можно ввести только месяц (например 2.2012) или только год (например 2012)
              </p>
            </div>

            <div class="form-group">
              <label for="comment" class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
              <textarea
                id="comment"
                v-model="climbData.comment"
                rows="6"
                maxlength="1000"
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                placeholder="Расскажите о вашем восхождении..."
              ></textarea>
              <div class="flex justify-between mt-2">
                <p class="text-sm text-gray-500">
                  Необязательное поле.
                </p>
                <p class="text-sm text-gray-500">
                  {{ climbData.description.length }}/1000
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end">
          <button
            type="submit"
            class="inline-flex justify-center py-2 px-6 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.climb-form-view {
  padding: 20px;
}

.form-group {
  @apply space-y-1;
}

input, textarea {
  @apply transition-colors duration-200;
}

input:focus, textarea:focus {
  @apply ring-2 ring-blue-200;
}
</style>