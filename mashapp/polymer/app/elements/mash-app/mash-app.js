(function() {

var SAMPLE_RATE = 44100;

Polymer({
  is: 'mash-app',
  
  properties: {
    config: {
      type: Object,
      value: null,
    },

    inputLines: {
      type: Array,
      value: [],
    },

    outputLines: {
      type: Array,
      value: [],
    },

    blocks: {
      type: Array,
      value: []
    },

    inputSelection: {
      type: Object,
      value: {
        track: null,
        startSample: null,
        endSample: null,
      },
    },

    outputSelection: {
      type: Object,
      value: {
        track: null,
        startSample: null,
        endSample: null,
      }
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
    this.push('inputLines', [{ sound: data, start: 0 }]);
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
    this.inputLines[0][0].sound = data;
    this.redrawAllLines();
  },

  // Generic services
  getService: function(e) {
    switch (e.detail.service) {
      case "selection": 
        selection = util.clone(this.isOutputMode() ? this.outputSelection : this.inputSelection);
        selection.isOutput = this.isOutputMode();
        e.detail.result = selection;
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
    if (this.isOutputMode()) {
      this.outputSelection = data;
    } else {
      this.inputSelection = data;
    }
    this.redrawAllLines();
  },

  handleFastRewind: function(e) {
    this.handleSetSelection({
      track: null,
      startSample: null,
      endSample: null,
    });
  },

  handleSetZoom: function(zoomLevel) {
    this.zoom = zoomLevel;
    this.indexStep = Math.pow(2, this.zoom);
    this.pixelsPerSample = 1 / this.indexStep;
    this.redrawAllLines();
  },

  handleAddInputTrack: function() {
    util.performAction('load-file', null, this);
  },

  handleAddOutputTrack: function() {
    // TODO: Use action instead?
    this.push('outputLines', []);
  },

  handleLoadFile: function(e) {
    this.$.loadFileDialog.open(this.loadFile.bind(this));
  },

  handleEditTrack: function(data) {
    this.$.editTrackDialog.details = data.track.details[0].sound.meta;
    this.$.editTrackDialog.open(this.editTrack.bind(this));
  },

  handleMuteAllExcept: function(data) {
    this.forEachInputLine(function(line) {
      // NOTE: mute-all-except null is a special case, resuling in nothing muted.
      line.isMuted = (data.track !== null && data.track != line);
    });
  },

  handleSelectBlock: function(e) {
    var blockId = e.target.dataset.blockId;
    if (e.detail.sourceEvent.ctrlKey && this.isOutputMode()) {
      if (this.outputLines.length == 0) {
        // First ensure 
        this.handleAddOutputTrack();
      }
      this.addBlockToOutput(blockId, this.outputSelection.track || 0, this.outputSelection.startSample || 0)
    } else {
      console.log("TODO: Select block " + blockId);
    }
  },

  handleCreateBlock: function(e) {
    if (this.inputSelection && this.inputSelection.startSample && this.inputSelection.endSample && this.inputSelection.track) {
      var trackId = this.inputSelection.track.details[0].sound.meta.id | 0; // HACK - normalize on read, not on write.
      this.$.textDialog.value = "";
      this.$.textDialog.title = "Enter name";
      this.$.textDialog.open(function(name) {
        this.createBlock(trackId, name);
      }.bind(this));
    } else {
      // TODO: Toasty.
      window.alert("Oops, need a selection on a single input track.");
    }
  },

  createBlock: function(trackId, name) {
    if (name === null) {
      return; // Dialog closed.
    }

    var blockDetails = {
      inputId: trackId,
      name: name,
      startSample: this.inputSelection.startSample,
      endSample: this.inputSelection.endSample,
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
  },

  handleCreateBlockResult: function(result) {
    this.push('blocks', result.block);
  },

  addBlockToOutput: function(blockId, track, startSample) {
    var block = this.getBlockById(blockId);
    if (!block) {
      util.whoops("Can't find the block, something really bad has happened...");
    }
    var input = this.getInputById(block.inputId);
    if (!input) {
      util.whoops("Can't find input block refers to...");
    }
    var samples = input.sound.samples.slice(block.startSample, block.endSample);

    // NOTE: Can't be done as a single push, for some reason that doesn't propagate correctly.
    var trackDetails = util.clone(this.outputLines[track.trackIndex]);
    trackDetails.push({
      sound: { samples: samples, },
      start: startSample,
    });
    var key = 'outputLines.' + track.trackIndex;
    this.set(key, trackDetails);
  },

  redrawAllLines: function() {
    // TODO - polymerize.
    this.forEachInputLine(function(line) {
      line.redraw();
    }.bind(this));
  },


  // TODO: Migrate into sound service.
  playSelection: function(frameCallback, endCallback) {
    var startFrame = 0;
    if (this.isOutputMode()) {
      startFrame = this.outputSelection.startSample || 0;
    } else {
      startFrame = this.inputSelection.startSample || 0;
    }
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
    var start, end;
    if (this.isOutputMode()) {
      start = this.outputSelection.startSample || 0;
      end = this.outputSelection.endSample !== null ? this.outputSelection.endSample : this.totalModeSampleLength();
    } else {
      start = this.inputSelection.startSample || 0;
      end = this.inputSelection.endSample !== null ? this.inputSelection.endSample : this.totalModeSampleLength();
    }

    var totalSamples = null;
    this.forEachModeLine(function(line) {
      var lineSamples = line.getSamples(start, end);
      totalSamples = util.mergeSamplesInPlace(totalSamples, lineSamples);
    });
    return totalSamples || [];
  },

  totalModeSampleLength: function() {
    var result = 0;
    this.forEachModeLine(function(line) {
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

  isOutputMode: function() {
    return this.selectedTab == 1;
  },

  forEachModeLine: function(cb) {
    if (this.isOutputMode()) {
      this.forEachOutputLine(cb);
    } else {
      this.forEachInputLine(cb);
    }
  },
  forEachInputLine: function(cb) {
    $("track-line.input").each(function(index, element) { cb(element); });
  },
  forEachOutputLine: function(cb) {
    $("track-line.output").each(function(index, element) { cb(element); });
  },

  getInputById: function(id) {
    for (var i in this.inputLines) {
      if (this.inputLines[i][0].sound.meta.id == id) {
        return this.inputLines[i][0];
      }
    }
    return null;
  },
  getBlockById: function(id) {
    for (var i in this.blocks) {
      if (this.blocks[i].id == id) {
        return this.blocks[i];
      }
    }
    return null;
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
