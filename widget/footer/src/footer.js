define(function(require, exports, module) {
  'use strict';

  require('core');
  require('ui.modal');

  var addToHS = require('ui.add2home'),
      cookie = require('util.cookie'),
      $ = window.Zepto,
      footerInit = function() {
        // modal mode
        $('.am-footer-ysp').on('click', function() {
          $('#am-footer-mode').modal();
        });

        addToHS();

        // switch mode
        // switch to desktop
        $('[data-rel="desktop"]').on('click', function(e) {
          e.preventDefault();
          if (window.AMPlatform) { // front end
            window.AMPlatform.util.goDesktop();
          } else { // back end
            cookie.set('allmobilize', 'desktop', '', '/');
            window.location = window.location;
          }
        });
      };

  $(window).on('load', function() {
    footerInit();
  });

  exports.init = footerInit;
});
