define(function(require, exports, module) {
    require('core');

    var PureView = require('ui.pureview');

    var $ = window.Zepto,
        UI = $.AMUI;

    var galleryInit = function() {
        var $gallery = $('[data-am-widget="gallery"]'),
            $galleryOne = $gallery.filter('.am-gallery-one');

        $gallery.each(function() {
            var options = UI.utils.parseOptions($(this).attr('data-am-gallery'));
            
            if (options.pureview) {
                (typeof options.pureview === 'object') ? $(this).pureview(options.pureview) : $(this).pureview();
            }
        });

        $galleryOne.each(function() {
            galleryMore($(this));
        });
    };

    function galleryMore(object) {
        var moreData = $('<li class=\'am-gallery-more\'><a href=\'javascript:;\'>更多 &gt;&gt;</a></li>');

        if (object.children().length > 6) {

            object.children().each(function(index) {
                if (index > 5) {
                    $(this).hide();
                }
            });

            object.find('.am-gallery-more').remove();
            object.append(moreData);
        }

        object.find('.am-gallery-more').on('click', function() {
            object.children().show();
            $(this).hide();
        });
    }

    $(function() {
        galleryInit();
    });

    exports.init = galleryInit;
});
