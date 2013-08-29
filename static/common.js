L.Icon.Default.imagePath = "https://raw.github.com/Leaflet/Leaflet/master/dist/images"

var geoErrorFunction = function(permissionErrorMsg) {
  return function(error) {
    if (1== error.code) { // Permission denied
      alert(permissionErrorMsg);
    } else {
       alert("Oops, location unavailable, please check your browser settings or try again.");
    }
  };
};

function getCurrentPosition(handler, permissionErrorMsg, positionOptions) {
  navigator.geolocation.getCurrentPosition(
    handler,
    geoErrorFunction(permissionErrorMsg),
    jQuery.extend({ timeout: 5000 }, positionOptions)
  );
};
