define(function(require, exports, module) {

    'use strict';

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    /**
     * @via https://github.com/Minwe/bootstrap/blob/master/js/alert.js
     * @copyright Copyright 2013 Twitter, Inc.
     * @license Apache 2.0
     */

    // Alert Class
    // NOTE: removeElement option is unavailable now

    var Alert = function(element, options) {
        this.options = $.extend({}, Alert.DEFAULTS, options);
        this.$element = $(element);

        this.$element.addClass('am-fade am-in').on('click.alert.amui', '.am-close', $.proxy(this.close, this));
    };

    Alert.DEFAULTS = {
        removeElement: true
    };

    Alert.prototype.close = function() {
        var $this = $(this),
            $target = $this.hasClass('am-alert') ? $this : $this.parent('.am-alert');

        $target.trigger('close:alert:amui');
        
        $target.removeClass('am-in');

        function processAlert() {
            $target.off().trigger('closed:alert:amui').remove();
        }

        UI.support.transition && $target.hasClass('am-fade') ?
            $target
                .one(UI.support.transition.end, processAlert)
                .emulateTransitionEnd(200) : processAlert();
    };

    UI.alert = Alert;

    // Alert Plugin
    $.fn.alert = function(option) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('amui.alert'),
                options = typeof option == 'object' && option;
            
            if (!data) {
                $this.data('amui.alert', (data = new Alert(this, options || {})));
            }

            if (typeof option == 'string') {
                data[option].call($this);
            }
        });
    };


    // Init code
    $(document).on('click.alert.amui', '[data-am-alert]', function(e){
        var $target = $(e.target);
        $(this).addClass('am-fade am-in');
        $target.is('.am-close') && $(this).alert('close');
    });

    module.exports = Alert;
});
