(function($) {
  'use strict';

  $(function() {
    $('#admin-fullscreen').on('click', function() {
      $.AMUI.fullscreen.toggle();
    });
  });
})(window.Zepto);
