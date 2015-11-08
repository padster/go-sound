window.viewmodel = {};

(function(VM) {

VM.lines = [];
VM.cachedLength = null;
VM.startSample = null;
VM.endSample = null;

VM.totalSampleLength = function() {
  if (VM.cachedLength == null) {
    VM.cachedLength = 0;
    for (var i in VM.lines) {
      for (var j in VM.lines[i]) {
        var block = VM.lines[i][j];
        VM.cachedLength = Math.max(VM.cachedLength, block.start + block.sound.samples.length);
      }
    }
  }
  return VM.cachedLength;
};

VM.addLine = function(sound, xStart) {
  VM.lines.push([{ sound: sound, start: xStart }])
  VM.cachedLength = null;
};


VM.getSelectedSamples = function() {
  var start = VM.startSample !== null ? VM.startSample : 0;
  var end = VM.endSample !== null ? VM.endSample : VM.totalSampleLength();

  var totalSamples = null;
  for (var i in VM.lines) {
    var lineSamples = getSamplesForLine(i, start, end);
    totalSamples = mergeSamplesInPlace(totalSamples, lineSamples);
  }
  return totalSamples || [];
};


var mergeSamplesInPlace = function(s1, s2) {
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

var getSamplesForLine = function(i, start, end) {
  var totalSamples = null;
  for (var j = 0; j < VM.lines[i].length; j++) {
    var blockSamples = getSamplesForBlock(i, j, start, end);
    totalSamples = mergeSamplesInPlace(totalSamples, blockSamples);
  }
  return totalSamples;
};

var getSamplesForBlock = function(i, j, start, end) {
  var block = VM.lines[i][j];
  var bS = block.start, bE = block.start + block.sound.samples.length;
  if (!intersect(start, end, bS, bE)) {
    return null;
  }

  var result = [];
  for (var at = start; at < end; at++) {
    var s = 0.0;
    if (at >= bS && at < bE) {
      s = block.sound.samples[at - bS];
    }
    result.push(s);
  }
  return result;
};


var intersect = function(s1, e1, s2, e2) {
  return !(e1 <= s2 || e2 <= s1);
};

})(window.viewmodel);
