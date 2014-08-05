define(function(require, exports, module) {
    var $ = window.Zepto;

    var sohuCSInit = function() {

        if (!$('#SOHUCS').length) return;

        var $sohucs = $('[data-am-widget="sohucs"]'),
            appid = $sohucs.attr('data-am-sohucs-appid'),
            conf = $sohucs.attr('data-am-sohucs-conf'),
            $cy = $('<script></script>', {
                charset: 'utf-8',
                id: 'changyan_mobile_js',
                src: 'http://changyan.sohu.com/upload/mobile/wap-js/changyan_mobile.js?client_id=' + appid + '&conf=' + conf
            });

        $('body').append($cy);
    };

    // Lazy load
    $(window).on('load', sohuCSInit);

    exports.init = sohuCSInit;
});
