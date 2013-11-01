// Returns a function, that, as long as it continues to be invoked, will not
// be triggered. The function will be called after it stops being called for
// N milliseconds. If `immediate` is passed, trigger the function on the
// leading edge, instead of the trailing.
// Copied from underscore.js
var debounce = function(func, wait) {
  var timeout;
  return function() {
    var context = this, args = arguments;
    var later = function() {
      timeout = null;
      func.apply(context, args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

uriSerialize = function(obj) {
  var str = [];
  for(var p in obj){
    if (obj.hasOwnProperty(p)) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
    }
  }
  return str.join("&");
}

postHTTP = function(path, params, onreadystatechange, async) {
  var paramStr = uriSerialize(params);
  var req = new XMLHttpRequest();
  if (onreadystatechange != undefined) {
    req.onreadystatechange = function(){onreadystatechange(req);};
  }
  req.open("POST", path, !!async);
  req.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  req.send(paramStr);
}
