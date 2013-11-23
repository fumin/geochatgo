$(document).ready(function() {

document.querySelector("#say_form").onsubmit = function(evt){
  evt.preventDefault(); // disable the default form submission action
  if (typeof g_source == "undefined") { return; }
  if (g_source.readyState != EventSource.OPEN) { return; }
  if (typeof g_username == "undefined") { return; }
  if ((typeof g_latitude == "undefined") &&
      (typeof g_longitude == "undefined")) { return; }
      
  var msg = document.querySelector("#msg").value;
  var data      = {
    username:  g_username,
    msg:       msg,
    latitude:  g_latitude,
    longitude: g_longitude,
    skipSelf: true,
  };
  postHTTP(document.querySelector("#say_form").action, data, function(req){
    if (req.readyState != 4) { return; }
    if (req.status != 200) { console.log("HTTP POST error:", req); }
  });
  document.querySelector("#msg").value = "";

  // Animations
  // We add a dummy marker into markers to create an illusion
  // that there's a slot created for the chat we just typed in.
  markerId = Math.floor(Math.random() * 1000000) + "";
  data["markerId"] = markerId;
  data["created_at"] = Date.now() / 1000; // in seconds
  var dummyMarker = g_map.markers.addChat(data);
  var p = g_map.latLngToContainerPoint(new L.LatLng(g_latitude, g_longitude));
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
        g_map.markers.removeLayer(dummyMarker);
        delete data.markerId;
        g_map.markers.addChat(data);

        document.getElementById("msg").placeholder ="Say something to the world?";
        $(this).remove();
    });
}; // document.querySelector("#say_form").onsubmit

}); // $(document).ready
