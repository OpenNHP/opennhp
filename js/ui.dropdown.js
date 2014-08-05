define(function(require, exports, module) {

    'use strict';

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    var toggle = '[data-am-dropdown] > .am-dropdown-toggle';

    var Dropdown = function(element, options) {
        $(element).on('click.dropdown.amui', this.toggle);
    };

    Dropdown.prototype.toggle = function(e) {
        var $this = $(this);

        if ($this.is('.am-disabled, :disabled')) {
            return;
        }

        var $parent = $this.parent(),
            isActive = $parent.hasClass('am-active');

        clearDropdowns();

        if (!isActive) {

            var relatedTarget = {
                relatedTarget: this
            };

            $parent.trigger(e = $.Event('open:dropdown:amui', relatedTarget));

            if (e.isDefaultPrevented()) {
                return;
            }

            $this.trigger('focus');

            $parent
                .toggleClass('am-active')
                .trigger(e = $.Event('opened:dropdown:amui', relatedTarget));
        } else {
            $this.blur();
        }

        return false
    };

    Dropdown.prototype.keydown = function(e) {

        if (!/(38|40|27)/.test(e.keyCode)) return;

        var $this = $(this);

        e.preventDefault();
        e.stopPropagation();

        if ($this.is('.am-disabled, :disabled')) {
            return;
        }

        var $parent = $this.parent(),
            isActive = $parent.hasClass('am-active');

        if (!isActive || (isActive && e.keyCode == 27)) {
            if (e.which == 27) {
                $parent.find(toggle).trigger('focus');
            }
            return $this.trigger('click')
        }
    };

    function clearDropdowns(e) {
        if (e && e.which === 3) return;
        $(toggle).each(function() {
            var $parent = $(this).parent(),
                relatedTarget = {
                    relatedTarget: this
                };

            if (!$parent.hasClass('am-active')) {
                return;
            }

            $parent.trigger(e = $.Event('close:dropdown:amui', relatedTarget));

            if (e.isDefaultPrevented()) return;

            $parent.removeClass('am-active')
                .trigger(e = $.Event('closed:dropdown:amui', relatedTarget));
        });
    }


    UI.dropdown = Dropdown;

    // Dropdown Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.dropdown');

            if (!data) {
                $this.data('amui.dropdown', (data = new Dropdown(this)));
            }

            if (typeof option == 'string') {
                data[option].call($this);
            }
        });
    }


    $.fn.dropdown = Plugin;


    // Init code

    $(document)
        .on('click.dropdown.amui', '.am-dropdown form', function(e) {
            e.stopPropagation()
        })
        .on('click.dropdown.amui', toggle, Dropdown.prototype.toggle)
        .on('keydown.dropdown.amui', toggle, Dropdown.prototype.keydown);
});

// TODO: 1. 处理链接 focus
//       2. 增加 mouseenter / mouseleave 选项
