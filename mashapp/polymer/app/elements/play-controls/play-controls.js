(function() {

Polymer({
  is: 'play-controls',

  properties: {
    isPlaying: Boolean,
  },

  handleFastRewind: function(e) {
    util.performAction('fast-rewind', null, this);
  },

  handlePlayButton: function(e) {
    util.performAction('play', null, this);
  },
});

})();
