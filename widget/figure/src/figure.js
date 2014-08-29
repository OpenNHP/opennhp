define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto;

    // PinchZoom Plugin
    var PinchZoom = require('zepto.pinchzoom');
    /**
     * Is Images zoomable
     * @return {Boolean}
     */
    $.isImgZoomAble = function(imgElement) {
        var t = new Image();
        t.src = imgElement.src;

        var zoomAble = ($(imgElement).width() < t.width);

        if (zoomAble) {
            $(imgElement).parent('.am-figure').addClass('am-figure-zoomable');
        }
        return zoomAble;
    };

    $.fn.imgZoomToggle = function() {
        return this.each(function() {
            var zoomAble = $.isImgZoomAble(this),
                $wrapDom = $('<div class="am-figure-wrap"><div class="pinch-zoom"></div></div>');
                $zoomWrap = $('.am-figure-wrap');

            if ($zoomWrap.length == 0) {
                $('body').append($wrapDom);
                $zoomWrap = $('.am-figure-wrap');
                $pinch = $zoomWrap.find('.pinch-zoom');

                $pinch.each(function() {
                    new PinchZoom($(this), {});
                });

            }

            if (zoomAble) {
                //$zoomWrap.empty().html(this.outerHTML);
                $pinch.empty().html(this.outerHTML);

                $zoomWrap.find('img').width($(window).width());
                $(this).parent('.am-figure').on('click', function() {
                    $zoomWrap.toggleClass('am-active');
                });

                $zoomWrap.on('click', function(e) {
                    e.preventDefault();
                    var target = e.target;
                    // Img is using pinch zoom
                    if (!$(target).is('img')) {
                        $(this).toggleClass('am-active');
                    }
                });
            }
        });
    };

    var figureInit = function() {
        $('.am-figure img').imgZoomToggle();
    };

    $(window).on('load', function() {
        figureInit();
    });

    exports.init = figureInit;
});