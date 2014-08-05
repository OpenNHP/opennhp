define(function (require, exports, module) {
    var $ = window.Zepto;

    require("zepto.flexslider");

    var sliderInit = function() {
        $('.am-slider').not(".am-slider-manual").each(function (i, item) {
            var options = $(item).attr('data-slider-config');
            if (options) {
                $(item).flexslider($.parseJSON(options));
            } else {
                $(item).flexslider();
            }
        });
    };

    $(document).on('ready', sliderInit);

    exports.init = sliderInit;
});
