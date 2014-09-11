define(function(require, exports, module) {
    require('core');

    // add2home
    var addToHS = require('ui.add2home');

    var cookie = require('util.cookie'),
        modal = require('ui.modal'),
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
                    AMPlatform.util.goDesktop();
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
