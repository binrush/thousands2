import { config } from '../config'

export function getImageUrl(path) {
  if (!path) return '/climber_no_photo.svg'
  return `${config.imageUrl}${path}`
} 