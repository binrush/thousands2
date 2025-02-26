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
                'icon-image': '{icon}', // Reference the 'icon' property in the GeoJSON
                'icon-size': 1,
                'icon-allow-overlap': true // Optional, to allow icons to overlap
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

    /*summits.forEach(s => {
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
    });*/

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

<style></style>
