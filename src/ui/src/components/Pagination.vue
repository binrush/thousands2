<template>
  <div v-if="totalPages > 1" class="mt-6 flex justify-center">
    <div class="flex items-center w-[320px] justify-center">
      <!-- Previous button -->
      <button 
        @click="goToPreviousPage" 
        :class="[
          'flex items-center justify-center w-10 h-10 mx-1 text-gray-700 transition-colors duration-300 transform bg-white rounded-md rtl:-scale-x-100',
          currentPage === 1 ? 'cursor-not-allowed text-gray-500' : 'hover:bg-blue-500 hover:text-white'
        ]"
        :disabled="currentPage === 1"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
        </svg>
      </button>

      <!-- Center pagination area with fixed width -->
      <div class="flex items-center justify-center w-[220px]">
        <template v-for="(item, index) in paginationItems" :key="index">
          <button
            v-if="item.type === 'page'"
            @click="handlePageChange(item.value)"
            class="w-10 h-10 mx-1 transition-colors duration-300 transform bg-white rounded-md sm:inline hover:bg-blue-500 hover:text-white flex items-center justify-center"
            :class="currentPage === item.value ? 'bg-blue-500 text-white' : 'text-gray-700'"
          >
            {{ item.value }}
          </button>
          
          <span 
            v-else-if="item.type === 'ellipsis'" 
            class="w-10 h-10 mx-1 text-gray-700 flex items-center justify-center"
          >
            {{ item.value }}
          </span>

          <span
            v-else-if="item.type === 'empty'"
            class="w-10 h-10 mx-1 flex items-center justify-center invisible"
          >
            &nbsp;
          </span>
        </template>
      </div>

      <!-- Next button -->
      <button 
        @click="goToNextPage" 
        :class="[
          'flex items-center justify-center w-10 h-10 mx-1 text-gray-700 transition-colors duration-300 transform bg-white rounded-md rtl:-scale-x-100',
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
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  currentPage: {
    type: Number,
    required: true
  },
  totalPages: {
    type: Number,
    required: true
  },
  visiblePagesCount: {
    type: Number,
    default: 5
  }
})

const emit = defineEmits(['page-change'])

// Determine which pages to show in pagination
const paginationItems = computed(() => {
  const { currentPage, totalPages, visiblePagesCount } = props
  const items = []
  
  if (totalPages <= visiblePagesCount) {
    // If fewer pages than visiblePagesCount, show all pages
    for (let i = 1; i <= totalPages; i++) {
      items.push({ type: 'page', value: i })
    }
    
    // Fill remaining slots with empty items for consistent width
    for (let i = totalPages + 1; i <= visiblePagesCount; i++) {
      items.push({ type: 'empty', value: '' })
    }
  } else {
    // More than visiblePagesCount pages, we need to be selective
    if (currentPage <= 3) {
      // Near the start: show 1, 2, 3, ..., totalPages
      items.push({ type: 'page', value: 1 })
      items.push({ type: 'page', value: 2 })
      items.push({ type: 'page', value: 3 })
      items.push({ type: 'ellipsis', value: '...' })
      items.push({ type: 'page', value: totalPages })
    } else if (currentPage >= totalPages - 2) {
      // Near the end: show 1, ..., totalPages-2, totalPages-1, totalPages
      items.push({ type: 'page', value: 1 })
      items.push({ type: 'ellipsis', value: '...' })
      items.push({ type: 'page', value: totalPages - 2 })
      items.push({ type: 'page', value: totalPages - 1 })
      items.push({ type: 'page', value: totalPages })
    } else {
      // In the middle: show 1, ..., currentPage, ..., totalPages
      items.push({ type: 'page', value: 1 })
      items.push({ type: 'ellipsis', value: '...' })
      items.push({ type: 'page', value: currentPage })
      items.push({ type: 'ellipsis', value: '...' })
      items.push({ type: 'page', value: totalPages })
    }
  }
  
  return items
})

function handlePageChange(page) {
  if (props.currentPage !== page) {
    emit('page-change', page)
  }
}

function goToPreviousPage() {
  if (props.currentPage > 1) {
    emit('page-change', props.currentPage - 1)
  }
}

function goToNextPage() {
  if (props.currentPage < props.totalPages) {
    emit('page-change', props.currentPage + 1)
  }
}
</script> 