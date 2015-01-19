'use strict';

var $ = require('jquery');
var UI = require('./core');
var dimmer = require('./ui.dimmer');
var $doc = $(document);
var supportTransition = UI.support.transition;

/**
 * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
 * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
 */

var Modal = function(element, options) {
  this.options = $.extend({}, Modal.DEFAULTS, options || {});
  this.$element = $(element);
  this.$dialog =   this.$element.find('.am-modal-dialog');

  if (!this.$element.attr('id')) {
    this.$element.attr('id', UI.utils.generateGUID('am-modal'));
  }

  this.isPopup = this.$element.hasClass('am-popup');
  this.isActions = this.$element.hasClass('am-modal-actions');
  this.active = this.transitioning = this.relatedTarget = null;

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
  closeViaDimmer: true,
  cancelable: true,
  onConfirm: function() {
  },
  onCancel: function() {
  },
  height: undefined,
  width: undefined,
  duration: 300, // must equal the CSS transition duration
  transitionEnd: supportTransition && supportTransition.end + '.modal.amui'
};

Modal.prototype.toggle = function(relatedTarget) {
  return this.active ? this.close() : this.open(relatedTarget);
};

Modal.prototype.open = function(relatedTarget) {
  var $element = this.$element;
  var options = this.options;
  var isPopup = this.isPopup;
  var width = options.width;
  var height = options.height;
  var style = {};

  if (this.active) {
    return;
  }

  if (!this.$element.length) {
    return;
  }

  // callback hook
  relatedTarget && (this.relatedTarget = relatedTarget);

  // 判断如果还在动画，就先触发之前的closed事件
  if (this.transitioning) {
    clearTimeout($element.transitionEndTimmer);
    $element.transitionEndTimmer = null;
    $element.trigger(options.transitionEnd).off(options.transitionEnd);
  }

  isPopup && this.$element.show();

  this.active = true;

  $element.trigger($.Event('open.modal.amui', {relatedTarget: relatedTarget}));

  dimmer.open($element);

  $element.show().redraw();

  // apply Modal width/height if set
  if (!isPopup && !this.isActions) {
    if (width) {
      width = parseInt(width, 10);
      style.width =  width + 'px';
      style.marginLeft =  -parseInt(width / 2) + 'px';
    }

    if (height) {
      height = parseInt(height, 10);
      // style.height = height + 'px';
      style.marginTop = -parseInt(height / 2) + 'px';

      // the background color is styled to $dialog
      // so the height should set to $dialog
      this.$dialog.css({height: height + 'px'});
    } else {
      style.marginTop = -parseInt($element.height() / 2, 10) + 'px';
    }

    $element.css(style);
  }

  $element.
    removeClass(options.className.out).
    addClass(options.className.active);

  this.transitioning = 1;

  var complete = function() {
    $element.trigger($.Event('opened.modal.amui',
      {relatedTarget: relatedTarget}));
    this.transitioning = 0;
  };

  if (!supportTransition) {
    return complete.call(this);
  }

  $element.
    one(options.transitionEnd, $.proxy(complete, this)).
    emulateTransitionEnd(options.duration);
};

Modal.prototype.close = function(relatedTarget) {
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
    dimmer.close($element, true);
  }

  this.$element.trigger($.Event('close.modal.amui',
    {relatedTarget: relatedTarget}));

  this.transitioning = 1;

  var complete = function() {
    $element.trigger('closed.modal.amui');
    isPopup && $element.removeClass(options.className.out);
    $element.hide();
    this.transitioning = 0;
    // 不强制关闭 Dimmer，以便多个 Modal 可以共享 Dimmer
    dimmer.close($element, false);
    this.active = false;
  };

  $element.removeClass(options.className.active).
    addClass(options.className.out);

  if (!supportTransition) {
    return complete.call(this);
  }

  $element.one(options.transitionEnd, $.proxy(complete, this)).
    emulateTransitionEnd(options.duration);
};

Modal.prototype.events = function() {
  var that = this;
  var $element = this.$element;
  var $ipt = $element.find('.am-modal-prompt-input');
  var getData = function() {
    var data = [];
    $ipt.each(function() {
      data.push($(this).val());
    });

    return (data.length === 0) ? undefined :
      ((data.length === 1) ? data[0] : data);
  };

  // close via Esc key
  if (this.options.cancelable) {
    $element.on('keyup.modal.amui', function(e) {
        if (that.active && e.which === 27) {
          $element.trigger('cancel.modal.amui');
          that.close();
        }
      });
  }

  // Close Modal when dimmer clicked
  if (this.options.closeViaDimmer) {
    dimmer.$element.on('click.dimmer.modal.amui', function(e) {
      that.close();
    });
  }

  // Close Modal when button clicked
  $element.find('[data-am-modal-close], .am-modal-btn').
    on('click.close.modal.amui', function(e) {
    e.preventDefault();
    that.close();
  });

  $element.find('[data-am-modal-confirm]').on('click.confirm.modal.amui',
    function() {
      $element.trigger($.Event('confirm.modal.amui', {
        trigger: this
      }));
    });

  $element.find('[data-am-modal-cancel]').
    on('click.cancel.modal.amui', function() {
      $element.trigger($.Event('cancel.modal.amui', {
        trigger: this
      }));
    });

  $element.on('confirm.modal.amui', function(e) {
    e.data = getData();
    that.options.onConfirm.call(that, e);
  }).on('cancel.modal.amui', function(e) {
    e.data = getData();
    that.options.onCancel.call(that, e);
  });
};

function Plugin(option, relatedTarget) {
  return this.each(function() {
    var $this = $(this);
    var data = $this.data('amui.modal');
    var options = $.extend({},
      Modal.DEFAULTS, typeof option == 'object' && option);

    if (!data) {
      $this.data('amui.modal', (data = new Modal(this, options)));
    }

    if (typeof option == 'string') {
      data[option] && data[option](relatedTarget);
    } else {
      data.toggle(option && option.relatedTarget || undefined);
    }
  });
}

$.fn.modal = Plugin;

// Init
$doc.on('click.modal.amui.data-api', '[data-am-modal]', function() {
  var $this = $(this);
  var options = UI.utils.parseOptions($this.attr('data-am-modal'));
  var $target = $(options.target ||
  (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
  var option = $target.data('amui.modal') ? 'toggle' : options;

  Plugin.call($target, option, this);
});

$.AMUI.modal = Modal;

module.exports = Modal;
