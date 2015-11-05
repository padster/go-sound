console.log('loaded!')

// Visual context
C = document.getElementById('surface');
W = C.getBoundingClientRect().width;
H = C.getBoundingClientRect().height;
C.width = W; C.height = H;
CTX = C.getContext('2d');
CTX.imageSmoothingEnabled = false;

// Audio context
AC = new AudioContext();


playSamples = function(samples) {
  var buffer = AC.createBuffer(1, samples.length, 44100);
  var channel = buffer.getChannelData(0);
  for (var i = 0; i < samples.length; i++) { 
    channel[i] = samples[i];
  }
  var source = AC.createBufferSource();
  source.buffer = buffer;
  source.connect(AC.destination);
  source.start();
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
  // console.log("Ys: %O", ys);
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
  drawSamples(data.samples, 1000, 0, 0.5, 0, H);
  playSamples(data.samples);
};

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
