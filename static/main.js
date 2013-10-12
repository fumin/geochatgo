function updateMapbounds(username, mapBounds) {
  postHTTP("update_mapbounds",
           { username: username,
             west: mapBounds.getWest(), south: mapBounds.getSouth(),
             east: mapBounds.getEast(), north: mapBounds.getNorth(), },
           function(data){ console.log("update_mapbounds: " + data); });
}

$(document).ready(function() {

g_map = L.map("map");
L.tileLayer('http://{s}.tile.cloudmade.com/BC9A493B41014CAABB98F0471D759707/1714/256/{z}/{x}/{y}.png', {
  maxZoom: 18,
  attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>'
}).addTo(g_map);
g_map.setView([25.041846, 121.539001], 13); // Initialize map to Taipei
L.control.scale().addTo(g_map);
L.control.locate({position: "bottomright"}).addTo(g_map);

// markers is the layer that displays the chatlogs on the map.
// To avoid cluttering the map with too many chats, we display only the latest
// five chats in each cluster.
var markers = new L.MarkerClusterGroup({
  zoomToBoundsOnClick: false,
  spiderfyOnMaxZoom:   false,
  singleMarkerMode:    true,
  iconCreateFunction: function(cluster) {
    var ms = cluster.getAllChildMarkers().sort(function(a, b) {
      if (b.chatData.created_at && a.chatData.created_at) {
        return b.chatData.created_at - a.chatData.created_at;
      } else if (b.chatData.created_at && !a.chatData.created_at) {
        return 1;
      } else { return -1; }
    }).slice(0, 5);

    var box = document.createElement("div");

    // For UI creation of the Taipei 101
    var roof = document.createElement("div");
    roof.classList.add("roof");
    box.appendChild(roof);

    for (var i = 0; i != ms.length; ++i) {
      var chat = document.createElement("div");
      chat.classList.add("msg-history");

      if (i == 0) { chat.classList.add("newchat"); }
      if (i == ms.length-1) { chat.classList.add("oldest-chat"); }

      if (ms[i].chatData.markerId) {
        chat.id = ms[i].chatData.markerId;
        chat.innerHTML = "&nbsp;";
      } else {
        chat.innerHTML = linkify(ms[i].chatData.msg);
      }
      box.appendChild(chat);
    }

    // For UI creation of the chat triangle.
    var tail = document.createElement("div");
    tail.classList.add("chat-tail");
    box.appendChild(tail);

    return new L.DivIcon({
      className: "",
      // html: (new L.Icon.Default()).createIcon().outerHTML + box.outerHTML,
      html: box.outerHTML,
    });
  },
});
g_map.addLayer(markers);
g_map["markers"] = markers;

// `addChat` is the helper that we should ALWAYS use to add a marker
markers["addChat"] = function(datum) {
  var marker = L.marker([datum.latitude, datum.longitude]);
  marker["chatData"] = datum;
  markers.addLayer(marker);
  return marker;
};

// chatlogLayer isn't for display but for retrieving historical chatlogs for
// each tile from the server by reusing code from Leaflet's TileLayer.
var chatlogLayer = L.ajaxTileLayer("/chatlogs/{z}/{x}/{y}.json", {
  maxZoom: 18,
  success: function(data_unparsed) {
    var data = JSON.parse(data_unparsed);
    var len = data.length;
    for (var i = 0; i != len; ++i) {
      markers.addChat(data[i]);
    }
  },
  httpMethod: "POST", // Circumvent the bastard openshift cache...
}).addTo(g_map);
g_map.on('viewreset', function(e) { // called upon zoom level change
  // Since the data we want to present is probably different at different
  // zoom levels, clear all data belonging to the previous zoom level.
  if (g_map.getZoom() <= chatlogLayer.options.maxZoom) {
    markers.clearLayers();
  }
});

var currentLocationIcon = L.divIcon({
  className: 'current-location-icon',
  html: "●",
});
var currentLocationMarker = L.marker([25.041846, 121.539001], {icon: currentLocationIcon}).addTo(g_map);

function geo_success(position){
  g_latitude = position.coords.latitude;
  g_longitude = position.coords.longitude;

  if (typeof g_source == "undefined") {
    g_map.setView([g_latitude, g_longitude], 13); // Initialize map

    g_source = new EventSource("/stream");
    g_source.addEventListener("username", function(e){
      g_username = e.data;
      mapBounds = g_map.getBounds(); // in latitude and longitude
      updateMapbounds(g_username, mapBounds);
    }, false);
    g_source.addEventListener("custom", function(e){
      var data = JSON.parse(e.data);
      console.log(data);
      markers.addChat(data);
    }, false);
  }

  currentLocationMarker.setLatLng(new L.LatLng(g_latitude, g_longitude));
}

function geo_error(err) {
}

var geo_options = {
  enableHighAccuracy: true,
};

navigator.geolocation.getCurrentPosition(function(position) {
    document.querySelector("#msg").placeholder = "Say something to the world?";
    geo_success(position);
    navigator.geolocation.watchPosition(geo_success, geo_error, geo_options);
  },function(err) {
    var errMsg = "Error(" + err.code + '): ' + err.message;
    document.querySelector("#msg").value = errMsg;
    document.querySelector("#location-instructions> .title").innerHTML = errMsg;
    document.querySelector("#location-instructions").style.display = "";
  }
);

g_map.on('moveend', debounce(function(e) {
  if (typeof g_username == "undefined") { return; }
  mapBounds = g_map.getBounds(); // in latitude and longitude
  updateMapbounds(g_username, mapBounds);
}, 5000));

}); // $(document).ready
