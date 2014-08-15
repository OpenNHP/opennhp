define(function (require, exports, module) {
    var $ = window.Zepto,

        listNewsInit = function () {
            $('.am-list-news-one').each(function() {
                amListNewsMore($(this));
            });
        };

    function amListNewsMore(object) {
        var $amList = object.find(".am-list");

        var $listMore = "<a class='am-list-news-more am-btn am-btn-default' href='###'>更多 &gt;&gt;</a>";

        if ($amList.children().length > 6) {

            $amList.children().each(function (index) {
                if (index > 5) {
                    $(this).hide();
                }
            });

            object.find('.am-list-news-more').remove();
            object.append($listMore);
        }

        object.find(".am-list-news-more").on("click", function() {
            $amList.children().show();
            $(this).hide();
        });
    }


    $(function () {
        listNewsInit();
    });

    exports.init = listNewsInit;
});