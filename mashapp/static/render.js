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
zoomValue.innerText = zoomSlider.value;
var INDEX_STEP = Math.pow(2, zoomSlider.value);
$(zoomSlider).on('change', function() {
  zoomValue.innerText = zoomSlider.value;
  INDEX_STEP = Math.pow(2, zoomSlider.value);
  R.drawRows();
});

var ROW_HEIGHT = 120;
var ROW_GAP = 30;

R.drawRows = function() {
  fixRowHeight(VM.lines.length);
  for (var i = 0; i < VM.lines.length; i++) {
    drawRow(i, VM.lines[i]);
  }
};

var drawRow = function(index, row) {
  for (var i = 0; i < row.length; i++) {
    height = ROW_GAP + index * (ROW_HEIGHT + ROW_GAP);
    drawSamples(row[i].sound.samples, INDEX_STEP, row[i].start, 1.0, height, height + ROW_HEIGHT);
  }
};

var drawSamples = function(samples, idStep, xStart, xStep, yLo, yHi) {
  CTX.beginPath();

  ys = [];
  x = xStart;
  for (var i = 0; i < samples.length; i += idStep) {
    y = yLo + (yHi - yLo) * (1 - 4 * samples[i]) / 2.0; // 1 -> yLo, -1 -> yHi
    if (i == 0) {
      CTX.moveTo(x, y);
    } else {
      CTX.lineTo(x, y);
    }
    x += xStep;
  }

  CTX.lineWidth = 1;
  CTX.strokeStyle = '#0000ff';
  CTX.stroke();
};

var fixRowHeight = function(rows) {
  if (rows < 1) { rows = 1; }
  H = rows * ROW_HEIGHT + (rows + 1) * ROW_GAP;
  C.height = H;
  C.style.height = H + "px";
};

})(window.render, window.viewmodel);