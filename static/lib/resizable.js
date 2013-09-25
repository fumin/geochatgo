// To make an element resizable:
// var removeListeners = makeResizable(el, {handle: resizeHandler});
//
// When we're done, simply run:
// removeListeners();
// Note that in the case where we want to remove either the `el` or the
// `resizeHandler`, we should also call `removeListeners()` since it's
// considered good practice to remove event listeners to avoid memory leaks.
//
// Note also we may also need to make el's style be "position: absolute;".
//
// Options:
// * handle: the element that will trigger the resize.
// * minWidth: the minimum width that will be resized.
// * minHeight: the minimum height that will be resize.
function makeResizable(el, args) {
  var m = makeResizableMouse(el, args);
  var t = makeResizableTouch(el, args);
  var removeListeners = function() {
    m.handle.removeEventListener("mousedown", m.mouseDownListener);
    t.handle.removeEventListener("touchstart", t.touchStartListener);
  };
  return removeListeners;
}

function makeResizableMouse(el, args) {
  var widthEl = args.widthEl || el;
  var heightEl = args.heightEl || el;
  var handle = args.handle || el;
  var minWidth = args.minWidth || 0;
  var minHeight = args.minHeight || 0;

  var lastMouseX = -1, lastMouseY = -1;
  var mouseMoveListener = function(e) {
    var diffX = e.pageX - lastMouseX,
        diffY = e.pageY - lastMouseY,
        currentWidth = parseInt(widthEl.clientWidth, 10),
        currentHeight = parseInt(heightEl.clientHeight, 10);
    widthEl.style.width = Math.max(currentWidth + diffX, minWidth) + "px";
    heightEl.style.height = Math.max(currentHeight + diffY, minHeight) + "px";

    lastMouseX = e.pageX; lastMouseY = e.pageY;
  };

  var mouseUpListener = function(e) {
    document.documentElement.removeEventListener("mousemove", mouseMoveListener);
    document.documentElement.removeEventListener("mouseup", mouseUpListener);
  }

  var mouseDownListener = function(e) {
    lastMouseX = e.pageX; lastMouseY = e.pageY;
    document.documentElement.addEventListener("mousemove", mouseMoveListener);
    document.documentElement.addEventListener("mouseup", mouseUpListener);
    // Disable the default text selection behaviour on unexpected elements.
    e.preventDefault();
  }
  handle.addEventListener("mousedown", mouseDownListener);

  return {handle: handle, mouseDownListener: mouseDownListener};
}

function makeResizableTouch(el, args) {
  var widthEl = args.widthEl || el;
  var heightEl = args.heightEl || el;
  var handle = args.handle || el;
  var minWidth = args.minWidth || 0;
  var minHeight = args.minHeight || 0;

  var lastTouchX = -1, lastTouchY = -1, fingerId = null;
  var touchMoveListener = function(e) {
    var touch = null;
    for (var i = 0; i != e.touches.length; i++) {
      if (e.touches[i].identifier == fingerId) {
        touch = e.touches[i];
        break;
      }
    }
    if (touch == null) { return; }
    var diffX = touch.pageX - lastTouchX,
        diffY = touch.pageY - lastTouchY,
        currentWidth = parseInt(widthEl.clientWidth, 10),
        currentHeight = parseInt(heightEl.clientHeight, 10);
    widthEl.style.width = Math.max(currentWidth + diffX, minWidth) + "px";
    heightEl.style.height = Math.max(currentHeight + diffY, minHeight) + "px";

    lastTouchX = e.pageX; lastTouchY = e.pageY;
  };

  var touchEndListener = function(e) {
    document.documentElement.removeEventListener("touchmove", touchMoveListener);
    document.documentElement.removeEventListener("touchend", touchEndListener);
  }

  var touchStartListener = function(e) {
    var touch = e.targetTouches[0];
    fingerId = touch.identifier;
    lastTouchX = touch.pageX; lastTouchY = touch.pageY;
    document.documentElement.addEventListener("touchmove", touchMoveListener);
    document.documentElement.addEventListener("touchend", touchEndListener);
    // Disable the default text selection behaviour on unexpected elements.
    e.preventDefault();
  }
  handle.addEventListener("touchstart", touchStartListener);

  return {handle: handle, touchStartListener: touchStartListener};
}
