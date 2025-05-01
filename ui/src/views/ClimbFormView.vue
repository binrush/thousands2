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
      
      <div v-else-if="summit" class="mb-8 bg-white">
        <div class="flex items-center space-x-2">
          <RouterLink 
            :to="{ name: 'summit', params: { ridge_id: route.params.ridge_id, summit_id: route.params.summit_id }}"
            class="text-lg font-medium text-blue-600 hover:text-blue-800"
          >
            {{ summit.name ? `${summit.name}, ` : '' }}{{ summit.height }}м
          </RouterLink>
          <span class="text-gray-500">|</span>
          <span class="text-gray-600">хребет {{ summit.ridge.name }}</span>
        </div>
      </div>
      
      <form @submit.prevent="submitClimb" class="space-y-6">
        <div class="bg-white">
          <div class="space-y-6">
            <div class="form-group">
              <label for="date" class="block text-sm text-gray-500 dark:text-gray-300">Дата восхождения</label>
              <input
                id="date"
                type="text"
                v-model="climbData.date"
                @input="climbData.date = formatDate($event.target.value)"
                class="block  mt-2 w-full placeholder-gray-400/70 dark:placeholder-gray-500 rounded-lg border border-gray-200 bg-white px-5 py-2.5 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-40 dark:border-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:focus:border-blue-300"
                placeholder="дд.мм.гггг"
              >
              <p class="mt-2 text-xs text-gray-500">
                Необязательное поле. Если точная дата неизвестна, можно ввести только месяц (например 2.2012) или только год (например 2012)
              </p>
            </div>

            <div class="form-group">
              <label for="comment" class="block text-sm text-gray-500 dark:text-gray-300">Описание</label>
              <textarea
                id="comment"
                v-model="climbData.comment"
                rows="6"
                maxlength="1000"
                class="block  mt-2 w-full  placeholder-gray-400/70 dark:placeholder-gray-500 rounded-lg border border-gray-200 bg-white px-4 h-32 py-2.5 text-gray-700 focus:border-blue-400 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-40 dark:border-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:focus:border-blue-300"
                placeholder="Расскажите о вашем восхождении..."
              ></textarea>
              <div class="flex justify-between mt-2">
                <p class="text-xs text-gray-500">
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
            class="px-6 py-2 font-medium tracking-wide text-white capitalize transition-colors duration-300 transform bg-blue-600 rounded-lg hover:bg-blue-500 focus:outline-none focus:ring focus:ring-blue-300 focus:ring-opacity-80"
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