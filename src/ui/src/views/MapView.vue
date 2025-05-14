<template>
    <div id="map" class="w-full h-full"></div>
</template>

<script setup>
import mapboxgl from 'mapbox-gl';
import MapboxLanguage from '@mapbox/mapbox-gl-language';
import { ref, onMounted, defineOptions } from 'vue'
import { useRouter } from 'vue-router'

// Specify layout to use
defineOptions({
    layout: 'map'
})

const router = useRouter()

async function loadMarkers(map) {
    // Wait for map to be fully loaded
    if (!map.loaded()) {
        await new Promise(resolve => {
            map.on('load', resolve);
        });
    }
    
    const res = await fetch("/api/summits")
    let summits = (await res.json()).summits

    // Define marker SVGs for regular and main summits (30% smaller)
    const regularMarkerSvg = `
        <svg width="17" height="25" viewBox="0 0 17 25" xmlns="http://www.w3.org/2000/svg">
            <path d="M8.5 0C3.8 0 0 3.8 0 8.5c0 5.2 8.5 16.5 8.5 16.5s8.5-11.3 8.5-16.5c0-4.7-3.8-8.5-8.5-8.5z" fill="{color}" stroke="white" stroke-width="1.2"/>
            {centerpiece}
        </svg>
    `;
    
    const mainMarkerSvg = `
        <svg width="22" height="34" viewBox="0 0 22 34" xmlns="http://www.w3.org/2000/svg">
            <path d="M11 0C4.9 0 0 4.9 0 11c0 6.7 11 23 11 23s11-16.3 11-23c0-6.1-4.9-11-11-11z" fill="{color}" stroke="white" stroke-width="1.5"/>
            {centerpiece}
        </svg>
    `;
    
    // Create markers directly
    summits.forEach(summit => {
        // Use color from API response
        const color = summit.color ? `#${summit.color}` : '#888888';
        const isMainPeak = summit.is_main;
        
        // For climbed summits, use a white flag instead of a circle
        let centerpiece = '';
        if (summit.climbed) {
            if (isMainPeak) {
                // Main marker: thinner pole
                centerpiece = `<g>
                    <rect x="9" y="6" width="2.2" height="15" fill="white"/>
                    <path d="M11 6.5 L18 11 L11 15.5 Z" fill="white"/>
                </g>`;
            } else {
                // Regular marker: thinner pole
                centerpiece = `<g>
                    <rect x="7" y="4.5" width="1.4" height="10" fill="white"/>
                    <path d="M8.5 5 L14 8 L8.5 11 Z" fill="white"/>
                </g>`;
            }
        } else {
            // Not climbed: use the original white circle
            centerpiece = isMainPeak
                ? '<circle cx="11" cy="11" r="4.5" fill="white"/>'
                : '<circle cx="8.5" cy="8.5" r="3.5" fill="white"/>';
        }
        
        // Choose the appropriate SVG template
        const svgTemplate = isMainPeak ? mainMarkerSvg : regularMarkerSvg;
        const markerSvg = svgTemplate
            .replace('{color}', color)
            .replace('{centerpiece}', centerpiece);
        
        // Create a DOM element for the marker
        const el = document.createElement('div');
        el.innerHTML = markerSvg;
        el.style.width = isMainPeak ? '22px' : '17px';
        el.style.height = isMainPeak ? '34px' : '25px';
        el.style.cursor = 'pointer';
        
        // Create the popup with appropriate offset
        const popupOffset = isMainPeak ? [0, -28] : [0, -21];
        
        // Create popup HTML with Tailwind classes
        const popupHTML = `
            <div class="font-sans rounded-lg p-1">
                <div class="text-sm font-medium mb-1">
                    <a href="${router.resolve({ name: 'summit', params: { ridge_id: summit.ridge_id, summit_id: summit.id } }).href}" 
                       class="text-blue-600 hover:text-blue-800 font-bold truncate block">
                        ${summit.name === null ? summit.height : summit.name}
                    </a>
                </div>
                <div class="text-xs text-gray-700">
                    <div class="flex items-center">
                        <span class="font-medium">Высота:</span>
                        <span class="ml-1">${summit.height} м</span>
                    </div>
                    <div class="flex items-center">
                        <span class="font-medium">Хребет:</span>
                        <span class="ml-1 truncate">${summit.ridge}</span>
                    </div>
                </div>
            </div>
        `;
        
        const popup = new mapboxgl.Popup({ 
            offset: popupOffset,
            className: 'mapbox-popup-tailwind' // Add a custom class for additional styling if needed
        })
            .setHTML(popupHTML);
        
        // Add marker to map
        new mapboxgl.Marker({ element: el })
            .setLngLat([summit.lng, summit.lat])
            .setPopup(popup)
            .addTo(map);
    });
}

function createMap() {
    mapboxgl.accessToken = 'pk.eyJ1IjoiYmlucnVzaCIsImEiOiJjbGk5dHB4YzIybDJjM2ZvM2FxZzhodmZrIn0.63GDcGk_4KwJlrBpvQVAVg';
    var map = new mapboxgl.Map({
        container: 'map',
        style: 'mapbox://styles/mapbox/outdoors-v12',
        center: [59.041, 54.480],
        zoom: 8
    });
    const language = new MapboxLanguage({
        defaultLanguage: 'ru'
    });
    map.addControl(language);
    return map;
}

onMounted(() => {
    let map = createMap()
    loadMarkers(map)
})
</script>

<style scoped>
/* Custom styles for Mapbox popups */
:global(.mapbox-popup-tailwind .mapboxgl-popup-content) {
  border-radius: 0.5rem;  /* rounded-lg equivalent */
  padding: 0.5rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: relative; /* Ensure positioning context for the close button */
  min-width: 180px; /* Ensure minimum width for long content */
  max-width: 220px; /* Limit maximum width */
}

/* Ensure popup content doesn't overflow */
:global(.mapbox-popup-tailwind .mapboxgl-popup-content > div) {
  width: 100%;
  overflow-wrap: break-word;
  word-break: break-word;
}

:global(.mapbox-popup-tailwind .mapboxgl-popup-tip) {
  border-top-color: white;  /* Match popup background color */
  border-bottom-color: white;  /* Match popup background color */
}

/* Improve close button styling and positioning */
:global(.mapbox-popup-tailwind .mapboxgl-popup-close-button) {
  position: absolute;
  top: 5px;
  right: 5px;
  padding: 4px 8px;
  font-size: 16px;
  color: #6b7280; /* text-gray-500 equivalent */
  background: transparent;
  border: none;
  border-radius: 9999px; /* rounded-full */
  cursor: pointer;
  transition: background-color 0.2s;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

:global(.mapbox-popup-tailwind .mapboxgl-popup-close-button:hover) {
  color: #1f2937; /* text-gray-800 */
  background-color: #f3f4f6; /* bg-gray-100 */
}
</style>
