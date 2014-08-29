define(function(require, exports, module) {
    require('core');
    require('ui.collapse');

    var $ = window.Zepto,
        UI = $.AMUI,
        accordionInit = function() {
            var $accordion = $('[data-am-widget="accordion"]'),
                selector = {
                    item: '.am-accordion-item',
                    title: '.am-accordion-title',
                    content: '.am-accordion-content'
                };

            $accordion.each(function(i, item) {
                var options = UI.utils.parseOptions($(item).attr('data-am-accordion')),
                    $title = $accordion.find(selector.title);

                $title.on('click', function() {
                    var $content = $(this).next(selector.content),
                        $parent = $(this).parent(selector.item),
                        data = $content.data('amui.collapse');

                    $parent.toggleClass('am-active');

                    if (!data) {
                        $content.collapse();
                    } else {
                        $content.collapse('toggle');
                    }

                    !options.multiple &&
                    $(item).children('.am-active').not($parent).removeClass('am-active').find('.am-accordion-content.am-in').collapse('close');

                });
            });
        };

    // Init on DOM ready
    $(function() {
        accordionInit();
    });

    exports.init = accordionInit;
});
