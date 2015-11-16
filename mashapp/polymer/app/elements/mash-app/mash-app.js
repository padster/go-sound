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

    blocks: {
      type: Array,
      value: []
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

    selectedTab: {
      type: Number,
      value: 0,
    }
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
      // Ignore, this is called when the dialog is closed.
      return;
    }
    this.startRpc("Loading " + path);
    $.ajax({
      url: "/_/input/load",
      type: "POST",
      contentType: "application/json",
      data: JSON.stringify({path: path}),
      dataType: "json",
      success: (function(result) {
        this.endRpc();
        this.handleNewInput(result);
      }).bind(this),
      error: (function(result) {
        this.endRpc();
        console.log("OOPS! %O", result)
      }).bind(this),
    });
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

  editTrack: function(data) {
    if (!data) {
      // Ignore, called when the edit is cancelled.
      return;
    }
    this.startRpc("Modifying track...");
    $.ajax({
      url: "/_/input/edit",
      type: "POST",
      contentType: "application/json",
      data: JSON.stringify({meta: data}),
      dataType: "json",
      success: (function(result) {
        this.endRpc();
        this.handleEditTrackResult(result);
      }).bind(this),
      error: (function(result) {
        this.endRpc();
        console.log("OOPS! %O", result)
      }).bind(this),
    });
  },

  handleEditTrackResult: function(data) {
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
    // HACK
    this.lines[0][0].sound = data;
    this.redrawAllLines();
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
      case "edit-track":
        this.handleEditTrack(e.detail.data);
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
      case "create-block":
        this.handleCreateBlock(e);
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

  handleAddTrack: function() {
    util.performAction('load-file', null, this);
  },

  handleLoadFile: function(e) {
    this.$.loadFileDialog.open(this.loadFile.bind(this));
  },

  handleEditTrack: function(data) {
    this.$.editTrackDialog.details = data.track.details[0].sound.meta;
    this.$.editTrackDialog.open(this.editTrack.bind(this));
  },

  handleMuteAllExcept: function(data) {
    this.forEachLine(function(line) {
      // NOTE: mute-all-except null is a special case, resuling in nothing muted.
      line.isMuted = (data.track !== null && data.track != line);
    });
  },

  handleCreateBlock: function(e) {
    if (this.selection && this.selection.startSample && this.selection.endSample && this.selection.track) {
      var trackDetails = this.selection.track.details[0].sound.meta;
      var name = window.prompt("Block name...");

      var blockDetails = {
        inputId: trackDetails.id | 0, // HACK - normalize on read, not on write.
        name: name,
        startSample: this.selection.startSample,
        endSample: this.selection.endSample,
      }

      this.startRpc("Creating block...");
      $.ajax({
        url: "/_/block/new",
        type: "POST",
        contentType: "application/json",
        data: JSON.stringify({block: blockDetails}),
        dataType: "json",
        success: (function(result) {
          this.endRpc();
          this.handleCreateBlockResult(result);
        }).bind(this),
        error: (function(result) {
          this.endRpc();
          console.log("Oops! %O", result);
        }).bind(this),
      });
    } else {
      // TODO: Toasty.
      window.alert("Oops, need a selection on a single input track.");
    }
  },
  handleCreateBlockResult: function(result) {
    this.push('blocks', result.block);
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

  startRpc: function(message) {
    this.$.toast.text = message;
    this.$.toast.show();
  },
  endRpc: function() {
    this.$.toast.hide();
    this.$.toast.text = "";
  },

});

})();
