L.Icon.Default.imagePath = "https://raw.github.com/Leaflet/Leaflet/master/dist/images"

// Calculates the tile x, y numbers of a LatLng.
// The result could then be passed to a tile provider whose URL may look
// something like "/#{map.getZoom()}/x/y.png".
//
// Note, in leaflet-0.6.4, this is almost equivalent to L.CRS.latLngToPoint
// except without the magic number 256.
// Actually, I suspect the existence of this magic number implies that the
// tileSize of our maps' layers MUST always be 256.
function latLngToTileNumber(latLng, zoom) {
  var projectedPoint = L.CRS.EPSG3857.projection.project(latLng),
      scale = Math.pow(2, zoom);

  return L.CRS.EPSG3857.transformation._transform(projectedPoint, scale).floor();
}

function tileNumberToTopLeftLatLng(tilePoint, zoom) {
  var scale = Math.pow(2, zoom),
      untransformedPoint = L.CRS.EPSG3857.transformation.untransform(tilePoint, scale);

  return L.CRS.EPSG3857.projection.unproject(untransformedPoint);
}

function tilePointToLatLng(tilePoint, zoom) {
  var southWestP = new L.Point(tilePoint.x, tilePoint.y + 1);
  var southWest = tileNumberToTopLeftLatLng(southWestP, zoom);

  var northEastP = new L.Point(tilePoint.x + 1, tilePoint.y);
  var northEast = tileNumberToTopLeftLatLng(northEastP, zoom);

  return new L.LatLngBounds(southWest, northEast);
}

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
