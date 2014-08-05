define(function(require, exports, module) {

    'use strict';

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    // Button Class

    var Button = function(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, Button.DEFAULTS, options);
        this.isLoading = false;
        this.hasSpinner = false;
    };

    Button.DEFAULTS = {
        loadingText: 'loading...',
        loadingClass: 'am-btn-loading',
        loadingWithSpinner: '<span class="am-icon-refresh am-icon-spin"></span> loading...'
    };

    Button.prototype.setState = function(state) {
        var d = 'disabled',
            $el = this.$element,
            val = $el.is('input') ? 'val' : 'html',
            data = $el.data();

        state = state + 'Text';

        if (data.resetText == null) {
            $el.data('resetText', $el[val]());
        }

        // add spinner for element with html()
        if (UI.support.animation && !this.hasSpinner && val === 'html') {
            this.options.loadingText = this.options.loadingWithSpinner;
            this.hasSpinner = true;
        }

        $el[val](data[state] == null ? this.options[state] : data[state]);

        // push to event loop to allow forms to submit
        setTimeout($.proxy(function() {
            if (state == 'loadingText') {
                this.isLoading = true;
                $el.addClass(d + ' ' + this.options.loadingClass).attr(d, d);
            } else if (this.isLoading) {
                this.isLoading = false;
                $el.removeClass(d + ' ' + this.options.loadingClass).removeAttr(d);
            }
        }, this), 0);
    };

    Button.prototype.toggle = function() {
        var changed = true,
            $parent = this.$element.parent('.am-btn-group');

        if ($parent.length) {
            var $input = this.$element.find('input');

            if ($input.prop('type') == 'radio') {
                if ($input.prop('checked') && this.$element.hasClass('am-active')) {
                    changed = false;
                } else {
                    $parent.find('.am-active').removeClass('am-active')
                }
            }

            if (changed) {
                $input.prop('checked', !this.$element.hasClass('am-active')).trigger('change')
            }
        }

        if (changed) {
            this.$element.toggleClass('am-active')
        }
    };


    // Button plugin

    function Plugin(option) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.button');
            var options = typeof option == 'object' && option;

            if (!data) {
                $this.data('amui.button', (data = new Button(this, options)));
            }

            if (option == 'toggle') {
                data.toggle();
            } else if (option) {
                data.setState(option)
            }
        });
    }

    $.fn.button = Plugin;


    // Init code

    $(document).on('click.button.amui', '[data-am-button]', function(e) {
        var $btn = $(e.target);

        if (!$btn.hasClass('am-btn')) {
            $btn = $btn.closest('.am-btn');
        }

        Plugin.call($btn, 'toggle');
        e.preventDefault();
    });

    module.exports = Button;
    // TODO: 样式复查
});
