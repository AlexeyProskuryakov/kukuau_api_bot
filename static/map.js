 function initMap() {
  var map = new google.maps.Map(document.getElementById('map'), {
    center: {lat: 0, lng: 0},
    zoom: 3,
    styles: [{
      featureType: 'poi',
      stylers: [{ visibility: 'off' }]  // Turn off points of interest.
    }, {
      featureType: 'transit.station',
      stylers: [{ visibility: 'on' }]
    }],
    disableDoubleClickZoom: true
  });
  map.addListener('click', function(e) {
        var position = {lat: e.latLng.lat(), lng: e.latLng.lng()}
        var marker = new google.maps.Marker({
        position: position,
        map: map
        });
        console.log("map position: ", position);
  });
}

