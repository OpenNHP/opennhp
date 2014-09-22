define(function(require, exports, module) {
    require('core');

    require('ui.smooth-scroll');
    var $ = window.Zepto;

    var UI = $.AMUI;

    var goTopInit = function() {
        var $goTop = $('[data-am-widget="gotop"]'),
            $fixed = $goTop.filter('.am-gotop-fixed'),
            $win= $(window);

        $goTop.find('a').on('click', function(e) {
            e.preventDefault();
            $win.smoothScroll();
        });

        function checkPosition() {
            $fixed[($win.scrollTop() > 50 ? 'add' : 'remove') + 'Class']('am-active');
        }

        checkPosition();

        $win.on('scroll.gotop.amui', $.AMUI.utils.debounce(checkPosition, 100));
    };


    $(function() {
        goTopInit();
    });

    exports.init = goTopInit;
});
