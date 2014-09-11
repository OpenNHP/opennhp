define(function (require, exports, module) {
    require('core');
    require('ui.tabs');

    var $ = window.Zepto;

    var tabsInit = function () {
        $('[data-am-widget="tabs"]').tabs();
    };

    $(function () {
        tabsInit();
    });

    exports.init = tabsInit;
});
