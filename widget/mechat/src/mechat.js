define(function(require, exports, module) {
    var $ = window.Zepto;

    var mechatInit = function() {

        if (!$('#mechat').length) return;

        var $mechat = $('[data-am-widget="mechat"]'),
            unitid = $mechat.data('am-mechat-unitid'),
            $mechatData = $('<script>', {
                charset: 'utf-8',
                src: 'http://mechatim.com/js/unit/button.js?id=' + unitid
            });

        $('body').append($mechatData);
    };

    // Lazy load
    $(window).on('load', mechatInit);

    exports.init = mechatInit;
});
