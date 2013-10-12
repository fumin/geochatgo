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
  closeBtn.innerHTML = "×";

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

  // Resort to JS since some agents don't support the CSS3 resize property.
  var resizeHandler = document.createElement("div");
  resizeHandler.classList.add("resize-handler");
  box.appendChild(resizeHandler);
  var removeListeners = makeResizable(listContainer,
                                      { handle: resizeHandler,
                                        heightEl: boxContent,
                                        minWidth: 30, minHeight: 40 });

  var closeListener = function(el){
    removeListeners();

    draggie.disable();
    box.parentNode.removeChild(box);
  };
  closeBtn.addEventListener("click", closeListener);
  closeBtn.addEventListener("touchend", closeListener);

  return box;
}

function prepadZero(i) {
  return i < 10 ? "0" + i : i;
}

function formatChatlog(data) {
  var date = new Date(data.created_at*1000);
  var year = date.getFullYear();
  var month = prepadZero(date.getMonth() + 1);
  var day = prepadZero(date.getDate());
  var hours = prepadZero(date.getHours());
  var minutes = prepadZero(date.getMinutes());
  var seconds = prepadZero(date.getSeconds());

  var dt = year + "-" + month + "-" + day + " " + hours + ":" + minutes + ":" + seconds;
  return dt + ": " + linkify(data.msg);
}

function getChatlogs(latLng, box) {
  var selector = ".chat-history > .content > .list-container"
  var listContainer = box.querySelector(selector);

  var zoom = g_map.getZoom();
  var tilePoint = latLngToTileNumber(latLng, zoom);
  var req = new XMLHttpRequest();
  req.onreadystatechange = function() {
    if (req.readyState != 4) { return; }
    if (req.status == 200) {
      var data = JSON.parse(req.responseText);
      for (var i = 0; i != data.length; ++i) {
        var d = document.createElement("div");
        d.innerHTML = formatChatlog(data[i]);
        listContainer.appendChild(d);
      }
    }
  };
  // We should be using GET here, this is just to circumvent openshift's cache.
  req.open("POST", "/chatlogs/"+zoom+"/"+tilePoint.x+"/"+tilePoint.y+".json?limit=200");
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
          div.innerHTML = formatChatlog(data);
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

var markerClickListener = function(mouseEvent){
  var latlng = mouseEvent.latlng;
  var box = createChatlogPopupUI(latlng);
  getChatlogs(latlng, box);
};

$(document).ready(function() {
  g_map.markers.on("clusterclick", markerClickListener);
  g_map.markers.on("click", markerClickListener);
});
