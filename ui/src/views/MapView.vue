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
    const res = await fetch("/api/summits")
    let summits = (await res.json()).summits

    let markersLayer = {
        "type": "FeatureCollection",
        "features": summits.map(s => ({
            "type": "Feature",
            "geometry": {
                "type": "Point",
                "coordinates": [s.lng, s.lat]
            },
            "properties": {
                "icon": "embassy",
                "description": `<a href="${router.resolve({ name: 'summit', params: { ridge_id: s.ridge_id, summit_id: s.id } }).href}" class="underline">${s.name === null ? s.height : s.name}</a><br>Высота: ${s.height}<br>Хребет: ${s.ridge}`
            }
        }))
    };

    map.on('load', function () {
        map.addSource('markers', {
            type: 'geojson',
            data: markersLayer
        });
        map.addLayer({
            id: 'markers',
            type: 'symbol',
            source: 'markers',
            layout: {
                'icon-image': '{icon}',
                'icon-size': 1,
                'icon-allow-overlap': true
            }
        });
    });

    map.on('click', 'markers', function (e) {
        var coordinates = e.features[0].geometry.coordinates.slice();
        var description = e.features[0].properties.description;

        new mapboxgl.Popup()
            .setLngLat(coordinates)
            .setHTML(description)
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
    return map
}

onMounted(() => {
    let map = createMap()
    loadMarkers(map)
})
</script>

<style scoped>
</style>
