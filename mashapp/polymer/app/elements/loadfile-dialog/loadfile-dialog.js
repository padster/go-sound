(function() {

Polymer({
  is: 'loadfile-dialog',

  properties: {
    options: {
      type: Array,
      value: [],
    },

    callback: Object,
  },

  open: function(cb) {
    this.callback = cb;
    this.$.dialog.open();
  },

  handleUploadFile: function() {
    this.handleResult(this.$.filePath.selectedItem.innerText.trim());
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
