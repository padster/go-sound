(function(W, R, VM) {

// Audio context
var AC = new AudioContext();

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
  
  //console.log("TODO: reenable sound play");
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

  VM.addLine(data, 0);
  R.drawRows();
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

})(window, window.render, window.viewmodel);
