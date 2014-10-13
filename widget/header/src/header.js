define(function(require, exports, module) {
    var $ = window.Zepto;

    function headerInit() {
        $('[data-am-widget="header"]').each(function() {
            if ($(this).hasClass('am-header-fixed')) {
                $('body').addClass('am-with-fixed-header');
                return false;
            }
        });
    }

    $(function() {
        headerInit();
    });

    exports.init = headerInit;
});
