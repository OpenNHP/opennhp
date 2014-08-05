define(function (require, exports, module) {
    var accordion = require('ui.accordion');

    var $ = window.Zepto,
        accordionInit = function() {
            $('.am-accordion').each(function(index, item) {
                var settings = $(item).attr('data-accordion-settings');
                try {
                    settings = JSON.parse(settings);
                    $(item).accordion(settings);
                } catch(e) {
                    $(item).accordion();
                }
            });
        };

    // Init on DOM ready
    $(function () {
        accordionInit();
    });

    exports.init = accordionInit;
});
