/**
 * Returns the text specified with URLs replaced by an html link.
 * does not honor whitespace. Escapes HTML in result. Example:
 *
 * linkify("Go to twitter.com/rcanine to read my <i>status</i>.");
 * -> "Go to <a href="http://twitter.com/rcanine">twitter.com/rcanine</a>
 * -> to read my &lt;i&gt;status&lt;/i&gt;."
 *
 * Nandalu change logs:
 * added target="_blank" to hint open a new tab for the external link - 2013-08-18 fumin
 */
var linkify = (function () {
  var GRUBERS_URL_RE = /\b((?:[a-z][\w-]+:(?:\/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}\/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))/i,
      HAS_PROTOCOL = /^[a-z][\w-]+:/;
  
  function escapeHTML(text) {
    return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;').replace(/>/g, '&gt;');
  }
  
  function wordToURL(word, index, array) {
    var m = word.match(GRUBERS_URL_RE),
        result = escapeHTML(word),
        url, escapedURL, text;
    
    if (m) {
      text = escapeHTML(m[1]);
      url = HAS_PROTOCOL.test(text) ? text : 'http://' + text;
      result = result.replace(text, '<a href="' + url + '" target="_blank">' + text + '</a>');
    }
    return result;
  }
  
  var map = Array.prototype.map ? function (arr, callback) {
    return arr.map(callback);
  } : function (arr, callback) {
    var arr2 = [], i, l;
    for (i = 0, l = arr.length; i < l; i = i + 1) {
      arr2[i] = callback(arr[i], i, arr);
    }
    return arr2;
  };
  
  return function (text) {
    return map(text.split(/\s+/), wordToURL).join(' ');
  };
}());
