$(document).ready(function() {

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
}); // $("#say_form").submit

}); // $(document).ready
