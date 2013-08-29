var index_page_init = function(){
if (window.location.pathname == "/") {

$("#screen_name").val(localStorage["screen_name"]);
// Submit form when enter is pressed
$('#msg').keydown(function(e){  
//  if (e.keyCode == 13) { $('#say_form').submit(); }
});

$("#say_form").submit(function(){
  if (typeof g_username == "undefined") { return false; }
  // We allow users to continuously send message for now.
  // if ($("#say-btn").prop("disabled")) { return false; }
  if((typeof g_latitude != "undefined") && (typeof g_longitude != "undefined")){
    //console.log(slider.getValue());
    var precision = 5;
    var msg = $("#msg").val();
    if ($("#screen_name").val()) {
      localStorage["screen_name"] = $("#screen_name").val();
      msg = $("#screen_name").val() + ": " + msg;
    }
    var data      = {
      username:  g_username,
      msg:       msg,
      latitude:  g_latitude,
      longitude: g_longitude,
      precision: precision
    };
    $("#say-btn").button("loading");
    jQuery.ajax({
      type: "POST",
      url: $("#say_form").attr("action"),
      data: data,
      success: function(data){
        console.log(data);
        $("#say-btn").button("reset");
      },
      error: function(jqXHR, textStatus, errorThrown){
        console.log(errorThrown);
        $("#say-btn").button("reset");
      }});
    $("#msg").val("");

    // Animations
    // We add a dummy marker into markers to create an illusion
    // that there's a slot created for the chat we just typed in.
    markerId = Math.floor(Math.random() * 1000000) + "";
    data["markerId"] = markerId;
    data["created_at"] = Date.now() / 1000; // in seconds
    var dummyMarker = markers.addChat(data);
    var p = map.latLngToContainerPoint(new L.LatLng(g_latitude, g_longitude));
    document.getElementById("msg").placeholder ="";
    // Do the animation
    $("<div>", {
        class: "messagebox",
           id: "newmsg" 
      }).css({
          top: "0px",
          left: "0px", 
          "text-align": "center",
          fontSize: "24px",
      }).html(data.msg).appendTo("#say_form").animate({
          width: "120px",
          "text-align": "center",
          padding: "0px", 
          fontSize: "14px", 
          /* compensate for div.leaflet-marker-icon (6,15) and for manually center align(60px) */
          left: p.x - 6 - 60, 
          top: p.y - 15 ,

          // left: p.x - window.innerWidth/2 - 6 ,
          // top: p.y - 10 - 16,
        },
        1000,
        function(){ // Animation has completed
          // Create the illusion that the chat we just typed in has
          // fit into the slot by removing the dummyMaker and adding the real
          markers.removeLayer(dummyMarker);
          delete data.markerId;
          markers.addChat(data);

          document.getElementById("msg").placeholder ="Say something to the world?";
          $(this).remove();
      });
  }
  return false; // disable the default form submission action
});

// Create map
var map = L.map("map");
L.tileLayer('http://{s}.tile.cloudmade.com/BC9A493B41014CAABB98F0471D759707/997/256/{z}/{x}/{y}.png', {
  maxZoom: 18,
  attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="http://cloudmade.com">CloudMade</a>',
}).addTo(map);
L.control.scale().addTo(map);
// Initialize the broadcast circle with an arbitrary position.
// We'll update it later when we have location information.
var broadcastCircle = L.circle([45, 45], 0).addTo(map);

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
map.addLayer(markers);

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
}).addTo(map);
map.on('viewreset', function(e) { // called upon zoom level change
  // Since the data we want to present is probably different at different
  // zoom levels, clear all data belonging to the previous zoom level.
  if (map.getZoom() <= chatlogLayer.options.maxZoom) {
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
    map.setView([g_latitude, g_longitude], 13);

    // Prepare broadcast radius slider
    broadcastCircle.setLatLng(new L.LatLng(g_latitude, g_longitude));
    //broadcastCircle.setRadius(slider.getValueInMeters());
    //$("#slider").slider().on("slide", function(ev){
    //  broadcastCircle.setRadius(slider.getValueInMeters());
    //});

    // Create streaming event source
    var path = "stream?latitude=" + g_latitude + "&longitude=" + g_longitude;
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
  if (typeof g_username == "undefined") { return; }
  jQuery.post("report_location",
              { username: g_username,
                latitude: g_latitude, longitude: g_longitude },
              function(data){
    console.log(data);
  });
  
}); // navigator.geolocation.watchPosition

} // if (window.location.pathname == "/") {
}; // var index_page_init = function(){


// jQuery(document).on("page:load", index_page_init); // Turbolinks
jQuery(document).ready(index_page_init);
