define(function (require, exports, module) {
    require('core');

    var $ = window.Zepto;
    // PinchZoom Plugin
    var PinchZoom = require('zepto.pinchzoom');

    var paragraphInit;

    $.fn.paragraphZoomToggle = function() {
        var $warpHead,
            $pinch,
            $zoomWrap,
            onOff = true, // 防止重复创建
            $wrapDom = $('<div class=\'am-paragraph-wrap\'><header></header><div class=\'pinch-zoom\'></div></div>');

        $zoomWrap = $('.am-paragraph-wrap');
        $warpHead = $('.am-paragraph-wrap header');
        $pinch = $zoomWrap.find('.pinch-zoom');

        this.each(function () {
            $(this).on('click', function () {
                if (onOff && $('.am-paragraph').length) {
                    $('body').append($wrapDom);
                    $zoomWrap = $('.am-paragraph-wrap');
                    $pinch = $zoomWrap.find('.pinch-zoom');
                    $warpHead = $zoomWrap.find('header');

                    $pinch.each(function () {
                        new PinchZoom($(this), {});
                    });

                    onOff = false;
                }

                $pinch.html(this.outerHTML);
                if ($(this).attr('alt')) {
                    $warpHead.html($(this).attr('alt'));
                } else {
                    $warpHead.html('返回');
                }

                $zoomWrap.addClass('am-active').find('img').width($(window).width());

            });
        })
    };

    $.fn.paragraphTable = function (objWidth) {
        var This = $(this),
            distX = 0,
            disX = 0,
            disY = 0,
            downX,
            downY,
            $parent,
            scrollY;

        if (objWidth > $('body').width()) {
            This.wrap('<div class=\'am-paragraph-table-container\'><div class=\'am-paragraph-table-scroller\'></div></div>');
            $parent = This.parent();
            $parent.width(objWidth);
            $parent.height(This.height());
            $parent.parent().height(This.height() + 20);

            $parent.on('touchstart MSPointerDown pointerdown', function (ev) {
                var oTarget = ev.targetTouches[0];
                distX = oTarget.clientX - $(this).offset().left;
                downX = oTarget.clientX;
                downY = oTarget.clientY;
                scrollY = undefined;

                $(document).on('touchmove MSPointerMove pointermove', fnMove);
                $(document).on('touchend MSPointerUp pointerup', fnUp);

            })

        }

        function fnUp(ev) {
            ev.preventDefault();
            var oTarget = ev.changedTouches[0];
            var L = $parent.offset().left;
            // ->
            if (L > 10) {
                $parent.animate({
                    left: 10
                }, 500, 'ease-out')
            }
            //<-
            if (L < -$parent.width() + $(window).width() - 10) {
                $parent.animate({
                    left: -$parent.width() + $(window).width() - 10
                }, 500, 'ease-out')
            }

            $(document).off('touchend MSPointerUp pointerup', fnUp);
            $(document).off('touchmove MSPointerMove pointermove', fnMove);

        }

        function fnMove(ev) {
            var oTarget = ev.targetTouches[0];
            disX = oTarget.clientX - downX;
            disY = oTarget.clientY - downY;

            if (typeof scrollY == 'undefined') {
                scrollY = !!( scrollY || Math.abs(disX) < Math.abs(disY) );
            }

            if (!scrollY) {
                ev.preventDefault();
                This.parent().css('left', oTarget.clientX - distX);
            }

        }

    };

    paragraphInit = function() {
        var $body = $('body'),
            $paragraph = $('.am-paragraph'),
            $tableWidth;

        if ($paragraph.length && $paragraph.attr('data-am-imgParagraph')) {

            $paragraph.find('img').paragraphZoomToggle();

            $body.on('click', '.am-paragraph-wrap', function (e) {
                e.preventDefault();
                var target = e.target;
                // Img is using pinch zoom
                if (!$(target).is('img')) {
                    $(this).toggleClass('am-active');
                }
            });
        }

        if ($paragraph.length && $paragraph.attr('data-am-tableParagraph')) {
            $paragraph.find('table').each(function () {
                $tableWidth = $(this).width();
                $(this).paragraphTable($tableWidth);
            })
        }
    };

    $(window).on('load', function() {
        paragraphInit();
    });

    exports.init = paragraphInit;
});