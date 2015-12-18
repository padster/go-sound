(function() {

Polymer({
  is: 'text-dialog',

  properties: {
    value: String,
    title: String,
    callback: Object,
  },

  open: function(cb) {
    this.callback = cb;
    this.$.dialog.open();
    this.$.value.focus();
  },

  handleSave: function() {
    this.handleResult(this.value);
  },

  handleClose: function() {
    this.handleResult(null);
  },

  handleResult: function(result) {
    this.callback(result);
    this.callback = null;
    this.$.dialog.close();
  },

  handleKey: function(e) {
    // Save on Enter. NOTE: Close on Escape is automatic.
    if (e.which == 13) {
      this.handleSave();
    }
  },
});

})();
