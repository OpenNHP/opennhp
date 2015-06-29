'use strict';

var $ = require('jquery');
var UI = require('./core');

var Validator = function(element, options) {
  this.options = $.extend({}, Validator.DEFAULTS, options);
  this.options.patterns = $.extend({}, Validator.patterns,
    this.options.patterns);
  var locales = this.options.locales;
  !Validator.validationMessages[locales] && (this.options.locales = 'zh_CN');
  this.$element = $(element);
  this.init();
};

Validator.DEFAULTS = {
  debug: false,
  locales: 'zh_CN',
  H5validation: false,
  H5inputType: ['email', 'url', 'number'],
  patterns: {},
  patternClassPrefix: 'js-pattern-',
  activeClass: 'am-active',
  inValidClass: 'am-field-error',
  validClass: 'am-field-valid',

  validateOnSubmit: true,
  alwaysRevalidate: false,
  // Elements to validate with allValid (only validating visible elements)
  // :input: selects all input, textarea, select and button elements.
  allFields: ':input:visible:not(:submit, :button, :disabled, .am-novalidate)',

  // Custom events
  customEvents: 'validate',

  // Keyboard events
  keyboardFields: ':input:not(:submit, :button, :disabled, .am-novalidate)',
  keyboardEvents: 'focusout, change', // keyup, focusin

  // bind `keyup` event to active field
  activeKeyup: false,
  textareaMaxlenthKeyup: true,

  // Mouse events
  pointerFields: 'input[type="range"]:not(:disabled, .am-novalidate), ' +
  'input[type="radio"]:not(:disabled, .am-novalidate), ' +
  'input[type="checkbox"]:not(:disabled, .am-novalidate), ' +
  'select:not(:disabled, .am-novalidate), ' +
  'option:not(:disabled, .am-novalidate)',
  pointerEvents: 'click',

  onValid: function(validity) {
  },

  onInValid: function(validity) {
  },

  markValid: function(validity) {
    // this is Validator instance
    var options = this.options;
    var $field = $(validity.field);
    var $parent = $field.closest('.am-form-group');

    $field.addClass(options.validClass).removeClass(options.inValidClass);
    $parent.addClass('am-form-success').removeClass('am-form-error');
    options.onValid.call(this, validity);
  },

  markInValid: function(validity) {
    var options = this.options;
    var $field = $(validity.field);
    var $parent = $field.closest('.am-form-group');

    $field.addClass(options.inValidClass + ' ' + options.activeClass).
      removeClass(options.validClass);
    $parent.addClass('am-form-error').removeClass('am-form-success');
    options.onInValid.call(this, validity);
  },

  validate: function(validity) {
    // return validity;
  },

  submit: null
};

Validator.VERSION = '2.4.1';

/* jshint -W101 */
Validator.patterns = {
  email: /^((([a-zA-Z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+(\.([a-zA-Z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+)*)|((\x22)((((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(([\x01-\x08\x0b\x0c\x0e-\x1f\x7f]|\x21|[\x23-\x5b]|[\x5d-\x7e]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(\\([\x01-\x09\x0b\x0c\x0d-\x7f]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF]))))*(((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(\x22)))@((([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.)+(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.?$/,

  url: /^(https?|ftp):\/\/(((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:)*@)?(((\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5]))|((([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.)+(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.?)(:\d*)?)(\/((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)+(\/(([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)*)*)?)?(\?((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|[\uE000-\uF8FF]|\/|\?)*)?(\#((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|\/|\?)*)?$/,

  // Number, including positive, negative, and floating decimal
  number: /^-?(?:\d+|\d{1,3}(?:,\d{3})+)?(?:\.\d+)?$/,
  dateISO: /^\d{4}[\/\-]\d{1,2}[\/\-]\d{1,2}$/,
  integer: /^-?\d+$/
};
/* jshint +W101 */

Validator.validationMessages = {
  zh_CN: {
    valueMissing: '请填写（选择）此字段',
    customError: {
      tooShort: '至少填写 %s 个字符',
      checkedOverflow: '至多选择 %s 项',
      checkedUnderflow: '至少选择 %s 项'
    },
    patternMismatch: '请按照要求的格式填写',
    rangeOverflow: '请填写小于等于 %s 的值',
    rangeUnderflow: '请填写大于等于 %s 的值',
    stepMismatch: '',
    tooLong: '至多填写 %s 个字符',
    typeMismatch: '请按照要求的类型填写'
  }
};

Validator.ERROR_MAP = {
  tooShort: 'minlength',
  checkedOverflow: 'maxchecked',
  checkedUnderflow: 'minchecked',
  rangeOverflow: 'max',
  rangeUnderflow: 'min',
  tooLong: 'maxlength'
};

// TODO: 考虑表单元素不是 form 子元素的情形
// TODO: change/click/focusout 同时触发时处理重复
// TODO: 显示提示信息

Validator.prototype.init = function() {
  var _this = this;
  var $element = this.$element;
  var options = this.options;

  // using H5 form validation if option set and supported
  if (options.H5validation && UI.support.formValidation) {
    return false;
  }

  // disable HTML5 form validation
  $element.attr('novalidate', 'novalidate');

  function regexToPattern(regex) {
    var pattern = regex.toString();
    return pattern.substring(1, pattern.length - 1);
  }

  // add pattern to H5 input type
  $.each(options.H5inputType, function(i, type) {
    var $field = $element.find('input[type=' + type + ']');
    if (!$field.attr('pattern') &&
      !$field.is('[class*=' + options.patternClassPrefix + ']')) {
      $field.attr('pattern', regexToPattern(options.patterns[type]));
    }
  });

  // add pattern to .js-pattern-xx
  $.each(options.patterns, function(key, value) {
    var $field = $element.find('.' + options.patternClassPrefix + key);
    !$field.attr('pattern') && $field.attr('pattern', regexToPattern(value));
  });

  $element.submit(function(e) {
    // user custom submit handler
    if (typeof options.submit === 'function') {
      return options.submit.call(_this, e);
    }

    if (options.validateOnSubmit) {
      var formValidity = _this.isFormValid();

      // sync validate, return result
      if ($.type(formValidity) === 'boolean') {
        return formValidity;
      }

      if ($element.data('amui.checked')) {
        return true;
      } else {
        $.when(formValidity).then(function() {
          // done, submit form
          $element.data('amui.checked', true).submit();
        }, function() {
          // fail
          $element.data('amui.checked', false).
            find('.' + options.inValidClass).eq(0).focus();
        });
        return false;
      }
    }
  });

  function bindEvents(fields, eventFlags, debounce) {
    var events = eventFlags.split(',');
    var validate = function(e) {
      // console.log(e.type);
      _this.validate(this);
    };

    if (debounce) {
      validate = UI.utils.debounce(validate, debounce);
    }

    $.each(events, function(i, event) {
      $element.on(event + '.validator.amui', fields, validate);
    });
  }

  bindEvents(':input', options.customEvents);
  bindEvents(options.keyboardFields, options.keyboardEvents);
  bindEvents(options.pointerFields, options.pointerEvents);

  if (options.textareaMaxlenthKeyup) {
    bindEvents('textarea[maxlength]', 'keyup', 50);
  }

  if (options.activeKeyup) {
    bindEvents('.am-active', 'keyup', 50);
  }

  /*if (options.errorMessage === 'tooltip') {
    this.$tooltip = $('<div></div>', {
      'class': 'am-validator-message',
      id: UI.utils.generateGUID('am-validator-message')
    });

    $(document.body).append(this.$tooltip);
  }*/
};

Validator.prototype.isValid = function(field) {
  var $field = $(field);
  var options = this.options;
  // valid field not has been validated
  if ($field.data('validity') === undefined || options.alwaysRevalidate) {
    this.validate(field);
  }

  return $field.data('validity') && $field.data('validity').valid;
};

Validator.prototype.validate = function(field) {
  var _this = this;
  var $element = this.$element;
  var options = this.options;
  var $field = $(field);

  // Validate equal, e.g. confirm password
  var equalTo = $field.data('equalTo');
  if (equalTo) {
    $field.attr('pattern', '^' + $element.find(equalTo).val() + '$');
  }

  var pattern = $field.attr('pattern') || false;
  var re = new RegExp(pattern);
  var $radioGroup = null;
  var $checkboxGroup = null;
  // if checkbox, return `:chcked` length
  // NOTE: checkbox and radio should have name attribute
  var value = ($field.is('[type=checkbox]')) ?
    ($checkboxGroup = $element.find('input[name="' + field.name + '"]')).
      filter(':checked').length : ($field.is('[type=radio]') ?
  ($radioGroup = this.$element.find('input[name="' + field.name + '"]')).
    filter(':checked').length > 0 : $field.val());

  // if checkbox, valid the first input of checkbox group
  $field = ($checkboxGroup && $checkboxGroup.length) ?
    $checkboxGroup.first() : $field;

  var required = ($field.attr('required') !== undefined) &&
    ($field.attr('required') !== 'false');
  var maxLength = parseInt($field.attr('maxlength'), 10);
  var minLength = parseInt($field.attr('minlength'), 10);
  var min = Number($field.attr('min'));
  var max = Number($field.attr('max'));
  var validity = this.createValidity({field: $field[0], valid: true});

  // Debug
  if (options.debug && window.console) {
    console.log('Validate: value -> [' + value + ', regex -> [' + re +
    '], required -> ' + required);
    console.log('Regex test: ' + re.test(value) + ', Pattern: ' + pattern);
  }

  // check value length
  if (!isNaN(maxLength) && value.length > maxLength) {
    validity.valid = false;
    validity.tooLong = true;
  }

  if (!isNaN(minLength) && value.length < minLength) {
    validity.valid = false;
    validity.customError = 'tooShort';
  }

  // check minimum and maximum
  // https://developer.mozilla.org/en-US/docs/Web/HTML/Element/Input
  // TODO: 日期验证最小值和最大值 min/max
  if (!isNaN(min) && Number(value) < min) {
    validity.valid = false;
    validity.rangeUnderflow = true;
  }

  if (!isNaN(max) && Number(value) > max) {
    validity.valid = false;
    validity.rangeOverflow = true;
  }

  // check required
  if (required && !value) {
    validity.valid = false;
    validity.valueMissing = true;
  } else if (($checkboxGroup || $field.is('select[multiple="multiple"]')) &&
    value) {
    // check checkboxes / multiple select with `minchecked`/`maxchecked` attr
    // var $multipleField = $checkboxGroup ? $checkboxGroup.first() : $field;

    // if is select[multiple="multiple"], return selected length
    value = $checkboxGroup ? value : value.length;

    // at least checked
    var minChecked = parseInt($field.attr('minchecked'), 10);
    // at most checked
    var maxChecked = parseInt($field.attr('maxchecked'), 10);

    if (!isNaN(minChecked) && value < minChecked) {
      // console.log('At least [%d] items checked！', maxChecked);
      validity.valid = false;
      validity.customError = 'checkedUnderflow';
    }

    if (!isNaN(maxChecked) && value > maxChecked) {
      // console.log('At most [%d] items checked！', maxChecked);
      validity.valid = false;
      validity.customError = 'checkedOverflow';
    }
  } else if (pattern && !re.test(value) && value) { // check pattern
    validity.valid = false;
    validity.patternMismatch = true;
  }

  var validateComplete = function(validity) {
    this.markField(validity);

    $field.trigger('validated.field.validator.amui', validity).
      data('validity', validity);

    // validate the radios/checkboxes with the same name
    var $fields = $radioGroup || $checkboxGroup;
    if ($fields) {
      $fields.not($field).data('validity', validity).each(function() {
        validity.field = this;
        _this.markField(validity);
      });
    }

    return validity;
  };

  // Run custom validate
  // NOTE: async custom validate should return Deferred project
  var customValidate;
  (typeof options.validate === 'function') &&
    (customValidate = options.validate.call(this, validity));

  // Deferred
  if (customValidate) {
    var dfd = new $.Deferred();
    $field.data('amui.dfdValidity', dfd.promise());
    return $.when(customValidate).always(function(validity) {
      dfd[validity.valid ? 'resolve' : 'reject'](validity);
      validateComplete.call(_this, validity);
    });
  }

  validateComplete.call(this, validity);
};

Validator.prototype.markField = function(validity) {
  var options = this.options;
  var flag = 'mark' + (validity.valid ? '' : 'In') + 'Valid';
  options[flag] && options[flag].call(this, validity);
};

// check all fields in the form are valid
Validator.prototype.validateForm = function() {
  var _this = this;
  var $element = this.$element;
  var options = this.options;
  var $allFields = $element.find(options.allFields);
  var radioNames = [];
  var valid = true;
  var formValidity = [];
  var $inValidFields = $([]);
  var promises = [];
  // for async validate
  var async = false;

  $element.trigger('validate.form.validator.amui');

  // Filter radio with the same name and keep only one,
  //   since they will be checked as a group by validate()
  var $filteredFields = $allFields.filter(function(index) {
    var name;
    if (this.tagName === 'INPUT' && this.type === 'radio') {
      name = this.name;
      if (radioNames[name] === true) {
        return false;
      }
      radioNames[name] = true;
    }
    return true;
  });

  $filteredFields.each(function() {
    var $this = $(this);
    var fieldValid = _this.isValid(this);
    var fieldValidity = $this.data('validity');

    valid = !!fieldValid && valid;
    formValidity.push(fieldValidity);

    if (!fieldValid) {
      $inValidFields = $inValidFields.add($(this), $element);
    }

    // async validity
    var promise = $this.data('amui.dfdValidity');

    if (promise) {
      promises.push(promise);
      async = true;
    } else {
      // convert sync validity to Promise
      var dfd = new $.Deferred();
      promises.push(dfd.promise());
      dfd[fieldValid ? 'resolve' : 'reject'](fieldValidity);
    }
  });

  // NOTE: If there are async validity, the valid may be not exact result.
  var validity = {
    valid: valid,
    $invalidFields: $inValidFields,
    validity: formValidity,
    promises: promises,
    async: async
  };

  $element.trigger('validated.form.validator.amui', validity);

  return validity;
};

Validator.prototype.isFormValid = function() {
  var _this = this;
  var formValidity = this.validateForm();
  var triggerValid = function(type) {
    _this.$element.trigger(type + '.validator.amui');
  };

  if (formValidity.async) {
    var masterDfd = new $.Deferred();

    $.when.apply(null, formValidity.promises).then(function() {
      masterDfd.resolve();
      triggerValid('valid');
    }, function() {
      masterDfd.reject();
      triggerValid('invalid');
    });

    return masterDfd.promise();
  } else {
    if (!formValidity.valid) {
      formValidity.$invalidFields.first().focus();
      triggerValid('invalid');
      return false;
    }

    triggerValid('valid');
    return true;
  }
};

// customErrors:
//    1. tooShort
//    2. checkedOverflow
//    3. checkedUnderflow
Validator.prototype.createValidity = function(validity) {
  return $.extend({
    customError: validity.customError || false,
    patternMismatch: validity.patternMismatch || false,
    rangeOverflow: validity.rangeOverflow || false, // higher than maximum
    rangeUnderflow: validity.rangeUnderflow || false, // lower than  minimum
    stepMismatch: validity.stepMismatch || false,
    tooLong: validity.tooLong || false,
    // value is not in the correct syntax
    typeMismatch: validity.typeMismatch || false,
    valid: validity.valid || true,
    // Returns true if the element has no value but is a required field
    valueMissing: validity.valueMissing || false
  }, validity);
};

Validator.prototype.getValidationMessage = function(validity) {
  var messages = Validator.validationMessages[this.options.locales];
  var error;
  var message;
  var placeholder = '%s';
  var $field = $(validity.field);

  if ($field.is('[type="checkbox"]') || $field.is('[type="radio"]')) {
    $field = this.$element.find('[name=' + $field.attr('name') + ']').first();
  }

  // get error name
  $.each(validity, function(key, val) {
    // skip `field` and `valid`
    if (key === 'field' || key === 'valid') {
      return key;
    }

    // Amaze UI custom error type
    if (key === 'customError' && val) {
      error = val;
      messages = messages.customError;
      return false;
    }

    // W3C specs error type
    if (val === true) {
      error = key;
      return false;
    }
  });

  message = messages[error] || undefined;

  if (message && Validator.ERROR_MAP[error]) {
    message = message.replace(placeholder,
      $field.attr(Validator.ERROR_MAP[error]) || '规定的');
  }

  return message;
};

// remove valid mark
Validator.prototype.removeMark = function() {
  this.$element.find('.am-form-success, .am-form-error, .am-field-error')
    .removeClass('am-form-success am-form-error am-field-error');
};

UI.plugin('validator', Validator);

// init code
UI.ready(function(context) {
  $('[data-am-validator]', context).validator();
});

module.exports = Validator;
