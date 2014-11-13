(function($) {
  'use strict';

  $(function() {
    var $fullText = $('.admin-fullText');
    $('#admin-fullscreen').on('click', function() {
      $.AMUI.fullscreen.toggle();
      $.AMUI.fullscreen.isFullscreen ? $fullText.text('关闭全屏') : $fullText.text('开启全屏');
    });
  });
})(jQuery);
