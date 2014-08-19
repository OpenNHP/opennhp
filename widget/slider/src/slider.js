define(function(require, exports, module) {
    var $ = window.Zepto,
        UI = $.AMUI;

    require("zepto.flexslider");

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
