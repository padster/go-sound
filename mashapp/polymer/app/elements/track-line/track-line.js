(function() {

// HACK - cleanup.
var CONTROL_WIDTH = 50;

Polymer({
  is: 'track-line',

  properties: {
    // Canvas context
    ctx: Object,

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

  detailsChanged: function() {
    this.sampleCount = 0;
    for (var i in this.details) {
      this.sampleCount = Math.max(this.sampleCount, this.details[i].sound.samples.length)
    }
    this.redraw();
  },

  playSampleAtChanged: function() {
    if (!this.ctx) {
      console.log("Skipping, not ready");
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

  getSamples: function(start, end) {
    if (this.details.length != 1) {
      util.whoops("Only support one sound per track so far...");
    }

    var totalSamples = null;
    for (var i = 0; i < this.details.length; i++) {
      var blockSamples = this.getSamplesForBlock(this.details[i], start, end);
      totalSamples = util.mergeSamplesInPlace(totalSamples, blockSamples);
    }
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
    if (selection.startSample !== null && selection.endSample !== null) {
      this.drawSelectionRange(selection);
    }

    if (this.details.length != 1) {
      util.whoops("Only support one sound per track so far...");
    }
    this.drawSamples(this.details[0].sound.samples, this.details[0].start);

    if (selection.startSample !== null && selection.endSample === null) {
      this.drawSelectionLine(selection);
    }
  },

  drawSelectionRange: function(selection) {
    var pps = util.getService('globals', this).pixelsPerSample;
    var x1 = pps * selection.startSample;
    var x2 = pps * selection.endSample;

    this.ctx.fillStyle = '#888';
    this.ctx.fillRect(x1, 0, x2 - x1, this.$.surface.height);
  },

  drawSelectionLine: function(selection) {
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
    // debugger;
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

    var x = e.offsetX, y = e.offsetY;
    var sampleX = (x / globals.pixelsPerSample) | 0;
    // TODO - track IDs?
    console.log("Clicked on sample %d in track %O", sampleX, this);

    var selection = util.getService('selection', this);

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
    var selection = {startSample: null, endSample: null};
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
});

})();
