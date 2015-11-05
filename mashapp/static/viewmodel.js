window.viewmodel = {};

(function(VM) {

VM.lines = [];

VM.addLine = function(sound, xStart) {
  VM.lines.push([{ sound: sound, start: xStart }])
};


})(window.viewmodel);