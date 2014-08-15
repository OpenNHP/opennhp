define(function (require, exports, module) {

    var touchGallery = require('zepto.touchgallery');

    var $ = window.Zepto;

    var galleryInit = function() {
        var $themeOne = $('.am-gallery-one');

        $('[data-am-gallery] a').touchTouch();

        $themeOne.each(function() {
            galleryMore($(this));
        });
    };

    function galleryMore(object) {
        var moreData = $('<li class=\'am-gallery-more\'><a href=\'javascript:;\'>更多 &gt;&gt;</a></li>');

        if (object.children().length > 6) {

            object.children().each(function (index) {
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

    $(function () {
        galleryInit();
    });

    exports.init = galleryInit;

});