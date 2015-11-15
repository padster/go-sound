(function() {

var SAMPLE_RATE = 44100;

Polymer({
  is: 'mash-app',
  
  properties: {
    config: {
      type: Object,
      value: null,
    },

    lines: {
      type: Array,
      value: [],
    },

    selection: {
      type: Object,
      value: {
        startSample: null,
        endSample: null,
      },
    },

    // TODO: migrate to sound service.
    AC: Object,
    currentSource: Object,
    ACStartTime: Number,
    playSampleAt: {
      type: Number,
      value: null,
    },
    isPlaying: {
      type: Boolean,
      value: false,
    },

    // TODO - migrate somewhere else too
    zoom: Number,
    indexStep: Number,
    pixelsPerSample: Number,
  },

  listeners: {
    'get-service': 'getService',
    'view-action': 'handleViewAction',
  },

  attached: function() {
    this.AC = new AudioContext();
    if (this.$.config.innerText != "") {
      this.config = JSON.parse(this.$.config.innerText);
    }
  },

  loadFile: function(path) {
    if (!path) {
      window.alert("Must have a path!");
    } else {
      $.ajax({
        url: "/_/load",
        type: "POST",
        contentType: "application/json",
        data: JSON.stringify({path: path}),
        dataType: "json",
        success: (function(result) {
          this.handleNewInput(result);
        }).bind(this),
        error: (function(result) {
          console.log("OOPS! %O", result)
        }).bind(this),
      });
    }
  },

  handleNewInput: function(data) {
    raw = window.atob(data.samples)
    byteData = Array.prototype.map.call(raw, function(x) { 
      return x.charCodeAt(0); 
    });
    buffer = new ArrayBuffer(byteData.length);
    intBuffer = new Uint8Array(buffer);
    for (var i = 0; i < byteData.length; i++) {
      intBuffer[i] = byteData[i];
    }

    // TODO - move into util?
    // rewrite base-64 to floats
    data.samples = new Float32Array(buffer);
    this.push('lines', [{ sound: data, start: 0 }]);
  },

  // Generic services
  getService: function(e) {
    switch (e.detail.service) {
      case "selection": 
        e.detail.result = this.selection;
        break;
      // HACK - remove once it's split up properly.
      case "globals":
        e.detail.result = this;
        break;
      default:
        return false;
    }
    return true;
  },

  // Generic action framework

  handleViewAction: function(e) {
    switch (e.detail.type) {
      case "play":
        this.handlePlayStop(e);
        break;
      case "load-file":
        this.handleLoadFile(e);
        break;
      case "fast-rewind":
        this.handleFastRewind(e);
        break;
      case "set-selection":
        this.handleSetSelection(e.detail.data);
        break;
      case "set-zoom":
        this.handleSetZoom(e.detail.data);
        break;
      case "mute-all-except":
        this.handleMuteAllExcept(e.detail.data);
        break;
      default:
        util.whoops("View action " + e.detail.type + " not supported :(")
        return false;
    }
    return true;
  },

  handlePlayStop: function(e) {
    // 'Stop' currently showing.
    if (this.isPlaying) {
      this.stopPlaying();
      // R.setPlaying(false);
    } else {
      this.isPlaying = this.playSelection(function(sampleAt) {
        this.playSampleAt = sampleAt;
      }.bind(this), function() {
        this.playSampleAt = null;
        this.isPlaying = false;
      }.bind(this));
    }
  },

  handleSetSelection: function(data) {
    this.selection = data;
    this.redrawAllLines();
  },

  handleFastRewind: function(e) {
    this.handleSetSelection({
      startSample: null,
      endSample: null,
      track: null,
    });
  },

  handleSetZoom: function(zoomLevel) {
    this.zoom = zoomLevel;
    this.indexStep = Math.pow(2, this.zoom);
    this.pixelsPerSample = 1 / this.indexStep;
    this.redrawAllLines();
  },

  handleLoadFile: function(e) {
    this.$.loadFileDialog.open();
  },

  handleCloseLoadFile: function(e) {
    this.$.loadFileDialog.close();
  },

  handleUploadFile: function(e) {
    var path = this.$.filePath.selectedItem.innerText.trim();
    this.loadFile(path);
    this.$.loadFileDialog.close();
  },

  handleMuteAllExcept: function(data) {
    this.forEachLine(function(line) {
      line.isMuted = (line != data.track);
    });
  },

  redrawAllLines: function() {
    // TODO - polymerize.
    this.forEachLine(function(line) {
      line.redraw();
    }.bind(this));
  },


  // TODO: Migrate into sound service.
  playSelection: function(frameCallback, endCallback) {
    var startFrame = this.selection.startSample !== null ? this.selection.startSample : 0;
    var frameCbWrapped = function(frameAt) { return frameCallback(startFrame + frameAt); }
    return this.playSamples(this.getSelectedSamples(), frameCbWrapped, endCallback);
  },
  stopPlaying: function() {
    this.playSampleAt = null;
    this.isPlaying = false;
    this.currentSource.stop();
    this.currentSource = null;
  },
  playSamples: function(samples, frameCallback, endCallback) {
    if (this.isPlaying) {
      return false; // TODO - stop & play new one? Or don't allow playing while already playing?
    }

    this.AC.resume();
    var buffer = this.AC.createBuffer(1, samples.length, SAMPLE_RATE);
    var channel = buffer.getChannelData(0);
    for (var i = 0; i < samples.length; i++) { 
      channel[i] = samples[i];
    }
    this.currentSource = this.AC.createBufferSource();
    this.currentSource.buffer = buffer;
    this.currentSource.connect(this.AC.destination);
    this.currentSource.onended = this.buildEndCallback(endCallback);
    this.ACStartTime = this.AC.currentTime;
    this.currentSource.start();

    window.requestAnimationFrame(this.buildFrameCallback(frameCallback, samples.length));
    return true;
  },

  getSelectedSamples: function() {
    var start = this.selection.startSample !== null ? this.selection.startSample : 0;
    var end = this.selection.endSample !== null ? this.selection.endSample : this.totalSampleLength();

    var totalSamples = null;
    this.forEachLine(function(line) {
      var lineSamples = line.getSamples(start, end);
      totalSamples = util.mergeSamplesInPlace(totalSamples, lineSamples);
    });
    return totalSamples || [];
  },

  totalSampleLength: function() {
    var result = 0;
    this.forEachLine(function(line) {
      result = Math.max(result, line.sampleCount);
    });
    return result;
  },

  buildEndCallback: function(cb) {
    return function() {
      // Chcek if already processed within buildFrameCallback...
      if (this.ACStartTime !== null) {
        this.ACStartTime = null;
        this.AC.suspend();
        cb();
      }
    }.bind(this);
  },

  buildFrameCallback: function(cb, maxSamples) {
    result = function() {
      if (this.isPlaying) {
        offsetTime = this.AC.currentTime - this.ACStartTime;
        offsetSample = (offsetTime * SAMPLE_RATE) | 0;
        if (offsetSample > maxSamples) {
          cb(maxSamples);
          this.currentSource.onended();
        } else {
          // For some reason, end doesn't always fire :/
          cb(offsetSample);
          window.requestAnimationFrame(result);
        }
      }
    }.bind(this);
    return result;
  },

  forEachLine: function(cb) {
    lines = this.getElementsByTagName("track-line");
    for (var i = 0; i < lines.length; i++) {
      cb(lines[i]);
    }
  },

});

})();
