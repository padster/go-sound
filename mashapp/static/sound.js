window.sound = {};

(function(W, S, VM) {

// Audio context
var AC = new AudioContext();
var SAMPLE_RATE = 44100;
var currentSource = null;

var startTime = null;

S.playSelection = function(frameCallback, endCallback) {
  startFrame = VM.startSample !== null ? VM.startSample : 0;
  frameCbWrapped = function(frameAt) { return frameCallback(startFrame + frameAt); }
  return playSamples(VM.getSelectedSamples(), frameCbWrapped, endCallback);
};

S.stop = function() {
  currentSource.stop();
};

S.isPlaying = function() {
  return startTime !== null;
};

var playSamples = function(samples, frameCallback, endCallback) {
  if (S.isPlaying()) {
    return false; // TODO - stop & play new one? Or don't allow playing while already playing?
  }

  AC.resume();
  var buffer = AC.createBuffer(1, samples.length, SAMPLE_RATE);
  var channel = buffer.getChannelData(0);
  for (var i = 0; i < samples.length; i++) { 
    channel[i] = samples[i];
  }
  currentSource = AC.createBufferSource();
  currentSource.buffer = buffer;
  currentSource.connect(AC.destination);
  currentSource.onended = buildEndCallback(endCallback);
  startTime = AC.currentTime;
  currentSource.start();

  W.requestAnimationFrame(buildFrameCallback(frameCallback, samples.length));
  return true;
};

var buildEndCallback = function(cb) {
  return function() {
    // Chcek if already processed within buildFrameCallback...
    if (startTime !== null) {
      startTime = null;
      AC.suspend();
      cb();
    }
  };
};

var buildFrameCallback = function(cb, maxSamples, endCb) {
  result = function() {
    if (S.isPlaying()) {
      offsetTime = AC.currentTime - startTime;
      offsetSample = (offsetTime * SAMPLE_RATE) | 0;
      if (offsetSample > maxSamples) {
        cb(maxSamples);
        currentSource.onended();
      } else {
        // For some reason, end doesn't always fire :/
        cb(offsetSample);
        W.requestAnimationFrame(result);
      }
    }
  };
  return result;
};

})(window, window.sound, window.viewmodel);
