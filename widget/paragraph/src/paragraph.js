define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    // Iscroll-lite Plugin
    var IScroll = require('ui.iscroll-lite');

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

        this.each(function() {
            $(this).on('click', function() {
                if (onOff && $('.am-paragraph').length) {
                    $('body').append($wrapDom);
                    $zoomWrap = $('.am-paragraph-wrap');
                    $pinch = $zoomWrap.find('.pinch-zoom');
                    $warpHead = $zoomWrap.find('header');

                    $pinch.each(function() {
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

    /*
    * index : ID 标识，多个 paragraph 里面多个 table
    */

    $.fn.scrollTable = function(index) {
        var This = $(this),
            $parent;

        This.wrap('<div class=\'am-paragraph-table-container\' id=\'am-paragraph-table-'+ index + '\'><div class=\'am-paragraph-table-scroller\'></div></div>');

        $parent = This.parent();
        $parent.width(This.width());
        $parent.height(This.height());

        var tableScroll = new IScroll('#am-paragraph-table-' + index, {
            eventPassthrough: true,
            scrollX: true,
            scrollY: false,
            preventDefault: false
        });
    };

    paragraphInit = function() {
        var $body = $('body'),
            $paragraph = $('[data-am-widget="paragraph"]');

        $paragraph.each(function(index) {
            var $this = $(this),
                options = UI.utils.parseOptions($this.attr('data-am-paragraph')),
                $index = index;

            if (options.imgLightbox) {
                $this.find('img').paragraphZoomToggle();

                $body.on('click', '.am-paragraph-wrap', function(e) {
                    e.preventDefault();
                    var target = e.target;
                    // Img is using pinch zoom
                    if (!$(target).is('img')) {
                        $(this).toggleClass('am-active');
                    }
                });
            }

            if (options.tableScrollable) {
                $this.find('table').each(function(index) {
                    if ($(this).width() > $(window).width()) {
                        $(this).scrollTable($index + '-' + index);
                    }
                });
            }
        });

    };

    $(window).on('load', function() {
        paragraphInit();
    });

    exports.init = paragraphInit;
});