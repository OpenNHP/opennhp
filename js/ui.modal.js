define(function(require, exports, module) {
    require('core');

    var dimmer = require('ui.dimmer');

    var $ = window.Zepto;

    var UI = $.AMUI;

    var $win = $(window),
        $doc = $(document),
        $body = $('body'),
        supportTransition = UI.support.transition;

    /**
     * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
     * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
     */

    var Modal = function(element, options) {
        this.options = $.extend({}, Modal.DEFAULTS, options || {});
        this.$element = $(element);
        this.isPopup = this.$element.hasClass('am-popup');
        this.active = this.transitioning = null;

        this.events();
    };


    // Set template tabindex to tigger onfocus on div
    Modal.DEFAULTS = {
        selector: {
            modal: '.am-modal',
            active: '.am-modal-active'
        },
        cancelable: true,
        onConfirm: function() {
        },
        onCancel: function() {
        }
    };

    Modal.prototype.toggle = function(relatedElement) {
        return this.active ? this.close() : this.open(relatedElement)
    };

    Modal.prototype.open = function(relatedElement) {
        var $element = this.$element,
            isPopup = this.isPopup;

        if (this.transitioning || this.active) return;

        if (!this.$element.length)  return;

        isPopup && this.$element.show();

        this.active = true;

        $element.trigger($.Event('open:modal:amui', {relatedElement: relatedElement}));

        // trigger reflow;
        var clientLeft = $element[0].clientLeft;

        dimmer.open();

        $element.show();

        !isPopup && $element.redraw()
            .css({marginTop: -$element.height() / 2 + 'px'})

        $element.removeClass('am-modal-out').addClass('am-modal-active');

        this.transitioning = 1;

        var complete = function() {
            $element.trigger($.Event('opened:modal:amui', {relatedElement: relatedElement}));
            this.transitioning = 0;
        };

        if (!supportTransition) return complete.call(this);

        $element.one(supportTransition.end, $.proxy(complete, this));
    };

    Modal.prototype.close = function(relatedElement) {
        if (this.transitioning || !this.active) return;

        var $element = this.$element,
            isPopup = this.isPopup,
            that = this;

        this.$element.trigger($.Event('close:modal:amui', {relatedElement: relatedElement}));

        this.transitioning = 1;

        var complete = function() {
            $element.trigger('closed:amui:modal');
            isPopup && $element.removeClass('am-modal-out');
            $element.hide();
            this.transitioning = 0;
        };

        $element.removeClass('am-modal-active').addClass('am-modal-out');

        if (!supportTransition) return complete.call(this);

        $element.one(supportTransition.end, $.proxy(complete, this));

        // hide dimmer when all modal is closed
        if (!$body.find(Modal.DEFAULTS.selector.active).length) {
            dimmer.close();
        }

        this.active = false;
    };

    Modal.prototype.events = function() {
        var that = this,
            $element = this.$element,
            $ipt = $element.find('.am-modal-prompt-input');

        if (this.options.cancelable) {
            $element.on('keyup.modal.amui',
                $.proxy(function(e) {
                    if (this.active && e.which === 27) {
                        this.options.onCancel();
                        this.close();
                    }
                }, that));

            dimmer.$element.on('click', function(e) {
                that.close();
            });
        }

        // Close button
        $element.find('[data-am-modal-close]').on('click.modal.amui', function(e) {
            e.preventDefault();
            that.close();
        });

        $element.find('.am-modal-btn').on('click.modal.amui', function(e) {
            that.close();
        });


        $element.find('[data-am-modal-confirm]').on('click.modal.amui', function() {
            that.options.onConfirm($ipt.val());
        });

        $element.find('[data-am-modal-cancel]').on('click.modal.amui', function() {
            that.options.onCancel($ipt.val());
        });
    };


    function Plugin(option, relatedElement) {
        return this.each(function() {
            var $this = $(this),
                data = $this.data('am.modal'),
                options = $.extend({}, Modal.DEFAULTS, typeof option == 'object' && option);

            if (!data) {
                $this.data('am.modal', (data = new Modal(this, options)));
            }

            if (typeof option == 'string') {
                data[option](relatedElement);
            } else {
                data.toggle(option && option.relatedElement || undefined);
            }
        });
    }


    $.fn.modal = Plugin;

    $doc.on('click', '[data-am-modal]', function() {
        var $this = $(this),
            options = UI.utils.parseOptions($this.attr('data-am-modal')),
            $target = $(options.target || (this.href && this.href.replace(/.*(?=#[^\s]+$)/, ''))),
            option = $target.data('am.modal') ? 'toggle' : options;

        Plugin.call($target, option, this);
    });

    module.exports = Modal;
});

// TODO: parse tpl?