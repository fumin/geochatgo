function adjustPopupPosition(latLng, box) {
  box.style.position = "absolute";
  var p = g_map.latLngToContainerPoint(latLng);
  var px = p.x + Math.floor(Math.random() * 34) - 17;
  var py = p.y + Math.floor(Math.random() * 34) - 17;
  var x = (px > 100 ? px - 100 : px + 100);
  var y = (py > 30 ? py - 30 : Math.max(py, 20));
  box.style.left = x + "px";
  box.style.top = y + "px";
  console.log(box.style);
}

g_maxPopupZindex = 0;

// Creates a draggable popup and returns the top level div element.
function createChatlogPopupUI(latLng) {
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

  adjustPopupPosition(latLng, box);

  var draggie = new Draggabilly(box, {
    handle: "#" + id
  });
  draggie.on("dragStart", function(draggieInstance, event, pointer){
    var zIndexStr = box.style.zIndex;
    var zIndex = zIndexStr == "" ? g_maxPopupZindex : parseInt(zIndexStr)
    var newMaxPopupZindex = Math.max(zIndex, g_maxPopupZindex+1);
    box.style.zIndex = newMaxPopupZindex + "";
    g_maxPopupZindex = newMaxPopupZindex;
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
  postHTTP("open_popup", {
      username: g_username,
      south: tileLatLngBounds.getSouth(),
      west:  tileLatLngBounds.getWest(),
      north: tileLatLngBounds.getNorth(),
      east:  tileLatLngBounds.getEast(),
    },
    function(req) {
      if (req.readyState != 4) { return; }
      if (req.status == 200) {
        var data = JSON.parse(req.responseText);
        var popupId = data["popupId"];
        console.log("popupId = " + popupId);

        var listener = function(e) {
          var data = JSON.parse(e.data);
          var div = document.createElement("div");
          div.innerHTML = linkify(data.msg);
          listContainer.insertBefore(div, listContainer.firstChild);
        };
        g_source.addEventListener(popupId, listener, false);

        var closeBtn = box.querySelector(".chat-history > .handler > .close");
        closeBtn.addEventListener("click", function(el){
          g_source.removeEventListener(popupId, listener);
          postHTTP("close_popup", {username: g_username, popupId: popupId});
        });
      }
    });
}

$(document).ready(function() {
  g_map.markers.on("clusterclick", function(mouseEvent){
    var latlng = mouseEvent.latlng;
    var box = createChatlogPopupUI(latlng);
    getChatlogs(latlng, box);
  });
});
