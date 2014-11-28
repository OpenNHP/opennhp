/*! Amaze UI v2.0.0-beta1 ~ Old IE Fucker | by Amaze UI Team | (c) 2014 AllMobilize, Inc. | Licensed under MIT | 2014-11-28T02:11:52 UTC */
(function e(t, n, r) {
  function s(o, u) {
    if (!n[o]) {
      if (!t[o]) {
        var a = typeof require == "function" && require;
        if (!u && a) return a(o, !0);
        if (i) return i(o, !0);
        var f = new Error("Cannot find module '" + o + "'");
        throw f.code = "MODULE_NOT_FOUND", f
      }
      var l = n[o] = {
        exports: {}
      };
      t[o][0].call(l.exports, function(e) {
        var n = t[o][1][e];
        return s(n ? n : e)
      }, l, l.exports, e, t, n, r)
    }
    return n[o].exports
  }
  var i = typeof require == "function" && require;
  for (var o = 0; o < r.length; o++) s(r[o]);
  return s
})({
  1: [
    function(require, module, exports) {
      (function(global) {
        // Amaze UI JavaScript for IE8

        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);

        require('./core');
        require('./ui.alert');
        require('./ui.button');
        require('./ui.collapse');
        require('./ui.dimmer');
        require('./ui.dropdown');
        require('./ui.flexslider');
        require('./ui.modal');
        require('./ui.offcanvas');
        require('./ui.popover');
        require('./ui.progress');
        require('./ui.scrollspynav');
        require('./ui.sticky');
        require('./util.cookie');

        module.exports = $.AMUI;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2,
      "./ui.alert": 3,
      "./ui.button": 4,
      "./ui.collapse": 5,
      "./ui.dimmer": 6,
      "./ui.dropdown": 7,
      "./ui.flexslider": 8,
      "./ui.modal": 9,
      "./ui.offcanvas": 10,
      "./ui.popover": 11,
      "./ui.progress": 12,
      "./ui.scrollspynav": 13,
      "./ui.sticky": 15,
      "./util.cookie": 16
    }
  ],
  2: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        /* jshint -W040 */

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);

        if (typeof $ === 'undefined') {
          throw new Error('Amaze UI requires jQuery :-(\n' +
            '\u7231\u4e0a\u4e00\u5339\u91ce\u9a6c\uff0c\u53ef\u4f60' +
            '\u7684\u5bb6\u91cc\u6ca1\u6709\u8349\u539f\u2026');
        }

        var UI = $.AMUI || {};
        var $win = $(window);
        var doc = window.document;
        var $html = $('html');

        UI.VERSION = '2.0.0';

        UI.support = {};

        UI.support.transition = (function() {
          var transitionEnd = (function() {
            // https://developer.mozilla.org/en-US/docs/Web/Events/transitionend#Browser_compatibility
            var element = doc.body || doc.documentElement;
            var transEndEventNames = {
              WebkitTransition: 'webkitTransitionEnd',
              MozTransition: 'transitionend',
              OTransition: 'oTransitionEnd otransitionend',
              transition: 'transitionend'
            };
            var name;

            for (name in transEndEventNames) {
              if (element.style[name] !== undefined) {
                return transEndEventNames[name];
              }
            }
          })();

          return transitionEnd && {
            end: transitionEnd
          };
        })();

        UI.support.animation = (function() {
          var animationEnd = (function() {
            var element = doc.body || doc.documentElement;
            var animEndEventNames = {
              WebkitAnimation: 'webkitAnimationEnd',
              MozAnimation: 'animationend',
              OAnimation: 'oAnimationEnd oanimationend',
              animation: 'animationend'
            };
            var name;

            for (name in animEndEventNames) {
              if (element.style[name] !== undefined) {
                return animEndEventNames[name];
              }
            }
          })();

          return animationEnd && {
            end: animationEnd
          };
        })();

        UI.support.requestAnimationFrame = window.requestAnimationFrame ||
          window.webkitRequestAnimationFrame ||
          window.mozRequestAnimationFrame ||
          window.msRequestAnimationFrame ||
          window.oRequestAnimationFrame ||
          function(callback) {
            window.setTimeout(callback, 1000 / 60);
        };

        /* jshint -W069 */
        UI.support.touch = (
          ('ontouchstart' in window &&
            navigator.userAgent.toLowerCase().match(/mobile|tablet/)) ||
          (window.DocumentTouch && document instanceof window.DocumentTouch) ||
          (window.navigator['msPointerEnabled'] &&
            window.navigator['msMaxTouchPoints'] > 0) || //IE 10
          (window.navigator['pointerEnabled'] &&
            window.navigator['maxTouchPoints'] > 0) || //IE >=11
          false);

        // https://developer.mozilla.org/zh-CN/docs/DOM/MutationObserver
        UI.support.mutationobserver = (window.MutationObserver ||
          window.WebKitMutationObserver || window.MozMutationObserver ||
          null);

        UI.utils = {};

        /**
         * Debounce function
         * @param {function} func  Function to be debounced
         * @param {number} wait Function execution threshold in milliseconds
         * @param {bool} immediate  Whether the function should be called at
         *                          the beginning of the delay instead of the
         *                          end. Default is false.
         * @desc Executes a function when it stops being invoked for n seconds
         * @via  _.debounce() http://underscorejs.org
         */
        UI.utils.debounce = function(func, wait, immediate) {
          var timeout;
          return function() {
            var context = this;
            var args = arguments;
            var later = function() {
              timeout = null;
              if (!immediate) {
                func.apply(context, args);
              }
            };
            var callNow = immediate && !timeout;

            clearTimeout(timeout);
            timeout = setTimeout(later, wait);

            if (callNow) {
              func.apply(context, args);
            }
          };
        };

        UI.utils.isInView = function(element, options) {
          var $element = $(element);
          var visible = !!($element.width() || $element.height()) &&
            $element.css('display') !== 'none';

          if (!visible) {
            return false;
          }

          var windowLeft = $win.scrollLeft();
          var windowTop = $win.scrollTop();
          var offset = $element.offset();
          var left = offset.left;
          var top = offset.top;

          options = $.extend({
            topOffset: 0,
            leftOffset: 0
          }, options);

          return (top + $element.height() >= windowTop &&
            top - options.topOffset <= windowTop + $win.height() &&
            left + $element.width() >= windowLeft &&
            left - options.leftOffset <= windowLeft + $win.width());
        };

        /* jshint -W054 */
        UI.utils.parseOptions = UI.utils.options = function(string) {
          if ($.isPlainObject(string)) {
            return string;
          }

          var start = (string ? string.indexOf('{') : -1);
          var options = {};

          if (start != -1) {
            try {
              options = (new Function('',
                'var json = ' + string.substr(start) +
                '; return JSON.parse(JSON.stringify(json));'))();
            } catch (e) {}
          }

          return options;
        };

        /* jshint +W054 */

        UI.utils.generateGUID = function(namespace) {
          var uid = namespace + '-' || 'am-';

          do {
            uid += Math.random().toString(36).substring(2, 7);
          } while (document.getElementById(uid));

          return uid;
        };

        // http://blog.alexmaccaw.com/css-transitions
        $.fn.emulateTransitionEnd = function(duration) {
          var called = false;
          var $el = this;

          $(this).one(UI.support.transition.end, function() {
            called = true;
          });

          var callback = function() {
            if (!called) {
              $($el).trigger(UI.support.transition.end);
            }
            $el.transitionEndTimmer = undefined;
          };
          this.transitionEndTimmer = setTimeout(callback, duration);
          return this;
        };

        $.fn.redraw = function() {
          $(this).each(function() {
            /* jshint unused:false */
            var redraw = this.offsetHeight;
          });
          return this;
        };

        /* jshint unused:true */

        $.fn.transitionEnd = function(callback) {
          var endEvent = UI.support.transition.end;
          var dom = this;

          function fireCallBack(e) {
            callback.call(this, e);
            endEvent && dom.off(endEvent, fireCallBack);
          }

          if (callback && endEvent) {
            dom.on(endEvent, fireCallBack);
          }

          return this;
        };

        $.fn.removeClassRegEx = function() {
          return this.each(function(regex) {
            var classes = $(this).attr('class');

            if (!classes || !regex) {
              return false;
            }

            var classArray = [];
            classes = classes.split(' ');

            for (var i = 0, len = classes.length; i < len; i++) {
              if (!classes[i].match(regex)) {
                classArray.push(classes[i]);
              }
            }

            $(this).attr('class', classArray.join(' '));
          });
        };

        //
        $.fn.alterClass = function(removals, additions) {
          var self = this;

          if (removals.indexOf('*') === -1) {
            // Use native jQuery methods if there is no wildcard matching
            self.removeClass(removals);
            return !additions ? self : self.addClass(additions);
          }

          var classPattern = new RegExp('\\s' +
            removals.replace(/\*/g, '[A-Za-z0-9-_]+').split(' ').join(
              '\\s|\\s') +
            '\\s', 'g');

          self.each(function(i, it) {
            var cn = ' ' + it.className + ' ';
            while (classPattern.test(cn)) {
              cn = cn.replace(classPattern, ' ');
            }
            it.className = $.trim(cn);
          });

          return !additions ? self : self.addClass(additions);
        };

        // handle multiple browsers for requestAnimationFrame()
        // http://www.paulirish.com/2011/requestanimationframe-for-smart-animating/
        // https://github.com/gnarf/jquery-requestAnimationFrame
        UI.utils.rAF = (function() {
          return window.requestAnimationFrame ||
            window.webkitRequestAnimationFrame ||
            window.mozRequestAnimationFrame ||
            window.oRequestAnimationFrame ||
            // if all else fails, use setTimeout
            function(callback) {
              return window.setTimeout(callback, 1000 / 60); // shoot for 60 fps
            };
        })();

        // handle multiple browsers for cancelAnimationFrame()
        UI.utils.cancelAF = (function() {
          return window.cancelAnimationFrame ||
            window.webkitCancelAnimationFrame ||
            window.mozCancelAnimationFrame ||
            window.oCancelAnimationFrame ||
            function(id) {
              window.clearTimeout(id);
            };
        })();

        // via http://davidwalsh.name/detect-scrollbar-width
        UI.utils.measureScrollbar = function() {
          if (document.body.clientWidth >= window.innerWidth) {
            return 0;
          }

          // if ($html.width() >= window.innerWidth) return;
          // var scrollbarWidth = window.innerWidth - $html.width();
          var $measure = $('<div ' +
            'style="width: 100px;height: 100px;overflow: scroll;' +
            'position: absolute;top: -9999px;"></div>');

          $(document.body).append($measure);

          var scrollbarWidth = $measure[0].offsetWidth - $measure[0].clientWidth;

          $measure.remove();

          return scrollbarWidth;
        };

        UI.utils.imageLoader = function($image, callback) {
          function loaded() {
            callback($image[0]);
          }

          function bindLoad() {
            this.one('load', loaded);
            if (/MSIE (\d+\.\d+);/.test(navigator.userAgent)) {
              var src = this.attr('src'),
                param = src.match(/\?/) ? '&' : '?';

              param += 'random=' + (new Date()).getTime();
              this.attr('src', src + param);
            }
          }

          if (!$image.attr('src')) {
            loaded();
            return;
          }

          if ($image[0].complete || $image[0].readyState === 4) {
            loaded();
          } else {
            bindLoad.call($image);
          }
        };

        /**
         * https://github.com/cho45/micro-template.js
         * (c) cho45 http://cho45.github.com/mit-license
         */
        /* jshint -W109 */
        UI.template = function(id, data) {
          var me = UI.template;

          if (!me.cache[id]) {
            me.cache[id] = (function() {
              var name = id;
              var string = /^[\w\-]+$/.test(id) ?
                me.get(id) : (name = 'template(string)', id); // no warnings

              var line = 1;
              var body = ('try { ' + (me.variable ?
                  'var ' + me.variable + ' = this.stash;' :
                  'with (this.stash) { ') +
                "this.ret += '" +
                string.replace(/<%/g, '\x11').replace(/%>/g, '\x13'). // if you want other tag, just edit this line
                replace(/'(?![^\x11\x13]+?\x13)/g, '\\x27').replace(
                  /^\s*|\s*$/g, '').replace(/\n/g, function() {
                  return "';\nthis.line = " + (++line) +
                    "; this.ret += '\\n";
                }).replace(/\x11-(.+?)\x13/g, "' + ($1) + '").replace(
                  /\x11=(.+?)\x13/g, "' + this.escapeHTML($1) + '").replace(
                  /\x11(.+?)\x13/g, "'; $1; this.ret += '") +
                "'; " + (me.variable ? "" : "}") + "return this.ret;" +
                "} catch (e) { throw 'TemplateError: ' + e + ' (on " +
                name +
                "' + ' line ' + this.line + ')'; } " +
                "//@ sourceURL=" + name + "\n" // source map
              ).replace(/this\.ret \+= '';/g, '');
              /* jshint -W054 */
              var func = new Function(body);
              var map = {
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '\x22': '&#x22;',
                '\x27': '&#x27;'
              };
              var escapeHTML = function(string) {
                return ('' + string).replace(/[&<>\'\"]/g, function(_) {
                  return map[_];
                });
              };

              return function(stash) {
                return func.call(me.context = {
                  escapeHTML: escapeHTML,
                  line: 1,
                  ret: '',
                  stash: stash
                });
              };
            })();
          }

          return data ? me.cache[id](data) : me.cache[id];
        };
        /* jshint +W109 */
        /* jshint +W054 */

        UI.template.cache = {};

        UI.template.get = function(id) {
          if (id) {
            var element = document.getElementById(id);
            return element && element.innerHTML || '';
          }
        };

        // Attach FastClick on touch devices
        if (UI.support.touch) {
          $html.addClass('am-touch');

          $(function() {
            var FastClick = $.AMUI.FastClick;
            FastClick && FastClick.attach(document.body);
          });
        }

        $(function() {
          var $body = $('body');

          // trigger DOM ready event
          $(document).trigger('domready.amui');

          $html.removeClass('no-js').addClass('js');

          UI.support.animation && $html.addClass('cssanimations');

          // iOS standalone mode
          if (window.navigator.standalone) {
            $html.addClass('am-standalone');
          }

          $('.am-topbar-fixed-top').length &&
            $body.addClass('am-with-topbar-fixed-top');

          $('.am-topbar-fixed-bottom').length &&
            $body.addClass('am-with-topbar-fixed-bottom');

          // Remove responsive classes in .am-layout
          var $layout = $('.am-layout');
          $layout.find('[class*="md-block-grid"]').alterClass(
            'md-block-grid-*');
          $layout.find('[class*="lg-block-grid"]').alterClass(
            'lg-block-grid');

          // widgets not in .am-layout
          $('[data-am-widget]').each(function() {
            var $widget = $(this);
            // console.log($widget.parents('.am-layout').length)
            if ($widget.parents('.am-layout').length === 0) {
              $widget.addClass('am-no-layout');
            }
          });
        });

        $.AMUI = UI;

        module.exports = UI;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {}
  ],
  3: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

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

          this.$element.
          addClass('am-fade am-in').
          on('click.alert.amui', '.am-close', $.proxy(this.close, this));
        };

        Alert.DEFAULTS = {
          removeElement: true
        };

        Alert.prototype.close = function() {
          var $this = $(this);
          var $target = $this.hasClass('am-alert') ?
            $this :
            $this.parent('.am-alert');

          $target.trigger('close.alert.amui');

          $target.removeClass('am-in');

          function processAlert() {
            $target.trigger('closed.alert.amui').remove();
          }

          UI.support.transition && $target.hasClass('am-fade') ?
            $target.
          one(UI.support.transition.end, processAlert).
          emulateTransitionEnd(200) : processAlert();
        };

        // Alert Plugin
        $.fn.alert = function(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.alert');
            var options = typeof option == 'object' && option;

            if (!data) {
              $this.data('amui.alert', (data = new Alert(this, options || {})));
            }

            if (typeof option == 'string') {
              data[option].call($this);
            }
          });
        };

        // Init code
        $(document).on('click.alert.amui.data-api', '[data-am-alert]',
          function(e) {
            var $target = $(e.target);
            $(this).addClass('am-fade am-in');
            $target.is('.am-close') && $(this).alert('close');
          });

        $.AMUI.alert = Alert;

        module.exports = Alert;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  4: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

        /**
         * @via https://github.com/twbs/bootstrap/blob/master/js/button.js
         * @copyright (c) 2011-2014 Twitter, Inc
         * @license The MIT License
         */

        var Button = function(element, options) {
          this.$element = $(element);
          this.options = $.extend({}, Button.DEFAULTS, options);
          this.isLoading = false;
          this.hasSpinner = false;
        };

        Button.DEFAULTS = {
          loadingText: 'loading...',
          className: {
            loading: 'am-btn-loading',
            disabled: 'am-disabled'
          },
          spinner: undefined
        };

        Button.prototype.setState = function(state) {
          var disabled = 'disabled';
          var $element = this.$element;
          var options = this.options;
          var val = $element.is('input') ? 'val' : 'html';
          var loadingClassName = options.className.disabled + ' ' +
            options.className.loading;

          state = state + 'Text';

          if (!options.resetText) {
            options.resetText = $element[val]();
          }

          // add spinner for element with html()
          if (UI.support.animation && options.spinner &&
            val === 'html' && !this.hasSpinner) {
            options.loadingText = '<span class="am-icon-' +
              options.spinner +
              ' am-icon-spin"></span>' + options.loadingText;

            this.hasSpinner = true;
          }

          $element[val](options[state]);

          // push to event loop to allow forms to submit
          setTimeout($.proxy(function() {
            if (state == 'loadingText') {
              $element.addClass(loadingClassName).attr(disabled,
                disabled);
              this.isLoading = true;
            } else if (this.isLoading) {
              $element.removeClass(loadingClassName).removeAttr(
                disabled);
              this.isLoading = false;
            }
          }, this), 0);
        };

        Button.prototype.toggle = function() {
          var changed = true;
          var $element = this.$element;
          var $parent = this.$element.parent('.am-btn-group');

          if ($parent.length) {
            var $input = this.$element.find('input');

            if ($input.prop('type') == 'radio') {
              if ($input.prop('checked') && $element.hasClass('am-active')) {
                changed = false;
              } else {
                $parent.find('.am-active').removeClass('am-active');
              }
            }

            if (changed) {
              $input.prop('checked', !$element.hasClass('am-active')).trigger(
                'change');
            }
          }

          if (changed) {
            $element.toggleClass('am-active');
            if (!$element.hasClass('am-active')) {
              $element.blur();
            }
          }
        };

        // Button plugin
        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.button');
            var options = typeof option == 'object' && option || {};

            if (!data) {
              $this.data('amui.button', (data = new Button(this,
                options)));
            }

            if (option == 'toggle') {
              data.toggle();
            } else if (typeof option == 'string') {
              data.setState(option);
            }
          });
        }

        $.fn.button = Plugin;

        // Init code
        $(document).on('click.button.amui.data-api', '[data-am-button]',
          function(e) {
            var $btn = $(e.target);

            if (!$btn.hasClass('am-btn')) {
              $btn = $btn.closest('.am-btn');
            }

            Plugin.call($btn, 'toggle');
            e.preventDefault();
          });

        $(function() {
          $('[data-am-loading]').each(function() {
            $(this).button(UI.utils.parseOptions($(this).data(
              'amLoading')));
          });
        });

        $.AMUI.button = Button;

        module.exports = Button;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  5: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

        /**
         * @via https://github.com/twbs/bootstrap/blob/master/js/collapse.js
         * @copyright (c) 2011-2014 Twitter, Inc
         * @license The MIT License
         */

        var Collapse = function(element, options) {
          this.$element = $(element);
          this.options = $.extend({}, Collapse.DEFAULTS, options);
          this.transitioning = null;

          if (this.options.parent) {
            this.$parent = $(this.options.parent);
          }

          if (this.options.toggle) {
            this.toggle();
          }
        };

        Collapse.DEFAULTS = {
          toggle: true
        };

        Collapse.prototype.open = function() {
          if (this.transitioning || this.$element.hasClass('am-in')) {
            return;
          }

          var startEvent = $.Event('open.collapse.amui');
          this.$element.trigger(startEvent);

          if (startEvent.isDefaultPrevented()) {
            return;
          }

          var actives = this.$parent && this.$parent.find(
            '> .am-panel > .am-in');

          if (actives && actives.length) {
            var hasData = actives.data('amui.collapse');

            if (hasData && hasData.transitioning) {
              return;
            }

            Plugin.call(actives, 'close');

            hasData || actives.data('amui.collapse', null);
          }

          this.$element
            .removeClass('am-collapse')
            .addClass('am-collapsing').height(0);

          this.transitioning = 1;

          var complete = function() {
            this.$element.
            removeClass('am-collapsing').
            addClass('am-collapse am-in').
            height('');
            this.transitioning = 0;
            this.$element.trigger('opened.collapse.amui');
          };

          if (!UI.support.transition) {
            return complete.call(this);
          }

          var scrollHeight = this.$element[0].scrollHeight;

          this.$element
            .one(UI.support.transition.end, $.proxy(complete, this))
            .emulateTransitionEnd(300).
          css({
            height: scrollHeight
          }); // 当折叠的容器有 padding 时，如果用 height() 只能设置内容的宽度
        };

        Collapse.prototype.close = function() {
          if (this.transitioning || !this.$element.hasClass('am-in')) {
            return;
          }

          var startEvent = $.Event('close.collapse.amui');
          this.$element.trigger(startEvent);

          if (startEvent.isDefaultPrevented()) {
            return;
          }

          this.$element.height(this.$element.height()).redraw();

          this.$element.addClass('am-collapsing').
          removeClass('am-collapse am-in');

          this.transitioning = 1;

          var complete = function() {
            this.transitioning = 0;
            this.$element.trigger('closed.collapse.amui').
            removeClass('am-collapsing').
            addClass('am-collapse');
            // css({height: '0'});
          };

          if (!UI.support.transition) {
            return complete.call(this);
          }

          this.$element.height(0)
            .one(UI.support.transition.end, $.proxy(complete, this))
            .emulateTransitionEnd(300);
        };

        Collapse.prototype.toggle = function() {
          this[this.$element.hasClass('am-in') ? 'close' : 'open']();
        };

        // Collapse Plugin
        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.collapse');
            var options = $.extend({}, Collapse.DEFAULTS,
              UI.utils.options($this.attr('data-am-collapse')),
              typeof option == 'object' && option);

            if (!data && options.toggle && option == 'open') {
              option = !option;
            }
            if (!data) {
              $this.data('amui.collapse', (data = new Collapse(this,
                options)));
            }
            if (typeof option == 'string') {
              data[option]();
            }
          });
        }

        $.fn.collapse = Plugin;

        // Init code
        $(document).on('click.collapse.amui.data-api', '[data-am-collapse]',
          function(e) {
            var href;
            var $this = $(this);
            var options = UI.utils.options($this.attr('data-am-collapse'));
            var target = options.target ||
              e.preventDefault() ||
              (href = $this.attr('href')) &&
              href.replace(/.*(?=#[^\s]+$)/, '');
            var $target = $(target);
            var data = $target.data('amui.collapse');
            var option = data ? 'toggle' : options;
            var parent = options.parent;
            var $parent = parent && $(parent);

            if (!data || !data.transitioning) {
              if ($parent) {
                // '[data-am-collapse*="{parent: \'' + parent + '"]
                $parent.find('[data-am-collapse]').not($this).addClass(
                  'am-collapsed');
              }

              $this[$target.hasClass('am-in') ? 'addClass' :
                'removeClass']('am-collapsed');
            }

            Plugin.call($target, option);
          });

        $.AMUI.collapse = Collapse;

        module.exports = Collapse;

        // TODO: 更好的 target 选择方式
        //       折叠的容器必须没有 border/padding 才能正常处理，否则动画会有一些小问题
        //       寻找更好的未知高度 transition 动画解决方案，max-height 之类的就算了

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  6: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        var $doc = $(document);
        var transition = UI.support.transition;

        var Dimmer = function() {
          this.id = UI.utils.generateGUID('am-dimmer');
          this.$element = $(Dimmer.DEFAULTS.tpl, {
            id: this.id
          });

          this.inited = false;
          this.scrollbarWidth = 0;
          this.used = $([]);
        };

        Dimmer.DEFAULTS = {
          tpl: '<div class="am-dimmer" data-am-dimmer></div>'
        };

        Dimmer.prototype.init = function() {
          if (!this.inited) {
            $(document.body).append(this.$element);
            this.inited = true;
            $doc.trigger('init.dimmer.amui');
          }

          return this;
        };

        Dimmer.prototype.open = function(relatedElement) {
          if (!this.inited) {
            this.init();
          }

          var $element = this.$element;

          // 用于多重调用
          if (relatedElement) {
            this.used = this.used.add($(relatedElement));
          }

          this.checkScrollbar().setScrollbar();

          $element.show().trigger('open.dimmer.amui');

          setTimeout(function() {
            $element.addClass('am-active');
          }, 0);

          return this;
        };

        Dimmer.prototype.close = function(relatedElement, force) {
          this.used = this.used.not($(relatedElement));

          if (!force && this.used.length) {
            return this;
          }

          var $element = this.$element;

          $element.removeClass('am-active').trigger('close.dimmer.amui');

          function complete() {
            this.resetScrollbar();
            $element.hide();
          }

          transition ? $element.one(transition.end, $.proxy(complete, this)) :
            complete.call(this);

          return this;
        };

        Dimmer.prototype.checkScrollbar = function() {
          this.scrollbarWidth = UI.utils.measureScrollbar();

          return this;
        };

        Dimmer.prototype.setScrollbar = function() {
          var $body = $(document.body);
          var bodyPaddingRight = parseInt(($body.css('padding-right') || 0),
            10);

          if (this.scrollbarWidth) {
            $body.css('padding-right', bodyPaddingRight + this.scrollbarWidth);
          }

          $body.addClass('am-dimmer-active');

          return this;
        };

        Dimmer.prototype.resetScrollbar = function() {
          $(document.body).css('padding-right', '').removeClass(
            'am-dimmer-active');

          return this;
        };

        var dimmer = new Dimmer();

        $.AMUI.dimmer = dimmer;

        module.exports = dimmer;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  7: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        var animation = UI.support.animation;

        /**
         * @via https://github.com/Minwe/bootstrap/blob/master/js/dropdown.js
         * @copyright (c) 2011-2014 Twitter, Inc
         * @license The MIT License
         */

        // var toggle = '[data-am-dropdown] > .am-dropdown-toggle';

        var Dropdown = function(element, options) {
          this.options = $.extend({}, Dropdown.DEFAULTS, options);

          options = this.options;

          this.$element = $(element);
          this.$toggle = this.$element.find(options.selector.toggle);
          this.$dropdown = this.$element.find(options.selector.dropdown);
          this.$boundary = (options.boundary === window) ? $(window) :
            this.$element.closest(options.boundary);
          this.$justify = (options.justify && $(options.justify).length &&
            $(options.justify)) || undefined;

          !this.$boundary.length && (this.$boundary = $(window));

          this.active = this.$element.hasClass('am-active') ? true :
            false;
          this.animating = null;

          this.events();
        };

        Dropdown.DEFAULTS = {
          animation: 'am-animation-slide-top-fixed',
          boundary: window,
          justify: undefined,
          selector: {
            dropdown: '.am-dropdown-content',
            toggle: '.am-dropdown-toggle'
          },
          trigger: 'click'
        };

        Dropdown.prototype.toggle = function() {
          this.clear();

          if (this.animating) {
            return;
          }

          this[this.active ? 'close' : 'open']();
        };

        Dropdown.prototype.open = function(e) {
          var $toggle = this.$toggle;
          var $element = this.$element;
          var $dropdown = this.$dropdown;

          if ($toggle.is('.am-disabled, :disabled')) {
            return;
          }

          if (this.active) {
            return;
          }

          $element.trigger('open.dropdown.amui').addClass('am-active');

          $toggle.trigger('focus');

          this.checkDimensions();

          var complete = $.proxy(function() {
            $element.trigger('opened.dropdown.amui');
            this.active = true;
            this.animating = 0;
          }, this);

          if (animation) {
            this.animating = 1;
            $dropdown.addClass(this.options.animation).
            on(animation.end + '.open.dropdown.amui', $.proxy(function() {
              complete();
              $dropdown.removeClass(this.options.animation);
            }, this));
          } else {
            complete();
          }
        };

        Dropdown.prototype.close = function() {
          if (!this.active) {
            return;
          }

          var animationName = this.options.animation +
            ' am-animation-reverse';
          var $element = this.$element;
          var $dropdown = this.$dropdown;

          $element.trigger('close.dropdown.amui');

          var complete = $.proxy(function complete() {
            $element.
            removeClass('am-active').
            trigger('closed.dropdown.amui');
            this.active = false;
            this.animating = 0;
            this.$toggle.blur();
          }, this);

          if (animation) {
            $dropdown.addClass(animationName);
            this.animating = 1;
            // animation
            $dropdown.one(animation.end + '.close.dropdown.amui', function() {
              complete();
              $dropdown.removeClass(animationName);
            });
          } else {
            complete();
          }
        };

        Dropdown.prototype.checkDimensions = function() {
          if (!this.$dropdown.length) {
            return;
          }

          var $dropdown = this.$dropdown;
          var offset = $dropdown.offset();
          var width = $dropdown.outerWidth();
          var boundaryWidth = this.$boundary.width();
          var boundaryOffset = $.isWindow(this.boundary) && this.$boundary.offset() ?
            this.$boundary.offset().left : 0;

          if (this.$justify) {
            $dropdown.css({
              'min-width': this.$justify.width()
            });
          }

          if ((width + (offset.left - boundaryOffset)) > boundaryWidth) {
            this.$element.addClass('am-dropdown-flip');
          }
        };

        Dropdown.prototype.clear = function() {
          $('[data-am-dropdown]').not(this.$element).each(function() {
            var data = $(this).data('amui.dropdown');
            data && data['close']();
          });
        };

        Dropdown.prototype.events = function() {
          var eventNS = 'dropdown.amui';
          // triggers = this.options.trigger.split(' '),
          var $toggle = this.$toggle;

          $toggle.on('click.' + eventNS, $.proxy(this.toggle, this));

          /*for (var i = triggers.length; i--;) {
   var trigger = triggers[i];

   if (trigger === 'click') {
   $toggle.on('click.' + eventNS, $.proxy(this.toggle, this))
   }

   if (trigger === 'focus' || trigger === 'hover') {
   var eventIn  = trigger == 'hover' ? 'mouseenter' : 'focusin';
   var eventOut = trigger == 'hover' ? 'mouseleave' : 'focusout';

   this.$element.on(eventIn + '.' + eventNS, $.proxy(this.open, this))
   .on(eventOut + '.' + eventNS, $.proxy(this.close, this));
   }
   }*/

          $(document).on('keydown.dropdown.amui', $.proxy(function(e) {
            e.keyCode === 27 && this.active && this.close();
          }, this)).on('click.outer.dropdown.amui', $.proxy(function(e) {
            // var $target = $(e.target);

            if (this.active &&
              (this.$element[0] === e.target || !this.$element.find(e.target)
                .length)) {
              this.close();
            }
          }, this));
        };

        // Dropdown Plugin
        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('amui.dropdown');
            var options = $.extend({},
              UI.utils.parseOptions($this.attr('data-am-dropdown')),
              typeof option == 'object' && option);

            if (!data) {
              $this.data('amui.dropdown', (data = new Dropdown(this,
                options)));
            }

            if (typeof option == 'string') {
              data[option]();
            }
          });
        }

        $.fn.dropdown = Plugin;

        // Init code
        $(function() {
          $('[data-am-dropdown]').dropdown();
        });

        $(document).on('click.dropdown.amui.data-api', '.am-dropdown form',
          function(e) {
            e.stopPropagation();
          });

        $.AMUI.dropdown = Dropdown;

        module.exports = Dropdown;

        // TODO: 1. 处理链接 focus
        //       2. 增加 mouseenter / mouseleave 选项
        //       3. 宽度适应

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  8: [
    function(require, module, exports) {
      (function(global) {
        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

        // MODIFIED:
        // - LINE 226: add `<i></i>`
        // - namespace
        // - Init code
        // TODO: start after x ms when pause on actions

        /*
         * jQuery FlexSlider v2.2.2
         * Copyright 2012 WooThemes
         * Contributing Author: Tyler Smith
         */

        // FlexSlider: Object Instance
        $.flexslider = function(el, options) {
          var slider = $(el);

          // making variables public
          slider.vars = $.extend({}, $.flexslider.defaults, options);

          var namespace = slider.vars.namespace,
            msGesture = window.navigator && window.navigator.msPointerEnabled &&
            window.MSGesture,
            touch = (("ontouchstart" in window) || msGesture || window.DocumentTouch &&
              document instanceof DocumentTouch) && slider.vars.touch,
            // depricating this idea, as devices are being released with both of these events
            //eventType = (touch) ? "touchend" : "click",
            eventType = "click touchend MSPointerUp keyup",
            watchedEvent = "",
            watchedEventClearTimer,
            vertical = slider.vars.direction === "vertical",
            reverse = slider.vars.reverse,
            carousel = (slider.vars.itemWidth > 0),
            fade = slider.vars.animation === "fade",
            asNav = slider.vars.asNavFor !== "",
            methods = {},
            focused = true;

          // Store a reference to the slider object
          $.data(el, 'flexslider', slider);

          // Private slider methods
          methods = {
            init: function() {
              slider.animating = false;
              // Get current slide and make sure it is a number
              slider.currentSlide = parseInt((slider.vars.startAt ? slider.vars
                .startAt : 0), 10);
              if (isNaN(slider.currentSlide)) {
                slider.currentSlide = 0;
              }
              slider.animatingTo = slider.currentSlide;
              slider.atEnd = (slider.currentSlide === 0 || slider.currentSlide ===
                slider.last);
              slider.containerSelector = slider.vars.selector.substr(0,
                slider.vars.selector.search(' '));
              slider.slides = $(slider.vars.selector, slider);
              slider.container = $(slider.containerSelector, slider);
              slider.count = slider.slides.length;
              // SYNC:
              slider.syncExists = $(slider.vars.sync).length > 0;
              // SLIDE:
              if (slider.vars.animation === "slide") slider.vars.animation =
                "swing";
              slider.prop = (vertical) ? "top" : "marginLeft";
              slider.args = {};
              // SLIDESHOW:
              slider.manualPause = false;
              slider.stopped = false;
              //PAUSE WHEN INVISIBLE
              slider.started = false;
              slider.startTimeout = null;
              // TOUCH/USECSS:
              slider.transitions = !slider.vars.video && !fade && slider.vars
                .useCSS && (function() {
                  var obj = document.createElement('div'),
                    props = ['perspectiveProperty', 'WebkitPerspective',
                      'MozPerspective', 'OPerspective', 'msPerspective'
                    ];
                  for (var i in props) {
                    if (obj.style[props[i]] !== undefined) {
                      slider.pfx = props[i].replace('Perspective', '').toLowerCase();
                      slider.prop = "-" + slider.pfx + "-transform";
                      return true;
                    }
                  }
                  return false;
                }());
              slider.ensureAnimationEnd = '';
              // CONTROLSCONTAINER:
              if (slider.vars.controlsContainer !== "") slider.controlsContainer =
                $(slider.vars.controlsContainer).length > 0 && $(slider.vars
                  .controlsContainer);
              // MANUAL:
              if (slider.vars.manualControls !== "") slider.manualControls =
                $(slider.vars.manualControls).length > 0 && $(slider.vars.manualControls);

              // RANDOMIZE:
              if (slider.vars.randomize) {
                slider.slides.sort(function() {
                  return (Math.round(Math.random()) - 0.5);
                });
                slider.container.empty().append(slider.slides);
              }

              slider.doMath();

              // INIT
              slider.setup("init");

              // CONTROLNAV:
              if (slider.vars.controlNav) methods.controlNav.setup();

              // DIRECTIONNAV:
              if (slider.vars.directionNav) methods.directionNav.setup();

              // KEYBOARD:
              if (slider.vars.keyboard && ($(slider.containerSelector).length ===
                1 || slider.vars.multipleKeyboard)) {
                $(document).bind('keyup', function(event) {
                  var keycode = event.keyCode;
                  if (!slider.animating && (keycode === 39 || keycode ===
                    37)) {
                    var target = (keycode === 39) ? slider.getTarget(
                        'next') :
                      (keycode === 37) ? slider.getTarget('prev') : false;
                    slider.flexAnimate(target, slider.vars.pauseOnAction);
                  }
                });
              }
              // MOUSEWHEEL:
              if (slider.vars.mousewheel) {
                slider.bind('mousewheel', function(event, delta, deltaX,
                  deltaY) {
                  event.preventDefault();
                  var target = (delta < 0) ? slider.getTarget('next') :
                    slider.getTarget('prev');
                  slider.flexAnimate(target, slider.vars.pauseOnAction);
                });
              }

              // PAUSEPLAY
              if (slider.vars.pausePlay) methods.pausePlay.setup();

              //PAUSE WHEN INVISIBLE
              if (slider.vars.slideshow && slider.vars.pauseInvisible)
                methods.pauseInvisible.init();

              // SLIDSESHOW
              if (slider.vars.slideshow) {
                if (slider.vars.pauseOnHover) {
                  slider.hover(function() {
                    if (!slider.manualPlay && !slider.manualPause) slider
                      .pause();
                  }, function() {
                    if (!slider.manualPause && !slider.manualPlay && !
                      slider.stopped) slider.play();
                  });
                }
                // initialize animation
                //If we're visible, or we don't use PageVisibility API
                if (!slider.vars.pauseInvisible || !methods.pauseInvisible.isHidden()) {
                  (slider.vars.initDelay > 0) ? slider.startTimeout =
                    setTimeout(slider.play, slider.vars.initDelay) : slider
                    .play();
                }
              }

              // ASNAV:
              if (asNav) methods.asNav.setup();

              // TOUCH
              if (touch && slider.vars.touch) methods.touch();

              // FADE&&SMOOTHHEIGHT || SLIDE:
              if (!fade || (fade && slider.vars.smoothHeight)) $(window).bind(
                "resize orientationchange focus", methods.resize);

              slider.find("img").attr("draggable", "false");

              // API: start() Callback
              setTimeout(function() {
                slider.vars.start(slider);
              }, 200);
            },
            asNav: {
              setup: function() {
                slider.asNav = true;
                slider.animatingTo = Math.floor(slider.currentSlide /
                  slider.move);
                slider.currentItem = slider.currentSlide;
                slider.slides.removeClass(namespace + "active-slide").eq(
                  slider.currentItem).addClass(namespace + "active-slide");
                if (!msGesture) {
                  slider.slides.on(eventType, function(e) {
                    e.preventDefault();
                    var $slide = $(this),
                      target = $slide.index();
                    var posFromLeft = $slide.offset().left - $(slider).scrollLeft(); // Find position of slide relative to left of slider container
                    if (posFromLeft <= 0 && $slide.hasClass(namespace +
                      'active-slide')) {
                      slider.flexAnimate(slider.getTarget("prev"), true);
                    } else if (!$(slider.vars.asNavFor).data('flexslider')
                      .animating && !$slide.hasClass(namespace +
                        "active-slide")) {
                      slider.direction = (slider.currentItem < target) ?
                        "next" : "prev";
                      slider.flexAnimate(target, slider.vars.pauseOnAction,
                        false, true, true);
                    }
                  });
                } else {
                  el._slider = slider;
                  slider.slides.each(function() {
                    var that = this;
                    that._gesture = new MSGesture();
                    that._gesture.target = that;
                    that.addEventListener("MSPointerDown", function(e) {
                      e.preventDefault();
                      if (e.currentTarget._gesture)
                        e.currentTarget._gesture.addPointer(e.pointerId);
                    }, false);
                    that.addEventListener("MSGestureTap", function(e) {
                      e.preventDefault();
                      var $slide = $(this),
                        target = $slide.index();
                      if (!$(slider.vars.asNavFor).data('flexslider').animating &&
                        !$slide.hasClass('active')) {
                        slider.direction = (slider.currentItem < target) ?
                          "next" : "prev";
                        slider.flexAnimate(target, slider.vars.pauseOnAction,
                          false, true, true);
                      }
                    });
                  });
                }
              }
            },
            controlNav: {
              setup: function() {
                if (!slider.manualControls) {
                  methods.controlNav.setupPaging();
                } else { // MANUALCONTROLS:
                  methods.controlNav.setupManual();
                }
              },
              setupPaging: function() {
                var type = (slider.vars.controlNav === "thumbnails") ?
                  'control-thumbs' : 'control-paging',
                  j = 1,
                  item,
                  slide;

                slider.controlNavScaffold = $('<ol class="' + namespace +
                  'control-nav ' + namespace + type + '"></ol>');

                if (slider.pagingCount > 1) {
                  for (var i = 0; i < slider.pagingCount; i++) {
                    slide = slider.slides.eq(i);
                    item = (slider.vars.controlNav === "thumbnails") ?
                      '<img src="' + slide.attr('data-thumb') + '"/>' :
                      '<a>' + j + '</a>';
                    if ('thumbnails' === slider.vars.controlNav && true ===
                      slider.vars.thumbCaptions) {
                      var captn = slide.attr('data-thumbcaption');
                      if ('' != captn && undefined != captn) item +=
                        '<span class="' + namespace + 'caption">' + captn +
                        '</span>';
                    }
                    // slider.controlNavScaffold.append('<li>' + item + '</li>');
                    slider.controlNavScaffold.append('<li>' + item +
                      '<i></i></li>');
                    j++;
                  }
                }

                // CONTROLSCONTAINER:
                (slider.controlsContainer) ? $(slider.controlsContainer).append(
                  slider.controlNavScaffold) : slider.append(slider.controlNavScaffold);
                methods.controlNav.set();

                methods.controlNav.active();

                slider.controlNavScaffold.delegate('a, img', eventType,
                  function(event) {
                    event.preventDefault();

                    if (watchedEvent === "" || watchedEvent === event.type) {
                      var $this = $(this),
                        target = slider.controlNav.index($this);

                      if (!$this.hasClass(namespace + 'active')) {
                        slider.direction = (target > slider.currentSlide) ?
                          "next" : "prev";
                        slider.flexAnimate(target, slider.vars.pauseOnAction);
                      }
                    }

                    // setup flags to prevent event duplication
                    if (watchedEvent === "") {
                      watchedEvent = event.type;
                    }
                    methods.setToClearWatchedEvent();

                  });
              },
              setupManual: function() {
                slider.controlNav = slider.manualControls;
                methods.controlNav.active();

                slider.controlNav.bind(eventType, function(event) {
                  event.preventDefault();

                  if (watchedEvent === "" || watchedEvent === event.type) {
                    var $this = $(this),
                      target = slider.controlNav.index($this);

                    if (!$this.hasClass(namespace + 'active')) {
                      (target > slider.currentSlide) ? slider.direction =
                        "next" : slider.direction = "prev";
                      slider.flexAnimate(target, slider.vars.pauseOnAction);
                    }
                  }

                  // setup flags to prevent event duplication
                  if (watchedEvent === "") {
                    watchedEvent = event.type;
                  }
                  methods.setToClearWatchedEvent();
                });
              },
              set: function() {
                var selector = (slider.vars.controlNav === "thumbnails") ?
                  'img' : 'a';
                slider.controlNav = $('.' + namespace + 'control-nav li ' +
                  selector, (slider.controlsContainer) ? slider.controlsContainer :
                  slider);
              },
              active: function() {
                slider.controlNav.removeClass(namespace + "active").eq(
                  slider.animatingTo).addClass(namespace + "active");
              },
              update: function(action, pos) {
                if (slider.pagingCount > 1 && action === "add") {
                  slider.controlNavScaffold.append($('<li><a>' + slider.count +
                    '</a></li>'));
                } else if (slider.pagingCount === 1) {
                  slider.controlNavScaffold.find('li').remove();
                } else {
                  slider.controlNav.eq(pos).closest('li').remove();
                }
                methods.controlNav.set();
                (slider.pagingCount > 1 && slider.pagingCount !== slider.controlNav
                  .length) ? slider.update(pos, action) : methods.controlNav
                  .active();
              }
            },
            directionNav: {
              setup: function() {
                var directionNavScaffold = $('<ul class="' + namespace +
                  'direction-nav"><li><a class="' + namespace +
                  'prev" href="#">' + slider.vars.prevText +
                  '</a></li><li><a class="' + namespace +
                  'next" href="#">' + slider.vars.nextText +
                  '</a></li></ul>');

                // CONTROLSCONTAINER:
                if (slider.controlsContainer) {
                  $(slider.controlsContainer).append(directionNavScaffold);
                  slider.directionNav = $('.' + namespace +
                    'direction-nav li a', slider.controlsContainer);
                } else {
                  slider.append(directionNavScaffold);
                  slider.directionNav = $('.' + namespace +
                    'direction-nav li a', slider);
                }

                methods.directionNav.update();

                slider.directionNav.bind(eventType, function(event) {
                  event.preventDefault();
                  var target;

                  if (watchedEvent === "" || watchedEvent === event.type) {
                    target = ($(this).hasClass(namespace + 'next')) ?
                      slider.getTarget('next') : slider.getTarget('prev');
                    slider.flexAnimate(target, slider.vars.pauseOnAction);
                  }

                  // setup flags to prevent event duplication
                  if (watchedEvent === "") {
                    watchedEvent = event.type;
                  }
                  methods.setToClearWatchedEvent();
                });
              },
              update: function() {
                var disabledClass = namespace + 'disabled';
                if (slider.pagingCount === 1) {
                  slider.directionNav.addClass(disabledClass).attr(
                    'tabindex', '-1');
                } else if (!slider.vars.animationLoop) {
                  if (slider.animatingTo === 0) {
                    slider.directionNav.removeClass(disabledClass).filter(
                      '.' + namespace + "prev").addClass(disabledClass).attr(
                      'tabindex', '-1');
                  } else if (slider.animatingTo === slider.last) {
                    slider.directionNav.removeClass(disabledClass).filter(
                      '.' + namespace + "next").addClass(disabledClass).attr(
                      'tabindex', '-1');
                  } else {
                    slider.directionNav.removeClass(disabledClass).removeAttr(
                      'tabindex');
                  }
                } else {
                  slider.directionNav.removeClass(disabledClass).removeAttr(
                    'tabindex');
                }
              }
            },
            pausePlay: {
              setup: function() {
                var pausePlayScaffold = $('<div class="' + namespace +
                  'pauseplay"><a></a></div>');

                // CONTROLSCONTAINER:
                if (slider.controlsContainer) {
                  slider.controlsContainer.append(pausePlayScaffold);
                  slider.pausePlay = $('.' + namespace + 'pauseplay a',
                    slider.controlsContainer);
                } else {
                  slider.append(pausePlayScaffold);
                  slider.pausePlay = $('.' + namespace + 'pauseplay a',
                    slider);
                }

                methods.pausePlay.update((slider.vars.slideshow) ?
                  namespace + 'pause' : namespace + 'play');

                slider.pausePlay.bind(eventType, function(event) {
                  event.preventDefault();

                  if (watchedEvent === "" || watchedEvent === event.type) {
                    if ($(this).hasClass(namespace + 'pause')) {
                      slider.manualPause = true;
                      slider.manualPlay = false;
                      slider.pause();
                    } else {
                      slider.manualPause = false;
                      slider.manualPlay = true;
                      slider.play();
                    }
                  }

                  // setup flags to prevent event duplication
                  if (watchedEvent === "") {
                    watchedEvent = event.type;
                  }
                  methods.setToClearWatchedEvent();
                });
              },
              update: function(state) {
                (state === "play") ? slider.pausePlay.removeClass(namespace +
                  'pause').addClass(namespace + 'play').html(slider.vars.playText) :
                  slider.pausePlay.removeClass(namespace + 'play').addClass(
                    namespace + 'pause').html(slider.vars.pauseText);
              }
            },
            touch: function() {
              var startX,
                startY,
                offset,
                cwidth,
                dx,
                startT,
                scrolling = false,
                localX = 0,
                localY = 0,
                accDx = 0;

              if (!msGesture) {
                el.addEventListener('touchstart', onTouchStart, false);

                function onTouchStart(e) {
                  if (slider.animating) {
                    e.preventDefault();
                  } else if ((window.navigator.msPointerEnabled) || e.touches
                    .length === 1) {
                    slider.pause();
                    // CAROUSEL:
                    cwidth = (vertical) ? slider.h : slider.w;
                    startT = Number(new Date());
                    // CAROUSEL:

                    // Local vars for X and Y points.
                    localX = e.touches[0].pageX;
                    localY = e.touches[0].pageY;

                    offset = (carousel && reverse && slider.animatingTo ===
                      slider.last) ? 0 :
                      (carousel && reverse) ? slider.limit - (((slider.itemW +
                      slider.vars.itemMargin) * slider.move) * slider.animatingTo) :
                      (carousel && slider.currentSlide === slider.last) ?
                      slider.limit :
                      (carousel) ? ((slider.itemW + slider.vars.itemMargin) *
                        slider.move) * slider.currentSlide :
                      (reverse) ? (slider.last - slider.currentSlide +
                        slider.cloneOffset) * cwidth : (slider.currentSlide +
                        slider.cloneOffset) * cwidth;
                    startX = (vertical) ? localY : localX;
                    startY = (vertical) ? localX : localY;

                    el.addEventListener('touchmove', onTouchMove, false);
                    el.addEventListener('touchend', onTouchEnd, false);
                  }
                }

                function onTouchMove(e) {
                  // Local vars for X and Y points.

                  localX = e.touches[0].pageX;
                  localY = e.touches[0].pageY;

                  dx = (vertical) ? startX - localY : startX - localX;
                  scrolling = (vertical) ? (Math.abs(dx) < Math.abs(localX -
                    startY)) : (Math.abs(dx) < Math.abs(localY - startY));

                  var fxms = 500;

                  if (!scrolling || Number(new Date()) - startT > fxms) {
                    e.preventDefault();
                    if (!fade && slider.transitions) {
                      if (!slider.vars.animationLoop) {
                        dx = dx / ((slider.currentSlide === 0 && dx < 0 ||
                            slider.currentSlide === slider.last && dx > 0) ?
                          (Math.abs(dx) / cwidth + 2) : 1);
                      }
                      slider.setProps(offset + dx, "setTouch");
                    }
                  }
                }

                function onTouchEnd(e) {
                  // finish the touch by undoing the touch session
                  el.removeEventListener('touchmove', onTouchMove, false);

                  if (slider.animatingTo === slider.currentSlide && !
                    scrolling && !(dx === null)) {
                    var updateDx = (reverse) ? -dx : dx,
                      target = (updateDx > 0) ? slider.getTarget('next') :
                      slider.getTarget('prev');

                    if (slider.canAdvance(target) && (Number(new Date()) -
                      startT < 550 && Math.abs(updateDx) > 50 || Math.abs(
                        updateDx) > cwidth / 2)) {
                      slider.flexAnimate(target, slider.vars.pauseOnAction);
                    } else {
                      if (!fade) slider.flexAnimate(slider.currentSlide,
                        slider.vars.pauseOnAction, true);
                    }
                  }
                  el.removeEventListener('touchend', onTouchEnd, false);

                  startX = null;
                  startY = null;
                  dx = null;
                  offset = null;
                }
              } else {
                el.style.msTouchAction = "none";
                el._gesture = new MSGesture();
                el._gesture.target = el;
                el.addEventListener("MSPointerDown", onMSPointerDown, false);
                el._slider = slider;
                el.addEventListener("MSGestureChange", onMSGestureChange,
                  false);
                el.addEventListener("MSGestureEnd", onMSGestureEnd, false);

                function onMSPointerDown(e) {
                  e.stopPropagation();
                  if (slider.animating) {
                    e.preventDefault();
                  } else {
                    slider.pause();
                    el._gesture.addPointer(e.pointerId);
                    accDx = 0;
                    cwidth = (vertical) ? slider.h : slider.w;
                    startT = Number(new Date());
                    // CAROUSEL:

                    offset = (carousel && reverse && slider.animatingTo ===
                      slider.last) ? 0 :
                      (carousel && reverse) ? slider.limit - (((slider.itemW +
                      slider.vars.itemMargin) * slider.move) * slider.animatingTo) :
                      (carousel && slider.currentSlide === slider.last) ?
                      slider.limit :
                      (carousel) ? ((slider.itemW + slider.vars.itemMargin) *
                        slider.move) * slider.currentSlide :
                      (reverse) ? (slider.last - slider.currentSlide +
                        slider.cloneOffset) * cwidth : (slider.currentSlide +
                        slider.cloneOffset) * cwidth;
                  }
                }

                function onMSGestureChange(e) {
                  e.stopPropagation();
                  var slider = e.target._slider;
                  if (!slider) {
                    return;
                  }
                  var transX = -e.translationX,
                    transY = -e.translationY;

                  //Accumulate translations.
                  accDx = accDx + ((vertical) ? transY : transX);
                  dx = accDx;
                  scrolling = (vertical) ? (Math.abs(accDx) < Math.abs(-
                    transX)) : (Math.abs(accDx) < Math.abs(-transY));

                  if (e.detail === e.MSGESTURE_FLAG_INERTIA) {
                    setImmediate(function() {
                      el._gesture.stop();
                    });

                    return;
                  }

                  if (!scrolling || Number(new Date()) - startT > 500) {
                    e.preventDefault();
                    if (!fade && slider.transitions) {
                      if (!slider.vars.animationLoop) {
                        dx = accDx / ((slider.currentSlide === 0 && accDx <
                          0 || slider.currentSlide === slider.last &&
                          accDx > 0) ? (Math.abs(accDx) / cwidth + 2) : 1);
                      }
                      slider.setProps(offset + dx, "setTouch");
                    }
                  }
                }

                function onMSGestureEnd(e) {
                  e.stopPropagation();
                  var slider = e.target._slider;
                  if (!slider) {
                    return;
                  }
                  if (slider.animatingTo === slider.currentSlide && !
                    scrolling && !(dx === null)) {
                    var updateDx = (reverse) ? -dx : dx,
                      target = (updateDx > 0) ? slider.getTarget('next') :
                      slider.getTarget('prev');

                    if (slider.canAdvance(target) && (Number(new Date()) -
                      startT < 550 && Math.abs(updateDx) > 50 || Math.abs(
                        updateDx) > cwidth / 2)) {
                      slider.flexAnimate(target, slider.vars.pauseOnAction);
                    } else {
                      if (!fade) slider.flexAnimate(slider.currentSlide,
                        slider.vars.pauseOnAction, true);
                    }
                  }

                  startX = null;
                  startY = null;
                  dx = null;
                  offset = null;
                  accDx = 0;
                }
              }
            },
            resize: function() {
              if (!slider.animating && slider.is(':visible')) {
                if (!carousel) slider.doMath();

                if (fade) {
                  // SMOOTH HEIGHT:
                  methods.smoothHeight();
                } else if (carousel) { //CAROUSEL:
                  slider.slides.width(slider.computedW);
                  slider.update(slider.pagingCount);
                  slider.setProps();
                } else if (vertical) { //VERTICAL:
                  slider.viewport.height(slider.h);
                  slider.setProps(slider.h, "setTotal");
                } else {
                  // SMOOTH HEIGHT:
                  if (slider.vars.smoothHeight) methods.smoothHeight();
                  slider.newSlides.width(slider.computedW);
                  slider.setProps(slider.computedW, "setTotal");
                }
              }
            },
            smoothHeight: function(dur) {
              if (!vertical || fade) {
                var $obj = (fade) ? slider : slider.viewport;
                (dur) ? $obj.animate({
                  "height": slider.slides.eq(slider.animatingTo).height()
                }, dur) : $obj.height(slider.slides.eq(slider.animatingTo).height());
              }
            },
            sync: function(action) {
              var $obj = $(slider.vars.sync).data("flexslider"),
                target = slider.animatingTo;

              switch (action) {
                case "animate":
                  $obj.flexAnimate(target, slider.vars.pauseOnAction, false,
                    true);
                  break;
                case "play":
                  if (!$obj.playing && !$obj.asNav) {
                    $obj.play();
                  }
                  break;
                case "pause":
                  $obj.pause();
                  break;
              }
            },
            uniqueID: function($clone) {
              // Append _clone to current level and children elements with id attributes
              $clone.filter('[id]').add($clone.find('[id]')).each(function() {
                var $this = $(this);
                $this.attr('id', $this.attr('id') + '_clone');
              });
              return $clone;
            },
            pauseInvisible: {
              visProp: null,
              init: function() {
                var prefixes = ['webkit', 'moz', 'ms', 'o'];

                if ('hidden' in document) return 'hidden';
                for (var i = 0; i < prefixes.length; i++) {
                  if ((prefixes[i] + 'Hidden') in document)
                    methods.pauseInvisible.visProp = prefixes[i] + 'Hidden';
                }
                if (methods.pauseInvisible.visProp) {
                  var evtname = methods.pauseInvisible.visProp.replace(
                    /[H|h]idden/, '') + 'visibilitychange';
                  document.addEventListener(evtname, function() {
                    if (methods.pauseInvisible.isHidden()) {
                      if (slider.startTimeout) clearTimeout(slider.startTimeout); //If clock is ticking, stop timer and prevent from starting while invisible
                      else slider.pause(); //Or just pause
                    } else {
                      if (slider.started) slider.play(); //Initiated before, just play
                      else(slider.vars.initDelay > 0) ? setTimeout(slider
                        .play, slider.vars.initDelay) : slider.play(); //Didn't init before: simply init or wait for it
                    }
                  });
                }
              },
              isHidden: function() {
                return document[methods.pauseInvisible.visProp] || false;
              }
            },
            setToClearWatchedEvent: function() {
              clearTimeout(watchedEventClearTimer);
              watchedEventClearTimer = setTimeout(function() {
                watchedEvent = "";
              }, 3000);
            }
          };

          // public methods
          slider.flexAnimate = function(target, pause, override, withSync,
            fromNav) {
            if (!slider.vars.animationLoop && target !== slider.currentSlide) {
              slider.direction = (target > slider.currentSlide) ? "next" :
                "prev";
            }

            if (asNav && slider.pagingCount === 1) slider.direction = (
              slider.currentItem < target) ? "next" : "prev";

            if (!slider.animating && (slider.canAdvance(target, fromNav) ||
              override) && slider.is(":visible")) {
              if (asNav && withSync) {
                var master = $(slider.vars.asNavFor).data('flexslider');
                slider.atEnd = target === 0 || target === slider.count - 1;
                master.flexAnimate(target, true, false, true, fromNav);
                slider.direction = (slider.currentItem < target) ? "next" :
                  "prev";
                master.direction = slider.direction;

                if (Math.ceil((target + 1) / slider.visible) - 1 !== slider
                  .currentSlide && target !== 0) {
                  slider.currentItem = target;
                  slider.slides.removeClass(namespace + "active-slide").eq(
                    target).addClass(namespace + "active-slide");
                  target = Math.floor(target / slider.visible);
                } else {
                  slider.currentItem = target;
                  slider.slides.removeClass(namespace + "active-slide").eq(
                    target).addClass(namespace + "active-slide");
                  return false;
                }
              }

              slider.animating = true;
              slider.animatingTo = target;

              // SLIDESHOW:
              if (pause) slider.pause();

              // API: before() animation Callback
              slider.vars.before(slider);

              // SYNC:
              if (slider.syncExists && !fromNav) methods.sync("animate");

              // CONTROLNAV
              if (slider.vars.controlNav) methods.controlNav.active();

              // !CAROUSEL:
              // CANDIDATE: slide active class (for add/remove slide)
              if (!carousel) slider.slides.removeClass(namespace +
                'active-slide').eq(target).addClass(namespace +
                'active-slide');

              // INFINITE LOOP:
              // CANDIDATE: atEnd
              slider.atEnd = target === 0 || target === slider.last;

              // DIRECTIONNAV:
              if (slider.vars.directionNav) methods.directionNav.update();

              if (target === slider.last) {
                // API: end() of cycle Callback
                slider.vars.end(slider);
                // SLIDESHOW && !INFINITE LOOP:
                if (!slider.vars.animationLoop) slider.pause();
              }

              // SLIDE:
              if (!fade) {
                var dimension = (vertical) ? slider.slides.filter(':first')
                  .height() : slider.computedW,
                  margin, slideString, calcNext;

                // INFINITE LOOP / REVERSE:
                if (carousel) {
                  //margin = (slider.vars.itemWidth > slider.w) ? slider.vars.itemMargin * 2 : slider.vars.itemMargin;
                  margin = slider.vars.itemMargin;
                  calcNext = ((slider.itemW + margin) * slider.move) *
                    slider.animatingTo;
                  slideString = (calcNext > slider.limit && slider.visible !==
                    1) ? slider.limit : calcNext;
                } else if (slider.currentSlide === 0 && target === slider.count -
                  1 && slider.vars.animationLoop && slider.direction !==
                  "next") {
                  slideString = (reverse) ? (slider.count + slider.cloneOffset) *
                    dimension : 0;
                } else if (slider.currentSlide === slider.last && target ===
                  0 && slider.vars.animationLoop && slider.direction !==
                  "prev") {
                  slideString = (reverse) ? 0 : (slider.count + 1) *
                    dimension;
                } else {
                  slideString = (reverse) ? ((slider.count - 1) - target +
                    slider.cloneOffset) * dimension : (target + slider.cloneOffset) *
                    dimension;
                }
                slider.setProps(slideString, "", slider.vars.animationSpeed);
                if (slider.transitions) {
                  if (!slider.vars.animationLoop || !slider.atEnd) {
                    slider.animating = false;
                    slider.currentSlide = slider.animatingTo;
                  }

                  // Unbind previous transitionEnd events and re-bind new transitionEnd event
                  slider.container.unbind(
                    "webkitTransitionEnd transitionend");
                  slider.container.bind("webkitTransitionEnd transitionend",
                    function() {
                      clearTimeout(slider.ensureAnimationEnd);
                      slider.wrapup(dimension);
                    });

                  // Insurance for the ever-so-fickle transitionEnd event
                  clearTimeout(slider.ensureAnimationEnd);
                  slider.ensureAnimationEnd = setTimeout(function() {
                    slider.wrapup(dimension);
                  }, slider.vars.animationSpeed + 100);

                } else {
                  slider.container.animate(slider.args, slider.vars.animationSpeed,
                    slider.vars.easing, function() {
                      slider.wrapup(dimension);
                    });
                }
              } else { // FADE:
                if (!touch) {
                  //slider.slides.eq(slider.currentSlide).fadeOut(slider.vars.animationSpeed, slider.vars.easing);
                  //slider.slides.eq(target).fadeIn(slider.vars.animationSpeed, slider.vars.easing, slider.wrapup);

                  slider.slides.eq(slider.currentSlide).css({
                    "zIndex": 1
                  }).animate({
                    "opacity": 0
                  }, slider.vars.animationSpeed, slider.vars.easing);
                  slider.slides.eq(target).css({
                    "zIndex": 2
                  }).animate({
                      "opacity": 1
                    }, slider.vars.animationSpeed, slider.vars.easing,
                    slider.wrapup);

                } else {
                  slider.slides.eq(slider.currentSlide).css({
                    "opacity": 0,
                    "zIndex": 1
                  });
                  slider.slides.eq(target).css({
                    "opacity": 1,
                    "zIndex": 2
                  });
                  slider.wrapup(dimension);
                }
              }
              // SMOOTH HEIGHT:
              if (slider.vars.smoothHeight) methods.smoothHeight(slider.vars
                .animationSpeed);
            }
          };
          slider.wrapup = function(dimension) {
            // SLIDE:
            if (!fade && !carousel) {
              if (slider.currentSlide === 0 && slider.animatingTo ===
                slider.last && slider.vars.animationLoop) {
                slider.setProps(dimension, "jumpEnd");
              } else if (slider.currentSlide === slider.last && slider.animatingTo ===
                0 && slider.vars.animationLoop) {
                slider.setProps(dimension, "jumpStart");
              }
            }
            slider.animating = false;
            slider.currentSlide = slider.animatingTo;
            // API: after() animation Callback
            slider.vars.after(slider);
          };

          // SLIDESHOW:
          slider.animateSlides = function() {
            if (!slider.animating && focused) slider.flexAnimate(slider.getTarget(
              "next"));
          };
          // SLIDESHOW:
          slider.pause = function() {
            clearInterval(slider.animatedSlides);
            slider.animatedSlides = null;
            slider.playing = false;
            // PAUSEPLAY:
            if (slider.vars.pausePlay) methods.pausePlay.update("play");
            // SYNC:
            if (slider.syncExists) methods.sync("pause");
          };
          // SLIDESHOW:
          slider.play = function() {
            if (slider.playing) clearInterval(slider.animatedSlides);
            slider.animatedSlides = slider.animatedSlides || setInterval(
              slider.animateSlides, slider.vars.slideshowSpeed);
            slider.started = slider.playing = true;
            // PAUSEPLAY:
            if (slider.vars.pausePlay) methods.pausePlay.update("pause");
            // SYNC:
            if (slider.syncExists) methods.sync("play");
          };
          // STOP:
          slider.stop = function() {
            slider.pause();
            slider.stopped = true;
          };
          slider.canAdvance = function(target, fromNav) {
            // ASNAV:
            var last = (asNav) ? slider.pagingCount - 1 : slider.last;
            return (fromNav) ? true :
              (asNav && slider.currentItem === slider.count - 1 && target ===
                0 && slider.direction === "prev") ? true :
              (asNav && slider.currentItem === 0 && target === slider.pagingCount -
                1 && slider.direction !== "next") ? false :
              (target === slider.currentSlide && !asNav) ? false :
              (slider.vars.animationLoop) ? true :
              (slider.atEnd && slider.currentSlide === 0 && target === last &&
                slider.direction !== "next") ? false :
              (slider.atEnd && slider.currentSlide === last && target === 0 &&
                slider.direction === "next") ? false :
              true;
          };
          slider.getTarget = function(dir) {
            slider.direction = dir;
            if (dir === "next") {
              return (slider.currentSlide === slider.last) ? 0 : slider.currentSlide +
                1;
            } else {
              return (slider.currentSlide === 0) ? slider.last : slider.currentSlide -
                1;
            }
          };

          // SLIDE:
          slider.setProps = function(pos, special, dur) {
            var target = (function() {
              var posCheck = (pos) ? pos : ((slider.itemW + slider.vars
                  .itemMargin) * slider.move) * slider.animatingTo,
                posCalc = (function() {
                  if (carousel) {
                    return (special === "setTouch") ? pos :
                      (reverse && slider.animatingTo === slider.last) ?
                      0 :
                      (reverse) ? slider.limit - (((slider.itemW +
                          slider.vars.itemMargin) * slider.move) *
                        slider.animatingTo) :
                      (slider.animatingTo === slider.last) ? slider.limit :
                      posCheck;
                  } else {
                    switch (special) {
                      case "setTotal":
                        return (reverse) ? ((slider.count - 1) -
                            slider.currentSlide + slider.cloneOffset) *
                          pos : (slider.currentSlide + slider.cloneOffset) *
                          pos;
                      case "setTouch":
                        return (reverse) ? pos : pos;
                      case "jumpEnd":
                        return (reverse) ? pos : slider.count * pos;
                      case "jumpStart":
                        return (reverse) ? slider.count * pos : pos;
                      default:
                        return pos;
                    }
                  }
                }());

              return (posCalc * -1) + "px";
            }());

            if (slider.transitions) {
              target = (vertical) ? "translate3d(0," + target + ",0)" :
                "translate3d(" + target + ",0,0)";
              dur = (dur !== undefined) ? (dur / 1000) + "s" : "0s";
              slider.container.css("-" + slider.pfx +
                "-transition-duration", dur);
              slider.container.css("transition-duration", dur);
            }

            slider.args[slider.prop] = target;
            if (slider.transitions || dur === undefined) slider.container.css(
              slider.args);

            slider.container.css('transform', target);
          };

          slider.setup = function(type) {
            // SLIDE:
            if (!fade) {
              var sliderOffset, arr;

              if (type === "init") {
                slider.viewport = $('<div class="' + namespace +
                  'viewport"></div>').css({
                  "overflow": "hidden",
                  "position": "relative"
                }).appendTo(slider).append(slider.container);
                // INFINITE LOOP:
                slider.cloneCount = 0;
                slider.cloneOffset = 0;
                // REVERSE:
                if (reverse) {
                  arr = $.makeArray(slider.slides).reverse();
                  slider.slides = $(arr);
                  slider.container.empty().append(slider.slides);
                }
              }
              // INFINITE LOOP && !CAROUSEL:
              if (slider.vars.animationLoop && !carousel) {
                slider.cloneCount = 2;
                slider.cloneOffset = 1;
                // clear out old clones
                if (type !== "init") slider.container.find('.clone').remove();
                slider.container.append(methods.uniqueID(slider.slides.first()
                  .clone().addClass('clone')).attr('aria-hidden', 'true'))
                  .prepend(methods.uniqueID(slider.slides.last().clone().addClass(
                    'clone')).attr('aria-hidden', 'true'));
              }
              slider.newSlides = $(slider.vars.selector, slider);

              sliderOffset = (reverse) ? slider.count - 1 - slider.currentSlide +
                slider.cloneOffset : slider.currentSlide + slider.cloneOffset;
              // VERTICAL:
              if (vertical && !carousel) {
                slider.container.height((slider.count + slider.cloneCount) *
                  200 + "%").css("position", "absolute").width("100%");
                setTimeout(function() {
                  slider.newSlides.css({
                    "display": "block"
                  });
                  slider.doMath();
                  slider.viewport.height(slider.h);
                  slider.setProps(sliderOffset * slider.h, "init");
                }, (type === "init") ? 100 : 0);
              } else {
                slider.container.width((slider.count + slider.cloneCount) *
                  200 + "%");
                slider.setProps(sliderOffset * slider.computedW, "init");
                setTimeout(function() {
                  slider.doMath();
                  slider.newSlides.css({
                    "width": slider.computedW,
                    "float": "left",
                    "display": "block"
                  });
                  // SMOOTH HEIGHT:
                  if (slider.vars.smoothHeight) methods.smoothHeight();
                }, (type === "init") ? 100 : 0);
              }
            } else { // FADE:
              slider.slides.css({
                "width": "100%",
                "float": "left",
                "marginRight": "-100%",
                "position": "relative"
              });
              if (type === "init") {
                if (!touch) {
                  //slider.slides.eq(slider.currentSlide).fadeIn(slider.vars.animationSpeed, slider.vars.easing);
                  if (slider.vars.fadeFirstSlide == false) {
                    slider.slides.css({
                      "opacity": 0,
                      "display": "block",
                      "zIndex": 1
                    }).eq(slider.currentSlide).css({
                      "zIndex": 2
                    }).css({
                      "opacity": 1
                    });
                  } else {
                    slider.slides.css({
                      "opacity": 0,
                      "display": "block",
                      "zIndex": 1
                    }).eq(slider.currentSlide).css({
                      "zIndex": 2
                    }).animate({
                      "opacity": 1
                    }, slider.vars.animationSpeed, slider.vars.easing);
                  }
                } else {
                  slider.slides.css({
                    "opacity": 0,
                    "display": "block",
                    "webkitTransition": "opacity " + slider.vars.animationSpeed /
                      1000 + "s ease",
                    "zIndex": 1
                  }).eq(slider.currentSlide).css({
                    "opacity": 1,
                    "zIndex": 2
                  });
                }
              }
              // SMOOTH HEIGHT:
              if (slider.vars.smoothHeight) methods.smoothHeight();
            }
            // !CAROUSEL:
            // CANDIDATE: active slide
            if (!carousel) slider.slides.removeClass(namespace +
              "active-slide").eq(slider.currentSlide).addClass(namespace +
              "active-slide");

            //FlexSlider: init() Callback
            slider.vars.init(slider);
          };

          slider.doMath = function() {
            var slide = slider.slides.first(),
              slideMargin = slider.vars.itemMargin,
              minItems = slider.vars.minItems,
              maxItems = slider.vars.maxItems;

            slider.w = (slider.viewport === undefined) ? slider.width() :
              slider.viewport.width();
            slider.h = slide.height();
            slider.boxPadding = slide.outerWidth() - slide.width();

            // CAROUSEL:
            if (carousel) {
              slider.itemT = slider.vars.itemWidth + slideMargin;
              slider.minW = (minItems) ? minItems * slider.itemT : slider.w;
              slider.maxW = (maxItems) ? (maxItems * slider.itemT) -
                slideMargin : slider.w;
              slider.itemW = (slider.minW > slider.w) ? (slider.w - (
                slideMargin * (minItems - 1))) / minItems :
                (slider.maxW < slider.w) ? (slider.w - (slideMargin * (
                maxItems - 1))) / maxItems :
                (slider.vars.itemWidth > slider.w) ? slider.w : slider.vars
                .itemWidth;

              slider.visible = Math.floor(slider.w / (slider.itemW));
              slider.move = (slider.vars.move > 0 && slider.vars.move <
                slider.visible) ? slider.vars.move : slider.visible;
              slider.pagingCount = Math.ceil(((slider.count - slider.visible) /
                slider.move) + 1);
              slider.last = slider.pagingCount - 1;
              slider.limit = (slider.pagingCount === 1) ? 0 :
                (slider.vars.itemWidth > slider.w) ? (slider.itemW * (
                slider.count - 1)) + (slideMargin * (slider.count - 1)) : (
                (slider.itemW + slideMargin) * slider.count) - slider.w -
                slideMargin;
            } else {
              slider.itemW = slider.w;
              slider.pagingCount = slider.count;
              slider.last = slider.count - 1;
            }
            slider.computedW = slider.itemW - slider.boxPadding;
          };

          slider.update = function(pos, action) {
            slider.doMath();

            // update currentSlide and slider.animatingTo if necessary
            if (!carousel) {
              if (pos < slider.currentSlide) {
                slider.currentSlide += 1;
              } else if (pos <= slider.currentSlide && pos !== 0) {
                slider.currentSlide -= 1;
              }
              slider.animatingTo = slider.currentSlide;
            }

            // update controlNav
            if (slider.vars.controlNav && !slider.manualControls) {
              if ((action === "add" && !carousel) || slider.pagingCount >
                slider.controlNav.length) {
                methods.controlNav.update("add");
              } else if ((action === "remove" && !carousel) || slider.pagingCount <
                slider.controlNav.length) {
                if (carousel && slider.currentSlide > slider.last) {
                  slider.currentSlide -= 1;
                  slider.animatingTo -= 1;
                }
                methods.controlNav.update("remove", slider.last);
              }
            }
            // update directionNav
            if (slider.vars.directionNav) methods.directionNav.update();

          };

          slider.addSlide = function(obj, pos) {
            var $obj = $(obj);

            slider.count += 1;
            slider.last = slider.count - 1;

            // append new slide
            if (vertical && reverse) {
              (pos !== undefined) ? slider.slides.eq(slider.count - pos).after(
                $obj) : slider.container.prepend($obj);
            } else {
              (pos !== undefined) ? slider.slides.eq(pos).before($obj) :
                slider.container.append($obj);
            }

            // update currentSlide, animatingTo, controlNav, and directionNav
            slider.update(pos, "add");

            // update slider.slides
            slider.slides = $(slider.vars.selector + ':not(.clone)', slider);
            // re-setup the slider to accomdate new slide
            slider.setup();

            //FlexSlider: added() Callback
            slider.vars.added(slider);
          };
          slider.removeSlide = function(obj) {
            var pos = (isNaN(obj)) ? slider.slides.index($(obj)) : obj;

            // update count
            slider.count -= 1;
            slider.last = slider.count - 1;

            // remove slide
            if (isNaN(obj)) {
              $(obj, slider.slides).remove();
            } else {
              (vertical && reverse) ? slider.slides.eq(slider.last).remove() :
                slider.slides.eq(obj).remove();
            }

            // update currentSlide, animatingTo, controlNav, and directionNav
            slider.doMath();
            slider.update(pos, "remove");

            // update slider.slides
            slider.slides = $(slider.vars.selector + ':not(.clone)', slider);
            // re-setup the slider to accomdate new slide
            slider.setup();

            // FlexSlider: removed() Callback
            slider.vars.removed(slider);
          };

          //FlexSlider: Initialize
          methods.init();
        };

        // Ensure the slider isn't focussed if the window loses focus.
        $(window).blur(function(e) {
          focused = false;
        }).focus(function(e) {
          focused = true;
        });

        // FlexSlider: Default Settings
        $.flexslider.defaults = {
          namespace: "am-", //{NEW} String: Prefix string attached to the class of every element generated by the plugin
          selector: ".am-slides > li", //{NEW} Selector: Must match a simple pattern. '{container} > {slide}' -- Ignore pattern at your own peril
          animation: "slide", //String: Select your animation type, "fade" or "slide"
          easing: "swing", //{NEW} String: Determines the easing method used in jQuery transitions. jQuery easing plugin is supported!
          direction: "horizontal", //String: Select the sliding direction, "horizontal" or "vertical"
          reverse: false, //{NEW} Boolean: Reverse the animation direction
          animationLoop: true, //Boolean: Should the animation loop? If false, directionNav will received "disable" classes at either end
          smoothHeight: false, //{NEW} Boolean: Allow height of the slider to animate smoothly in horizontal mode
          startAt: 0, //Integer: The slide that the slider should start on. Array notation (0 = first slide)
          slideshow: true, //Boolean: Animate slider automatically
          slideshowSpeed: 5000, //Integer: Set the speed of the slideshow cycling, in milliseconds
          animationSpeed: 600, //Integer: Set the speed of animations, in milliseconds
          initDelay: 0, //{NEW} Integer: Set an initialization delay, in milliseconds
          randomize: false, //Boolean: Randomize slide order
          fadeFirstSlide: true, //Boolean: Fade in the first slide when animation type is "fade"
          thumbCaptions: false, //Boolean: Whether or not to put captions on thumbnails when using the "thumbnails" controlNav.

          // Usability features
          pauseOnAction: true, //Boolean: Pause the slideshow when interacting with control elements, highly recommended.
          pauseOnHover: false, //Boolean: Pause the slideshow when hovering over slider, then resume when no longer hovering
          pauseInvisible: true, //{NEW} Boolean: Pause the slideshow when tab is invisible, resume when visible. Provides better UX, lower CPU usage.
          useCSS: true, //{NEW} Boolean: Slider will use CSS3 transitions if available
          touch: true, //{NEW} Boolean: Allow touch swipe navigation of the slider on touch-enabled devices
          video: false, //{NEW} Boolean: If using video in the slider, will prevent CSS3 3D Transforms to avoid graphical glitches

          // Primary Controls
          controlNav: true, //Boolean: Create navigation for paging control of each slide? Note: Leave true for manualControls usage
          directionNav: true, //Boolean: Create navigation for previous/next navigation? (true/false)
          prevText: "Previous", //String: Set the text for the "previous" directionNav item
          nextText: "Next", //String: Set the text for the "next" directionNav item

          // Secondary Navigation
          keyboard: true, //Boolean: Allow slider navigating via keyboard left/right keys
          multipleKeyboard: false, //{NEW} Boolean: Allow keyboard navigation to affect multiple sliders. Default behavior cuts out keyboard navigation with more than one slider present.
          mousewheel: false, //{UPDATED} Boolean: Requires jquery.mousewheel.js (https://github.com/brandonaaron/jquery-mousewheel) - Allows slider navigating via mousewheel
          pausePlay: false, //Boolean: Create pause/play dynamic element
          pauseText: "Pause", //String: Set the text for the "pause" pausePlay item
          playText: "Play", //String: Set the text for the "play" pausePlay item

          // Special properties
          controlsContainer: "", //{UPDATED} jQuery Object/Selector: Declare which container the navigation elements should be appended too. Default container is the FlexSlider element. Example use would be $(".flexslider-container"). Property is ignored if given element is not found.
          manualControls: "", //{UPDATED} jQuery Object/Selector: Declare custom control navigation. Examples would be $(".flex-control-nav li") or "#tabs-nav li img", etc. The number of elements in your controlNav should match the number of slides/tabs.
          sync: "", //{NEW} Selector: Mirror the actions performed on this slider with another slider. Use with care.
          asNavFor: "", //{NEW} Selector: Internal property exposed for turning the slider into a thumbnail navigation for another slider

          // Carousel Options
          itemWidth: 0, //{NEW} Integer: Box-model width of individual carousel items, including horizontal borders and padding.
          itemMargin: 0, //{NEW} Integer: Margin between carousel items.
          minItems: 1, //{NEW} Integer: Minimum number of carousel items that should be visible. Items will resize fluidly when below this.
          maxItems: 0, //{NEW} Integer: Maxmimum number of carousel items that should be visible. Items will resize fluidly when above this limit.
          move: 0, //{NEW} Integer: Number of carousel items that should move on animation. If 0, slider will move all visible items.
          allowOneSlide: true, //{NEW} Boolean: Whether or not to allow a slider comprised of a single slide

          // Callback API
          start: function() {}, //Callback: function(slider) - Fires when the slider loads the first slide
          before: function() {}, //Callback: function(slider) - Fires asynchronously with each slider animation
          after: function() {}, //Callback: function(slider) - Fires after each slider animation completes
          end: function() {}, //Callback: function(slider) - Fires when the slider reaches the last slide (asynchronous)
          added: function() {}, //{NEW} Callback: function(slider) - Fires after a slide is added
          removed: function() {}, //{NEW} Callback: function(slider) - Fires after a slide is removed
          init: function() {} //{NEW} Callback: function(slider) - Fires after the slider is initially setup
        };

        // FlexSlider: Plugin Function
        $.fn.flexslider = function(options) {
          if (options === undefined) options = {};

          if (typeof options === "object") {
            return this.each(function() {
              var $this = $(this),
                selector = (options.selector) ? options.selector :
                ".am-slides > li",
                $slides = $this.find(selector);

              if (($slides.length === 1 && options.allowOneSlide ===
                true) || $slides.length === 0) {
                $slides.fadeIn(400);
                if (options.start) options.start($this);
              } else if ($this.data('flexslider') === undefined) {
                new $.flexslider(this, options);
              }
            });
          } else {
            // Helper strings to quickly pecdrform functions on the slider
            var $slider = $(this).data('flexslider');
            switch (options) {
              case 'play':
                $slider.play();
                break;
              case 'pause':
                $slider.pause();
                break;
              case 'stop':
                $slider.stop();
                break;
              case 'next':
                $slider.flexAnimate($slider.getTarget('next'), true);
                break;
              case 'prev':
              case 'previous':
                $slider.flexAnimate($slider.getTarget('prev'), true);
                break;
              default:
                if (typeof options === 'number') {
                  $slider.flexAnimate(options, true);
                }
            }
          }
        };

        // Init code
        $(function() {
          $('[data-am-flexslider]').each(function(i, item) {
            var $slider = $(item);
            var options = UI.utils.parseOptions($slider.data(
              'amFlexslider'));

            options.before = function(slider) {
              if (slider._pausedTimer) {
                window.clearTimeout(slider._pausedTimer);
                slider._pausedTimer = null;
              }
            };

            options.after = function(slider) {
              var pauseTime = slider.vars.playAfterPaused;
              if (pauseTime && !isNaN(pauseTime) && !slider.playing) {
                if (!slider.manualPause && !slider.manualPlay && !
                  slider.stopped) {
                  slider._pausedTimer = window.setTimeout(function() {
                    slider.play();
                  }, pauseTime);
                }
              }
            };

            $slider.flexslider(options);
          });
        });

        // if (!slider.manualPause && !slider.manualPlay && !slider.stopped) slider.play();

        module.exports = $.flexslider;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  9: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
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
          onConfirm: function() {},
          onCancel: function() {},
          duration: 300, // must equal the CSS transition duration
          transitionEnd: supportTransition && supportTransition.end +
            '.modal.amui'
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

          $element.trigger($.Event('open.modal.amui', {
            relatedElement: relatedElement
          }));

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
            $element.trigger($.Event('opened.modal.amui', {
              relatedElement: relatedElement
            }));
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
            dimmer.close($element, true);
          }

          this.$element.trigger($.Event('close.modal.amui', {
            relatedElement: relatedElement
          }));

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

          $element.
          removeClass(options.className.active).
          addClass(options.className.out);

          if (!supportTransition) {
            return complete.call(this);
          }

          $element.
          one(options.transitionEnd, $.proxy(complete, this)).
          emulateTransitionEnd(options.duration);
        };

        Modal.prototype.events = function() {
          var that = this;
          var $element = this.$element;
          var $ipt = $element.find('.am-modal-prompt-input');

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
          $element.find('[data-am-modal-close]').on('click.modal.amui',
            function(e) {
              e.preventDefault();
              that.close();
            });

          $element.find('.am-modal-btn').on('click.modal.amui', function(e) {
            that.close();
          });

          $element.find('[data-am-modal-confirm]').on('click.modal.amui',
            function() {
              that.options.onConfirm($ipt.val());
            });

          $element.find('[data-am-modal-cancel]').on('click.modal.amui',
            function() {
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

        $doc.on('click.modal.amui.data-api', '[data-am-modal]', function() {
          var $this = $(this);
          var options = UI.utils.parseOptions($this.attr('data-am-modal'));
          var $target = $(options.target ||
            (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
          var option = $target.data('am.modal') ? 'toggle' : options;

          Plugin.call($target, option, this);
        });

        $.AMUI.modal = Modal;

        module.exports = Modal;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2,
      "./ui.dimmer": 6
    }
  ],
  10: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        var $win = $(window);
        var $doc = $(document);
        var scrollPos;

        /**
         * @via https://github.com/uikit/uikit/blob/master/src/js/offcanvas.js
         * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
         */

        var OffCanvas = function(element, options) {
          this.$element = $(element);
          this.options = options;
          this.active = null;
          this.events();
        };

        OffCanvas.DEFAULTS = {
          duration: 300,
          effect: 'overlay' // {push|overlay}, push is too expensive
        };

        OffCanvas.prototype.open = function(relatedElement) {
          var $element = this.$element;

          if (!$element.length || $element.hasClass('am-active')) {
            return;
          }

          var effect = this.options.effect;
          var $html = $('html');
          var $body = $('body');
          var $bar = $element.find('.am-offcanvas-bar').first();
          var dir = $bar.hasClass('am-offcanvas-bar-flip') ? -1 : 1;

          $bar.addClass('am-offcanvas-bar-' + effect);

          scrollPos = {
            x: window.scrollX,
            y: window.scrollY
          };

          $element.addClass('am-active');

          $body.
          css({
            width: window.innerWidth,
            height: $win.height()
          }).
          addClass('am-offcanvas-page');

          if (effect !== 'overlay') {
            $body.css({
              'margin-left': $bar.outerWidth() * dir
            }).width(); // force redraw
          }

          $html.css('margin-top', scrollPos.y * -1);

          setTimeout(function() {
            $bar.addClass('am-offcanvas-bar-active').width();
          }, 0);

          $element.trigger('open.offcanvas.amui');

          this.active = 1;

          // Close OffCanvas when none content area clicked
          $element.off('click.offcanvas.amui').
          on('click.offcanvas.amui', $.proxy(function(e) {
            var $target = $(e.target);

            if (!e.type.match(/swipe/)) {
              if ($target.hasClass('am-offcanvas-bar')) {
                return;
              }

              if ($target.parents('.am-offcanvas-bar').first().length) {
                return;
              }
            }

            // https://developer.mozilla.org/zh-CN/docs/DOM/event.stopImmediatePropagation
            e.stopImmediatePropagation();

            this.close();
          }, this));

          $html.on('keydown.offcanvas.amui', $.proxy(function(e) {
            if (e.keyCode === 27) { // ESC
              this.close();
            }
          }, this));
        };

        OffCanvas.prototype.close = function(relatedElement) {
          var me = this;
          var $html = $('html');
          var $body = $('body');
          var $element = this.$element;
          var $bar = $element.find('.am-offcanvas-bar').first();

          if (!$element.length || !$element.hasClass('am-active')) {
            return;
          }

          $element.trigger('close.offcanvas.amui');

          function complete() {
            $body.removeClass('am-offcanvas-page').
            css({
              width: '',
              height: '',
              'margin-left': '',
              'margin-right': ''
            });
            $element.removeClass('am-active');
            $bar.removeClass('am-offcanvas-bar-active');
            $html.css('margin-top', '');
            window.scrollTo(scrollPos.x, scrollPos.y);
            $element.trigger('closed.offcanvas.amui');
            me.active = 0;
          }

          if (UI.support.transition) {
            setTimeout(function() {
              $bar.removeClass('am-offcanvas-bar-active');
            }, 0);

            $body.css('margin-left', '').one(UI.support.transition.end,
              function() {
                complete();
              }).emulateTransitionEnd(this.options.duration);
          } else {
            complete();
          }

          $element.off('click.offcanvas.amui');
          $html.off('.offcanvas.amui');
        };

        OffCanvas.prototype.events = function() {
          $doc.on('click.offcanvas.amui', '[data-am-dismiss="offcanvas"]',
            $.proxy(function(e) {
              e.preventDefault();
              this.close();
            }, this));

          $win.on('resize.offcanvas.amui orientationchange.offcanvas.amui',
            $.proxy(function(e) {
              this.active && this.close();
            }, this));

          return this;
        };

        function Plugin(option, relatedElement) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('am.offcanvas');
            var options = $.extend({}, OffCanvas.DEFAULTS,
              typeof option == 'object' && option);

            if (!data) {
              $this.data('am.offcanvas', (data = new OffCanvas(this,
                options)));
              data.open(relatedElement);
            }

            if (typeof option == 'string') {
              data[option] && data[option](relatedElement);
            }
          });
        }

        $.fn.offCanvas = Plugin;

        // Init code
        $doc.on('click.offcanvas.amui', '[data-am-offcanvas]', function(e) {
          e.preventDefault();
          var $this = $(this);
          var options = UI.utils.parseOptions($this.data('amOffcanvas'));
          var $target = $(options.target ||
            (this.href && this.href.replace(/.*(?=#[^\s]+$)/, '')));
          var option = $target.data('am.offcanvas') ? 'open' : options;

          Plugin.call($target, option, this);
        });

        $.AMUI.offcanvas = OffCanvas;

        module.exports = OffCanvas;

        // TODO: 优化动画效果
        // http://dbushell.github.io/Responsive-Off-Canvas-Menu/step4.html

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  11: [
    function(require, module, exports) {
      (function(global) {
        'use strict';
        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        var $w = $(window);

        /**
         * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
         * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
         */

        var Popover = function(element, options) {
          this.options = $.extend({}, Popover.DEFAULTS, options || {});
          this.$element = $(element);
          this.active = null;
          this.$popover = (this.options.target && $(this.options.target)) ||
            null;

          this.init();
          this.events();
        };

        Popover.DEFAULTS = {
          trigger: 'click',
          content: '',
          open: false,
          target: undefined,
          tpl: '<div class="am-popover">' +
            '<div class="am-popover-inner"></div>' +
            '<div class="am-popover-caret"></div></div>'
        };

        Popover.prototype.init = function() {
          var me = this;
          var $element = this.$element;
          var $popover;

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

          // TODO: 监听页面内容变化，重新调整位置

          $element.on('open.popover.amui', function() {
            $(window).on('resize.popover.amui', UI.utils.debounce(
              sizePopover, 50));
          });

          $element.on('close.popover.amui', function() {
            $(window).off('resize.popover.amui', sizePopover);
          });

          this.options.open && this.open();
        };

        Popover.prototype.sizePopover = function sizePopover() {
          var $element = this.$element;
          var $popover = this.$popover;

          if (!$popover || !$popover.length) {
            return;
          }

          var popWidth = $popover.outerWidth();
          var popHeight = $popover.outerHeight();
          var $popCaret = $popover.find('.am-popover-caret');
          var popCaretSize = ($popCaret.outerWidth() / 2) || 8;
          // 取不到 $popCaret.outerHeight() 的值，所以直接加 8
          var popTotalHeight = popHeight + 8; // $popCaret.outerHeight();

          var triggerWidth = $element.outerWidth();
          var triggerHeight = $element.outerHeight();
          var triggerOffset = $element.offset();
          var triggerRect = $element[0].getBoundingClientRect();

          var winHeight = $w.height();
          var winWidth = $w.width();
          var popTop = 0;
          var popLeft = 0;
          var diff = 0;
          var spacing = 2;
          var popPosition = 'top';

          $popover.css({
            left: '',
            top: ''
          }).removeClass('am-popover-left ' +
            'am-popover-right am-popover-top am-popover-bottom');

          $popCaret.css({
            left: '',
            top: ''
          });

          if (popTotalHeight - spacing < triggerRect.top + spacing) {
            // Popover on the top of trigger
            popTop = triggerOffset.top - popTotalHeight - spacing;
          } else if (popTotalHeight <
            winHeight - triggerRect.top - triggerRect.height) {
            // On bottom
            popPosition = 'bottom';
            popTop = triggerOffset.top + triggerHeight + popCaretSize +
              spacing;
          } else { // On middle
            popPosition = 'middle';
            popTop = triggerHeight / 2 + triggerOffset.top - popHeight / 2;
          }

          // Horizontal Position
          if (popPosition === 'top' || popPosition === 'bottom') {
            popLeft = triggerWidth / 2 + triggerOffset.left - popWidth / 2;

            diff = popLeft;

            if (popLeft < 5) {
              popLeft = 5;
            }

            if (popLeft + popWidth > winWidth) {
              popLeft = (winWidth - popWidth - 20);
              // console.log('left %d, win %d, popw %d', popLeft, winWidth, popWidth);
            }

            if (popPosition === 'top') {
              // This is the Popover position, NOT caret position
              // Popover on the Top of trigger, caret on the bottom of Popover
              $popover.addClass('am-popover-top');
            }

            if (popPosition === 'bottom') {
              $popover.addClass('am-popover-bottom');
            }

            diff = diff - popLeft;
            $popCaret.css({
              left: (popWidth / 2 - popCaretSize + diff) + 'px'
            });

          } else if (popPosition === 'middle') {
            popLeft = triggerOffset.left - popWidth - popCaretSize;
            $popover.addClass('am-popover-left');
            if (popLeft < 5) {
              popLeft = triggerOffset.left + triggerWidth + popCaretSize;
              $popover.removeClass('am-popover-left').addClass(
                'am-popover-right');
            }

            if (popLeft + popWidth > winWidth) {
              popLeft = winWidth - popWidth - 5;
              $popover.removeClass('am-popover-left').addClass(
                'am-popover-right');
            }
            $popCaret.css({
              top: (popHeight / 2 - popCaretSize / 2) + 'px'
            });
          }

          // Apply position style
          $popover.css({
            top: popTop + 'px',
            left: popLeft + 'px'
          });
        };

        Popover.prototype.toggle = function() {
          return this[this.active ? 'close' : 'open']();
        };

        Popover.prototype.open = function() {
          var $popover = this.$popover;

          this.$element.trigger('open.popover.amui');
          this.sizePopover();
          $popover.show().addClass('am-active');
          this.active = true;
        };

        Popover.prototype.close = function() {
          var $popover = this.$popover;

          this.$element.trigger('close.popover.amui');

          $popover.
          removeClass('am-active').
          trigger('closed.popover.amui').
          hide();

          this.active = false;
        };

        Popover.prototype.getPopover = function() {
          var uid = UI.utils.generateGUID('am-popover');
          return $(this.options.tpl).attr('id', uid);
        };

        Popover.prototype.setContent = function() {
          this.$popover && this.$popover.find('.am-popover-inner').empty().
          html(this.options.content);
        };

        Popover.prototype.events = function() {
          var eventNS = 'popover.amui';
          var triggers = this.options.trigger.split(' ');

          for (var i = triggers.length; i--;) {
            var trigger = triggers[i];

            if (trigger === 'click') {
              this.$element.on('click.' + eventNS, $.proxy(this.toggle,
                this));
            } else { // hover or focus
              var eventIn = trigger == 'hover' ? 'mouseenter' : 'focusin';
              var eventOut = trigger == 'hover' ? 'mouseleave' : 'focusout';

              this.$element.on(eventIn + '.' + eventNS, $.proxy(this.open,
                this));
              this.$element.on(eventOut + '.' + eventNS, $.proxy(this.close,
                this));
            }
          }
        };

        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('am.popover');
            var options = $.extend({},
              UI.utils.parseOptions($this.attr('data-am-popover')),
              typeof option == 'object' && option);

            if (!data) {
              $this.data('am.popover', (data = new Popover(this,
                options)));
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

        $.AMUI.popover = Popover;

        module.exports = Popover;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  12: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

        var Progress = (function() {
          /**
           * NProgress (c) 2013, Rico Sta. Cruz
           * @via http://ricostacruz.com/nprogress
           */

          var NProgress = {};
          var $html = $('html');

          NProgress.version = '0.1.6';

          var Settings = NProgress.settings = {
            minimum: 0.08,
            easing: 'ease',
            positionUsing: '',
            speed: 200,
            trickle: true,
            trickleRate: 0.02,
            trickleSpeed: 800,
            showSpinner: true,
            parent: 'body',
            barSelector: '[role="nprogress-bar"]',
            spinnerSelector: '[role="nprogress-spinner"]',
            template: '<div class="nprogress-bar" role="nprogress-bar">' +
              '<div class="nprogress-peg"></div></div>' +
              '<div class="nprogress-spinner" role="nprogress-spinner">' +
              '<div class="nprogress-spinner-icon"></div></div>'
          };

          /**
           * Updates configuration.
           *
           *     NProgress.configure({
           *       minimum: 0.1
           *     });
           */
          NProgress.configure = function(options) {
            var key, value;
            for (key in options) {
              value = options[key];
              if (value !== undefined && options.hasOwnProperty(key))
                Settings[key] = value;
            }

            return this;
          };

          /**
           * Last number.
           */

          NProgress.status = null;

          /**
           * Sets the progress bar status, where `n` is a number from `0.0` to `1.0`.
           *
           *     NProgress.set(0.4);
           *     NProgress.set(1.0);
           */

          NProgress.set = function(n) {
            var started = NProgress.isStarted();

            n = clamp(n, Settings.minimum, 1);
            NProgress.status = (n === 1 ? null : n);

            var progress = NProgress.render(!started),
              bar = progress.querySelector(Settings.barSelector),
              speed = Settings.speed,
              ease = Settings.easing;

            progress.offsetWidth;
            /* Repaint */

            queue(function(next) {
              // Set positionUsing if it hasn't already been set
              if (Settings.positionUsing === '') Settings.positionUsing =
                NProgress.getPositioningCSS();

              // Add transition
              css(bar, barPositionCSS(n, speed, ease));

              if (n === 1) {
                // Fade out
                css(progress, {
                  transition: 'none',
                  opacity: 1
                });
                progress.offsetWidth;
                /* Repaint */

                setTimeout(function() {
                  css(progress, {
                    transition: 'all ' + speed + 'ms linear',
                    opacity: 0
                  });
                  setTimeout(function() {
                    NProgress.remove();
                    next();
                  }, speed);
                }, speed);
              } else {
                setTimeout(next, speed);
              }
            });

            return this;
          };

          NProgress.isStarted = function() {
            return typeof NProgress.status === 'number';
          };

          /**
           * Shows the progress bar.
           * This is the same as setting the status to 0%, except that it doesn't go backwards.
           *
           *     NProgress.start();
           *
           */
          NProgress.start = function() {
            if (!NProgress.status) NProgress.set(0);

            var work = function() {
              setTimeout(function() {
                if (!NProgress.status) return;
                NProgress.trickle();
                work();
              }, Settings.trickleSpeed);
            };

            if (Settings.trickle) work();

            return this;
          };

          /**
           * Hides the progress bar.
           * This is the *sort of* the same as setting the status to 100%, with the
           * difference being `done()` makes some placebo effect of some realistic motion.
           *
           *     NProgress.done();
           *
           * If `true` is passed, it will show the progress bar even if its hidden.
           *
           *     NProgress.done(true);
           */

          NProgress.done = function(force) {
            if (!force && !NProgress.status) return this;

            return NProgress.inc(0.3 + 0.5 * Math.random()).set(1);
          };

          /**
           * Increments by a random amount.
           */

          NProgress.inc = function(amount) {
            var n = NProgress.status;

            if (!n) {
              return NProgress.start();
            } else {
              if (typeof amount !== 'number') {
                amount = (1 - n) * clamp(Math.random() * n, 0.1, 0.95);
              }

              n = clamp(n + amount, 0, 0.994);
              return NProgress.set(n);
            }
          };

          NProgress.trickle = function() {
            return NProgress.inc(Math.random() * Settings.trickleRate);
          };


          /**
           * (Internal) renders the progress bar markup based on the `template`
           * setting.
           */

          NProgress.render = function(fromStart) {
            if (NProgress.isRendered()) return document.getElementById(
              'nprogress');

            $html.addClass('nprogress-busy');

            var progress = document.createElement('div');
            progress.id = 'nprogress';
            progress.innerHTML = Settings.template;

            var bar = progress.querySelector(Settings.barSelector),
              perc = fromStart ? '-100' : toBarPerc(NProgress.status ||
                0),
              parent = document.querySelector(Settings.parent),
              spinner;

            css(bar, {
              transition: 'all 0 linear',
              transform: 'translate3d(' + perc + '%,0,0)'
            });

            if (!Settings.showSpinner) {
              spinner = progress.querySelector(Settings.spinnerSelector);
              spinner && $(spinner).remove();
            }

            if (parent != document.body) {
              $(parent).addClass('nprogress-custom-parent');
            }

            parent.appendChild(progress);
            return progress;
          };

          /**
           * Removes the element. Opposite of render().
           */

          NProgress.remove = function() {
            $html.removeClass('nprogress-busy');
            $(Settings.parent).removeClass('nprogress-custom-parent');

            var progress = document.getElementById('nprogress');
            progress && $(progress).remove();
          };

          /**
           * Checks if the progress bar is rendered.
           */

          NProgress.isRendered = function() {
            return !!document.getElementById('nprogress');
          };

          /**
           * Determine which positioning CSS rule to use.
           */

          NProgress.getPositioningCSS = function() {
            // Sniff on document.body.style
            var bodyStyle = document.body.style;

            // Sniff prefixes
            var vendorPrefix = ('WebkitTransform' in bodyStyle) ?
              'Webkit' :
              ('MozTransform' in bodyStyle) ? 'Moz' :
              ('msTransform' in bodyStyle) ? 'ms' :
              ('OTransform' in bodyStyle) ? 'O' : '';

            if (vendorPrefix + 'Perspective' in bodyStyle) {
              // Modern browsers with 3D support, e.g. Webkit, IE10
              return 'translate3d';
            } else if (vendorPrefix + 'Transform' in bodyStyle) {
              // Browsers without 3D support, e.g. IE9
              return 'translate';
            } else {
              // Browsers without translate() support, e.g. IE7-8
              return 'margin';
            }
          };

          /**
           * Helpers
           */

          function clamp(n, min, max) {
            if (n < min) return min;
            if (n > max) return max;
            return n;
          }

          /**
           * (Internal) converts a percentage (`0..1`) to a bar translateX
           * percentage (`-100%..0%`).
           */

          function toBarPerc(n) {
            return (-1 + n) * 100;
          }


          /**
           * (Internal) returns the correct CSS for changing the bar's
           * position given an n percentage, and speed and ease from Settings
           */

          function barPositionCSS(n, speed, ease) {
            var barCSS;

            if (Settings.positionUsing === 'translate3d') {
              barCSS = {
                transform: 'translate3d(' + toBarPerc(n) + '%,0,0)'
              };
            } else if (Settings.positionUsing === 'translate') {
              barCSS = {
                transform: 'translate(' + toBarPerc(n) + '%,0)'
              };
            } else {
              barCSS = {
                'margin-left': toBarPerc(n) + '%'
              };
            }

            barCSS.transition = 'all ' + speed + 'ms ' + ease;

            return barCSS;
          }


          /**
           * (Internal) Queues a function to be executed.
           */

          var queue = (function() {
            var pending = [];

            function next() {
              var fn = pending.shift();
              if (fn) {
                fn(next);
              }
            }

            return function(fn) {
              pending.push(fn);
              if (pending.length == 1) next();
            };
          })();


          /**
           * (Internal) Applies css properties to an element, similar to the jQuery
           * css method.
           *
           * While this helper does assist with vendor prefixed property names, it
           * does not perform any manipulation of values prior to setting styles.
           */

          var css = (function() {
            var cssPrefixes = ['Webkit', 'O', 'Moz', 'ms'],
              cssProps = {};

            function camelCase(string) {
              return string.replace(/^-ms-/, 'ms-').replace(
                /-([\da-z])/gi, function(match, letter) {
                  return letter.toUpperCase();
                });
            }

            function getVendorProp(name) {
              var style = document.body.style;
              if (name in style) return name;

              var i = cssPrefixes.length,
                capName = name.charAt(0).toUpperCase() + name.slice(
                  1),
                vendorName;
              while (i--) {
                vendorName = cssPrefixes[i] + capName;
                if (vendorName in style) return vendorName;
              }

              return name;
            }

            function getStyleProp(name) {
              name = camelCase(name);
              return cssProps[name] || (cssProps[name] =
                getVendorProp(name));
            }

            function applyCss(element, prop, value) {
              prop = getStyleProp(prop);
              element.style[prop] = value;
            }

            return function(element, properties) {
              var args = arguments,
                prop,
                value;

              if (args.length == 2) {
                for (prop in properties) {
                  value = properties[prop];
                  if (value !== undefined && properties.hasOwnProperty(
                    prop)) applyCss(element, prop, value);
                }
              } else {
                applyCss(element, args[1], args[2]);
              }
            }
          })();

          return NProgress;
        })();

        $.AMUI.progress = Progress;

        module.exports = Progress;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  13: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        require('./ui.smooth-scroll');

        /**
         * @via https://github.com/uikit/uikit/
         * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
         */

        // ScrollSpyNav Class
        var ScrollSpyNav = function(element, options) {
          this.options = $.extend({}, ScrollSpyNav.DEFAULTS, options);
          this.$element = $(element);
          this.anchors = [];

          this.$links = this.$element.find('a[href^="#"]').each(function(
            i, link) {
            this.anchors.push($(link).attr('href'));
          }.bind(this));

          this.$targets = $(this.anchors.join(', '));

          var processRAF = function() {
            UI.utils.rAF.call(window, $.proxy(this.process, this));
          }.bind(this);

          this.$window = $(window).on('scroll.scrollspynav.amui',
            processRAF)
            .on(
              'resize.scrollspynav.amui orientationchange.scrollspynav.amui',
              UI.utils.debounce(processRAF, 50));

          processRAF();
          this.scrollProcess();
        };

        ScrollSpyNav.DEFAULTS = {
          className: {
            active: 'am-active'
          },
          closest: false,
          smooth: true
        };

        ScrollSpyNav.prototype.process = function() {
          var scrollTop = this.$window.scrollTop();
          var options = this.options;
          var inViews = [];
          var $links = this.$links;

          var $targets = this.$targets;

          $targets.each(function(i, target) {
            if (UI.utils.isInView(target, options)) {
              inViews.push(target);
            }
          });

          // console.log(inViews.length);

          if (inViews.length) {
            var $target;

            $.each(inViews, function(i, item) {
              if ($(item).offset().top >= scrollTop) {
                $target = $(item);
                return false; // break
              }
            });

            if (!$target) {
              return;
            }

            if (options.closest) {
              $links.closest(options.closest).removeClass(options.className
                .active);
              $links.filter('a[href="#' + $target.attr('id') + '"]').
              closest(options.closest).addClass(options.className.active);
            } else {
              $links.removeClass(options.className.active).
              filter('a[href="#' + $target.attr('id') + '"]').
              addClass(options.className.active);
            }
          }
        };

        ScrollSpyNav.prototype.scrollProcess = function() {
          var $links = this.$links;

          // smoothScroll
          if (this.options.smooth) {
            $links.on('click', function(e) {
              e.preventDefault();

              var $this = $(this);
              var $target = $($this.attr('href'));

              if (!$target) {
                return;
              }

              $(window).smoothScroll({
                position: $target.offset().top
              });
            });
          }
        };

        // ScrollSpyNav Plugin
        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('am.scrollspynav');
            var options = typeof option == 'object' && option;

            if (!data) {
              $this.data('am.scrollspynav', (data = new ScrollSpyNav(
                this, options)));
            }

            if (typeof option == 'string') {
              data[option]();
            }
          });
        }

        $.fn.scrollspynav = Plugin;

        // Init code
        $(function() {
          $('[data-am-scrollspy-nav]').each(function() {
            var $this = $(this);
            var options = UI.utils.options($this.attr(
              'data-am-scrollspy-nav'));

            Plugin.call($this, options);
          });
        });

        $.AMUI.scrollspynav = ScrollSpyNav;

        module.exports = ScrollSpyNav;

        // TODO: 1. 算法改进
        //       2. 多级菜单支持
        //       3. smooth scroll pushState

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2,
      "./ui.smooth-scroll": 14
    }
  ],
  14: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');
        var rAF = UI.utils.rAF;
        var cAF = UI.utils.cancelAF;

        /**
         * Smooth Scroll
         * @param position
         * @via http://mir.aculo.us/2014/01/19/scrolling-dom-elements-to-the-top-a-zepto-plugin/
         */

        // Usage: $(window).smoothScroll([options])

        // only allow one scroll to top operation to be in progress at a time,
        // which is probably what you want
        var smoothScrollInProgress = false;

        var SmoothScroll = function(element, options) {
          options = options || {};

          var $this = $(element);
          var targetY = parseInt(options.position) || SmoothScroll.DEFAULTS
            .position;
          var initialY = $this.scrollTop();
          var lastY = initialY;
          var delta = targetY - initialY;
          // duration in ms, make it a bit shorter for short distances
          // this is not scientific and you might want to adjust this for
          // your preferences
          var speed = options.speed ||
            Math.min(750, Math.min(1500, Math.abs(initialY - targetY)));
          // temp variables (t will be a position between 0 and 1, y is the calculated scrollTop)
          var start;
          var t;
          var y;
          var cancelScroll = function() {
            abort();
          };

          // abort if already in progress or nothing to scroll
          if (smoothScrollInProgress) {
            return;
          }

          if (delta === 0) {
            return;
          }

          // quint ease-in-out smoothing, from
          // https://github.com/madrobby/scripty2/blob/master/src/effects/transitions/penner.js#L127-L136
          function smooth(pos) {
            if ((pos /= 0.5) < 1) {
              return 0.5 * Math.pow(pos, 5);
            }

            return 0.5 * (Math.pow((pos - 2), 5) + 2);
          }

          function abort() {
            $this.off('touchstart.smoothscroll.amui', cancelScroll);
            smoothScrollInProgress = false;
          }

          // when there's a touch detected while scrolling is in progress, abort
          // the scrolling (emulates native scrolling behavior)
          $this.on('touchstart.smoothscroll.amui', cancelScroll);
          smoothScrollInProgress = true;

          // start rendering away! note the function given to frame
          // is named "render" so we can reference it again further down
          function render(now) {
            if (!smoothScrollInProgress) {
              return;
            }
            if (!start) {
              start = now;
            }

            // calculate t, position of animation in [0..1]
            t = Math.min(1, Math.max((now - start) / speed, 0));
            // calculate the new scrollTop position (don't forget to smooth)
            y = Math.round(initialY + delta * smooth(t));
            // bracket scrollTop so we're never over-scrolling
            if (delta > 0 && y > targetY) {
              y = targetY;
            }
            if (delta < 0 && y < targetY) {
              y = targetY;
            }

            // only actually set scrollTop if there was a change fromt he last frame
            if (lastY != y) {
              $this.scrollTop(y);
            }

            lastY = y;
            // if we're not done yet, queue up an other frame to render,
            // or clean up
            if (y !== targetY) {
              cAF(scrollRAF);
              scrollRAF = rAF(render);
            } else {
              cAF(scrollRAF);
              abort();
            }
          }

          var scrollRAF = rAF(render);
        };

        SmoothScroll.DEFAULTS = {
          position: 0
        };

        $.fn.smoothScroll = function(option) {
          return this.each(function() {
            new SmoothScroll(this, option);
          });
        };

        // Init code
        $(document).on('click.smoothScroll.amui.data-api',
          '[data-am-smooth-scroll]',
          function(e) {
            e.preventDefault();
            var options = UI.utils.parseOptions($(this).attr(
              'data-am-smooth-scroll'));

            $(window).smoothScroll(options);
          });

        module.exports = SmoothScroll;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  15: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        var UI = require('./core');

        /**
         * @via https://github.com/uikit/uikit/blob/master/src/js/addons/sticky.js
         * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
         */

        // Sticky Class
        var Sticky = function(element, options) {
          var me = this;

          this.options = $.extend({}, Sticky.DEFAULTS, options);
          this.$element = $(element);
          this.sticked = null;
          this.inited = null;
          this.$holder = undefined;

          this.$window = $(window).
          on('scroll.sticky.amui',
            UI.utils.debounce($.proxy(this.checkPosition, this), 10)).
          on('resize.sticky.amui orientationchange.sticky.amui',
            UI.utils.debounce(function() {
              me.reset(true, function() {
                me.checkPosition();
              });
            }, 50)).
          on('load.sticky.amui', $.proxy(this.checkPosition, this));

          // the `.offset()` is diff between jQuery & Zepto.js
          // jQuery: return `top` and `left`
          // Zepto.js: return `top`, `left`, `width`, `height`
          this.offset = this.$element.offset();

          this.init();
        };

        Sticky.DEFAULTS = {
          top: 0,
          bottom: 0,
          animation: '',
          className: {
            sticky: 'am-sticky',
            resetting: 'am-sticky-resetting',
            stickyBtm: 'am-sticky-bottom',
            animationRev: 'am-animation-reverse'
          }
        };

        Sticky.prototype.init = function() {
          var result = this.check();

          if (!result) {
            return false;
          }

          var $element = this.$element;
          var $holder = $('<div class="am-sticky-placeholder"></div>').css({
            height: $element.css('position') != 'absolute' ?
              $element.outerHeight() : '',
            float: $element.css('float') != 'none' ? $element.css(
              'float') : '',
            margin: $element.css('margin')
          });

          this.$holder = $element.css('margin', 0).wrap($holder).parent();

          this.inited = 1;

          return true;
        };

        Sticky.prototype.reset = function(force, cb) {
          var options = this.options;
          var $element = this.$element;
          var animation = (options.animation) ?
            ' am-animation-' + options.animation : '';
          var complete = function() {
            $element.css({
              position: '',
              top: '',
              width: '',
              left: '',
              margin: 0
            });
            $element.removeClass([
              animation,
              options.className.animationRev,
              options.className.sticky,
              options.className.resetting
            ].join(' '));

            this.animating = false;
            this.sticked = false;
            this.offset = $element.offset();
            cb && cb();
          }.bind(this);

          $element.addClass(options.className.resetting);

          if (!force && options.animation && UI.support.animation) {

            this.animating = true;

            $element.removeClass(animation).one(UI.support.animation.end,
              function() {
                complete();
              }).width(); // force redraw

            $element.addClass(animation + ' ' + options.className.animationRev);
          } else {
            complete();
          }
        };

        Sticky.prototype.check = function() {
          if (!this.$element.is(':visible')) {
            return false;
          }

          var media = this.options.media;

          if (media) {
            switch (typeof(media)) {
              case 'number':
                if (window.innerWidth < media) {
                  return false;
                }
                break;

              case 'string':
                if (window.matchMedia && !window.matchMedia(media).matches) {
                  return false;
                }
                break;
            }
          }

          return true;
        };

        Sticky.prototype.checkPosition = function() {
          if (!this.inited) {
            var initialized = this.init();
            if (!initialized) {
              return;
            }
          }

          var options = this.options;
          var scrollTop = this.$window.scrollTop();
          var offsetTop = options.top;
          var offsetBottom = options.bottom;
          var $element = this.$element;
          var animation = (options.animation) ?
            ' am-animation-' + options.animation : '';
          var className = [options.className.sticky, animation].join(' ');

          if (typeof offsetBottom == 'function') {
            offsetBottom = offsetBottom(this.$element);
          }

          var checkResult = (scrollTop > this.$holder.offset().top);

          if (!this.sticked && checkResult) {
            $element.addClass(className);
          } else if (this.sticked && !checkResult) {
            this.reset();
          }

          this.$holder.height($element.is(':visible') ? $element.height() :
            0);

          if (checkResult) {
            $element.css({
              top: offsetTop,
              left: this.$holder.offset().left,
              width: this.$holder.width()
            });

            /*
     if (offsetBottom) {
     // （底部边距 + 元素高度 > 窗口高度） 时定位到底部
     if ((offsetBottom + this.offset.height > $(window).height()) &&
     (scrollTop + $(window).height() >= scrollHeight - offsetBottom)) {
     $element.addClass(options.className.stickyBtm).
     css({top: $(window).height() - offsetBottom - this.offset.height});
     } else {
     $element.removeClass(options.className.stickyBtm).css({top: offsetTop});
     }
     }
     */
          }

          this.sticked = checkResult;
        };

        // Sticky Plugin
        function Plugin(option) {
          return this.each(function() {
            var $this = $(this);
            var data = $this.data('am.sticky');
            var options = typeof option == 'object' && option;

            if (!data) {
              $this.data('am.sticky', (data = new Sticky(this, options)));
            }

            if (typeof option == 'string') {
              data[option]();
            }
          });
        }

        $.fn.sticky = Plugin;

        // Init code
        $(window).on('load', function() {
          $('[data-am-sticky]').each(function() {
            var $this = $(this);
            var options = UI.utils.options($this.attr('data-am-sticky'));

            Plugin.call($this, options);
          });
        });

        $.AMUI.sticky = Sticky;

        module.exports = Sticky;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ],
  16: [
    function(require, module, exports) {
      (function(global) {
        'use strict';

        var $ = (typeof window !== "undefined" ? window.jQuery : typeof global !==
          "undefined" ? global.jQuery : null);
        require('./core');

        var cookie = {
          get: function(name) {
            var cookieName = encodeURIComponent(name) + '=';
            var cookieStart = document.cookie.indexOf(cookieName);
            var cookieValue = null;
            var cookieEnd;

            if (cookieStart > -1) {
              cookieEnd = document.cookie.indexOf(';', cookieStart);
              if (cookieEnd == -1) {
                cookieEnd = document.cookie.length;
              }
              cookieValue = decodeURIComponent(document.cookie.substring(
                cookieStart +
                cookieName.length, cookieEnd));
            }

            return cookieValue;
          },

          set: function(name, value, expires, path, domain, secure) {
            var cookieText = encodeURIComponent(name) + '=' +
              encodeURIComponent(value);

            if (expires instanceof Date) {
              cookieText += '; expires=' + expires.toGMTString();
            }

            if (path) {
              cookieText += '; path=' + path;
            }

            if (domain) {
              cookieText += '; domain=' + domain;
            }

            if (secure) {
              cookieText += '; secure';
            }

            document.cookie = cookieText;
          },

          unset: function(name, path, domain, secure) {
            this.set(name, '', new Date(0), path, domain, secure);
          }
        };

        $.AMUI.utils.cookie = cookie;

        module.exports = cookie;

      }).call(this, typeof global !== "undefined" ? global : typeof self !==
        "undefined" ? self : typeof window !== "undefined" ? window : {})
    }, {
      "./core": 2
    }
  ]
}, {}, [1]);
