define(function(require, exports, module) {

    require('core');

    var $ = window.Zepto,
        UI = $.AMUI,
        $w = $(window),
        $doc = $(document);

    /**
     * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
     * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
     */

    var Popover = function(element, options) {
        this.options = $.extend({}, Popover.DEFAULTS, options || {});
        this.$element = $(element);
        this.active = null;
        this.$popover = (this.options.target && $(this.options.target)) || null;

        this.init();
        this.events();
    };

    Popover.DEFAULTS = {
        trigger: 'click',
        content: '',
        open: false,
        target: undefined,
        tpl: '<div class="am-popover"><div class="am-popover-inner"></div><div class="am-popover-caret"></div></div>'
    };

    Popover.prototype.init = function() {
        var me = this,
            $element = this.$element,
            $popover;

        if (!this.options.target) {
            this.$popover = this.getPopover();
            this.setContent();
        }

        $popover = this.$popover;

        $popover.appendTo($('body'));

        this.sizePopover();

        function sizePopover() {
            me.sizePopover();
        }

        $(window).on('resize:popover:amui', UI.utils.debounce(sizePopover, 50));

        $element.on('open:popover:amui', function() {
            $(window).on('resize:popover:amui', UI.utils.debounce(sizePopover, 50));
        });

        $element.on('close:popover:amui', function() {
            $(window).off('resize:popover:amui', sizePopover);
        });

        this.options.open && this.open();
    };

    Popover.prototype.sizePopover = function sizePopover() {
        var $element = this.$element,
            $popover = this.$popover;

        if (!$popover || !$popover.length) return;

        var popSize = $popover.getSize(),
            popWidth = $popover.width() || popSize.width,
            popHeight = $popover.height() || popSize.height,
            $popCaret = $popover.find('.am-popover-caret'),
            popCaretSize = ($popCaret.width() / 2) || 10,
            popTotalHeight = popHeight + popCaretSize;

        var triggerWidth = $element.outerWidth(),
            triggerHeight = $element.outerHeight(),
            triggerOffset = $element.offset(),
            triggerRect = $element[0].getBoundingClientRect();

        var winHeight = $w.height(),
            winWidth = $w.width();

        var popTop = 0,
            popLeft = 0,
            diff = 0,
            spacing = 3,
            popPosition = 'top';

        $popover.css({left: '', top: ''})
            .removeClass('am-popover-left am-popover-right am-popover-top am-popover-bottom');

        $popCaret.css({left: '', top: ''});

        if (popTotalHeight - spacing < triggerRect.top + spacing) { // on Top
            popTop = triggerOffset.top - popTotalHeight - spacing;
        } else if (popTotalHeight < winHeight - triggerRect.top - triggerRect.height) {
            // On bottom
            popPosition = 'bottom';
            popTop = triggerOffset.top + triggerHeight + popCaretSize + spacing;
        } else { // On middle
            popPosition = 'middle';
            popTop = triggerHeight / 2 + triggerOffset.top - popHeight / 2;
        }


        // Horizontal Position

        if (popPosition === 'top' || popPosition === 'bottom') {
            popLeft = triggerWidth / 2 + triggerOffset.left - popWidth / 2;

            diff = popLeft;

            if (popLeft < 5) popLeft = 5;
            if (popLeft + popWidth > winWidth) {
                popLeft = (winWidth - popWidth - 20);
                // console.log('戳到边边了 left %d, win %d, popw %d', popLeft, winWidth, popWidth);
            }
            if (popPosition === 'top') $popover.addClass('am-popover-bottom');
            if (popPosition === 'bottom') $popover.addClass('am-popover-top');
            diff = diff - popLeft;
            $popCaret.css({left: (popWidth / 2 - popCaretSize + diff) + 'px'});

        } else if (popPosition === 'middle') {
            popLeft = triggerOffset.left - popWidth - popCaretSize;
            $popover.addClass('am-popover-left');
            if (popLeft < 5) {
                popLeft = triggerOffset.left + triggerWidth + popCaretSize;
                $popover.removeClass('am-popover-left').addClass('am-popover-right');
            }

            if (popLeft + popWidth > winWidth) {
                popLeft = winWidth - popWidth - 5;
                $popover.removeClass('am-popover-left').addClass('am-popover-right');
            }
            $popCaret.css({top: (popHeight / 2 - popCaretSize / 2) + 'px'});
        }

        // Apply position style
        $popover.css({top: popTop + 'px', left: popLeft + 'px'});
    };

    Popover.prototype.toggle = function() {
        return this[this.active ? 'close' : 'open']();
    };

    Popover.prototype.open = function() {
        var $popover = this.$popover;
        this.$element.trigger('open:popover:amui');
        this.sizePopover();
        $popover.show().addClass('am-active');
        this.active = true;
    };

    Popover.prototype.close = function() {
        var $popover = this.$popover;

        this.$element.trigger('close:popover:amui');

        $popover
            .removeClass('am-active')
            .trigger('closed:popover:amui')
            .hide();
        this.active = false;
    };

    Popover.prototype.getPopover = function() {
        var uid = UI.utils.generateGUID('am-popover');
        return $(this.options.tpl, {
            id: uid
        });
    };

    Popover.prototype.setContent = function() {
        this.$popover && this.$popover.find('.am-popover-inner').empty().html(this.options.content);
    };

    Popover.prototype.events = function() {
        var eventNS = 'popover.amui',
            triggers = this.options.trigger.split(' ');

        for (var i = triggers.length; i--;) {
            var trigger = triggers[i];

            if (trigger === 'click') {
                this.$element.on('click.' + eventNS, $.proxy(this.toggle, this))
            } else { // hover or focus
                var eventIn  = trigger == 'hover' ? 'mouseenter' : 'focusin';
                var eventOut = trigger == 'hover' ? 'mouseleave' : 'focusout';

                this.$element.on(eventIn + '.' + eventNS, $.proxy(this.open, this));
                this.$element.on(eventOut + '.' + eventNS, $.proxy(this.close, this));
            }
        }
    };


        UI.popover = Popover;

        function Plugin(option) {
            return this.each(function() {
                var $this = $(this),
                    data = $this.data('am.popover'),
                    options = $.extend({}, UI.utils.parseOptions($this.attr('data-am-popover')), typeof option == 'object' && option);

                if (!data) {
                    $this.data('am.popover', (data = new Popover(this, options)));
                }

                if (typeof option == 'string') {
                    data[option]();
                }
            });
        }

        $.fn.popover = Plugin;


        // Init code
        $(function() {
            $('[data-am-popover]').popover();
        });

        module.exports = Popover;
    }
    )
    ;
