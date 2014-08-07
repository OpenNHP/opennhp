define(function(require, exports, module) {
    require('core');

    var $ = window.Zepto,
        UI = $.AMUI;

    var Progress = (function() {

        /**
         * NProgress (c) 2013, Rico Sta. Cruz
         * @via http://ricostacruz.com/nprogress
         */

        var NProgress = {},
            $html = $('html');

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
                if (value !== undefined && options.hasOwnProperty(key)) Settings[key] = value;
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
                bar      = progress.querySelector(Settings.barSelector),
                speed    = Settings.speed,
                ease     = Settings.easing;

            progress.offsetWidth; /* Repaint */

            queue(function(next) {
                // Set positionUsing if it hasn't already been set
                if (Settings.positionUsing === '') Settings.positionUsing = NProgress.getPositioningCSS();

                // Add transition
                css(bar, barPositionCSS(n, speed, ease));

                if (n === 1) {
                    // Fade out
                    css(progress, {
                        transition: 'none',
                        opacity: 1
                    });
                    progress.offsetWidth; /* Repaint */

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
            if (NProgress.isRendered()) return document.getElementById('nprogress');

            $html.addClass('nprogress-busy');

            var progress = document.createElement('div');
            progress.id = 'nprogress';
            progress.innerHTML = Settings.template;

            var bar      = progress.querySelector(Settings.barSelector),
                perc     = fromStart ? '-100' : toBarPerc(NProgress.status || 0),
                parent   = document.querySelector(Settings.parent),
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
            var vendorPrefix = ('WebkitTransform' in bodyStyle) ? 'Webkit' :
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
                barCSS = { transform: 'translate3d('+toBarPerc(n)+'%,0,0)' };
            } else if (Settings.positionUsing === 'translate') {
                barCSS = { transform: 'translate('+toBarPerc(n)+'%,0)' };
            } else {
                barCSS = { 'margin-left': toBarPerc(n)+'%' };
            }

            barCSS.transition = 'all '+speed+'ms '+ease;

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
            var cssPrefixes = [ 'Webkit', 'O', 'Moz', 'ms' ],
                cssProps    = {};

            function camelCase(string) {
                return string.replace(/^-ms-/, 'ms-').replace(/-([\da-z])/gi, function(match, letter) {
                    return letter.toUpperCase();
                });
            }

            function getVendorProp(name) {
                var style = document.body.style;
                if (name in style) return name;

                var i = cssPrefixes.length,
                    capName = name.charAt(0).toUpperCase() + name.slice(1),
                    vendorName;
                while (i--) {
                    vendorName = cssPrefixes[i] + capName;
                    if (vendorName in style) return vendorName;
                }

                return name;
            }

            function getStyleProp(name) {
                name = camelCase(name);
                return cssProps[name] || (cssProps[name] = getVendorProp(name));
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
                        if (value !== undefined && properties.hasOwnProperty(prop)) applyCss(element, prop, value);
                    }
                } else {
                    applyCss(element, args[1], args[2]);
                }
            }
        })();

        return NProgress;
    })();

    UI.progress = Progress;

    module.exports = Progress;

    //TODO: $.fn.css 添加自动前缀功能，替换 css() 方法，
});