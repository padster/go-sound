(function() {

// HACK - cleanup.
var CONTROL_WIDTH = 50;

Polymer({
  is: 'track-line',

  properties: {
    // Canvas context
    ctx: Object,

    // TODO - clean up; mash-app uses isOutput
    isOutput: {
      type: Boolean,
      value: false,
    },

    trackIndex: Number,

    details: {
      type: Array,
      observer: 'detailsChanged',
    },

    sampleCount: {
      type: Number,
      value: 0,
    },

    playSampleAt: {
      type: Number,
      observer: 'playSampleAtChanged',
    },
    showPlayLine: Boolean,

    isMuted: {
      type: Boolean,
      observer: 'muteChanged',
    },

    mouseDownE: Object,
    mouseIsDrag: Boolean,
    mouseIsDown: Boolean,
  },

  attached: function() {
    rect = this.$.surface.getBoundingClientRect();
    this.$.surface.width = rect.width;
    this.$.surface.height = rect.height;

    this.ctx = this.$.surface.getContext('2d');
    this.ctx.imageSmoothingEnabled = false;

    this.redraw();
  },

  handleMouseDown: function(e) {
    this.mouseDownE = e;
    this.mouseIsDrag = false;
    this.mouseIsDown = true;
  },

  handleMouseMove: function(e) {
    if (this.mouseIsDown) {
      if (this.isLargeDrag(this.mouseDownE, e)) {
        this.mouseIsDrag = true;
        this.handleCanvasDrag(this.mouseDownE, e);
      }
    }
  },

  handleMouseUp: function() {
    if (!this.mouseIsDrag) {
      this.handleCanvasClick(this.mouseDownE);
    } 
    this.mouseDownE = null;
    this.mouseIsDrag = false;
    this.mouseIsDown = false;
  },

  handleSolo: function(e) {
    var isShift = e.detail.sourceEvent.shiftKey;

    if (isShift) {
      util.performAction('mute-all-except', {track: null}, this);
    } else {
      this.isMuted = true;
    }
  },

  handleMute: function(e) {
    var isShift = e.detail.sourceEvent.shiftKey;
    if (isShift) {
      util.performAction('mute-all-except', {track: this}, this);
    } else {
      this.isMuted = false;
    }
  },

  handleEdit: function(e) {
    util.performAction('edit-track', {track: this}, this);
  },

  detailsChanged: function() {
    this.sampleCount = this.calculateSampleCount();
    this.isMuted = this.calculateMuted();
    this.redraw();
  },

  playSampleAtChanged: function() {
    if (!this.ctx) {
      return;
    }
    this.showPlayLine = (this.playSampleAt != null);

    var pps = util.getService('globals', this).pixelsPerSample;

    if (this.showPlayLine) {
      var scrollElt = document.body;
      var windowWidth = $(window).width();
      var firstVisible = scrollElt.scrollLeft;
      var lastVisible = firstVisible + windowWidth;
      var placement = (this.playSampleAt * pps) + CONTROL_WIDTH;

      this.$.playLine.style.left = placement + 'px';

      if (placement < firstVisible) {
        scrollElt.scrollLeft = placement;
      } else if (placement > lastVisible) {
        scrollElt.scrollLeft = Math.min(lastVisible, this.$.surface.width - windowWidth);
      }  
    }
  },

  muteChanged: function() {
    // Need to denormalize mute to the single input if needed.
    if (!this.isOutput) {
      this.details[0].sound.meta.muted = this.isMuted;
    }
  },

  getSamples: function(start, end) {
    if (this.isMuted) {
      return null;
    }

    var totalSamples = null;
    this.forEachBlock(function(block) {
      var blockSamples = this.getSamplesForBlock(block, start, end);
      totalSamples = util.mergeSamplesInPlace(totalSamples, blockSamples);
    }.bind(this));
    return totalSamples;
  },

  getSamplesForBlock: function(block, start, end) {
    var bS = block.start, bE = block.start + block.sound.samples.length;
    if (!util.intersect(start, end, bS, bE)) {
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
  },

  redraw: function() {
    if (!this.ctx) {
      return;
    }

    this.fixWidth();

    var selection = util.getService('selection', this);
    if (selection.isOutput == this.isOutput && selection.startSample !== null && selection.endSample !== null) {
      this.drawSelectionRange(selection);
    }

    this.forEachBlock(function(block) {
      this.drawSamples(block.sound.samples, block.start);
    }.bind(this));

    if (selection.isOutput == this.isOutput && selection.startSample !== null && selection.endSample === null) {
      this.drawSelectionLine(selection);
    }
  },

  drawSelectionRange: function(selection) {
    if (selection.track !== null && selection.track != this) {
      return;
    }

    var pps = util.getService('globals', this).pixelsPerSample;
    var x1 = pps * selection.startSample;
    var x2 = pps * selection.endSample;

    this.ctx.fillStyle = '#888';
    this.ctx.fillRect(x1, 0, x2 - x1, this.$.surface.height);
  },

  drawSelectionLine: function(selection) {
    if (selection.track !== null && selection.track != this) {
      return;
    }

    var pps = util.getService('globals', this).pixelsPerSample;
    var x = pps * selection.startSample;

    this.ctx.beginPath();
    this.ctx.moveTo(x, 0);
    this.ctx.lineTo(x, this.$.surface.height);
    this.ctx.lineWidth = 2;
    this.ctx.strokeStyle = '#000';
    this.ctx.stroke();
  },

  fixWidth: function() {
    var pps = util.getService('globals', this).pixelsPerSample;
    sampleWidth = (this.sampleCount * pps) | 0;
    width = Math.max(document.body.clientWidth - CONTROL_WIDTH, sampleWidth);
    if (this.$.surface.width != width) {
      this.style.width = (width + CONTROL_WIDTH) + 'px';
      this.$.surface.width = width;
      this.$.surface.style.width = width + "px";
    } else {
      // HACK - need better way to clear canvas
      this.$.surface.width = width;
    }
  },

  drawSamples: function(samples, sampleStart) {
    // HACK - clean up.
    var yLo = 0;
    var yHi = 100;
    var globals = util.getService('globals', this);

    this.ctx.beginPath();

    for (var i = 0; i < samples.length; i += globals.indexStep) {
      x = globals.pixelsPerSample * (sampleStart + i);
      y = yLo + (yHi - yLo) * (1 - 4 * samples[i]) / 2.0; // 1 -> yLo, -1 -> yHi
      if (i == 0) {
        this.ctx.moveTo(x, y);
      } else {
        this.ctx.lineTo(x, y);
      }
    }

    this.ctx.lineWidth = 1;
    this.ctx.strokeStyle = '#00f';
    this.ctx.stroke();
  },

  handleCanvasClick: function(e) {
    var globals = util.getService('globals', this);
    var selection = util.getService('selection', this);

    var x = e.offsetX, y = e.offsetY;
    var sampleX = (x / globals.pixelsPerSample) | 0;

    if (e.shiftKey && selection.startSample !== null) {
      if (selection.endSample === null) {
        this.setSelectedSamples(selection.startSample, sampleX);
      } else {
        if (selection.startSample > sampleX) {
          this.setSelectedSamples(sampleX, selection.endSample);
        } else {
          this.setSelectedSamples(selection.startSample, sampleX);
        }
      }
    } else {
      this.setSelectedSamples(sampleX, null);
    }

    // HACK - redraw all through listening to selection-change event
    this.redraw();
  },

  handleCanvasDrag: function(e1, e2) {
    var pps = util.getService('globals', this).pixelsPerSample;
    var x1 = e1.offsetX, y1 = e1.offsetY;
    var x2 = e2.offsetX, y2 = e2.offsetY;
    var sampleX1 = (x1 / pps) | 0;
    var sampleX2 = (x2 / pps) | 0;
    this.setSelectedSamples(sampleX1, sampleX2);
    this.redraw();
  },

  setSelectedSamples: function(s1, s2) {
    var selection = {
      startSample: null, 
      endSample: null,
      track: this,
    };
    if (s1 === null) {
      selection.startSample = s2;
    } else if (s2 === null) {
      selection.startSample = s1;
    } else {
      selection.startSample = Math.min(s1, s2);
      selection.endSample = Math.max(s1, s2);
    }
    util.performAction('set-selection', selection, this)
  },

  isLargeDrag: function(e1, e2) {
    return util.dist(e1.offsetX, e1.offsetY, e2.offsetX, e2.offsetY) > 5;
  },

  forEachBlock: function(cb) {
    for (var i in this.details) {
      cb(this.details[i]);
    }
  },

  calculateSampleCount: function() {
    var count = 0;
    this.forEachBlock(function(block) {
      count = Math.max(count, block.start + block.sound.samples.length);
    });
    return count;
  },

  // Input/Output distinction:
  calculateMuted: function() {
    if (this.isOutput) {
      return this.isMuted;
    } else {
      return this.details[0].sound.meta.muted;
    }
  },
});

})();
