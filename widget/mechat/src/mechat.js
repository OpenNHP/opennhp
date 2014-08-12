define(function(require, exports, module) {
    var $ = window.Zepto;

    var mechatInit = function() {

        if (!$('#mechat').length) return;

        var $mechat = $('[data-am-widget="mechat"]'),
            unitid = $mechat.attr('data-am-mechat-unitid'),
            $mechatData = $('<script></script>', {
                charset: 'utf-8',
                id: 'mechat_button_js',
                src: 'http://mechatim.com/js/unit/button.js?id=' + unitid
            });

        $('body').append($mechatData);
    };

    // Lazy load
    $(window).on('load', mechatInit);

    exports.init = mechatInit;
});
