define(function(require, exports, module) {
  'use strict';

  require('core');

  var dimmer = require('ui.dimmer');
  var $ = window.Zepto;
  var UI = $.AMUI;
  var $doc = $(document);
  var supportTransition = UI.support.transition;

  /**
   * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
   * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
   */

  var Modal = function(element, options) {
    this.options = $.extend({}, Modal.DEFAULTS, options || {});
    this.$element = $(element);

    if (!this.$element.attr('id')) {
      this.$element.attr('id', UI.utils.generateGUID('am-modal'));
    }

    this.isPopup = this.$element.hasClass('am-popup');
    this.active = this.transitioning = null;

    this.events();
  };

  Modal.DEFAULTS = {
    className: {
      active: 'am-modal-active',
      out: 'am-modal-out'
    },
    selector: {
      modal: '.am-modal',
      active: '.am-modal-active'
    },
    cancelable: true,
    onConfirm: function() {
    },
    onCancel: function() {
    },
    duration: 300, // must equal the CSS transition duration
    transitionEnd: supportTransition.end &&
        supportTransition.end + '.modal.amui'
  };

  Modal.prototype.toggle = function(relatedElement) {
    return this.active ? this.close() : this.open(relatedElement);
  };

  Modal.prototype.open = function(relatedElement) {
    var $element = this.$element;
    var options = this.options;
    var isPopup = this.isPopup;

    if (this.active) {
      return;
    }

    if (!this.$element.length) {
      return;
    }

    // 判断如果还在动画，就先触发之前的closed事件
    if (this.transitioning) {
      clearTimeout($element.transitionEndTimmer);
      $element.transitionEndTimmer = null;
      $element.trigger(options.transitionEnd).off(options.transitionEnd);
    }

    isPopup && this.$element.show();

    this.active = true;

    $element.trigger($.Event('open:modal:amui',
        {relatedElement: relatedElement}));

    dimmer.open($element);

    $element.show().redraw();

    !isPopup && $element.css({
      marginTop: -parseInt($element.height() / 2, 10) + 'px'
    });

    $element.
        removeClass(options.className.out).
        addClass(options.className.active);

    this.transitioning = 1;

    var complete = function() {
      $element.trigger($.Event('opened:modal:amui',
          {relatedElement: relatedElement}));
      this.transitioning = 0;
    };

    if (!supportTransition) {
      return complete.call(this);
    }

    $element.
        one(options.transitionEnd, $.proxy(complete, this)).
        emulateTransitionEnd(options.duration);
  };

  Modal.prototype.close = function(relatedElement) {
    if (!this.active) {
      return;
    }

    var $element = this.$element;
    var options = this.options;
    var isPopup = this.isPopup;

    // 判断如果还在动画，就先触发之前的opened事件
    if (this.transitioning) {
      clearTimeout($element.transitionEndTimmer);
      $element.transitionEndTimmer = null;
      $element.trigger(options.transitionEnd).off(options.transitionEnd);
    }

    this.$element.trigger($.Event('close:modal:amui',
        {relatedElement: relatedElement}));

    this.transitioning = 1;

    var complete = function() {
      $element.trigger('closed:amui:modal');
      isPopup && $element.removeClass(options.className.out);
      $element.hide();
      this.transitioning = 0;
    };

    $element.
        removeClass(options.className.active).
        addClass(options.className.out);

    if (!supportTransition) {
      return complete.call(this);
    }

    $element
        .one(options.transitionEnd, $.proxy(complete, this))
        .emulateTransitionEnd(options.duration);

    dimmer.close($element, true);

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
      var $this = $(this);
      var data = $this.data('am.modal');
      var options = $.extend({},
              Modal.DEFAULTS, typeof option == 'object' && option);

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
    var $this = $(this);
    var options = UI.utils.parseOptions($this.attr('data-am-modal'));
    var $target = $(options.target ||
            (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
    var option = $target.data('am.modal') ? 'toggle' : options;

    Plugin.call($target, option, this);
  });

  UI.modal = Modal;

  module.exports = Modal;
});
