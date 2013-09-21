// Creates a draggable popup and returns the top level div element.
function createChatlogPopupUI() {
  var box = document.createElement("div");
  box.classList.add("chat-history");

  var handler = document.createElement("div");
  handler.classList.add("handler");
  var id = "log-handler" + Math.floor(Math.random() * 1000000);
  handler.id = id;

  var closeBtn = document.createElement("div");
  closeBtn.classList.add("close");
  closeBtn.innerHTML = "X";

  var boxContent = document.createElement("div");
  boxContent.classList.add("content");
  var listContainer = document.createElement("div");
  listContainer.classList.add("list-container");
  boxContent.appendChild(listContainer);

  handler.appendChild(closeBtn);
  box.appendChild(handler);
  box.appendChild(boxContent);

  document.querySelector("#historical-arena").appendChild(box);
  var draggie = new Draggabilly(box, {
    handle: "#" + id
  });

  closeBtn.addEventListener("click", function(el){
    draggie.disable();
    box.remove();
  });

  return box;
}

function getChatlogs(latLng, box) {
  var selector = ".chat-history > .content > .list-container"
  var listContainer = box.querySelector(selector);

  var zoom = g_map.getZoom() + 1; // +1 to increase the granularity of popups.
  var tilePoint = latLngToTileNumber(latLng, zoom);
  var req = new XMLHttpRequest();
  req.onreadystatechange = function() {
    if (req.readyState != 4) { return; }
    if (req.status == 200) {
      var data = JSON.parse(req.responseText);
      for (var i = 0; i != data.length; ++i) {
        var d = document.createElement("div");
        d.innerHTML = linkify(data[i].msg);
        listContainer.appendChild(d);
      }
    }
  };
  req.open("GET", "/chatlogs/"+zoom+"/"+tilePoint.x+"/"+tilePoint.y+".json?limit=200");
  req.send();

  var tileLatLngBounds = tilePointToLatLng(tilePoint, zoom);
  var source = new EventSource("/stream?rtreeType=intersectRTree");
  source.addEventListener("username", function(e){
    var username = e.data;
    updateMapbounds(username, tileLatLngBounds, "intersectRTree");
  }, false);
  source.addEventListener("custom", function(e){
    var data = JSON.parse(e.data);
    var div = document.createElement("div");
    div.innerHTML = linkify(data.msg);
    listContainer.insertBefore(div, listContainer.firstChild);
  }, false);

  var closeBtn = box.querySelector(".chat-history > .handler > .close");
  closeBtn.addEventListener("click", function(el){
    source.close();
  });
}

$(document).ready(function() {
  g_map.markers.on("clusterclick", function(mouseEvent){
    var box = createChatlogPopupUI();
    getChatlogs(mouseEvent.latlng, box);
  });
});
