L.Control.Locate = L.Control.extend({
  options: {
    position: "topleft",
    btnHTML: "<img src='static/img/locate.png'>",
    btnTitle: "Go to my location",
  },

  onAdd: function (map) {
    this._map = map;
    var className = 'leaflet-control-locate',
        container = L.DomUtil.create('div', className);

    this._locateButton = this._createButton(
      this.options.btnHTML,
      this.options.btnTitle,
      className + "-btn",
      container,
      this._locate,
      this);

    return container;
  },

  _locate: function(e) {
    console.log(e);
    this._map.setView([g_latitude, g_longitude], this._map.getZoom());
  },

  _createButton: L.Control.Zoom.prototype._createButton,
});

L.control.locate = function(options) {
  return new L.Control.Locate(options);
};
