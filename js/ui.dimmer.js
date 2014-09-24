define(function(require, exports, module) {

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    var $dimmer = $('<div class="am-dimmer" data-am-dimmer></div>'),
        $doc = $(document),
        $html = $('html');

    var Dimmer = function() {
        this.hasDimmer = $('[data-am-dimmer]').length ? true : false;

        this.$element = $dimmer;

        this.scrollbarWidth = 0;

        $(document).on('ready', $.proxy(this.init, this));
    };

    Dimmer.prototype.init = function() {
        if (!this.hasDimmer) {
            $dimmer.appendTo($('body'));
            this.events();
            this.hasDimmer = true;
        }
        $doc.trigger('init:dimmer:amui');
        return this;
    };

    Dimmer.prototype.open = function(relatedElement) {

        this.measureScrollbar();

        $html.css('margin-left', -this.scrollbarWidth)
             .addClass('am-dimmer-active');

        $dimmer.addClass('am-active');
        $(relatedElement).length && $(relatedElement).show();
        $doc.trigger('open:dimmer:amui');
        return this;
    };

    Dimmer.prototype.close = function(relatedElement) {

        $html.css('margin-left', '')
             .removeClass('am-dimmer-active');

        $dimmer.removeClass('am-active');

        $(relatedElement).length && $(relatedElement).hide();
        $doc.trigger('close:dimmer:amui');
        return this;
    };

    Dimmer.prototype.events = function() {
        var that = this;
        $dimmer.on('click.dimmer.amui', function() {
            //that.hide();
        })
    };

    Dimmer.prototype.measureScrollbar = function() {

        if ($html.width() >= window.innerWidth) return;

        var scrollbarWidth = window.innerWidth - $html.width();

        this.scrollbarWidth = this.scrollbarWidth || scrollbarWidth;

    };
    var dimmer = new Dimmer();

    UI.dimmer = dimmer;

    module.exports = dimmer;
});
