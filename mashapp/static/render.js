window.render = {};

(function(R, VM) {

// Visual context
var C = document.getElementById('surface');
var W = C.getBoundingClientRect().width;
var H = C.getBoundingClientRect().height;
C.width = W; C.height = H;
var CTX = C.getContext('2d');
CTX.imageSmoothingEnabled = false;

// Visual controls
var zoomSlider = document.getElementById('zoomSlider');
var zoomValue = document.getElementById('zoomValue');
var INDEX_STEP = 1;

var ROW_HEIGHT = 120;
var ROW_GAP = 30;

var PIXELS_PER_SAMPLE = 1 / INDEX_STEP;

R.drawRows = function() {
  fixRowHeight(VM.lines.length);
  fixWidth();

  if (VM.startSample !== null && VM.endSample !== null) {
    drawSelectionRange();
  }

  for (var i = 0; i < VM.lines.length; i++) {
    drawRow(i, VM.lines[i]);
  }

  if (VM.startSample !== null && VM.endSample === null) {
    drawSelectionLine();
  }
};

var drawSelectionRange = function() {
  x1 = PIXELS_PER_SAMPLE * VM.startSample;
  x2 = PIXELS_PER_SAMPLE * VM.endSample;

  CTX.fillStyle = '#888';
  CTX.fillRect(x1, 0, x2 - x1, H);
  console.log("(%d, %d), size (%d, %d)", x1, 0, x2 - x1, H)
};

var drawSelectionLine = function() {
  x = PIXELS_PER_SAMPLE * VM.startSample;

  CTX.beginPath();
  CTX.moveTo(x, 0);
  CTX.lineTo(x, H);
  CTX.lineWidth = 2;
  CTX.strokeStyle = '#000';
  CTX.stroke();
};

var drawRow = function(index, row) {
  for (var i = 0; i < row.length; i++) {
    height = ROW_GAP / 2 + index * (ROW_HEIGHT + ROW_GAP);
    drawSamples(row[i].sound.samples, row[i].start, height, height + ROW_HEIGHT);
  }
};

var drawSamples = function(samples, sampleStart, yLo, yHi) {
  CTX.beginPath();

  for (var i = 0; i < samples.length; i += INDEX_STEP) {
    x = PIXELS_PER_SAMPLE * (sampleStart + i);
    y = yLo + (yHi - yLo) * (1 - 4 * samples[i]) / 2.0; // 1 -> yLo, -1 -> yHi
    if (i == 0) {
      CTX.moveTo(x, y);
    } else {
      CTX.lineTo(x, y);
    }
  }

  CTX.lineWidth = 1;
  CTX.strokeStyle = '#00f';
  CTX.stroke();
};

var fixWidth = function() {
  W = Math.max(document.body.clientWidth, calcWidthPx());
  if (C.width != W) {
    C.width = W;
    C.style.width = W + "px";
  }
};

var calcWidthPx = function() {
  lastSample = 0;
  for (var i = 0; i < VM.lines.length; i++) {
    lastRowSample = 0;
    for (var j = 0; j < VM.lines[i].length; j++) {
      block = VM.lines[i][j];
      lastBlockSample = block.sound.samples.length + block.start
      lastRowSample = Math.max(lastRowSample, lastBlockSample);
    }
    lastSample = Math.max(lastSample, lastRowSample);
  }
  return (lastSample * PIXELS_PER_SAMPLE) | 0;
};

var fixRowHeight = function(rows) {
  if (rows < 1) { rows = 1; }
  H = rows * ROW_HEIGHT + rows * ROW_GAP;
  C.height = H;
  C.style.height = H + "px";
};


var mouseDownE, mouseIsDrag, mouseIsDown;
$(C).on('mousedown', function(e) {
  mouseDownE = e;
  mouseIsDrag = false;
  mouseIsDown = true;
});
$(C).on('mousemove', function(e) {
  if (mouseIsDown) {
    if (isLargeDrag(mouseDownE, e)) {
      mouseIsDrag = true;
      handleCanvasDrag(mouseDownE, e);
    }
  }
});
$(C).on('mouseup', function(e) {
  if (!mouseIsDrag) {
    handleCanvasClick(mouseDownE);
  } 
  mouseDownE = null;
  mouseIsDrag = false;
  mouseIsDown = false;
});


var isLargeDrag = function(e1, e2) {
  return dist(e1.offsetX, e1.offsetY, e2.offsetX, e2.offsetY) > 5;
};
var dist = function(x1, y1, x2, y2) {
  var dx = x1 - x2, dy = y1 - y2;
  return Math.sqrt(dx * dx + dy * dy);
};


var handleCanvasClick = function(e) {
  var x = e.offsetX, y = e.offsetY;
  var sampleX = (x / PIXELS_PER_SAMPLE) | 0;
  var trackY = (y / (ROW_HEIGHT + ROW_GAP)) | 0;
  console.log("Clicked on sample %d in track %d", sampleX, trackY);

  if (e.shiftKey && VM.startSample !== null) {
    if (VM.endSample === null) {
      setSelectedSamples(VM.startSample, sampleX);
    } else {
      if (VM.startSample > sampleX) {
        setSelectedSamples(sampleX, VM.endSample);
      } else {
        setSelectedSamples(VM.startSample, sampleX);
      }

    }
  } else {
    VM.startSample = sampleX;
    VM.endSample = null;
  }
  R.drawRows();
};

var handleCanvasDrag = function(e1, e2) {
  var x1 = e1.offsetX, y1 = e1.offsetY;
  var x2 = e2.offsetX, y2 = e2.offsetY;
  var sampleX1 = (x1 / PIXELS_PER_SAMPLE) | 0;
  var sampleX2 = (x2 / PIXELS_PER_SAMPLE) | 0;
  setSelectedSamples(sampleX1, sampleX2);
  R.drawRows();
};

var setSelectedSamples = function(s1, s2) {
  VM.startSample = Math.min(s1, s2);
  VM.endSample   = Math.max(s1, s2);  
};

var updateSlider = function() {
  zoomValue.innerText = zoomSlider.value;
  INDEX_STEP = Math.pow(2, zoomSlider.value);
  PIXELS_PER_SAMPLE = 1 / INDEX_STEP;
  R.drawRows();
};


// Initialize state:
updateSlider();
$(zoomSlider).on('change', updateSlider);

})(window.render, window.viewmodel);