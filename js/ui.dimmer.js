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
        $html.addClass('am-dimmer-active');
        $dimmer.addClass('am-active');
        $(relatedElement).length && $(relatedElement).show();
        $doc.trigger('open:dimmer:amui');
        return this;
    };

    Dimmer.prototype.close = function(relatedElement) {
        $dimmer.removeClass('am-active');
        $html.removeClass('am-dimmer-active');
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


    var dimmer = new Dimmer();

    UI.dimmer = dimmer;

    module.exports = dimmer;
});
