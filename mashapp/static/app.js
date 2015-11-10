(function(W, R, S, VM) {

// Audio context
var AC = new AudioContext();

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
};


$(document.forms.loadFile.load).on('click', function(e) {
  path = document.forms.loadFile.path.value;
  if (!path) {
    window.alert("Must have a path!");
  } else {
    // TODO - modal 'loading' popup.
    $(this).blur();
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
  e.stopPropagation();
  e.preventDefault();
  return false;
});


// TODO - split out into 'actions'?
var handlePlayStop = function() {
  // 'Stop' currently showing.
  if (S.isPlaying()) {
    S.stop();
    R.setPlaying(false);
  } else {
    R.setPlaying(S.playSelection(function(sampleAt) {
      R.showPlayLineAtSample(sampleAt);
    }, function() {
      R.hidePlayLine();
      R.setPlaying(false);
    }));
  }
};


$(document.getElementById('playButton')).on('click', handlePlayStop);
document.body.addEventListener('keypress', function(e) {
  if (e.keyCode == 32) {
    handlePlayStop();
  }
}, true);

// $('input, select').keypress(function(event) { return event.keyCode != 32; });


})(window, window.render, window.sound, window.viewmodel);
