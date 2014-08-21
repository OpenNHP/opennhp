define(function(require, exports, module) {
    require('core');
    require('zepto.flexslider');

    var $ = window.Zepto,
        UI = $.AMUI;

    var sliderInit = function() {
        var $sliders = $('[data-am-widget="slider"]');
        $sliders.not(".am-slider-manual").each(function(i, item) {
            var options = UI.utils.parseOptions($(item).attr('data-am-slider'));
            $(item).flexslider(options);
        });
    };

    $(document).on('ready', sliderInit);

    exports.init = sliderInit;
});
