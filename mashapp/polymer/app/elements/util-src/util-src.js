// TODO: Not use a global util if possible with the new module stuff?
window.util = {};

(function(M) {

M.getService = function(serviceName, elt) {
 if (!elt) {
    return M.whoops("You forgot to pass 'this' to getService :(");
  }
  var service = elt.fire('get-service', {'service': serviceName}).detail.result;
  if (!service) {
    return M.whoops("Service " + serviceName + " not registered!");
  }
  return service;
}

M.performAction = function(type, data, elt) {
  if (!elt) {
    return M.whoops("You forgot to pass 'this' to perform :(");
  }
  elt.fire('view-action', {'type': type, 'data': data});
}

M.asClassAttribute = function(classSet) {
  var values = [];
  for (var key in classSet) {
    if (classSet[key]) {
      values.push(key);
    }
  }
  return values.join(" ");
};

M.asStyleAttribute = function(styleObject) {
  var values = [];
  for (var key in styleObject) {
    if (styleObject[key] !== null) {
      values.push(key + ':' + styleObject[key]);
    }
  }
  return values.join(";");
};

M.momentInRange = function(toTest, startMoment, endMoment) {
  return !toTest.isBefore(startMoment) && toTest.isBefore(endMoment);
};

M.filter = function(from, properties) {
  result = {};
  properties.forEach(function(key) {
    result[key] = from[key];
  });
  return result;
};

M.intersect = function(s1, e1, s2, e2) {
  return !(e1 <= s2 || e2 <= s1);
};

M.mergeSamplesInPlace = function(s1, s2) {
  if (s2 === null) { return s1; }
  if (s1 === null) { return s2; }
  if (s1.length != s2.length) {
    console.error("Can't merge samples of different lengths...");
    return null;
  }
  for (var i = 0; i < s1.length; i++) {
    s1[i] = (s1[i] + s2[i]) / 2.0;
  }
  return s1;
};

M.dist = function(x1, y1, x2, y2) {
  var dx = x1 - x2, dy = y1 - y2;
  return Math.sqrt(dx * dx + dy * dy);
};

M.deepEquals = function(a, b) {
  if (a === null != b === null) { return false; }
  if (a === null && b === null) { return true; }
  if (a === undefined != b === undefined) { return false; }
  if (a === undefined && b === undefined) { return true; }

  var aArr = (a instanceof Array ), bArr = (b instanceof Array );
  var aObj = (a instanceof Object), bObj = (b instanceof Object);
  if (aArr != bArr || aObj != bObj) { return false; }

  if (aArr) {
    if (a.length != b.length) { return false; }
    for (var i in a) { if (!M.deepEquals(a[i], b[i])) { return false; } }
    return true;
  } else if (aObj) {
    var aLength = 0, bLength = 0;
    for (var i in a) { aLength++; }
    for (var i in b) { bLength++; }
    if (aLength != bLength) { return false; }
    for (var i in a) { if (!M.deepEquals(a[i], b[i])) { return false; } }
    return true;
  } else { // Primitives?
    return a === b;
  }
};

M.clone = function(o) {
  if ($.isArray(o)) {
    return o.slice();
  }
  return $.extend(true, {}, o);
};

M.whoops = function(msg) {
  console.error(msg);
  debugger;
  return null;
};

})(window.util);
