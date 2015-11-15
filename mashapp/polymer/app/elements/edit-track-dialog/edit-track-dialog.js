(function() {

Polymer({
  is: 'edit-track-dialog',

  properties: {
    details: {
      type: Array,
      value: [],
    },

    callback: Object,
  },

  open: function(cb) {
    this.callback = cb;
    this.$.dialog.open();
  },

  handleSave: function() {
    // PICK: Whitelist properties?
    var dataClone = $.extend(true, {}, this.details);
    this.handleResult(dataClone);
  },

  handleClose: function() {
    this.handleResult(null);
  },

  handleResult: function(result) {
    this.callback(result);
    this.callback = null;
    this.$.dialog.close();
  },
});

})();
