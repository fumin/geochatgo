$(document).ready(function() {

g_map = L.map("map");
L.tileLayer('http://{s}.tile.cloudmade.com/BC9A493B41014CAABB98F0471D759707/997/256/{z}/{x}/{y}.png', {
  maxZoom: 18,
  attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
}).addTo(g_map);
L.control.scale().addTo(g_map);
// Initialize the broadcast circle with an arbitrary position.
// We'll update it later when we have location information.
var broadcastCircle = L.circle([45, 45], 0).addTo(g_map);

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
    var content = "";
    for (var i = 0; i != ms.length; ++i) {
      // if markerId exists, it's a dummy chatlog created by our javascript
      // as a means to facilitate animation
      var htmlContent = "&nbsp;";
      if (ms[i].chatData.markerId) {
        content += ("<div id=" + ms[i].chatData.markerId + " class='msg-history ");
      } else {
        htmlContent = linkify(ms[i].chatData.msg);
        content += "<div class='msg-history ";
      }

      if (i == 0) {
        content += "newchat"
      }
      content += "'>"

      content += (htmlContent + "</div>");
    }
    return new L.DivIcon({
      className: "",
      html: (new L.Icon.Default()).createIcon().outerHTML + content
    })
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

// Create chatlogs layer
var chatlogLayer = L.ajaxTileLayer("/chatlogs/{z}/{x}/{y}.json", {
  maxZoom: 15,
  success: function(data_unparsed) {
    var data = JSON.parse(data_unparsed);
    var len = data.length;
    for (var i = 0; i != len; ++i) {
      markers.addChat(data[i]);
    }
  },
}).addTo(g_map);
g_map.on('viewreset', function(e) { // called upon zoom level change
  // Since the data we want to present is probably different at different
  // zoom levels, clear all data belonging to the previous zoom level.
  if (g_map.getZoom() <= chatlogLayer.options.maxZoom) {
    markers.clearLayers();
  }
});

// slider results circle here
// var slider = $('#slider').slider({
//               step: 0.1, tooltip: "hide", value: 0 }).data("slider");
// Utility method that gives the value in meters assuming the full stretch
// of the slider equals half the bounds of the map.
//slider["getValueInMeters"] = function() {
//  var bounds = map.getBounds(),
//      centerLat = bounds.getCenter().lat,
//      halfWorldMeters = 6378137 * Math.PI * Math.cos(centerLat * Math.PI / 180),
//      dist = halfWorldMeters * (bounds.getNorthEast().lng - bounds.getSouthWest().lng) / 180 / 2;
//  return dist / slider.max * slider.getValue();
//};

navigator.geolocation.watchPosition(function(position){
  g_latitude = position.coords.latitude;
  g_longitude = position.coords.longitude;

  if (typeof g_source == "undefined") {
    // Initialize map
    g_map.setView([g_latitude, g_longitude], 13);

    // Prepare broadcast radius slider
    broadcastCircle.setLatLng(new L.LatLng(g_latitude, g_longitude));
    //broadcastCircle.setRadius(slider.getValueInMeters());
    //$("#slider").slider().on("slide", function(ev){
    //  broadcastCircle.setRadius(slider.getValueInMeters());
    //});

    // Create streaming event source
    mapBounds = g_map.getBounds(); // in latitude and longitude
    var path = "stream?west=" + mapBounds.getWest() +
                 "&south=" + mapBounds.getSouth() +
                 "&east="  + mapBounds.getEast() +
                 "&north=" + mapBounds.getNorth();
    g_source = new EventSource(path);
    g_source.addEventListener("username", function(e){
      g_username = e.data;
    }, false);
    g_source.addEventListener("custom", function(e){
      var data = JSON.parse(e.data);
      console.log(data);
      markers.addChat(data);
    }, false);
  }

  // Routine work on every location change: update circle and report location
  broadcastCircle.setLatLng(new L.LatLng(g_latitude, g_longitude));
  
}); // navigator.geolocation.watchPosition

g_map.on('moveend', debounce(function(e) {
  if (typeof g_username == "undefined") { return; }
  mapBounds = g_map.getBounds(); // in latitude and longitude
  jQuery.post("update_mapbounds",
              { username: g_username,
                west: mapBounds.getWest(), south: mapBounds.getSouth(),
                east: mapBounds.getEast(), north: mapBounds.getNorth(), },
              function(data){ console.log("update_mapbounds: " + data); });
}, 5000));

}); // $(document).ready
