<script setup>

import mapboxgl from 'mapbox-gl';
import MapboxLanguage from '@mapbox/mapbox-gl-language';


import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'


//const summits = ref(null)

const router = useRouter()
const route = useRoute()

async function loadMarkers(map) {
    const res = await fetch("/api/summits")
    let summits = (await res.json()).summits
    
    summits.forEach(s => {
        let name = s.name === null ? s.height : s.name
        let url = router.resolve({name: 'summit', params: { ridge_id: s.ridge_id, summit_id: s.id}}).href
        let href = `<a href="${url}" class="underline">${name}</a>`
        new mapboxgl.Marker({
            //color: s.color,
            color: '#' + s.color,
            scale: 0.75
        })
        .setLngLat([s.lng, s.lat])
        .setPopup(new mapboxgl.Popup().setHTML(
            `${href}<br>Высота: ${s.height}<br>Хребет: ${s.ridge}`)
            )
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
    //language.setLanguage('mapbox://styles/mapbox/outdoors-v12', 'ru')
    map.addControl(language);
    return map
}

onMounted(function () {
    let map = createMap()
    loadMarkers(map)
})

</script>

<template>
  <div id="map" class="flex-grow"></div>
</template>

<style>
</style>
