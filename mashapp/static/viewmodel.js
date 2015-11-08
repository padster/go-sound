window.viewmodel = {};

(function(VM) {

VM.lines = [];

VM.startSample = null;
VM.endSample = null;

VM.addLine = function(sound, xStart) {
  VM.lines.push([{ sound: sound, start: xStart }])
};

VM.getSelectedSamples = function() {
  if (VM.startSample === null || VM.endSample === null) {
    // TODO: Default to startSample = 0, endSample = length.
    return [];
  }

  var totalSamples = null;
  for (var i in VM.lines) {
    var lineSamples = getSelectedSamplesForLine(i);
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

var getSelectedSamplesForLine = function(i) {
  var totalSamples = null;
  for (var j = 0; j < VM.lines[i].length; j++) {
    var blockSamples = getSelectedSamplesForBlock(i, j);
    totalSamples = mergeSamplesInPlace(totalSamples, blockSamples);
  }
  return totalSamples;
};

var getSelectedSamplesForBlock = function(i, j) {
  var block = VM.lines[i][j];
  var bS = block.start, bE = block.start + block.sound.samples.length;
  var sS = VM.startSample, sE = VM.endSample;
  if (!intersect(sS, sE, bS, bE)) {
    return null;
  }

  var result = [];
  for (var at = sS; at < sE; at++) {
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
