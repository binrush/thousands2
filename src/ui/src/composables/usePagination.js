import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

/**
 * Composable for handling pagination with URL query parameters
 * @param {Function} fetchFunction - Function to call when page changes
 * @param {Object} options - Configuration options
 * @returns {Object} Pagination state and handlers
 */
export function usePagination(fetchFunction, options = {}) {
  const route = useRoute()
  const router = useRouter()
  
  // Default options
  const { 
    scrollRef = null,
    preserveScroll = true
  } = options
  
  const currentPage = ref(parseInt(route.query.page) || 1)
  const totalPages = ref(1)
  
  // Initial setup - read from URL
  watch(() => route.query.page, (newPage) => {
    const pageNum = newPage ? parseInt(newPage) : 1
    if (pageNum !== currentPage.value && pageNum > 0) {
      currentPage.value = pageNum
      fetchFunction(pageNum)
    }
  }, { immediate: true })
  
  // Handle page change
  const handlePageChange = (page) => {
    if (currentPage.value === page) return
    
    // Capture current scroll position if preserveScroll is enabled
    let currentScrollPos = 0
    if (preserveScroll && scrollRef?.value) {
      currentScrollPos = scrollRef.value.getBoundingClientRect().top + window.scrollY
    }
    
    // Update URL with the new page
    const query = { ...route.query }
    if (page === 1) {
      // Remove page parameter for page 1 (cleaner URL)
      delete query.page
    } else {
      query.page = page
    }
    
    // Use router replace to update URL without creating new history entries
    router.replace({ 
      query,
      params: route.params
    }, undefined, { 
      preserveState: true,
      preventScroll: true 
    })
    
    // Update page and fetch data
    currentPage.value = page
    fetchFunction(page)
    
    // Restore scroll position after data is loaded
    if (preserveScroll && scrollRef?.value) {
      setTimeout(() => {
        window.scrollTo({
          top: currentScrollPos,
          behavior: 'auto'
        })
      }, 100)
    }
  }
  
  return {
    currentPage,
    totalPages,
    handlePageChange
  }
} 