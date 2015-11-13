(function() {

Polymer({
  is: 'play-controls',

  properties: {
    isPlaying: Boolean,
    zoomValue: {
      type: Number,
      value: 7,
    }
  },

  attached: function() {
    this.handleZoomChange();
  },

  handleFastRewind: function(e) {
    util.performAction('fast-rewind', null, this);
  },

  handlePlayButton: function(e) {
    util.performAction('play', null, this);
  },

  handleZoomChange: function(e) {
    util.performAction('set-zoom', this.$.zoom.value, this);
  },

  handleLoadFile: function(e) {
    util.performAction('load-file', null, this);
  },
});

})();
