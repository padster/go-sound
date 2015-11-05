(function(W) {

W.console.log('loaded!')

// Visual context
C = document.getElementById('surface');
W = C.getBoundingClientRect().width;
H = C.getBoundingClientRect().height;
C.width = W; C.height = H;
CTX = C.getContext('2d');
CTX.imageSmoothingEnabled = false;

// Audio context
AC = new AudioContext();

lines = [];

ROW_HEIGHT = 120;
ROW_GAP = 30;
INDEX_STEP = 256;

addLine = function(sound, xStart) {
  newLine = [{ sound: sound, start: xStart }];
  lines.push(newLine)
};

fixRowHeight = function(rows) {
  if (rows < 1) { rows = 1; }
  H = rows * ROW_HEIGHT + (rows + 1) * ROW_GAP;
  C.height = H;
  C.style.height = H + "px";
};

drawRows = function() {
  fixRowHeight(lines.length);
  for (var i = 0; i < lines.length; i++) {
    drawRow(i, lines[i]);
  }
};

drawRow = function(index, row) {
  for (var i = 0; i < row.length; i++) {
    height = ROW_GAP + index * (ROW_HEIGHT + ROW_GAP);
    drawSamples(row[i].sound.samples, INDEX_STEP, row[i].start, 1.0, height, height + ROW_HEIGHT);
  }
};

playSamples = function(samples) {
  /*
  var buffer = AC.createBuffer(1, samples.length, 44100);
  var channel = buffer.getChannelData(0);
  for (var i = 0; i < samples.length; i++) { 
    channel[i] = samples[i];
  }
  var source = AC.createBufferSource();
  source.buffer = buffer;
  source.connect(AC.destination);
  source.start();
  */
  console.log("TODO: reenable sound play");
};

drawSamples = function(samples, idStep, xStart, xStep, yLo, yHi) {
  S = samples;

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

handleNewInput = function(data) {
  raw = window.atob(data.samples)
  byteData = Array.prototype.map.call(raw, function(x) { 
    return x.charCodeAt(0); 
  });

  buffer = new ArrayBuffer(byteData.length);
  intBuffer = new Uint8Array(buffer);
  for (var i = 0; i < byteData.length; i++) {
    intBuffer[i] = byteData[i];
  }

  // rewrite base-64 to floats
  data.samples = new Float32Array(buffer);
  console.log("LOADED! %O", data);

  addLine(data, 0);
  drawRows();
  // drawSamples(data.samples, 1000, 0, 0.5, 0, H);
  // playSamples(data.samples);
};

zoomSlider = document.getElementById('zoomSlider');
zoomValue = document.getElementById('zoomValue');
zoomValue.innerText = zoomSlider.value;
INDEX_STEP = Math.pow(2, zoomSlider.value);
$(zoomSlider).on('change', function() {
  zoomValue.innerText = zoomSlider.value;
  INDEX_STEP = Math.pow(2, zoomSlider.value);
  drawRows();
});


$(document.forms.loadFile).on('submit', function() {
  path = document.forms.loadFile.path.value;
  if (!path) {
    window.alert("Must have a path!");
  } else {
    console.log("Loading " + path + "...");
    $.ajax({
      url: "/_/load",
      type: "POST",
      contentType: "application/json",
      data: JSON.stringify({path: path}),
      dataType: "json",
      success: function(result) {
        handleNewInput(result);
      },
      error: function(result) {
        console.log("OOPS! %O", result)
      }
    });
  }
  return false;
});

W.C = C;
W.CTX = CTX;
W.AC = AC;

// HACK - split drawing code into separate file
W.lines = lines;

})(window);
