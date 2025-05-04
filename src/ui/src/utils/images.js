import { config } from '../config'

export function getImageUrl(path) {
  if (!path) return '/default-avatar.png'
  return `${config.imageUrl}${path}`
} 