define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto;

    function duoshuoInit() {
        var $dsThread = $('.ds-thread'),
            dsShortName = $dsThread.parent('[data-am-widget="duoshuo"]').attr('data-ds-short-name');

        if (!$dsThread.length || !dsShortName) return;

        window.duoshuoQuery = {short_name: dsShortName};

        var $dsJS = $('<script>', {
            async: true,
            type: 'text/javascript',
            src: (document.location.protocol == 'https:' ? 'https:' : 'http:') + '//static.duoshuo.com/embed.js',
            charset: 'utf-8'
        });

        $('body').append($dsJS);
    }

    $(window).on('load', duoshuoInit);

    exports.init = duoshuoInit;
});
