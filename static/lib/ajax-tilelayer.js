// A Leaflet extension on L.TileLayer that makes Ajax requests
// instead of loading tile images.
//
// Usage:
// -----------------------------------------------------------
// var map = L.map("map");
// var markers = new L.MarkerClusterGroup();
// map.addLayer(markers);
//
// L.ajaxTileLayer("/chatlogs/{z}/{x}/{y}.json",
//   maxZoom: 15,
//   success: function(data_unparsed) {
//     var data = JSON.parse(data_unparsed);
//     var len = data.length;
//     for (var i = 0; i != len; ++i) {
//       var datum = data[i];
//       var marker = L.marker([datum.latitude, datum.longitude])
//                     .bindPopup(datum.msg);
//       markers.addLayer(marker);
//     }
//   }
// }).addTo(map);
//
// map.on('viewreset', function(e){ 
//   // Since the data we want to present is probably different at different
//   // zoom levels, clear all data belonging to the previous zoom level.
//   if (map.getZoom() <= chatlogLayer.options.maxZoom) {
//     markers.clearLayers();
//   }
// });

L.TileLayer.Ajax = L.TileLayer.extend({
  _requests: {},
  options: {
    unloadInvisibleTiles: false, // since we don't really unload ajax data yet
    success: function(){}
  },

  initialize: function (url, options) {
    L.TileLayer.prototype.initialize.call(this, url, options);
    L.setOptions(this, options);
  },

  _addTile: function (tilePoint, container) {
    var tile = {};
    // The instance variable `_tiles` is used in `_tileShouldBeLoaded`
    // to not load tiles twice. We adopt this protocol here.
    this._tiles[tilePoint.x + ':' + tilePoint.y] = tile;

    this._loadTile(tile, tilePoint);
  },

  _loadTile: function (tile, tilePoint) {
    var req = new XMLHttpRequest();
    this._requests[tilePoint.x + ':' + tilePoint.y] = req;
    var that = this;
    req.onreadystatechange = function() {
      if (req.readyState != 4) { return; }
      // Call _tileLoaded regardless of whether the request succeeded or not
      // according to leaflet's protocol.
      that._tileLoaded();
      delete that._requests[tilePoint.x + ':' + tilePoint.y];

      if (req.status == 200) {
        that.options.success(req.responseText);
      }
    };
    this._adjustTilePoint(tilePoint); // Sets the z index
    req.open('GET', this.getTileUrl(tilePoint), true);
    req.send();
  },

  _reset: function() {
    L.TileLayer.prototype._reset.apply(this, arguments);

    // Abort requests for the current layer that's going to be replaced.
    for (var k in this._requests) {
      this._requests[k].abort();
    }
    this._requests = {};
  },
  
});

L.ajaxTileLayer = function(url, options) {
  return new L.TileLayer.Ajax(url, options);
};
