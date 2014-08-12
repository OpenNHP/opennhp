/*! AmazeUI - v1.0.0 | (c) 2014 AllMobilize, Inc. | @license MIT | 2014-08-12 14:08:17 */ /*! Sea.js 2.2.1 | seajs.org/LICENSE.md */
!function(a,b){function c(a){return function(b){return{}.toString.call(b)=="[object "+a+"]"}}function d(){return A++}function e(a){return a.match(D)[0]}function f(a){for(a=a.replace(E,"/");a.match(F);)a=a.replace(F,"/");return a=a.replace(G,"$1/")}function g(a){var b=a.length-1,c=a.charAt(b);return"#"===c?a.substring(0,b):".js"===a.substring(b-2)||a.indexOf("?")>0||".css"===a.substring(b-3)||"/"===c?a:a+".js"}function h(a){var b=v.alias;return b&&x(b[a])?b[a]:a}function i(a){var b=v.paths,c;return b&&(c=a.match(H))&&x(b[c[1]])&&(a=b[c[1]]+c[2]),a}function j(a){var b=v.vars;return b&&a.indexOf("{")>-1&&(a=a.replace(I,function(a,c){return x(b[c])?b[c]:a})),a}function k(a){var b=v.map,c=a;if(b)for(var d=0,e=b.length;e>d;d++){var f=b[d];if(c=z(f)?f(a)||a:a.replace(f[0],f[1]),c!==a)break}return c}function l(a,b){var c,d=a.charAt(0);if(J.test(a))c=a;else if("."===d)c=f((b?e(b):v.cwd)+a);else if("/"===d){var g=v.cwd.match(K);c=g?g[0]+a.substring(1):a}else c=v.base+a;return 0===c.indexOf("//")&&(c=location.protocol+c),c}function m(a,b){if(!a)return"";a=h(a),a=i(a),a=j(a),a=g(a);var c=l(a,b);return c=k(c)}function n(a){return a.hasAttribute?a.src:a.getAttribute("src",4)}function o(a,b,c){var d=S.test(a),e=L.createElement(d?"link":"script");if(c){var f=z(c)?c(a):c;f&&(e.charset=f)}p(e,b,d,a),d?(e.rel="stylesheet",e.href=a):(e.async=!0,e.src=a),T=e,R?Q.insertBefore(e,R):Q.appendChild(e),T=null}function p(a,c,d,e){function f(){a.onload=a.onerror=a.onreadystatechange=null,d||v.debug||Q.removeChild(a),a=null,c()}var g="onload"in a;return!d||!V&&g?(g?(a.onload=f,a.onerror=function(){C("error",{uri:e,node:a}),f()}):a.onreadystatechange=function(){/loaded|complete/.test(a.readyState)&&f()},b):(setTimeout(function(){q(a,c)},1),b)}function q(a,b){var c=a.sheet,d;if(V)c&&(d=!0);else if(c)try{c.cssRules&&(d=!0)}catch(e){"NS_ERROR_DOM_SECURITY_ERR"===e.name&&(d=!0)}setTimeout(function(){d?b():q(a,b)},20)}function r(){if(T)return T;if(U&&"interactive"===U.readyState)return U;for(var a=Q.getElementsByTagName("script"),b=a.length-1;b>=0;b--){var c=a[b];if("interactive"===c.readyState)return U=c}}function s(a){var b=[];return a.replace(X,"").replace(W,function(a,c,d){d&&b.push(d)}),b}function t(a,b){this.uri=a,this.dependencies=b||[],this.exports=null,this.status=0,this._waitings={},this._remain=0}if(!a.seajs){var u=a.seajs={version:"2.2.1"},v=u.data={},w=c("Object"),x=c("String"),y=Array.isArray||c("Array"),z=c("Function"),A=0,B=v.events={};u.on=function(a,b){var c=B[a]||(B[a]=[]);return c.push(b),u},u.off=function(a,b){if(!a&&!b)return B=v.events={},u;var c=B[a];if(c)if(b)for(var d=c.length-1;d>=0;d--)c[d]===b&&c.splice(d,1);else delete B[a];return u};var C=u.emit=function(a,b){var c=B[a],d;if(c)for(c=c.slice();d=c.shift();)d(b);return u},D=/[^?#]*\//,E=/\/\.\//g,F=/\/[^/]+\/\.\.\//,G=/([^:/])\/\//g,H=/^([^/:]+)(\/.+)$/,I=/{([^{]+)}/g,J=/^\/\/.|:\//,K=/^.*?\/\/.*?\//,L=document,M=e(L.URL),N=L.scripts,O=L.getElementById("seajsnode")||N[N.length-1],P=e(n(O)||M);u.resolve=m;var Q=L.head||L.getElementsByTagName("head")[0]||L.documentElement,R=Q.getElementsByTagName("base")[0],S=/\.css(?:\?|$)/i,T,U,V=+navigator.userAgent.replace(/.*(?:AppleWebKit|AndroidWebKit)\/(\d+).*/,"$1")<536;u.request=o;var W=/"(?:\\"|[^"])*"|'(?:\\'|[^'])*'|\/\*[\S\s]*?\*\/|\/(?:\\\/|[^\/\r\n])+\/(?=[^\/])|\/\/.*|\.\s*require|(?:^|[^$])\brequire\s*\(\s*(["'])(.+?)\1\s*\)/g,X=/\\\\/g,Y=u.cache={},Z,$={},_={},ab={},bb=t.STATUS={FETCHING:1,SAVED:2,LOADING:3,LOADED:4,EXECUTING:5,EXECUTED:6};t.prototype.resolve=function(){for(var a=this,b=a.dependencies,c=[],d=0,e=b.length;e>d;d++)c[d]=t.resolve(b[d],a.uri);return c},t.prototype.load=function(){var a=this;if(!(a.status>=bb.LOADING)){a.status=bb.LOADING;var c=a.resolve();C("load",c);for(var d=a._remain=c.length,e,f=0;d>f;f++)e=t.get(c[f]),e.status<bb.LOADED?e._waitings[a.uri]=(e._waitings[a.uri]||0)+1:a._remain--;if(0===a._remain)return a.onload(),b;var g={};for(f=0;d>f;f++)e=Y[c[f]],e.status<bb.FETCHING?e.fetch(g):e.status===bb.SAVED&&e.load();for(var h in g)g.hasOwnProperty(h)&&g[h]()}},t.prototype.onload=function(){var a=this;a.status=bb.LOADED,a.callback&&a.callback();var b=a._waitings,c,d;for(c in b)b.hasOwnProperty(c)&&(d=Y[c],d._remain-=b[c],0===d._remain&&d.onload());delete a._waitings,delete a._remain},t.prototype.fetch=function(a){function c(){u.request(g.requestUri,g.onRequest,g.charset)}function d(){delete $[h],_[h]=!0,Z&&(t.save(f,Z),Z=null);var a,b=ab[h];for(delete ab[h];a=b.shift();)a.load()}var e=this,f=e.uri;e.status=bb.FETCHING;var g={uri:f};C("fetch",g);var h=g.requestUri||f;return!h||_[h]?(e.load(),b):$[h]?(ab[h].push(e),b):($[h]=!0,ab[h]=[e],C("request",g={uri:f,requestUri:h,onRequest:d,charset:v.charset}),g.requested||(a?a[g.requestUri]=c:c()),b)},t.prototype.exec=function(){function a(b){return t.get(a.resolve(b)).exec()}var c=this;if(c.status>=bb.EXECUTING)return c.exports;c.status=bb.EXECUTING;var e=c.uri;a.resolve=function(a){return t.resolve(a,e)},a.async=function(b,c){return t.use(b,c,e+"_async_"+d()),a};var f=c.factory,g=z(f)?f(a,c.exports={},c):f;return g===b&&(g=c.exports),delete c.factory,c.exports=g,c.status=bb.EXECUTED,C("exec",c),g},t.resolve=function(a,b){var c={id:a,refUri:b};return C("resolve",c),c.uri||u.resolve(c.id,b)},t.define=function(a,c,d){var e=arguments.length;1===e?(d=a,a=b):2===e&&(d=c,y(a)?(c=a,a=b):c=b),!y(c)&&z(d)&&(c=s(""+d));var f={id:a,uri:t.resolve(a),deps:c,factory:d};if(!f.uri&&L.attachEvent){var g=r();g&&(f.uri=g.src)}C("define",f),f.uri?t.save(f.uri,f):Z=f},t.save=function(a,b){var c=t.get(a);c.status<bb.SAVED&&(c.id=b.id||a,c.dependencies=b.deps||[],c.factory=b.factory,c.status=bb.SAVED)},t.get=function(a,b){return Y[a]||(Y[a]=new t(a,b))},t.use=function(b,c,d){var e=t.get(d,y(b)?b:[b]);e.callback=function(){for(var b=[],d=e.resolve(),f=0,g=d.length;g>f;f++)b[f]=Y[d[f]].exec();c&&c.apply(a,b),delete e.callback},e.load()},t.preload=function(a){var b=v.preload,c=b.length;c?t.use(b,function(){b.splice(0,c),t.preload(a)},v.cwd+"_preload_"+d()):a()},u.use=function(a,b){return t.preload(function(){t.use(a,b,v.cwd+"_use_"+d())}),u},t.define.cmd={},a.define=t.define,u.Module=t,v.fetchedList=_,v.cid=d,u.require=function(a){var b=t.get(t.resolve(a));return b.status<bb.EXECUTING&&(b.onload(),b.exec()),b.exports};var cb=/^(.+?\/)(\?\?)?(seajs\/)+/;v.base=(P.match(cb)||["",P])[1],v.dir=P,v.cwd=M,v.charset="utf-8",v.preload=function(){var a=[],b=location.search.replace(/(seajs-\w+)(&|$)/g,"$1=1$2");return b+=" "+L.cookie,b.replace(/(seajs-\w+)=1/g,function(b,c){a.push(c)}),a}(),u.config=function(a){for(var b in a){var c=a[b],d=v[b];if(d&&w(d))for(var e in c)d[e]=c[e];else y(d)?c=d.concat(c):"base"===b&&("/"!==c.slice(-1)&&(c+="/"),c=l(c)),v[b]=c}return C("config",a),u}}}(this);

define("core", [ "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    // Zepto animate extend
    require("zepto.extend.fx");
    // Zepto data extend
    require("zepto.extend.data");
    // Zepto selector extend
    require("zepto.extend.selector");
    var $ = window.Zepto, UI = $.AMUI || {}, $win = $(window), doc = window.document, $html = $("html");
    if (UI.fn) {
        return UI;
    }
    UI.fn = function(command, options) {
        var args = arguments, cmd = command.match(/^([a-z\-]+)(?:\.([a-z]+))?/i), component = cmd[1], method = cmd[2];
        if (!UI[component]) {
            log.error('Amaze UI component [" + component + "] does not exist.');
            return this;
        }
        return this.each(function() {
            var $this = $(this), data = $this.data(component);
            if (!data) $this.data(component, data = UI[component](this, method ? undefined : options));
            if (method) data[method].apply(data, Array.prototype.slice.call(args, 1));
        });
    };
    UI.support = {};
    UI.support.transition = function() {
        var transitionEnd = function() {
            var element = doc.body || doc.documentElement, transEndEventNames = {
                WebkitTransition: "webkitTransitionEnd",
                MozTransition: "transitionend",
                OTransition: "oTransitionEnd otransitionend",
                transition: "transitionend"
            }, name;
            for (name in transEndEventNames) {
                if (element.style[name] !== undefined) return transEndEventNames[name];
            }
        }();
        return transitionEnd && {
            end: transitionEnd
        };
    }();
    UI.support.animation = function() {
        var animationEnd = function() {
            var element = doc.body || doc.documentElement, animEndEventNames = {
                WebkitAnimation: "webkitAnimationEnd",
                MozAnimation: "animationend",
                OAnimation: "oAnimationEnd oanimationend",
                animation: "animationend"
            }, name;
            for (name in animEndEventNames) {
                if (element.style[name] !== undefined) return animEndEventNames[name];
            }
        }();
        return animationEnd && {
            end: animationEnd
        };
    }();
    UI.support.requestAnimationFrame = window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || window.msRequestAnimationFrame || window.oRequestAnimationFrame || function(callback) {
        window.setTimeout(callback, 1e3 / 60);
    };
    UI.support.touch = "ontouchstart" in window && navigator.userAgent.toLowerCase().match(/mobile|tablet/) || window.DocumentTouch && document instanceof window.DocumentTouch || window.navigator["msPointerEnabled"] && window.navigator["msMaxTouchPoints"] > 0 || //IE 10
    window.navigator["pointerEnabled"] && window.navigator["maxTouchPoints"] > 0 || //IE >=11
    false;
    // https://developer.mozilla.org/zh-CN/docs/DOM/MutationObserver
    UI.support.mutationobserver = window.MutationObserver || window.WebKitMutationObserver || window.MozMutationObserver || null;
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
            var context = this, args = arguments;
            var later = function() {
                timeout = null;
                if (!immediate) func.apply(context, args);
            };
            var callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
            if (callNow) func.apply(context, args);
        };
    };
    UI.utils.isInView = function(element, options) {
        var $element = $(element);
        var visible = !!($element.width() || $element.height()) && $element.css("display") !== "none";
        if (!visible) {
            return false;
        }
        var window_left = $win.scrollLeft(), window_top = $win.scrollTop(), offset = $element.offset(), left = offset.left, top = offset.top;
        options = $.extend({
            topOffset: 0,
            leftOffset: 0
        }, options);
        return top + $element.height() >= window_top && top - options.topOffset <= window_top + $win.height() && left + $element.width() >= window_left && left - options.leftOffset <= window_left + $win.width();
    };
    UI.utils.parseOptions = UI.utils.options = function(string) {
        if ($.isPlainObject(string)) return string;
        var start = string ? string.indexOf("{") : -1, options = {};
        if (start != -1) {
            try {
                options = new Function("", "var json = " + string.substr(start) + "; return JSON.parse(JSON.stringify(json));")();
            } catch (e) {}
        }
        return options;
    };
    UI.utils.event = {};
    UI.utils.event.click = UI.support.touch ? "tap" : "click";
    $.AMUI = UI;
    $.fn.amui = UI.fn;
    $.AMUI.langdirection = $("html").attr("dir") == "rtl" ? "right" : "left";
    // http://blog.alexmaccaw.com/css-transitions
    $.fn.emulateTransitionEnd = function(duration) {
        var called = false, $el = this;
        $(this).one(UI.support.transition.end, function() {
            called = true;
        });
        var callback = function() {
            if (!called) {
                $($el).trigger(UI.support.transition.end);
            }
        };
        setTimeout(callback, duration);
        return this;
    };
    $.fn.redraw = function() {
        $(this).each(function() {
            var redraw = this.offsetHeight;
        });
        return this;
    };
    $.fn.transitionEnd = function(callback) {
        var endEvent = UI.support.transition.end, dom = this;
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
            var classes = $(this).attr("class");
            if (!classes || !regex) return false;
            var classArray = [];
            classes = classes.split(" ");
            for (var i = 0, len = classes.length; i < len; i++) if (!classes[i].match(regex)) classArray.push(classes[i]);
            $(this).attr("class", classArray.join(" "));
        });
    };
    //
    $.fn.alterClass = function(removals, additions) {
        var self = this;
        if (removals.indexOf("*") === -1) {
            // Use native jQuery methods if there is no wildcard matching
            self.removeClass(removals);
            return !additions ? self : self.addClass(additions);
        }
        var classPattern = new RegExp("\\s" + removals.replace(/\*/g, "[A-Za-z0-9-_]+").split(" ").join("\\s|\\s") + "\\s", "g");
        self.each(function(i, it) {
            var cn = " " + it.className + " ";
            while (classPattern.test(cn)) {
                cn = cn.replace(classPattern, " ");
            }
            it.className = $.trim(cn);
        });
        return !additions ? self : self.addClass(additions);
    };
    $.fn.getHeight = function() {
        var $ele = $(this), height = "auto";
        if ($ele.is(":visible")) {
            height = $ele.height();
        } else {
            var tmp = {
                position: $ele.css("position"),
                visibility: $ele.css("visibility"),
                display: $ele.css("display")
            };
            height = $ele.css({
                position: "absolute",
                visibility: "hidden",
                display: "block"
            }).height();
            $ele.css(tmp);
        }
        return height;
    };
    $.fn.getSize = function() {
        var $el = $(this);
        if ($el.css("display") !== "none") {
            return {
                width: $el.width(),
                height: $el.height()
            };
        }
        var old = {
            position: $el.css("position"),
            visibility: $el.css("visibility"),
            display: $el.css("display")
        }, tmpStyle = {
            display: "block",
            position: "absolute",
            visibility: "hidden"
        };
        $el.css(tmpStyle);
        var width = $el.width(), height = $el.height();
        $el.css(old);
        return {
            width: width,
            height: height
        };
    };
    // adding :visible and :hidden to zepto
    // https://github.com/jquery/jquery/blob/73e120116ce13b992d5229b3e10fcc19f9505a15/src/css/hiddenVisibleSelectors.js
    var _is = $.fn.is, _filter = $.fn.filter;
    function visible(elem) {
        elem = $(elem);
        return !!(elem.width() || elem.height()) && elem.css("display") !== "none";
    }
    $.fn.is = function(sel) {
        if (sel === ":visible") {
            return visible(this);
        }
        if (sel === ":hidden") {
            return !visible(this);
        }
        return _is.call(this, sel);
    };
    $.fn.filter = function(sel) {
        if (sel === ":visible") {
            return $([].filter.call(this, visible));
        }
        if (sel === ":hidden") {
            return $([].filter.call(this, function(elem) {
                return !visible(elem);
            }));
        }
        return _filter.call(this, sel);
    };
    // handle multiple browsers for requestAnimationFrame()
    // http://www.paulirish.com/2011/requestanimationframe-for-smart-animating/
    // https://github.com/gnarf/jquery-requestAnimationFrame
    UI.utils.rAF = function() {
        return window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || //window.oRequestAnimationFrame ||
        // if all else fails, use setTimeout
        function(callback) {
            return window.setTimeout(callback, 1e3 / 60);
        };
    }();
    // handle multiple browsers for cancelAnimationFrame()
    UI.utils.cancelAF = function() {
        return window.cancelAnimationFrame || window.webkitCancelAnimationFrame || window.mozCancelAnimationFrame || //window.oCancelAnimationFrame ||
        function(id) {
            window.clearTimeout(id);
        };
    }();
    // Require fastclick.js on touch devices
    if (UI.support.touch) {
        require.async([ "util.fastclick" ], function(FastClick) {
            $(function() {
                FastClick && FastClick.attach(document.body);
                $html.addClass("am-touch");
            });
        });
    }
    $(function() {
        // trigger domready event
        $(document).trigger("domready:amui");
        $html.removeClass("no-js").addClass("js");
        UI.support.animation && $html.addClass("cssanimations");
        // Remove responsive classes in .am-layout
        var $layout = $(".am-layout");
        $layout.find('[class*="md-block-grid"]').alterClass("md-block-grid-*");
        $layout.find('[class*="lg-block-grid"]').alterClass("lg-block-grid");
        // widgets not in .am-layout
        $("[data-am-widget]").each(function() {
            var $widget = $(this);
            // console.log($widget.parents('.am-layout').length)
            if ($widget.parents(".am-layout").length === 0) {
                $widget.addClass("am-no-layout");
            }
        });
    });
    module.exports = UI;
});
define("accordion", [ "ui.accordion" ], function(require, exports, module) {
    var accordion = require("ui.accordion");
    var $ = window.Zepto, accordionInit = function() {
        $(".am-accordion").each(function(index, item) {
            var settings = $(item).attr("data-accordion-settings");
            try {
                settings = JSON.parse(settings);
                $(item).accordion(settings);
            } catch (e) {
                $(item).accordion();
            }
        });
    };
    // Init on DOM ready
    $(function() {
        accordionInit();
    });
    exports.init = accordionInit;
});
define("divider", [], function(require, exports, module) {});
define("figure", [ "zepto.pinchzoom" ], function(require, exports, module) {
    var $ = window.Zepto;
    // PinchZoom Plugin
    var PinchZoom = require("zepto.pinchzoom");
    /**
     * Is Images zoomable
     * @return {Boolean}
     */
    $.isImgZoomAble = function(imgElement) {
        var t = new Image();
        t.src = imgElement.src;
        var zoomAble = $(imgElement).width() < t.width;
        if (zoomAble) {
            $(imgElement).parent(".am-figure").addClass("am-figure-zoomable");
        }
        return zoomAble;
    };
    $.fn.imgZoomToggle = function() {
        return this.each(function() {
            var zoomAble = $.isImgZoomAble(this), $wrapDom = $('<div class="am-figure-wrap"><div class="pinch-zoom"></div></div>');
            $zoomWrap = $(".am-figure-wrap");
            if ($zoomWrap.length == 0) {
                $("body").append($wrapDom);
                $zoomWrap = $(".am-figure-wrap");
                $pinch = $zoomWrap.find(".pinch-zoom");
                $pinch.each(function() {
                    new PinchZoom($(this), {});
                });
            }
            if (zoomAble) {
                //$zoomWrap.empty().html(this.outerHTML);
                $pinch.empty().html(this.outerHTML);
                $zoomWrap.find("img").width($(window).width());
                $(this).parent(".am-figure").on("click", function() {
                    $zoomWrap.toggleClass("am-active");
                });
                $zoomWrap.on("click", function(e) {
                    e.preventDefault();
                    var target = e.target;
                    // Img is using pinch zoom
                    if (!$(target).is("img")) {
                        $(this).toggleClass("am-active");
                    }
                });
            }
        });
    };
    var figureInit = function() {
        $(".am-figure img").imgZoomToggle();
    };
    $(window).on("load", function() {
        figureInit();
    });
    exports.init = figureInit;
});
define("footer", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector", "ui.add2home", "util.cookie", "ui.modal" ], function(require, exports, module) {
    require("core");
    // add2home
    require("ui.add2home");
    var cookie = require("util.cookie"), modal = require("ui.modal"), $ = window.Zepto, footerInit = function() {
        // modal mode
        $(".am-footer-ysp").on("click", function() {
            $("#am-switch-mode").modal();
        });
        // switch mode
        // switch to desktop
        $('[data-rel="desktop"]').on("click", function(e) {
            e.preventDefault();
            if (window.AMPlatform) {
                // front end
                AMPlatform.util.goDesktop();
            } else {
                // back end
                cookie.set("allmobilize", "desktop", "", "/");
                window.location = window.location;
            }
        });
    };
    $(window).on("load", function() {
        footerInit();
    });
    exports.init = footerInit;
});
define("gallery", [ "zepto.touchgallery" ], function(require, exports, module) {
    var touchGallery = require("zepto.touchgallery");
    var $ = window.Zepto;
    var galleryInit = function() {
        var $themeOne = $(".am-gallery-one");
        $("[data-am-gallery] a").touchTouch();
        $themeOne.each(function() {
            galleryMore($(this));
        });
    };
    function galleryMore(object) {
        var moreData = $("<li class='am-gallery-more'><a href='javascript:;' class='am-btn am-btn-default'>更多 &gt;&gt;</a></li>");
        if (object.children().length > 6) {
            object.children().each(function(index) {
                if (index > 5) {
                    $(this).hide();
                }
            });
            object.find(".am-gallery-more").remove();
            object.append(moreData);
        }
        $(".am-gallery-more").on("click", function() {
            object.children().show();
            $(this).hide();
        });
    }
    $(function() {
        galleryInit();
    });
    exports.init = galleryInit;
});
define("gotop", [ "./ui.smooth-scroll", "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("./ui.smooth-scroll");
    var $ = window.Zepto;
    var UI = $.AMUI;
    var goTopInit = function() {
        $(".am-gotop").find("a").on("click", function(e) {
            e.preventDefault();
            $("body").smoothScroll(0);
        });
    };
    $(function() {
        goTopInit();
    });
    exports.init = goTopInit;
});
define("intro", [], function(require, exports, module) {
    var $ = window.Zepto;
});
define("list_news", [], function(require, exports, module) {
    var $ = window.Zepto, listNewsInit = function() {
        $(".am-list-news-one").each(function() {
            amListNewsMore($(this));
        });
    };
    function amListNewsMore(object) {
        var $amList = object.find(".am-list");
        var $listMore = "<a class='am-list-news-more am-btn am-btn-default' href='###'>更多 &gt;&gt;</a>";
        if ($amList.children().length > 6) {
            $amList.children().each(function(index) {
                if (index > 5) {
                    $(this).hide();
                }
            });
            object.find(".am-list-news-more").remove();
            object.append($listMore);
        }
        $(".am-list-news-more").on("click", function() {
            $amList.children().show();
            $(this).hide();
        });
    }
    $(function() {
        listNewsInit();
    });
    exports.init = listNewsInit;
});
define("map", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto;
    function addMapApi(callback) {
        var $mapApi0 = $("<script/>", {
            src: "http://api.map.baidu.com/getscript?type=quick&file=api&ak=WVAXZ05oyNRXS5egLImmentg&t=20140109092002"
        });
        $("body").append($mapApi0);
        $mapApi0.on("load", function() {
            var $mapApi1 = $("<script/>", {
                src: "http://api.map.baidu.com/getscript?type=quick&file=feature&ak=WVAXZ05oyNRXS5egLImmentg&t=20140109092002"
            });
            $("body").append($mapApi1);
            $mapApi1.on("load", function() {
                var script = document.createElement("script");
                script.textContent = "(" + callback.toString() + ")();";
                $("body")[0].appendChild(script);
            });
        });
    }
    function addBdMap() {
        // 如果使用 $ 选择符，minify 以后会报错: $ is undefined
        // 即使传入 $ 也无效，改为使用原生方法
        // 这个函数作为 callback 会插入到 body 以后才执行，应该是 $ 引用错误导致
        var content = document.querySelector(".am-map"), defaultLng = 116.331398, //经度默认值
        defaultLat = 39.897445, //纬度默认值
        name = content.getAttribute("data-name"), address = content.getAttribute("data-address"), lng = content.getAttribute("data-longitude") || defaultLng, lat = content.getAttribute("data-latitude") || defaultLat;
        var map = new BMap.Map("bd-map");
        //实例化一个地理坐标点
        var point = new BMap.Point(lng, lat);
        //设初始化地图, options: 3-18
        map.centerAndZoom(point, 18);
        //添加地图缩放控件
        map.addControl(new BMap.ZoomControl());
        var opts = {
            width: 200,
            // 信息窗口宽度
            //height: 'auto',     // 信息窗口高度
            title: name
        };
        // 创建信息窗口对象
        var infoWindow = new BMap.InfoWindow("地址：" + address, opts);
        // 创建地址解析器实例
        var myGeo = new BMap.Geocoder();
        //判断有没有使用经纬度
        if (lng == defaultLng && lat == defaultLat) {
            // 使用地址反解析来设置地图
            // 将地址解析结果显示在地图上,并调整地图视野
            myGeo.getPoint(address, function(point) {
                if (point) {
                    map.centerAndZoom(point, 17);
                    map.addOverlay(new BMap.Marker(point));
                    map.openInfoWindow(infoWindow, point);
                }
            }, "");
        } else {
            // 使用经纬度来设置地图
            myGeo.getLocation(point, function(result) {
                map.centerAndZoom(point, 17);
                map.addOverlay(new BMap.Marker(point));
                if (address) {
                    map.openInfoWindow(infoWindow, point);
                } else {
                    map.openInfoWindow(new BMap.InfoWindow(address, opts), point);
                }
            });
        }
    }
    var mapInit = function() {
        $(".am-map").length && addMapApi(addBdMap);
    };
    $(document).on("ready", mapInit);
    exports.init = mapInit;
});
define("menu", [ "nav", "zepto.outerdemension", "zepto.extend.data", "zepto.extend.selector", "zepto.extend.fx", "core", "ui.offcanvas" ], function(require, exports, module) {
    require("nav");
    require("ui.offcanvas");
    var $ = window.Zepto;
    var UI = $.AMUI;
    var menuInit = function() {
        //one theme variable
        var $this, $next, $width, iNow, aNum, $menuLv2, $menuLv3;
        //排除掉one主题
        $(".am-menu").not("[data-am-nav]").not(".am-menu-one").each(function() {
            var nav = $(this);
            if (!nav.data("nav")) {
                var obj = new UI.nav(nav, nav.data("am-nav") ? UI.utils.options(nav.data("am-nav")) : {});
            }
        });
        // has one class
        if ($(".am-menu").hasClass("am-menu-one")) {
            $this = $(".am-menu-one");
            $next = $("<a>").attr({
                "class": "am-menu-next",
                href: "javascript:;"
            });
            $width = 0;
            iNow = -1;
            aNum = [];
            $menuLv2 = $(".am-menu-lv2");
            $menuLv3 = $(".am-menu-lv3");
            $this.children("li").find("a").eq(0).attr("href", "javascript:;").addClass("am-menu-prev am-menu-disabled");
            $this.find("li").eq(0).append($next);
            $this.find(".am-menu-lv2").wrap("<div class='am-menu-wrap'></div>");
            $menuLv2.children("li").children("a").each(function() {
                $(this).parent().width($(this).width());
                $width += $(this).width();
                aNum.push($width);
            });
            $menuLv2.find(".am-parent").each(function() {
                var $firstA = $(this).find("a"), $li = $("<li class='am-menu-item-more'><a class='am-menu-item-close' href='javascript:;'>×</a><a class='am-menu-item-into' href=" + $firstA.attr("href") + ">进入" + $firstA.html() + "</a></li>");
                $firstA.attr("href", "javascript:;");
                $(this).find(".am-menu-lv3").append($li);
            });
            $menuLv2.width($width);
            $menuLv3.width($(".am-menu-lv1").width() - 20);
            // 减去Menu 左右padding
            // FIXME: ide border
            //$menuLv3.width($('.am-menu-lv1').width() - 22);// 减去Menu 左右padding
            $(".am-menu-wrap .am-parent").children("a").on("click", function() {
                if ($(this).hasClass("active")) {
                    $(this).removeClass("active");
                    $(this).siblings(".am-menu-lv3").animate({
                        opacity: 0
                    }, "fast", "linear", function() {
                        $(this).css("display", "none");
                    });
                } else {
                    offAll();
                    $(this).addClass("active");
                    $(this).siblings(".am-menu-lv3").css("display", "block").animate({
                        left: -$(this).offset().left + 10,
                        opacity: 1
                    }, "fast", "linear");
                }
            });
            $(".am-menu-next").on("click", function() {
                offAll();
                if (-aNum[iNow] + parseInt($menuLv2.css("left")) < -$menuLv2.width() + $menuLv2.parent().width() - $next.width() * 2) {
                    $menuLv2.animate({
                        left: -$menuLv2.width() + $menuLv2.parent().width() - $next.width() * 2
                    }, "fast", "linear");
                    $(this).addClass("am-menu-disabled");
                } else {
                    iNow++;
                    $menuLv2.animate({
                        left: -aNum[iNow]
                    }, "fast", "linear");
                    $(".am-menu-prev").removeClass("am-menu-disabled");
                }
            });
            $(".am-menu-item-close").on("click", function() {
                offAll();
            });
            $(".am-menu-prev").on("click", function() {
                offAll();
                if (iNow <= -1) {
                    $menuLv2.animate({
                        left: 0
                    }, "fast", "linear");
                    $(this).addClass("am-menu-disabled");
                } else {
                    iNow--;
                    $menuLv2.animate({
                        left: -aNum[iNow]
                    }, "fast", "linear");
                    $(".am-menu-next").removeClass("am-menu-disabled");
                }
            });
            drag($menuLv2);
        }
        /*
         *  offAll menu children active
         */
        function offAll() {
            $(".am-menu-wrap .am-parent").children("a").removeClass("active").siblings(".am-menu-lv3").animate({
                opacity: 0
            }, "fast", "linear").css("display", "none");
        }
        /*
         *  drag menu children
         *  @obj Zepto object
         */
        function drag(obj) {
            var disX, downX, nOffsetLeft = 0;
            obj.on("touchstart MSPointerDown pointerdown", function(ev) {
                offAll();
                ev.preventDefault();
                var oTarget = ev.targetTouches[0];
                disX = oTarget.clientX - $(this).offset().left;
                downX = oTarget.clientX;
                $(document).on("touchmove MSPointerMove pointermove", fnMove);
                $(document).on("touchend MSPointerUp pointerup", fnUp);
            });
            function fnUp(ev) {
                $.each(aNum, function(index, item) {
                    nOffsetLeft += -aNum[index];
                    if (parseInt(obj.css("left")) >= nOffsetLeft) {
                        iNow = index;
                        return false;
                    }
                });
                nOffsetLeft = 0;
                $(document).off("touchend MSPointerUp pointerup", fnUp);
                $(document).off("touchmove MSPointerMove pointermove", fnMove);
            }
            function fnMove(ev) {
                ev.preventDefault();
                var oTarget = ev.targetTouches[0];
                var nLeft = oTarget.clientX - disX;
                // ->
                if (nLeft > 0) {
                    nLeft = 0;
                }
                // <-
                if (nLeft < -obj.width() + obj.parent().width() - $next.width() * 2) {
                    nLeft = -obj.width() + obj.parent().width() - $next.width() * 2;
                }
                obj.css("left", nLeft);
            }
        }
    };
    $(function() {
        menuInit();
    });
    exports.init = menuInit;
});
define("nav", [ "./zepto.outerdemension", "./zepto.extend.data", "./zepto.extend.selector", "./zepto.extend.fx", "./core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("./zepto.outerdemension");
    require("./zepto.extend.data");
    // selector extend
    require("./zepto.extend.selector");
    require("./zepto.extend.fx");
    require("./core");
    var $ = window.Zepto;
    var UI = $.AMUI;
    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/nav.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */
    var Nav = function(element, options) {
        var $this = this, $element = $(element);
        if ($element.data("nav")) return;
        this.options = $.extend({}, this.options, options);
        this.element = $element.on("click", this.options.trigger, function(e) {
            e.preventDefault();
            // trigger link
            var ele = $(this);
            $this.toggleNav(ele.parent(".am-parent"));
        });
        // wrap sub menu
        this.element.find(this.options.lists).each(function() {
            var $ele = $(this), parent = $ele.parent();
            // li.am-parent
            var navHeight = getHeight($ele);
            parent.data("list-container", $ele).attr("data-nav-height", navHeight);
        });
        this.element.data("nav", this);
    };
    $.extend(Nav.prototype, {
        options: {
            trigger: ".am-parent > a",
            lists: ".am-parent > ul",
            multiple: false
        },
        toggleNav: function(li, noanimation) {
            var element = this.element, $li = $(li);
            // 是否允许同时展开多个菜单
            if (!this.options.multiple) {
                $li.siblings(".am-open").each(function() {
                    if ($(this).data("list-container")) {
                        $(this).removeClass("am-open").data("list-container").animate({
                            height: 0
                        }, function() {});
                    }
                });
            }
            $li.toggleClass("am-open");
            var targetMenu = $li.data("list-container"), targetMenuPosition = targetMenu.css("position");
            if ($li.data("list-container")) {
                if (noanimation) {
                    $li.data("list-container").height($li.hasClass("am-open") ? "auto" : 0);
                } else {
                    // 三级菜单展开时增加二级菜单容器高度
                    var parentWrap = $li.parents(".am-parent");
                    // 二级菜单
                    if (parentWrap.length > 0) {
                        var parentNavWrap = parentWrap.eq(0).data("list-container");
                        // 三级菜单展开且三级菜单非绝对定位时增加父级容器高度
                        var addHeight = $li.hasClass("am-open") && targetMenuPosition != "absolute" ? Number($li.attr("data-nav-height")) : 0;
                        parentNavWrap.animate({
                            height: Number(parentWrap.attr("data-nav-height")) + addHeight
                        });
                        // 三级菜单绝对定位时
                        if (targetMenuPosition == "absolute") {
                            parentNavWrap.css({
                                overflow: $li.hasClass("am-open") ? "visible" : "hidden"
                            });
                        }
                    }
                    $li.data("list-container").animate({
                        height: $li.hasClass("am-open") ? $li.attr("data-nav-height") + "px" : 0
                    });
                    // 一级菜单闭合时闭合所有展开子菜单
                    var subNavs = $li.find(".am-menu-sub");
                    // console.log($li);
                    if (subNavs.length > 0 && !$li.hasClass("am-open")) {
                        // console.log(subNavs.length);
                        subNavs.each(function(index, item) {
                            $(item).animate({
                                height: 0,
                                overflow: "hidden"
                            });
                            $(item).parent(".am-parent.am-open").not($li).removeClass("am-open");
                        });
                    }
                }
            }
        }
    });
    UI["nav"] = Nav;
    // helper
    function getHeight(ele) {
        var $ele = ele, height = "auto";
        if ($ele.is(":visible")) {
            height = $ele.outerHeight();
        } else {
            var position = $ele.css("position");
            // show element if it is hidden (it is needed if display is none)
            $ele.show();
            // place it so it displays as usually but hidden
            $ele.css({
                position: "absolute",
                visibility: "hidden",
                height: "auto"
            });
            // get naturally height
            height = $ele.outerHeight();
            //console.log($ele.outerHeight(), $ele.height());
            // set initial css for animation
            $ele.css({
                position: position,
                visibility: "visible",
                overflow: "hidden",
                height: 0
            });
        }
        return height;
    }
    // init code
    $(function() {
        $("[data-am-nav]").each(function() {
            var nav = $(this);
            if (!nav.data("nav")) {
                var obj = new Nav(nav, UI.utils.options(nav.data("am-nav")));
            }
        });
    });
});
define("navbar", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector", "util.qrcode", "ui.modal" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto, qrcode = require("util.qrcode"), modal = require("ui.modal");
    var navbarInit = function() {
        var _parent = $(".am-navbar"), parentUl = _parent.find("ul"), subLi = _parent.find("li"), oneWidth = subLi.width(), minWidth = 100, //每个li最小宽度
        _more = null, _moreList = null, onOff = true, onOffCreat = true, // 防止多次的创建
        $body = $("body");
        var $share = $("[data-am-navbar-share]");
        var $qrcode = $("[data-am-navbar-qrcode]");
        var navbarPosition = _parent.css("position");
        if (navbarPosition == "fixed") {
            $body.addClass("with-fixed-navbar");
        }
        if ($qrcode.length) {
            var qrImg = $("[data-am-navbar-qrcode]").attr("data-am-navbar-qrcode");
            var url = window.location.href;
            var qrData = $("<div class=\"am-modal am-modal-no-btn\" id='am-navbar-boxqrcode'>" + "<div class='am-modal-dialog' id='am-navbar-qrcode-data'></div>" + "</div>");
            $body.append(qrData);
            //判断上传自定义的二维码没有，否则生成二维码
            if (qrImg) {
                $("#am-navbar-qrcode-data").html('<img src="' + qrImg + '"/>');
            } else {
                var qrnode = new qrcode({
                    render: "canvas",
                    correctLevel: 0,
                    text: url,
                    width: 190,
                    height: 190,
                    background: "#fff",
                    foreground: "#000"
                });
                $("#am-navbar-qrcode-data").html(qrnode);
            }
        }
        //添加share className
        $share.addClass("am-navbar-share");
        $qrcode.addClass("am-navbar-qrcode");
        if ($share.length) {
            //share start
            window._bd_share_config = {
                common: {
                    bdSnsKey: {},
                    bdText: "",
                    bdMini: "2",
                    bdMiniList: false,
                    bdPic: "",
                    bdStyle: "1",
                    bdSize: "16"
                },
                share: {
                    bdSize: 24
                }
            };
            $body.append($("<script />", {
                src: "http://bdimg.share.baidu.com/static/api/js/share.js?v=89343201.js?cdnversion=" + ~(-new Date() / 36e5)
            }));
            var shareData = '<div class="bdsharebuttonbox">' + '<div class="am-modal-actions am-modal-out" id="am-navbar-share">' + '<div class="am-modal-actions-group">' + '<ul class="am-list">' + '<li class="am-modal-actions-header" data-cmd="more">分享到</li>' + '<li><a href="#" class="am-icon-qq" data-cmd="qzone" title="分享到QQ空间">QQ空间</a></li>' + '<li><a href="#" class="am-icon-weibo" data-cmd="tsina" title="分享到新浪微博">新浪微博</a></li>' + '<li><a href="#" class="am-icon-tencent-weibo" data-cmd="tqq" title="分享到腾讯微博">腾讯微博</a></li>' + '<li><a href="#" class="am-icon-renren" data-cmd="renren" title="分享到人人网">人人网</a></li>' + '<li><a href="#" class="am-icon-wechat" data-cmd="weixin" title="分享到微信">微信</a></li>' + "</ul>" + "</div>" + '<div class="am-modal-actions-group"><button type="button" class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button></div>' + "</div>" + "</div>";
            $body.append(shareData);
            $share.on("click", function(event) {
                event.preventDefault();
                $("#am-navbar-share").modal();
            });
        }
        if ($qrcode.length) {
            //qrcode start
            $qrcode.on("click", function(event) {
                event.preventDefault();
                $("#am-navbar-boxqrcode").modal();
            });
        }
        //qrcode end
        if (_parent.length) {
            $body.append($('<ul class="am-navbar-moreList"></ul>'));
        }
        if (_parent.find("li").length * _parent.find("li").width() > $(window).width()) {
            //如果li没有完全展示
            //替换父级的class
            displaceClass(_parent.find("li").length, parentUl);
            var nowWidth = _parent.find("li").width();
            if (nowWidth < minWidth) {
                if (onOffCreat) {
                    addMore();
                    onOffCreat = false;
                }
                displaceClass(liLength(), parentUl);
                addMoreLi(liLength());
            }
        }
        //有问题的代码：
        /*$(window).on("resize",function(){

         if(_parent.find("li").length * _parent.find("li").width() > $(window).width()){ //如果li没有完全展示

         //替换父级的class
         displaceClass(_parent.find("li").length,parentUl);
         var nowWidth = _parent.find("li").width();

         //现在的宽度小于最小宽度
         if(nowWidth < minWidth){

         if(onOffCreat){
         addMore();
         onOffCreat = false;
         }
         displaceClass(liLength(),parentUl);
         addMoreLi(liLength());
         }else{

         addParentLi(liLength());

         if($(".am-navbar-moreList").children().length){
         removeMore();
         onOffCreat = true;
         }

         displaceClass(liLength(),parentUl);
         }
         }else{
         displaceClass(_parent.find("li").length,parentUl);
         if(_parent.find("li").width < minWidth){
         console.log("小于")
         }
         }

         })*/
        _more = $(".am-navbar-more");
        _moreList = $(".am-navbar-moreList");
        _parent.on("click", ".am-navbar-more", function() {
            if (onOff) {
                _moreList.css({
                    bottom: _moreList.height(),
                    display: "block"
                }).animate({
                    bottom: 49
                }, {
                    duration: "fast",
                    complete: function() {
                        _more.addClass("am-navbar-more-active");
                    }
                });
                onOff = !onOff;
            } else {
                _moreList.animate({
                    bottom: -_moreList.height()
                }, {
                    complete: function() {
                        $(this).css("display", "none");
                        _more.removeClass("am-navbar-more-active");
                    }
                });
                onOff = !onOff;
            }
        });
        //添加more
        function addMore() {
            parentUl.append($('<li class="am-navbar-item am-navbar-more"><a href="javascript:;"><span class="am-icon-chevron-up"></span>更多</a></li>'));
        }
        //删除more
        function removeMore() {
            parentUl.find(".am-navbar-more").remove();
        }
        //计算合适的长度
        function liLength() {
            return parseInt($(window).width() / minWidth);
        }
        //移出parent下的li,并添加到moreList里面
        function addMoreLi(len) {
            subLi.not(".am-navbar-more").each(function(index) {
                if (index > len - 2) {
                    $(this).appendTo($(".am-navbar-moreList"));
                }
            });
        }
        //移出moreList里面的li,并添加到parent下面
        function addParentLi(len) {
            $(".am-navbar-moreList").children().first().appendTo(parentUl);
        }
        //替换class
        function displaceClass(num, object) {
            var $className = object.attr("class").replace(/sm-block-grid-\d/, "sm-block-grid-" + num);
            object.attr("class", $className);
        }
    };
    // DOMContentLoaded
    $(function() {
        navbarInit();
    });
    exports.init = navbarInit;
});
define("pagination", [], function(require, exports, module) {});
define("paragraph", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector", "zepto.pinchzoom" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto;
    // PinchZoom Plugin
    var PinchZoom = require("zepto.pinchzoom");
    var paragraphInit;
    $.fn.paragraphZoomToggle = function() {
        var $warpHead, $pinch, $zoomWrap, onOff = true, // 防止重复创建
        $wrapDom = $("<div class='am-paragraph-wrap'><header></header><div class='pinch-zoom'></div></div>");
        $zoomWrap = $(".am-paragraph-wrap");
        $warpHead = $(".am-paragraph-wrap header");
        $pinch = $zoomWrap.find(".pinch-zoom");
        this.each(function() {
            $(this).on("click", function() {
                if (onOff && $(".am-paragraph").length) {
                    $("body").append($wrapDom);
                    $zoomWrap = $(".am-paragraph-wrap");
                    $pinch = $zoomWrap.find(".pinch-zoom");
                    $warpHead = $zoomWrap.find("header");
                    $pinch.each(function() {
                        new PinchZoom($(this), {});
                    });
                    onOff = false;
                }
                $pinch.html(this.outerHTML);
                if ($(this).attr("alt")) {
                    $warpHead.html($(this).attr("alt"));
                } else {
                    $warpHead.html("返回");
                }
                $zoomWrap.addClass("am-active").find("img").width($(window).width());
            });
        });
    };
    $.fn.paragraphTable = function(objWidth) {
        var This = $(this), distX = 0, disX = 0, disY = 0, downX, downY, $parent, scrollY;
        if (objWidth > $("body").width()) {
            This.wrap("<div class='am-paragraph-table-container'><div class='am-paragraph-table-scroller'></div></div>");
            $parent = This.parent();
            $parent.width(objWidth);
            $parent.height(This.height());
            $parent.parent().height(This.height() + 20);
            $parent.on("touchstart MSPointerDown pointerdown", function(ev) {
                var oTarget = ev.targetTouches[0];
                distX = oTarget.clientX - $(this).offset().left;
                downX = oTarget.clientX;
                downY = oTarget.clientY;
                scrollY = undefined;
                $(document).on("touchmove MSPointerMove pointermove", fnMove);
                $(document).on("touchend MSPointerUp pointerup", fnUp);
            });
        }
        function fnUp(ev) {
            ev.preventDefault();
            var oTarget = ev.changedTouches[0];
            var L = $parent.offset().left;
            // ->
            if (L > 10) {
                $parent.animate({
                    left: 10
                }, 500, "ease-out");
            }
            //<-
            if (L < -$parent.width() + $(window).width() - 10) {
                $parent.animate({
                    left: -$parent.width() + $(window).width() - 10
                }, 500, "ease-out");
            }
            $(document).off("touchend MSPointerUp pointerup", fnUp);
            $(document).off("touchmove MSPointerMove pointermove", fnMove);
        }
        function fnMove(ev) {
            var oTarget = ev.targetTouches[0];
            disX = oTarget.clientX - downX;
            disY = oTarget.clientY - downY;
            if (typeof scrollY == "undefined") {
                scrollY = !!(scrollY || Math.abs(disX) < Math.abs(disY));
            }
            if (!scrollY) {
                ev.preventDefault();
                This.parent().css("left", oTarget.clientX - distX);
            }
        }
    };
    paragraphInit = function() {
        var $body = $("body"), $paragraph = $(".am-paragraph"), $tableWidth;
        if ($paragraph.length && $paragraph.attr("data-am-imgParagraph")) {
            $paragraph.find("img").paragraphZoomToggle();
            $body.on("click", ".am-paragraph-wrap", function(e) {
                e.preventDefault();
                var target = e.target;
                // Img is using pinch zoom
                if (!$(target).is("img")) {
                    $(this).toggleClass("am-active");
                }
            });
        }
        if ($paragraph.length && $paragraph.attr("data-am-tableParagraph")) {
            $paragraph.find("table").each(function() {
                $tableWidth = $(this).width();
                $(this).paragraphTable($tableWidth);
            });
        }
    };
    $(window).on("load", function() {
        paragraphInit();
    });
    exports.init = paragraphInit;
});
define("slider", [ "zepto.flexslider" ], function(require, exports, module) {
    var $ = window.Zepto;
    require("zepto.flexslider");
    var sliderInit = function() {
        $(".am-slider").not(".am-slider-manual").each(function(i, item) {
            var options = $(item).attr("data-slider-config");
            if (options) {
                $(item).flexslider($.parseJSON(options));
            } else {
                $(item).flexslider();
            }
        });
    };
    $(document).on("ready", sliderInit);
    exports.init = sliderInit;
});
define("sohucs", [], function(require, exports, module) {
    var $ = window.Zepto;
    var sohuCSInit = function() {
        if (!$("#SOHUCS").length) return;
        var $sohucs = $('[data-am-widget="sohucs"]'), appid = $sohucs.attr("data-am-sohucs-appid"), conf = $sohucs.attr("data-am-sohucs-conf"), $cy = $("<script></script>", {
            charset: "utf-8",
            id: "changyan_mobile_js",
            src: "http://changyan.sohu.com/upload/mobile/wap-js/changyan_mobile.js?client_id=" + appid + "&conf=" + conf
        });
        $("body").append($cy);
    };
    // Lazy load
    $(window).on("load", sohuCSInit);
    exports.init = sohuCSInit;
});
define("tabs", [ "zepto.extend.touch", "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("zepto.extend.touch");
    require("core");
    var $ = window.Zepto;
    var tabsInit = function() {
        $(".am-tabs").each(function() {
            amTabs($(this));
        });
    };
    function amTabs(parent) {
        var $tabsContent = parent.find(".am-tabs-bd-content"), $tabsDiv = $tabsContent.children(), oneWidth, iNow = 0, disX, disY, downY, downX, $tabLi = parent.find(".am-tabs-hd").children();
        //设置tabsdiv宽度
        $tabsContent.width($tabsContent.parent().width() * $tabsDiv.length);
        $tabsDiv.width($tabsContent.parent().width());
        oneWidth = $tabsDiv.width();
        $(window).on("resize", function() {
            $tabsContent.width($tabsContent.parent().width() * $tabsDiv.length);
            $tabsDiv.width($tabsContent.parent().width());
            oneWidth = $tabsDiv.width();
        });
        /*$tabsContent.on("touchstart MSPointerDown pointerdown", function(ev){
         ev.preventDefault();
         var oTarget = ev.targetTouches[0];
         disX = oTarget.clientX - $tabsContent.offset().left;
         disY = oTarget.clientY - $tabsContent.offset().top;
         downX = oTarget.clientX;
         downY = oTarget.clientY;
         $( $tabsContent ).on("touchmove MSPointerMove pointermove", fnMove);
         $( $tabsContent ).on("touchend MSPointerUp pointerup", fnUp);
         });*/
        $tabsContent.swipeRight(function() {
            iNow--;
            if (iNow < 0) {
                iNow = 0;
            }
            $tabsContent.animate({
                left: -iNow * oneWidth
            });
            $tabLi.removeClass("am-tabs-hd-active");
            $tabLi.eq(iNow).addClass("am-tabs-hd-active");
        });
        $tabsContent.swipeLeft(function() {
            iNow++;
            if (iNow > $tabsDiv.length - 1) {
                iNow = $tabsDiv.length - 1;
            }
            $tabsContent.animate({
                left: -iNow * oneWidth
            });
            $tabLi.removeClass("am-tabs-hd-active");
            $tabLi.eq(iNow).addClass("am-tabs-hd-active");
        });
        $tabLi.on("click", function() {
            iNow = $(this).index();
            $tabLi.removeClass("am-tabs-hd-active");
            $tabLi.eq(iNow).addClass("am-tabs-hd-active");
            $tabsContent.animate({
                left: -iNow * oneWidth
            });
        });
    }
    $(function() {
        tabsInit();
    });
    exports.init = tabsInit;
});
define("titlebar", [], function(require, exports, module) {});
define("ui.accordion", [ "./zepto.extend.fx", "./zepto.extend.selector", "./zepto.extend.data" ], function(require, exports, module) {
    var $ = window.Zepto;
    require("./zepto.extend.fx");
    require("./zepto.extend.selector");
    require("./zepto.extend.data");
    /**
     * @via https://github.com/Semantic-Org/Semantic-UI/blob/master/src/modules/accordion.js
     * @license https://github.com/Semantic-Org/Semantic-UI/blob/master/LICENSE.md
     */
    $.fn.accordion = function(parameters) {
        var $allModules = $(this), query = arguments[0], methodInvoked = typeof query == "string", queryArguments = [].slice.call(arguments, 1), returnedValue;
        $allModules.each(function() {
            var settings = $.isPlainObject(parameters) ? $.extend(true, {}, $.fn.accordion.settings, parameters) : $.extend({}, $.fn.accordion.settings), className = settings.className, namespace = settings.namespace, selector = settings.selector, eventNamespace = "." + namespace, moduleNamespace = "module-" + namespace, $module = $(this), $title = $module.find(selector.title), $content = $module.find(selector.content), $item = $module.find(selector.item), element = this, instance = $module.data(moduleNamespace), module;
            module = {
                initialize: function() {
                    // initializing
                    $title.on("click" + eventNamespace, module.event.click);
                    module.instantiate();
                },
                instantiate: function() {
                    instance = module;
                    $module.data(moduleNamespace, module);
                },
                destroy: function() {
                    $module.removeData(moduleNamespace);
                    $title.off(eventNamespace);
                },
                event: {
                    click: function() {
                        var $activeTitle = $(this), index = $item.index($activeTitle.parent(selector.item));
                        module.toggle(index);
                    }
                },
                toggle: function(index) {
                    var $activeItem = $item.eq(index), contentIsOpen = $activeItem.hasClass(className.active);
                    if (contentIsOpen) {
                        if (settings.collapsible) {
                            module.close(index);
                        }
                    } else {
                        module.open(index);
                    }
                },
                open: function(index) {
                    var $activeItem = $item.eq(index), $activeContent = $activeItem.next(selector.content);
                    if (!settings.multiple) {
                        $item.not($activeItem).removeClass(className.active);
                    }
                    $activeItem.addClass(className.active);
                    $activeContent.animate(settings.duration, settings.easing, function() {
                        $.proxy(settings.onOpen, $activeContent)();
                        $.proxy(settings.onChange, $activeContent)();
                    });
                },
                close: function(index) {
                    var $activeItem = $item.eq(index), $activeContent = $activeItem.find(selector.content);
                    $activeItem.removeClass(className.active);
                    $activeContent.animate(settings.duration, settings.easing, function() {
                        $.proxy(settings.onClose, $activeContent)();
                        $.proxy(settings.onChange, $activeContent)();
                    });
                },
                setting: function(name, value) {
                    if ($.isPlainObject(name)) {
                        $.extend(true, settings, name);
                    } else if (value !== undefined) {
                        settings[name] = value;
                    } else {
                        return settings[name];
                    }
                },
                internal: function(name, value) {
                    if (value !== undefined) {
                        if ($.isPlainObject(name)) {
                            $.extend(true, module, name);
                        } else {
                            module[name] = value;
                        }
                    } else {
                        return module[name];
                    }
                },
                invoke: function(query, passedArguments, context) {
                    var object = instance, maxDepth, found, response;
                    passedArguments = passedArguments || queryArguments;
                    context = element || context;
                    if (typeof query == "string" && object !== undefined) {
                        query = query.split(/[\. ]/);
                        maxDepth = query.length - 1;
                        $.each(query, function(depth, value) {
                            var camelCaseValue = depth != maxDepth ? value + query[depth + 1].charAt(0).toUpperCase() + query[depth + 1].slice(1) : query;
                            if ($.isPlainObject(object[camelCaseValue]) && depth != maxDepth) {
                                object = object[camelCaseValue];
                            } else if (object[camelCaseValue] !== undefined) {
                                found = object[camelCaseValue];
                                return false;
                            } else if ($.isPlainObject(object[value]) && depth != maxDepth) {
                                object = object[value];
                            } else if (object[value] !== undefined) {
                                found = object[value];
                                return false;
                            } else {
                                return false;
                            }
                        });
                    }
                    if ($.isFunction(found)) {
                        response = found.apply(context, passedArguments);
                    } else if (found !== undefined) {
                        response = found;
                    }
                    if ($.isArray(returnedValue)) {
                        returnedValue.push(response);
                    } else if (returnedValue !== undefined) {
                        returnedValue = [ returnedValue, response ];
                    } else if (response !== undefined) {
                        returnedValue = response;
                    }
                    return found;
                }
            };
            if (methodInvoked) {
                if (instance === undefined) {
                    module.initialize();
                }
                module.invoke(query);
            } else {
                if (instance !== undefined) {
                    module.destroy();
                }
                module.initialize();
            }
        });
        return returnedValue !== undefined ? returnedValue : this;
    };
    $.fn.accordion.settings = {
        name: "Accordion",
        namespace: "accordion",
        multiple: false,
        collapsible: true,
        duration: 500,
        easing: "ease-in-out",
        onOpen: function() {},
        onClose: function() {},
        onChange: function() {},
        className: {
            active: "am-active"
        },
        selector: {
            item: ".am-accordion-item",
            title: ".am-accordion-title",
            content: ".am-accordion-content"
        }
    };
});
define("ui.add2home", [], function(require, exports, module) {
    /*!
     * Add to Homescreen v2.0.11 ~ Copyright (c) 2013 Matteo Spinelli, http://cubiq.org
     * Released under MIT license, http://cubiq.org/license
     */
    var addToHome = function(w) {
        var nav = w.navigator, isIDevice = "platform" in nav && /iphone|ipod|ipad/gi.test(nav.platform), isIPad, isRetina, isSafari, isStandalone, OSVersion, startX = 0, startY = 0, lastVisit = 0, isExpired, isSessionActive, isReturningVisitor, balloon, overrideChecks, positionInterval, closeTimeout, options = {
            autostart: true,
            // Automatically open the balloon
            returningVisitor: false,
            // Show the balloon to returning visitors only (setting this to true is highly recommended)
            animationIn: "drop",
            // drop || bubble || fade
            animationOut: "fade",
            // drop || bubble || fade
            startDelay: 2e3,
            // 2 seconds from page load before the balloon appears
            lifespan: 15e3,
            // 15 seconds before it is automatically destroyed
            bottomOffset: 14,
            // Distance of the balloon from bottom
            expire: 0,
            // Minutes to wait before showing the popup again (0 = always displayed)
            message: "",
            // Customize your message or force a language ('' = automatic)
            touchIcon: false,
            // Display the touch icon
            arrow: true,
            // Display the balloon arrow
            hookOnLoad: true,
            // Should we hook to onload event? (really advanced usage)
            closeButton: true,
            // Let the user close the balloon
            iterations: 100
        }, intl = {
            en_us: "Install this web app on your %device: tap %icon and then <strong>Add to Home Screen</strong>.",
            zh_cn: "您可以将此应用安装到您的 %device 上。请按 %icon 然后选择<strong>添加至主屏幕</strong>。",
            zh_tw: "您可以將此應用程式安裝到您的 %device 上。請按 %icon 然後點選<strong>加入主畫面螢幕</strong>。"
        };
        function init() {
            // Preliminary check, all further checks are performed on iDevices only
            if (!isIDevice) return;
            var now = Date.now(), i;
            // Merge local with global options
            if (w.addToHomeConfig) {
                for (i in w.addToHomeConfig) {
                    options[i] = w.addToHomeConfig[i];
                }
            }
            if (!options.autostart) options.hookOnLoad = false;
            isIPad = /ipad/gi.test(nav.platform);
            isRetina = w.devicePixelRatio && w.devicePixelRatio > 1;
            isSafari = /Safari/i.test(nav.appVersion) && !/CriOS/i.test(nav.appVersion);
            isStandalone = nav.standalone;
            OSVersion = nav.appVersion.match(/OS (\d+_\d+)/i);
            OSVersion = OSVersion && OSVersion[1] ? +OSVersion[1].replace("_", ".") : 0;
            lastVisit = +w.localStorage.getItem("addToHome");
            isSessionActive = w.sessionStorage.getItem("addToHomeSession");
            isReturningVisitor = options.returningVisitor ? lastVisit && lastVisit + 28 * 24 * 60 * 60 * 1e3 > now : true;
            if (!lastVisit) lastVisit = now;
            // If it is expired we need to reissue a new balloon
            isExpired = isReturningVisitor && lastVisit <= now;
            if (options.hookOnLoad) w.addEventListener("load", loaded, false); else if (!options.hookOnLoad && options.autostart) loaded();
        }
        function loaded() {
            w.removeEventListener("load", loaded, false);
            if (!isReturningVisitor) w.localStorage.setItem("addToHome", Date.now()); else if (options.expire && isExpired) w.localStorage.setItem("addToHome", Date.now() + options.expire * 6e4);
            if (!overrideChecks && (!isSafari || !isExpired || isSessionActive || isStandalone || !isReturningVisitor)) return;
            var touchIcon = "", platform = nav.platform.split(" ")[0], language = nav.language.replace("-", "_");
            balloon = document.createElement("div");
            balloon.id = "addToHomeScreen";
            balloon.style.cssText += "left:-9999px;-webkit-transition-property:-webkit-transform,opacity;-webkit-transition-duration:0;-webkit-transform:translate3d(0,0,0);position:" + (OSVersion < 5 ? "absolute" : "fixed");
            // Localize message
            if (options.message in intl) {
                // You may force a language despite the user's locale
                language = options.message;
                options.message = "";
            }
            if (options.message === "") {
                // We look for a suitable language (defaulted to en_us)
                options.message = language in intl ? intl[language] : intl["en_us"];
            }
            if (options.touchIcon) {
                touchIcon = isRetina ? document.querySelector('head link[rel^=apple-touch-icon][sizes="114x114"],head link[rel^=apple-touch-icon][sizes="144x144"],head link[rel^=apple-touch-icon]') : document.querySelector('head link[rel^=apple-touch-icon][sizes="57x57"],head link[rel^=apple-touch-icon]');
                if (touchIcon) {
                    touchIcon = '<span style="background-image:url(' + touchIcon.href + ')" class="addToHomeTouchIcon"></span>';
                }
            }
            balloon.className = (OSVersion >= 7 ? "addToHomeIOS7 " : "") + (isIPad ? "addToHomeIpad" : "addToHomeIphone") + (touchIcon ? " addToHomeWide" : "");
            balloon.innerHTML = touchIcon + options.message.replace("%device", platform).replace("%icon", OSVersion >= 4.2 ? '<span class="addToHomeShare"></span>' : '<span class="addToHomePlus">+</span>') + (options.arrow ? '<span class="addToHomeArrow"' + (OSVersion >= 7 && isIPad && touchIcon ? ' style="margin-left:-32px"' : "") + "></span>" : "") + (options.closeButton ? '<span class="addToHomeClose">×</span>' : "");
            document.body.appendChild(balloon);
            // Add the close action
            if (options.closeButton) balloon.addEventListener("click", clicked, false);
            if (!isIPad && OSVersion >= 6) window.addEventListener("orientationchange", orientationCheck, false);
            setTimeout(show, options.startDelay);
        }
        function show() {
            var duration, iPadXShift = 208;
            // Set the initial position
            if (isIPad) {
                if (OSVersion < 5) {
                    startY = w.scrollY;
                    startX = w.scrollX;
                } else if (OSVersion < 6) {
                    iPadXShift = 160;
                } else if (OSVersion >= 7) {
                    iPadXShift = 143;
                }
                balloon.style.top = startY + options.bottomOffset + "px";
                balloon.style.left = Math.max(startX + iPadXShift - Math.round(balloon.offsetWidth / 2), 9) + "px";
                switch (options.animationIn) {
                  case "drop":
                    duration = "0.6s";
                    balloon.style.webkitTransform = "translate3d(0," + -(w.scrollY + options.bottomOffset + balloon.offsetHeight) + "px,0)";
                    break;

                  case "bubble":
                    duration = "0.6s";
                    balloon.style.opacity = "0";
                    balloon.style.webkitTransform = "translate3d(0," + (startY + 50) + "px,0)";
                    break;

                  default:
                    duration = "1s";
                    balloon.style.opacity = "0";
                }
            } else {
                startY = w.innerHeight + w.scrollY;
                if (OSVersion < 5) {
                    startX = Math.round((w.innerWidth - balloon.offsetWidth) / 2) + w.scrollX;
                    balloon.style.left = startX + "px";
                    balloon.style.top = startY - balloon.offsetHeight - options.bottomOffset + "px";
                } else {
                    balloon.style.left = "50%";
                    balloon.style.marginLeft = -Math.round(balloon.offsetWidth / 2) - (w.orientation % 180 && OSVersion >= 6 && OSVersion < 7 ? 40 : 0) + "px";
                    balloon.style.bottom = options.bottomOffset + "px";
                }
                switch (options.animationIn) {
                  case "drop":
                    duration = "1s";
                    balloon.style.webkitTransform = "translate3d(0," + -(startY + options.bottomOffset) + "px,0)";
                    break;

                  case "bubble":
                    duration = "0.6s";
                    balloon.style.webkitTransform = "translate3d(0," + (balloon.offsetHeight + options.bottomOffset + 50) + "px,0)";
                    break;

                  default:
                    duration = "1s";
                    balloon.style.opacity = "0";
                }
            }
            balloon.offsetHeight;
            // repaint trick
            balloon.style.webkitTransitionDuration = duration;
            balloon.style.opacity = "1";
            balloon.style.webkitTransform = "translate3d(0,0,0)";
            balloon.addEventListener("webkitTransitionEnd", transitionEnd, false);
            closeTimeout = setTimeout(close, options.lifespan);
        }
        function manualShow(override) {
            if (!isIDevice || balloon) return;
            overrideChecks = override;
            loaded();
        }
        function close() {
            clearInterval(positionInterval);
            clearTimeout(closeTimeout);
            closeTimeout = null;
            // check if the popup is displayed and prevent errors
            if (!balloon) return;
            var posY = 0, posX = 0, opacity = "1", duration = "0";
            if (options.closeButton) balloon.removeEventListener("click", clicked, false);
            if (!isIPad && OSVersion >= 6) window.removeEventListener("orientationchange", orientationCheck, false);
            if (OSVersion < 5) {
                posY = isIPad ? w.scrollY - startY : w.scrollY + w.innerHeight - startY;
                posX = isIPad ? w.scrollX - startX : w.scrollX + Math.round((w.innerWidth - balloon.offsetWidth) / 2) - startX;
            }
            balloon.style.webkitTransitionProperty = "-webkit-transform,opacity";
            switch (options.animationOut) {
              case "drop":
                if (isIPad) {
                    duration = "0.4s";
                    opacity = "0";
                    posY += 50;
                } else {
                    duration = "0.6s";
                    posY += balloon.offsetHeight + options.bottomOffset + 50;
                }
                break;

              case "bubble":
                if (isIPad) {
                    duration = "0.8s";
                    posY -= balloon.offsetHeight + options.bottomOffset + 50;
                } else {
                    duration = "0.4s";
                    opacity = "0";
                    posY -= 50;
                }
                break;

              default:
                duration = "0.8s";
                opacity = "0";
            }
            balloon.addEventListener("webkitTransitionEnd", transitionEnd, false);
            balloon.style.opacity = opacity;
            balloon.style.webkitTransitionDuration = duration;
            balloon.style.webkitTransform = "translate3d(" + posX + "px," + posY + "px,0)";
        }
        function clicked() {
            w.sessionStorage.setItem("addToHomeSession", "1");
            isSessionActive = true;
            close();
        }
        function transitionEnd() {
            balloon.removeEventListener("webkitTransitionEnd", transitionEnd, false);
            balloon.style.webkitTransitionProperty = "-webkit-transform";
            balloon.style.webkitTransitionDuration = "0.2s";
            // We reached the end!
            if (!closeTimeout) {
                balloon.parentNode.removeChild(balloon);
                balloon = null;
                return;
            }
            // On iOS 4 we start checking the element position
            if (OSVersion < 5 && closeTimeout) positionInterval = setInterval(setPosition, options.iterations);
        }
        function setPosition() {
            var matrix = new WebKitCSSMatrix(w.getComputedStyle(balloon, null).webkitTransform), posY = isIPad ? w.scrollY - startY : w.scrollY + w.innerHeight - startY, posX = isIPad ? w.scrollX - startX : w.scrollX + Math.round((w.innerWidth - balloon.offsetWidth) / 2) - startX;
            // Screen didn't move
            if (posY == matrix.m42 && posX == matrix.m41) return;
            balloon.style.webkitTransform = "translate3d(" + posX + "px," + posY + "px,0)";
        }
        // Clear local and session storages (this is useful primarily in development)
        function reset() {
            w.localStorage.removeItem("addToHome");
            w.sessionStorage.removeItem("addToHomeSession");
        }
        function orientationCheck() {
            balloon.style.marginLeft = -Math.round(balloon.offsetWidth / 2) - (w.orientation % 180 && OSVersion >= 6 && OSVersion < 7 ? 40 : 0) + "px";
        }
        // Bootstrap!
        init();
        return {
            show: manualShow,
            close: close,
            reset: reset
        };
    }(window);
});
define("ui.alert", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
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
        this.$element.addClass("am-fade am-in").on("click", ".am-close", $.proxy(this.close, this));
    };
    Alert.DEFAULTS = {
        removeElement: true
    };
    Alert.prototype.close = function() {
        var $this = $(this), $target = $this.hasClass("am-alert") ? $this : $this.parent(".am-alert");
        $target.trigger("close:alert:amui");
        $target.removeClass("am-in");
        function processAlert() {
            $target.off().trigger("closed:alert:amui").remove();
        }
        UI.support.transition && $target.hasClass("am-fade") ? $target.one(UI.support.transition.end, processAlert).emulateTransitionEnd(200) : processAlert();
    };
    UI.alert = Alert;
    // Alert Plugin
    $.fn.alert = function(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("amui.alert"), options = typeof option == "object" && option;
            if (!data) {
                $this.data("amui.alert", data = new Alert(this, options));
            }
            if (typeof option == "string") {
                data[option].call($this);
            }
        });
    };
    // Init code
    $(document).on("click.alert.amui", "[data-am-alert]", function(e) {
        var $target = $(e.target);
        $(this).addClass("am-fade am-in");
        $target.is(".am-close") && $(this).alert("close");
    });
    module.exports = Alert;
});
define("ui.button", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/twbs/bootstrap/blob/master/js/button.js
     * @copyright Copyright 2013 Twitter, Inc.
     * @license Apache 2.0
     */
    var Button = function(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, Button.DEFAULTS, options);
        this.isLoading = false;
        this.hasSpinner = false;
    };
    Button.DEFAULTS = {
        loadingText: "loading...",
        loadingClass: "am-btn-loading",
        loadingWithSpinner: '<span class="am-icon-refresh am-icon-spin"></span> loading...'
    };
    Button.prototype.setState = function(state) {
        var d = "disabled", $el = this.$element, val = $el.is("input") ? "val" : "html", data = $el.data();
        state = state + "Text";
        if (data.resetText == null) {
            $el.data("resetText", $el[val]());
        }
        // add spinner for element with html()
        if (UI.support.animation && !this.hasSpinner && val === "html") {
            this.options.loadingText = this.options.loadingWithSpinner;
            this.hasSpinner = true;
        }
        $el[val](data[state] == null ? this.options[state] : data[state]);
        // push to event loop to allow forms to submit
        setTimeout($.proxy(function() {
            if (state == "loadingText") {
                this.isLoading = true;
                $el.addClass(d + " " + this.options.loadingClass).attr(d, d);
            } else if (this.isLoading) {
                this.isLoading = false;
                $el.removeClass(d + " " + this.options.loadingClass).removeAttr(d);
            }
        }, this), 0);
    };
    Button.prototype.toggle = function() {
        var changed = true, $parent = this.$element.parent(".am-btn-group");
        if ($parent.length) {
            var $input = this.$element.find("input");
            if ($input.prop("type") == "radio") {
                if ($input.prop("checked") && this.$element.hasClass("am-active")) {
                    changed = false;
                } else {
                    $parent.find(".am-active").removeClass("am-active");
                }
            }
            if (changed) {
                $input.prop("checked", !this.$element.hasClass("am-active")).trigger("change");
            }
        }
        if (changed) {
            this.$element.toggleClass("am-active");
        }
    };
    // Button plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data("amui.button");
            var options = typeof option == "object" && option;
            if (!data) {
                $this.data("amui.button", data = new Button(this, options));
            }
            if (option == "toggle") {
                data.toggle();
            } else if (option) {
                data.setState(option);
            }
        });
    }
    $.fn.button = Plugin;
    // Init code
    $(document).on("click.button.amui", "[data-am-button]", function(e) {
        var $btn = $(e.target);
        if (!$btn.hasClass("am-btn")) {
            $btn = $btn.closest(".am-btn");
        }
        Plugin.call($btn, "toggle");
        e.preventDefault();
    });
    module.exports = Button;
});
define("ui.collapse", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/twbs/bootstrap/blob/master/js/collapse.js
     * @copyright Copyright 2013 Twitter, Inc.
     * @license Apache 2.0
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
        if (this.transitioning || this.$element.hasClass("am-in")) return;
        var startEvent = $.Event("open:collapse:amui");
        this.$element.trigger(startEvent);
        if (startEvent.isDefaultPrevented()) return;
        var actives = this.$parent && this.$parent.find("> .am-panel > .am-in");
        if (actives && actives.length) {
            var hasData = actives.data("amui.collapse");
            if (hasData && hasData.transitioning) return;
            Plugin.call(actives, "close");
            hasData || actives.data("amui.collapse", null);
        }
        this.$element.removeClass("am-collapse").addClass("am-collapsing").height(0);
        this.transitioning = 1;
        var complete = function() {
            this.$element.removeClass("am-collapsing").addClass("am-collapse am-in").height("");
            this.transitioning = 0;
            this.$element.trigger("opened:collapse:amui");
        };
        if (!UI.support.transition) {
            return complete.call(this);
        }
        this.$element.one(UI.support.transition.end, $.proxy(complete, this)).emulateTransitionEnd(350).height(this.$element[0].scrollHeight);
    };
    Collapse.prototype.close = function() {
        if (this.transitioning || !this.$element.hasClass("am-in")) return;
        var startEvent = $.Event("close:collapse:amui");
        this.$element.trigger(startEvent);
        if (startEvent.isDefaultPrevented()) return;
        this.$element.height(this.$element.height());
        this.$element[0].offsetHeight;
        this.$element.addClass("am-collapsing").removeClass("am-collapse").removeClass("am-in");
        this.transitioning = 1;
        var complete = function() {
            this.transitioning = 0;
            this.$element.trigger("closed:collapse:amui").removeClass("am-collapsing").addClass("am-collapse");
        };
        if (!UI.support.transition) {
            return complete.call(this);
        }
        this.$element.height(0).one(UI.support.transition.end, $.proxy(complete, this)).emulateTransitionEnd(350);
    };
    Collapse.prototype.toggle = function() {
        this[this.$element.hasClass("am-in") ? "close" : "open"]();
    };
    UI.collapse = Collapse;
    // Collapse Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("amui.collapse"), options = $.extend({}, Collapse.DEFAULTS, UI.utils.options($this.attr("data-am-collapse")), typeof option == "object" && option);
            if (!data && options.toggle && option == "open") {
                option = !option;
            }
            if (!data) {
                $this.data("amui.collapse", data = new Collapse(this, options));
            }
            if (typeof option == "string") {
                data[option]();
            }
        });
    }
    $.fn.collapse = Plugin;
    // Init code
    $(document).on("click.collapse.amui", "[data-am-collapse]", function(e) {
        var href, $this = $(this), options = UI.utils.options($this.attr("data-am-collapse")), target = options.target || e.preventDefault() || (href = $this.attr("href")) && href.replace(/.*(?=#[^\s]+$)/, "");
        var $target = $(target);
        var data = $target.data("amui.collapse");
        var option = data ? "toggle" : options;
        var parent = options.parent;
        var $parent = parent && $(parent);
        if (!data || !data.transitioning) {
            if ($parent) {
                //'[data-am-collapse*="{parent: \'' + parent + '"]
                $parent.find("[data-am-collapse]").not($this).addClass("am-collapsed");
            }
            $this[$target.hasClass("am-in") ? "addClass" : "removeClass"]("am-collapsed");
        }
        Plugin.call($target, option);
    });
    module.exports = Collapse;
});
define("ui.dimmer", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    var $dimmer = $('<div class="am-dimmer" data-am-dimmer></div>'), $doc = $(document), $html = $("html");
    var Dimmer = function() {
        this.hasDimmer = $("[data-am-dimmer]").length ? true : false;
        this.$element = $dimmer;
        $(document).on("ready", $.proxy(this.init, this));
    };
    Dimmer.prototype.init = function() {
        if (!this.hasDimmer) {
            $dimmer.appendTo($("body"));
            this.events();
            this.hasDimmer = true;
        }
        $doc.trigger("init:dimmer:amui");
        return this;
    };
    Dimmer.prototype.open = function(relatedElement) {
        $html.addClass("am-dimmer-active");
        $dimmer.addClass("am-active");
        $(relatedElement).length && $(relatedElement).show();
        $doc.trigger("open:dimmer:amui");
        return this;
    };
    Dimmer.prototype.close = function(relatedElement) {
        $dimmer.removeClass("am-active");
        $html.removeClass("am-dimmer-active");
        $(relatedElement).length && $(relatedElement).hide();
        $doc.trigger("close:dimmer:amui");
        return this;
    };
    Dimmer.prototype.events = function() {
        var that = this;
        $dimmer.on("click.dimmer.amui", function() {});
    };
    var dimmer = new Dimmer();
    UI.dimmer = dimmer;
    module.exports = dimmer;
});
define("ui.dropdown", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/Minwe/bootstrap/blob/master/js/dropdown.js
     * @copyright Copyright 2013 Twitter, Inc.
     * @license Apache 2.0
     */
    var toggle = "[data-am-dropdown] > .am-dropdown-toggle";
    var Dropdown = function(element, options) {
        $(element).on("click.dropdown.amui", this.toggle);
    };
    Dropdown.prototype.toggle = function(e) {
        var $this = $(this);
        if ($this.is(".am-disabled, :disabled")) {
            return;
        }
        var $parent = $this.parent(), isActive = $parent.hasClass("am-active");
        clearDropdowns();
        if (!isActive) {
            var relatedTarget = {
                relatedTarget: this
            };
            $parent.trigger(e = $.Event("open:dropdown:amui", relatedTarget));
            if (e.isDefaultPrevented()) {
                return;
            }
            $this.trigger("focus");
            $parent.toggleClass("am-active").trigger(e = $.Event("opened:dropdown:amui", relatedTarget));
        } else {
            $this.blur();
        }
        return false;
    };
    Dropdown.prototype.keydown = function(e) {
        if (!/(38|40|27)/.test(e.keyCode)) return;
        var $this = $(this);
        e.preventDefault();
        e.stopPropagation();
        if ($this.is(".am-disabled, :disabled")) {
            return;
        }
        var $parent = $this.parent(), isActive = $parent.hasClass("am-active");
        if (!isActive || isActive && e.keyCode == 27) {
            if (e.which == 27) {
                $parent.find(toggle).trigger("focus");
            }
            return $this.trigger("click");
        }
    };
    function clearDropdowns(e) {
        if (e && e.which === 3) return;
        $(toggle).each(function() {
            var $parent = $(this).parent(), relatedTarget = {
                relatedTarget: this
            };
            if (!$parent.hasClass("am-active")) {
                return;
            }
            $parent.trigger(e = $.Event("close:dropdown:amui", relatedTarget));
            if (e.isDefaultPrevented()) return;
            $parent.removeClass("am-active").trigger(e = $.Event("closed:dropdown:amui", relatedTarget));
        });
    }
    UI.dropdown = Dropdown;
    // Dropdown Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data("amui.dropdown");
            if (!data) {
                $this.data("amui.dropdown", data = new Dropdown(this));
            }
            if (typeof option == "string") {
                data[option].call($this);
            }
        });
    }
    $.fn.dropdown = Plugin;
    // Init code
    $(document).on("click.dropdown.amui", ".am-dropdown form", function(e) {
        e.stopPropagation();
    }).on("click.dropdown.amui", toggle, Dropdown.prototype.toggle).on("keydown.dropdown.amui", toggle, Dropdown.prototype.keydown);
});
define("ui.modal", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector", "ui.dimmer" ], function(require, exports, module) {
    require("core");
    var dimmer = require("ui.dimmer");
    var $ = window.Zepto;
    var UI = $.AMUI;
    var $win = $(window), $doc = $(document), $body = $("body"), supportTransition = UI.support.transition;
    /**
     * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
     * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
     */
    var Modal = function(element, options) {
        this.options = $.extend({}, Modal.DEFAULTS, options || {});
        this.$element = $(element);
        this.isPopup = this.$element.hasClass("am-popup");
        this.active = this.transitioning = null;
        this.events();
    };
    // Set template tabindex to tigger onfocus on div
    Modal.DEFAULTS = {
        selector: {
            modal: ".am-modal",
            active: ".am-modal-active"
        },
        cancelable: true,
        onConfirm: function() {},
        onCancel: function() {}
    };
    Modal.prototype.toggle = function(relatedElement) {
        return this.active ? this.close() : this.open(relatedElement);
    };
    Modal.prototype.open = function(relatedElement) {
        var $element = this.$element, isPopup = this.isPopup;
        if (this.transitioning || this.active) return;
        if (!this.$element.length) return;
        isPopup && this.$element.show();
        this.active = true;
        $element.trigger($.Event("open:modal:amui", {
            relatedElement: relatedElement
        }));
        // trigger reflow;
        var clientLeft = $element[0].clientLeft;
        dimmer.open();
        $element.show();
        !isPopup && $element.redraw().css({
            marginTop: -$element.height() / 2 + "px"
        });
        $element.removeClass("am-modal-out").addClass("am-modal-active");
        this.transitioning = 1;
        var complete = function() {
            $element.trigger($.Event("opened:modal:amui", {
                relatedElement: relatedElement
            }));
            this.transitioning = 0;
        };
        if (!supportTransition) return complete.call(this);
        $element.one(supportTransition.end, $.proxy(complete, this));
    };
    Modal.prototype.close = function(relatedElement) {
        if (this.transitioning || !this.active) return;
        var $element = this.$element, isPopup = this.isPopup, that = this;
        this.$element.trigger($.Event("close:modal:amui", {
            relatedElement: relatedElement
        }));
        this.transitioning = 1;
        var complete = function() {
            $element.trigger("closed.amui.modal");
            isPopup && $element.removeClass("am-modal-out");
            $element.hide();
            this.transitioning = 0;
        };
        $element.removeClass("am-modal-active").addClass("am-modal-out");
        if (!supportTransition) return complete.call(this);
        $element.one(supportTransition.end, $.proxy(complete, this));
        // hide dimmer when all modal is closed
        if (!$body.find(Modal.DEFAULTS.selector.active).length) {
            dimmer.close();
        }
        this.active = false;
    };
    Modal.prototype.events = function() {
        var that = this, $element = this.$element, $ipt = $element.find(".am-modal-prompt-input");
        if (this.options.cancelable) {
            $element.on("keyup.modal.amui", $.proxy(function(e) {
                if (this.active && e.which === 27) {
                    this.options.onCancel();
                    this.close();
                }
            }, that));
            dimmer.$element.on("click", function(e) {
                that.close();
            });
        }
        // Close button
        $element.find("[data-am-modal-close]").on("click.modal.amui", function(e) {
            e.preventDefault();
            that.close();
        });
        $element.find(".am-modal-btn").on("click.modal.amui", function(e) {
            that.close();
        });
        $element.find("[data-am-modal-confirm]").on("click.modal.amui", function() {
            that.options.onConfirm($ipt.val());
        });
        $element.find("[data-am-modal-cancel]").on("click.modal.amui", function() {
            that.options.onCancel($ipt.val());
        });
    };
    function Plugin(option, relatedElement) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.modal"), options = $.extend({}, Modal.DEFAULTS, typeof option == "object" && option);
            if (!data) {
                $this.data("am.modal", data = new Modal(this, options));
            }
            if (typeof option == "string") {
                data[option](relatedElement);
            } else {
                data.toggle(option && option.relatedElement || undefined);
            }
        });
    }
    $.fn.modal = Plugin;
    $doc.on("click", "[data-am-modal]", function() {
        var $this = $(this), options = UI.utils.parseOptions($this.attr("data-am-modal")), $target = $(options.target || this.href && this.href.replace(/.*(?=#[^\s]+$)/, "")), option = $target.data("am.modal") ? "toggle" : options;
        Plugin.call($target, option, this);
    });
    module.exports = Modal;
});
define("ui.offcanvas", [ "zepto.outerdemension", "zepto.extend.data", "core", "zepto.extend.fx", "zepto.extend.selector" ], function(require, exports, module) {
    require("zepto.outerdemension");
    require("zepto.extend.data");
    require("core");
    var $ = window.Zepto, UI = $.AMUI, $win = $(window), $doc = $(document), scrollPos;
    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/offcanvas.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */
    var OffCanvas = function(element, options) {
        this.$element = $(element);
        this.options = options;
        this.events();
    };
    OffCanvas.DEFAULTS = {
        effect: "overlay"
    };
    OffCanvas.prototype.open = function(relatedElement) {
        var _self = this, $element = this.$element, openEvent = $.Event("open:offcanvas:amui");
        if (!$element.length || $element.hasClass("am-active")) return;
        var effect = this.options.effect, $html = $("html"), $bar = $element.find(".am-offcanvas-bar").first(), dir = $bar.hasClass("am-offcanvas-bar-flip") ? -1 : 1;
        $bar.addClass("am-offcanvas-bar-" + effect);
        scrollPos = {
            x: window.scrollX,
            y: window.scrollY
        };
        $element.addClass("am-active");
        $html.css({
            width: "100%",
            height: $win.height()
        }).addClass("am-offcanvas-page");
        if (!(effect === "overlay")) {
            $html.css({
                "margin-left": $bar.outerWidth() * dir
            }).width();
        }
        $html.css("margin-top", scrollPos.y * -1);
        UI.utils.debounce(function() {
            $bar.addClass("am-offcanvas-bar-active").width();
        }, 0)();
        $doc.trigger(openEvent);
        $element.off(".offcanvas.amui").on("click.offcanvas.amui swipeRight.offcanvas.amui swipeLeft.offcanvas.amui", function(e) {
            var $target = $(e.target);
            if (!e.type.match(/swipe/)) {
                if ($target.hasClass("am-offcanvas-bar")) return;
                if ($target.parents(".am-offcanvas-bar").first().length) return;
            }
            // https://developer.mozilla.org/zh-CN/docs/DOM/event.stopImmediatePropagation
            e.stopImmediatePropagation();
            _self.close();
        });
        $doc.on("keydown.offcanvas.amui", function(e) {
            if (e.keyCode === 27) {
                // ESC
                _self.close();
            }
        });
    };
    OffCanvas.prototype.close = function(relatedElement) {
        var $html = $("html"), $element = this.$element, $bar = $element.find(".am-offcanvas-bar").first();
        if (!$element.length || !$element.hasClass("am-active")) return;
        $element.trigger("close:offcanvas:amui");
        if (UI.support.transition) {
            $html.one(UI.support.transition.end, function() {
                $html.removeClass("am-offcanvas-page").css({
                    width: "",
                    height: "",
                    "margin-top": ""
                });
                $element.removeClass("am-active");
                window.scrollTo(scrollPos.x, scrollPos.y);
            }).css("margin-left", "");
            UI.utils.debounce(function() {
                $bar.removeClass("am-offcanvas-bar-active");
            }, 0)();
        } else {
            $html.removeClass("am-offcanvas-page").attr("style", "");
            $element.removeClass("am-active");
            $bar.removeClass("am-offcanvas-bar-active");
            window.scrollTo(scrollPos.x, scrollPos.y);
        }
        $element.off(".offcanvas.amui");
    };
    OffCanvas.prototype.events = function() {
        $doc.on("click.offcanvas.amui", '[data-am-dismiss="offcanvas"]', $.proxy(function(e) {
            e.preventDefault();
            this.close();
        }, this));
        return this;
    };
    UI.offcanvas = OffCanvas;
    function Plugin(option, relatedElement) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.offcanvas"), options = $.extend({}, OffCanvas.DEFAULTS, typeof option == "object" && option);
            if (!data) {
                $this.data("am.offcanvas", data = new OffCanvas(this, options));
                data.open(relatedElement);
            }
            if (typeof option == "string") {
                data[option] && data[option](relatedElement);
            }
        });
    }
    $.fn.offCanvas = Plugin;
    // Init code
    $doc.on("click.offcanvas.amui", "[data-am-offcanvas]", function(e) {
        e.preventDefault();
        var $this = $(this), options = UI.utils.parseOptions($this.attr("data-am-offcanvas")), $target = $(options.target || this.href && this.href.replace(/.*(?=#[^\s]+$)/, ""));
        option = $target.data("am.offcanvas") ? "open" : options;
        Plugin.call($target, option, this);
    });
    module.exports = OffCanvas;
});
define("ui.popover", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto, UI = $.AMUI, $w = $(window), $doc = $(document);
    /**
     * @reference https://github.com/nolimits4web/Framework7/blob/master/src/js/modals.js
     * @license https://github.com/nolimits4web/Framework7/blob/master/LICENSE
     */
    var Popover = function(element, options) {
        this.options = $.extend({}, Popover.DEFAULTS, options || {});
        this.$element = $(element);
        this.active = null;
        this.$popover = this.options.target && $(this.options.target) || null;
        this.init();
        this.events();
    };
    Popover.DEFAULTS = {
        trigger: "click",
        content: "",
        open: false,
        target: undefined,
        tpl: '<div class="am-popover"><div class="am-popover-inner"></div><div class="am-popover-caret"></div></div>'
    };
    Popover.prototype.init = function() {
        var $element = this.$element, $popover;
        if (!this.options.target) {
            this.$popover = this.getPopover();
            this.setContent();
        }
        $popover = this.$popover;
        $popover.appendTo($("body"));
        function sizePopover() {
            var popSize = $popover.getSize(), popWidth = $popover.width() || popSize.width, popHeight = $popover.height() || popSize.height, $popCaret = $popover.find(".am-popover-caret"), popCaretSize = $popCaret.width() / 2 || 10, popTotalHeight = popHeight + popCaretSize;
            var triggerWidth = $element.outerWidth(), triggerHeight = $element.outerHeight(), triggerOffset = $element.offset(), triggerRect = $element[0].getBoundingClientRect();
            var winHeight = $w.height(), winWidth = $w.width();
            var popTop = 0, popLeft = 0, diff = 0, spacing = 3, popPosition = "top";
            $popover.css({
                left: "",
                top: ""
            }).removeClass("am-popover-left am-popover-right am-popover-top am-popover-bottom");
            $popCaret.css({
                left: "",
                top: ""
            });
            if (popTotalHeight - spacing < triggerRect.top + spacing) {
                // on Top
                popTop = triggerOffset.top - popTotalHeight - spacing;
            } else if (popTotalHeight < winHeight - triggerRect.top - triggerRect.height) {
                // On bottom
                popPosition = "bottom";
                popTop = triggerOffset.top + triggerHeight + popCaretSize + spacing;
            } else {
                // On middle
                popPosition = "middle";
                popTop = triggerHeight / 2 + triggerOffset.top - popHeight / 2;
            }
            // Horizontal Position
            if (popPosition === "top" || popPosition === "bottom") {
                popLeft = triggerWidth / 2 + triggerOffset.left - popWidth / 2;
                diff = popLeft;
                if (popLeft < 5) popLeft = 5;
                if (popLeft + popWidth > winWidth) {
                    popLeft = winWidth - popWidth - 20;
                }
                if (popPosition === "top") $popover.addClass("am-popover-bottom");
                if (popPosition === "bottom") $popover.addClass("am-popover-top");
                diff = diff - popLeft;
                $popCaret.css({
                    left: popWidth / 2 - popCaretSize + diff + "px"
                });
            } else if (popPosition === "middle") {
                popLeft = triggerOffset.left - popWidth - popCaretSize;
                $popover.addClass("am-popover-left");
                if (popLeft < 5) {
                    popLeft = triggerOffset.left + triggerWidth + popCaretSize;
                    $popover.removeClass("am-popover-left").addClass("am-popover-right");
                }
                if (popLeft + popWidth > winWidth) {
                    popLeft = winWidth - popWidth - 5;
                    $popover.removeClass("am-popover-left").addClass("am-popover-right");
                }
                $popCaret.css({
                    top: popHeight / 2 - popCaretSize / 2 + "px"
                });
            }
            // Apply position style
            $popover.css({
                top: popTop + "px",
                left: popLeft + "px"
            });
        }
        sizePopover();
        $(window).on("resize", UI.utils.debounce(sizePopover, 50));
        $element.on("open:popover:amui", function() {
            $(window).on("resize", UI.utils.debounce(sizePopover, 50));
        });
        $element.on("close:popover:amui", function() {
            $(window).off("resize", sizePopover);
        });
        this.options.open && this.open();
    };
    Popover.prototype.toggle = function() {
        return this[this.active ? "close" : "open"]();
    };
    Popover.prototype.open = function() {
        var $popover = this.$popover;
        this.$element.trigger("open:popover:amui");
        $popover.show().addClass("am-active");
        this.active = true;
    };
    Popover.prototype.close = function() {
        var $popover = this.$popover;
        this.$element.trigger("close:popover:amui");
        $popover.removeClass("am-active").trigger("closed:popover:amui").hide();
        this.active = false;
    };
    Popover.prototype.getUID = function() {
        var ns = "am-popover-";
        do {
            ns += parseInt(Math.random() * 1e6);
        } while (document.getElementById(ns));
        return ns;
    };
    Popover.prototype.getPopover = function() {
        var uid = this.getUID();
        return $(this.options.tpl, {
            id: uid
        });
    };
    Popover.prototype.setContent = function() {
        this.$popover && this.$popover.find(".am-popover-inner").empty().html(this.options.content);
    };
    Popover.prototype.events = function() {
        var trigger = this.options.trigger, eventNS = "popover.amui";
        if (trigger === "click") {
            this.$element.on("click." + eventNS, $.proxy(this.toggle, this));
        } else if (trigger === "hover") {
            this.$element.on("mouseenter." + eventNS, $.proxy(this.open, this));
            this.$element.on("mouseleave." + eventNS, $.proxy(this.close, this));
        }
    };
    UI.popover = Popover;
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.popover"), options = $.extend({}, UI.utils.parseOptions($this.attr("data-am-popover")), typeof option == "object" && option);
            if (!data) {
                $this.data("am.popover", data = new Popover(this, options));
            }
            if (typeof option == "string") {
                data[option]();
            }
        });
    }
    $.fn.popover = Plugin;
    // Init code
    $(function() {
        $("[data-am-popover]").popover();
    });
    module.exports = Popover;
});
define("ui.progress", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    var Progress = function() {
        /**
         * NProgress (c) 2013, Rico Sta. Cruz
         * @via http://ricostacruz.com/nprogress
         */
        var NProgress = {}, $html = $("html");
        NProgress.version = "0.1.6";
        var Settings = NProgress.settings = {
            minimum: .08,
            easing: "ease",
            positionUsing: "",
            speed: 200,
            trickle: true,
            trickleRate: .02,
            trickleSpeed: 800,
            showSpinner: true,
            parent: "body",
            barSelector: '[role="nprogress-bar"]',
            spinnerSelector: '[role="nprogress-spinner"]',
            template: '<div class="nprogress-bar" role="nprogress-bar">' + '<div class="nprogress-peg"></div></div>' + '<div class="nprogress-spinner" role="nprogress-spinner">' + '<div class="nprogress-spinner-icon"></div></div>'
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
            NProgress.status = n === 1 ? null : n;
            var progress = NProgress.render(!started), bar = progress.querySelector(Settings.barSelector), speed = Settings.speed, ease = Settings.easing;
            progress.offsetWidth;
            /* Repaint */
            queue(function(next) {
                // Set positionUsing if it hasn't already been set
                if (Settings.positionUsing === "") Settings.positionUsing = NProgress.getPositioningCSS();
                // Add transition
                css(bar, barPositionCSS(n, speed, ease));
                if (n === 1) {
                    // Fade out
                    css(progress, {
                        transition: "none",
                        opacity: 1
                    });
                    progress.offsetWidth;
                    /* Repaint */
                    setTimeout(function() {
                        css(progress, {
                            transition: "all " + speed + "ms linear",
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
            return typeof NProgress.status === "number";
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
            return NProgress.inc(.3 + .5 * Math.random()).set(1);
        };
        /**
         * Increments by a random amount.
         */
        NProgress.inc = function(amount) {
            var n = NProgress.status;
            if (!n) {
                return NProgress.start();
            } else {
                if (typeof amount !== "number") {
                    amount = (1 - n) * clamp(Math.random() * n, .1, .95);
                }
                n = clamp(n + amount, 0, .994);
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
            if (NProgress.isRendered()) return document.getElementById("nprogress");
            $html.addClass("nprogress-busy");
            var progress = document.createElement("div");
            progress.id = "nprogress";
            progress.innerHTML = Settings.template;
            var bar = progress.querySelector(Settings.barSelector), perc = fromStart ? "-100" : toBarPerc(NProgress.status || 0), parent = document.querySelector(Settings.parent), spinner;
            css(bar, {
                transition: "all 0 linear",
                transform: "translate3d(" + perc + "%,0,0)"
            });
            if (!Settings.showSpinner) {
                spinner = progress.querySelector(Settings.spinnerSelector);
                spinner && $(spinner).remove();
            }
            if (parent != document.body) {
                $(parent).addClass("nprogress-custom-parent");
            }
            parent.appendChild(progress);
            return progress;
        };
        /**
         * Removes the element. Opposite of render().
         */
        NProgress.remove = function() {
            $html.removeClass("nprogress-busy");
            $(Settings.parent).removeClass("nprogress-custom-parent");
            var progress = document.getElementById("nprogress");
            progress && $(progress).remove();
        };
        /**
         * Checks if the progress bar is rendered.
         */
        NProgress.isRendered = function() {
            return !!document.getElementById("nprogress");
        };
        /**
         * Determine which positioning CSS rule to use.
         */
        NProgress.getPositioningCSS = function() {
            // Sniff on document.body.style
            var bodyStyle = document.body.style;
            // Sniff prefixes
            var vendorPrefix = "WebkitTransform" in bodyStyle ? "Webkit" : "MozTransform" in bodyStyle ? "Moz" : "msTransform" in bodyStyle ? "ms" : "OTransform" in bodyStyle ? "O" : "";
            if (vendorPrefix + "Perspective" in bodyStyle) {
                // Modern browsers with 3D support, e.g. Webkit, IE10
                return "translate3d";
            } else if (vendorPrefix + "Transform" in bodyStyle) {
                // Browsers without 3D support, e.g. IE9
                return "translate";
            } else {
                // Browsers without translate() support, e.g. IE7-8
                return "margin";
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
            if (Settings.positionUsing === "translate3d") {
                barCSS = {
                    transform: "translate3d(" + toBarPerc(n) + "%,0,0)"
                };
            } else if (Settings.positionUsing === "translate") {
                barCSS = {
                    transform: "translate(" + toBarPerc(n) + "%,0)"
                };
            } else {
                barCSS = {
                    "margin-left": toBarPerc(n) + "%"
                };
            }
            barCSS.transition = "all " + speed + "ms " + ease;
            return barCSS;
        }
        /**
         * (Internal) Queues a function to be executed.
         */
        var queue = function() {
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
        }();
        /**
         * (Internal) Applies css properties to an element, similar to the jQuery
         * css method.
         *
         * While this helper does assist with vendor prefixed property names, it
         * does not perform any manipulation of values prior to setting styles.
         */
        var css = function() {
            var cssPrefixes = [ "Webkit", "O", "Moz", "ms" ], cssProps = {};
            function camelCase(string) {
                return string.replace(/^-ms-/, "ms-").replace(/-([\da-z])/gi, function(match, letter) {
                    return letter.toUpperCase();
                });
            }
            function getVendorProp(name) {
                var style = document.body.style;
                if (name in style) return name;
                var i = cssPrefixes.length, capName = name.charAt(0).toUpperCase() + name.slice(1), vendorName;
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
                var args = arguments, prop, value;
                if (args.length == 2) {
                    for (prop in properties) {
                        value = properties[prop];
                        if (value !== undefined && properties.hasOwnProperty(prop)) applyCss(element, prop, value);
                    }
                } else {
                    applyCss(element, args[1], args[2]);
                }
            };
        }();
        return NProgress;
    }();
    UI.progress = Progress;
    module.exports = Progress;
});
define("ui.scrollspy", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/scrollspy.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */
    var ScrollSpy = function(element, options) {
        if (!UI.support.animation) return;
        this.options = $.extend({}, ScrollSpy.DEFAULTS, options);
        this.$element = $(element);
        var checkViewRAF = function() {
            UI.utils.rAF.call(window, $.proxy(this.checkView, this));
        }.bind(this);
        this.$window = $(window).on("scroll.scrollspy.amui", checkViewRAF).on("resize.scrollspy.amui orientationchange.scrollspy.amui", UI.utils.debounce(checkViewRAF, 50));
        this.timer = this.inViewState = this.initInView = null;
        checkViewRAF();
    };
    ScrollSpy.DEFAULTS = {
        animation: "fade",
        inViewCls: "am-scrollspy-inview",
        initCls: "am-scrollspy-init",
        repeat: true,
        delay: 0,
        topOffset: 0,
        leftOffset: 0
    };
    ScrollSpy.prototype.checkView = function() {
        var $element = this.$element, options = this.options, inView = UI.utils.isInView($element, options), animation = options.animation ? " am-animation-" + options.animation : "";
        if (inView && !this.inViewState) {
            if (this.timer) clearTimeout(this.timer);
            if (!this.initInView) {
                $element.addClass(options.initCls);
                this.offset = $element.offset();
                this.initInView = true;
                $element.trigger("init:scrollspy:amui");
            }
            this.timer = setTimeout(function() {
                if (inView) {
                    $element.addClass(options.inViewCls + animation).width();
                }
            }, options.delay);
            this.inViewState = true;
            $element.trigger("inview:scrollspy:amui");
        }
        if (!inView && this.inViewState && options.repeat) {
            $element.removeClass(options.inViewCls + animation);
            this.inViewState = false;
            $element.trigger("outview:scrollspy:amui");
        }
    };
    ScrollSpy.prototype.check = function() {
        UI.utils.rAF.call(window, $.proxy(this.checkView, this));
    };
    UI.scrollspy = ScrollSpy;
    // Sticky Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.scrollspy"), options = typeof option == "object" && option;
            if (!data) $this.data("am.scrollspy", data = new ScrollSpy(this, options));
            if (typeof option == "string") data[option]();
        });
    }
    $.fn.scrollspy = Plugin;
    // Init code
    $(function() {
        $("[data-am-scrollspy]").each(function() {
            var $this = $(this), options = UI.utils.options($this.attr("data-am-scrollspy"));
            Plugin.call($this, options);
        });
    });
    module.exports = ScrollSpy;
});
define("ui.scrollspynav", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/uikit/uikit/
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */
    // ScrollSpyNav Class
    var ScrollSpyNav = function(element, options) {
        this.options = $.extend({}, ScrollSpyNav.DEFAULTS, options);
        this.$element = $(element);
        this.anchors = [];
        this.$links = this.$element.find('a[href^="#"]').each(function(i, link) {
            this.anchors.push($(link).attr("href"));
        }.bind(this));
        this.$targets = $(this.anchors.join(", "));
        var processRAF = function() {
            UI.utils.rAF.call(window, $.proxy(this.process, this));
        }.bind(this);
        this.$window = $(window).on("scroll.scrollspynav.amui", processRAF).on("resize.scrollspynav.amui orientationchange.scrollspynav.amui", UI.utils.debounce(processRAF, 50));
        processRAF();
        this.scrollProcess();
    };
    ScrollSpyNav.DEFAULTS = {
        cls: "am-active",
        topOffset: 0,
        leftOffset: 0,
        closest: false,
        smooth: true
    };
    ScrollSpyNav.prototype.process = function() {
        var scrollTop = this.$window.scrollTop(), options = this.options, inViews = [], $links = this.$links;
        var $targets = this.$targets;
        $targets.each(function(i, target) {
            if (UI.utils.isInView(target, options)) {
                inViews.push(target);
            }
        });
        //console.log(inViews.length);
        if (inViews.length) {
            var $target;
            $.each(inViews, function(i, item) {
                if ($(item).offset().top >= scrollTop) {
                    // console.log($(item));
                    $target = $(item);
                    return false;
                }
            });
            if (!$target) return;
            if (options.closest) {
                $links.closest(options.closest).removeClass(options.cls);
                $links.filter('a[href="#' + $target.attr("id") + '"]').closest(options.closest).addClass(options.cls);
            } else {
                $links.removeClass(options.cls).filter('a[href="#' + $target.attr("id") + '"]').addClass(options.cls);
            }
        }
    };
    ScrollSpyNav.prototype.scrollProcess = function() {
        var $links = this.$links;
        // smoothScroll
        if (this.options.smooth) {
            require.async([ "ui.smooth-scroll" ], function() {
                $links.on("click", function(e) {
                    e.preventDefault();
                    var $this = $(this), target = $this.attr("href"), position = $this.data("am.smoothScroll");
                    !position && $this.data("am.smoothScroll", position = $(target).offset().top);
                    $(window).smoothScroll(position);
                });
            });
        }
    };
    UI.scrollspynav = ScrollSpyNav;
    // ScrollSpyNav Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.scrollspynav"), options = typeof option == "object" && option;
            if (!data) $this.data("am.scrollspynav", data = new ScrollSpyNav(this, options));
            if (typeof option == "string") data[option]();
        });
    }
    $.fn.scrollspynav = Plugin;
    // Init code
    $(function() {
        $("[data-am-scrollspy-nav]").each(function() {
            var $this = $(this), options = UI.utils.options($this.attr("data-am-scrollspy-nav"));
            Plugin.call($this, options);
        });
    });
    module.exports = ScrollSpyNav;
});
define("ui.smooth-scroll", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    var UI = require("core"), rAF = UI.utils.rAF;
    var $ = window.Zepto;
    /**
     * Smooth Scroll
     * @param position
     * @via http://mir.aculo.us/2014/01/19/scrolling-dom-elements-to-the-top-a-zepto-plugin/
     */
    // Usage: $(element).smoothScroll([position])
    // only allow one scroll to top operation to be in progress at a time,
    // which is probably what you want
    var smoothScrollInProgress = false;
    $.fn.smoothScroll = function(position) {
        var $this = this, targetY = position || 0, initialY = $this.scrollTop(), lastY = initialY, delta = targetY - initialY, // duration in ms, make it a bit shorter for short distances
        // this is not scientific and you might want to adjust this for
        // your preferences
        speed = Math.min(750, Math.min(1500, Math.abs(initialY - targetY))), // temp variables (t will be a position between 0 and 1, y is the calculated scrollTop)
        start, t, y, cancelScroll = function() {
            abort();
        };
        // abort if already in progress or nothing to scroll
        if (smoothScrollInProgress) return;
        if (delta == 0) return;
        // quint ease-in-out smoothing, from
        // https://github.com/madrobby/scripty2/blob/master/src/effects/transitions/penner.js#L127-L136
        function smooth(pos) {
            if ((pos /= .5) < 1) return .5 * Math.pow(pos, 5);
            return .5 * (Math.pow(pos - 2, 5) + 2);
        }
        function abort() {
            $this.off("touchstart", cancelScroll);
            smoothScrollInProgress = false;
        }
        // when there's a touch detected while scrolling is in progress, abort
        // the scrolling (emulates native scrolling behavior)
        $this.on("touchstart", cancelScroll);
        smoothScrollInProgress = true;
        // start rendering away! note the function given to frame
        // is named "render" so we can reference it again further down
        rAF(function render(now) {
            if (!smoothScrollInProgress) return;
            if (!start) start = now;
            // calculate t, position of animation in [0..1]
            t = Math.min(1, Math.max((now - start) / speed, 0));
            // calculate the new scrollTop position (don't forget to smooth)
            y = Math.round(initialY + delta * smooth(t));
            // bracket scrollTop so we're never over-scrolling
            if (delta > 0 && y > targetY) y = targetY;
            if (delta < 0 && y < targetY) y = targetY;
            // only actually set scrollTop if there was a change fromt he last frame
            if (lastY != y) $this.scrollTop(y);
            lastY = y;
            // if we're not done yet, queue up an other frame to render,
            // or clean up
            if (y !== targetY) {
                rAF(render);
            } else {
                abort();
            }
        });
    };
    // Init code
    $(document).on("click.smoothScroll.amui", "[data-am-smooth-scroll]", function(e) {
        e.preventDefault();
        var $this = $(this), ssTo = Number($this.attr("data-am-smooth-scroll"));
        $(window).smoothScroll(isNaN(ssTo) ? 0 : ssTo);
    });
});
define("ui.sticky", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/uikit/uikit/blob/master/src/js/addons/sticky.js
     * @license https://github.com/uikit/uikit/blob/master/LICENSE.md
     */
    // Sticky Class
    var Sticky = function(element, options) {
        this.options = $.extend({}, Sticky.DEFAULTS, options);
        this.$element = $(element);
        this.$window = $(window).on("scroll.sticky.amui", UI.utils.debounce($.proxy(this.checkPosition, this), 50)).on("click.sticky.amui", UI.utils.debounce($.proxy(this.checkPosition, this), 1));
        this.original = {
            offsetTop: this.$element.offset().top,
            width: this.$element.width()
        };
        this.sticked = null;
        this.checkPosition();
    };
    Sticky.DEFAULTS = {
        top: 0,
        cls: "am-sticky"
    };
    Sticky.prototype.checkPosition = function() {
        if (!this.$element.is(":visible")) return;
        var scrollHeight = $(document).height(), scrollTop = this.$window.scrollTop(), options = this.options, offsetTop = options.top, $element = this.$element, animation = options.animation ? " am-animation-" + options.animation : "";
        this.sticked = scrollTop > this.original.offsetTop ? "sticky" : false;
        if (this.sticked) {
            $element.addClass(options.cls + animation).css({
                top: offsetTop
            });
        } else {
            $element.removeClass(options.cls + animation).css({
                top: ""
            });
        }
    };
    UI.sticky = Sticky;
    // Sticky Plugin
    function Plugin(option) {
        return this.each(function() {
            var $this = $(this), data = $this.data("am.sticky"), options = typeof option == "object" && option;
            if (!data) $this.data("am.sticky", data = new Sticky(this, options));
            if (typeof option == "string") data[option]();
        });
    }
    $.fn.sticky = Plugin;
    // Init code
    $(window).on("load", function() {
        $("[data-am-sticky]").each(function() {
            var $this = $(this), options = UI.utils.options($this.attr("data-am-sticky"));
            Plugin.call($this, options);
        });
    });
    module.exports = Sticky;
});
define("util.cookie", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    "use strict";
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    var cookie = {
        get: function(name) {
            var cookieName = encodeURIComponent(name) + "=", cookieStart = document.cookie.indexOf(cookieName), cookieValue = null, cookieEnd;
            if (cookieStart > -1) {
                cookieEnd = document.cookie.indexOf(";", cookieStart);
                if (cookieEnd == -1) {
                    cookieEnd = document.cookie.length;
                }
                cookieValue = decodeURIComponent(document.cookie.substring(cookieStart + cookieName.length, cookieEnd));
            }
            return cookieValue;
        },
        set: function(name, value, expires, path, domain, secure) {
            var cookieText = encodeURIComponent(name) + "=" + encodeURIComponent(value);
            if (expires instanceof Date) {
                cookieText += "; expires=" + expires.toGMTString();
            }
            if (path) {
                cookieText += "; path=" + path;
            }
            if (domain) {
                cookieText += "; domain=" + domain;
            }
            if (secure) {
                cookieText += "; secure";
            }
            document.cookie = cookieText;
        },
        unset: function(name, path, domain, secure) {
            this.set(name, "", new Date(0), path, domain, secure);
        }
    };
    UI.utils.cookie = cookie;
    module.exports = cookie;
});
define("util.fastclick", [], function(require, exports, module) {
    var $ = window.Zepto;
    /**
     * FastClick: polyfill to remove click delays on browsers with touch UIs.
     *
     * @version 1.0.2
     * @codingstandard ftlabs-jsv2
     * @copyright The Financial Times Limited [All Rights Reserved]
     * @license MIT License (see LICENSE.txt)
     */
    /*jslint browser:true, node:true*/
    /*global define, Event, Node*/
    /**
     * Instantiate fast-clicking listeners on the specified layer.
     *
     * @constructor
     * @param {Element} layer The layer to listen on
     * @param {Object} options The options to override the defaults
     */
    function FastClick(layer, options) {
        "use strict";
        var oldOnClick;
        options = options || {};
        /**
         * Whether a click is currently being tracked.
         *
         * @type boolean
         */
        this.trackingClick = false;
        /**
         * Timestamp for when click tracking started.
         *
         * @type number
         */
        this.trackingClickStart = 0;
        /**
         * The element being tracked for a click.
         *
         * @type EventTarget
         */
        this.targetElement = null;
        /**
         * X-coordinate of touch start event.
         *
         * @type number
         */
        this.touchStartX = 0;
        /**
         * Y-coordinate of touch start event.
         *
         * @type number
         */
        this.touchStartY = 0;
        /**
         * ID of the last touch, retrieved from Touch.identifier.
         *
         * @type number
         */
        this.lastTouchIdentifier = 0;
        /**
         * Touchmove boundary, beyond which a click will be cancelled.
         *
         * @type number
         */
        this.touchBoundary = options.touchBoundary || 10;
        /**
         * The FastClick layer.
         *
         * @type Element
         */
        this.layer = layer;
        /**
         * The minimum time between tap(touchstart and touchend) events
         *
         * @type number
         */
        this.tapDelay = options.tapDelay || 200;
        if (FastClick.notNeeded(layer)) {
            return;
        }
        // Some old versions of Android don't have Function.prototype.bind
        function bind(method, context) {
            return function() {
                return method.apply(context, arguments);
            };
        }
        var methods = [ "onMouse", "onClick", "onTouchStart", "onTouchMove", "onTouchEnd", "onTouchCancel" ];
        var context = this;
        for (var i = 0, l = methods.length; i < l; i++) {
            context[methods[i]] = bind(context[methods[i]], context);
        }
        // Set up event handlers as required
        if (deviceIsAndroid) {
            layer.addEventListener("mouseover", this.onMouse, true);
            layer.addEventListener("mousedown", this.onMouse, true);
            layer.addEventListener("mouseup", this.onMouse, true);
        }
        layer.addEventListener("click", this.onClick, true);
        layer.addEventListener("touchstart", this.onTouchStart, false);
        layer.addEventListener("touchmove", this.onTouchMove, false);
        layer.addEventListener("touchend", this.onTouchEnd, false);
        layer.addEventListener("touchcancel", this.onTouchCancel, false);
        // Hack is required for browsers that don't support Event#stopImmediatePropagation (e.g. Android 2)
        // which is how FastClick normally stops click events bubbling to callbacks registered on the FastClick
        // layer when they are cancelled.
        if (!Event.prototype.stopImmediatePropagation) {
            layer.removeEventListener = function(type, callback, capture) {
                var rmv = Node.prototype.removeEventListener;
                if (type === "click") {
                    rmv.call(layer, type, callback.hijacked || callback, capture);
                } else {
                    rmv.call(layer, type, callback, capture);
                }
            };
            layer.addEventListener = function(type, callback, capture) {
                var adv = Node.prototype.addEventListener;
                if (type === "click") {
                    adv.call(layer, type, callback.hijacked || (callback.hijacked = function(event) {
                        if (!event.propagationStopped) {
                            callback(event);
                        }
                    }), capture);
                } else {
                    adv.call(layer, type, callback, capture);
                }
            };
        }
        // If a handler is already declared in the element's onclick attribute, it will be fired before
        // FastClick's onClick handler. Fix this by pulling out the user-defined handler function and
        // adding it as listener.
        if (typeof layer.onclick === "function") {
            // Android browser on at least 3.2 requires a new reference to the function in layer.onclick
            // - the old one won't work if passed to addEventListener directly.
            oldOnClick = layer.onclick;
            layer.addEventListener("click", function(event) {
                oldOnClick(event);
            }, false);
            layer.onclick = null;
        }
    }
    /**
     * Android requires exceptions.
     *
     * @type boolean
     */
    var deviceIsAndroid = navigator.userAgent.indexOf("Android") > 0;
    /**
     * iOS requires exceptions.
     *
     * @type boolean
     */
    var deviceIsIOS = /iP(ad|hone|od)/.test(navigator.userAgent);
    /**
     * iOS 4 requires an exception for select elements.
     *
     * @type boolean
     */
    var deviceIsIOS4 = deviceIsIOS && /OS 4_\d(_\d)?/.test(navigator.userAgent);
    /**
     * iOS 6.0(+?) requires the target element to be manually derived
     *
     * @type boolean
     */
    var deviceIsIOSWithBadTarget = deviceIsIOS && /OS ([6-9]|\d{2})_\d/.test(navigator.userAgent);
    /**
     * Determine whether a given element requires a native click.
     *
     * @param {EventTarget|Element} target Target DOM element
     * @returns {boolean} Returns true if the element needs a native click
     */
    FastClick.prototype.needsClick = function(target) {
        "use strict";
        switch (target.nodeName.toLowerCase()) {
          // Don't send a synthetic click to disabled inputs (issue #62)
            case "button":
          case "select":
          case "textarea":
            if (target.disabled) {
                return true;
            }
            break;

          case "input":
            // File inputs need real clicks on iOS 6 due to a browser bug (issue #68)
            if (deviceIsIOS && target.type === "file" || target.disabled) {
                return true;
            }
            break;

          case "label":
          case "video":
            return true;
        }
        return /\bneedsclick\b/.test(target.className);
    };
    /**
     * Determine whether a given element requires a call to focus to simulate click into element.
     *
     * @param {EventTarget|Element} target Target DOM element
     * @returns {boolean} Returns true if the element requires a call to focus to simulate native click.
     */
    FastClick.prototype.needsFocus = function(target) {
        "use strict";
        switch (target.nodeName.toLowerCase()) {
          case "textarea":
            return true;

          case "select":
            return !deviceIsAndroid;

          case "input":
            switch (target.type) {
              case "button":
              case "checkbox":
              case "file":
              case "image":
              case "radio":
              case "submit":
                return false;
            }
            // No point in attempting to focus disabled inputs
            return !target.disabled && !target.readOnly;

          default:
            return /\bneedsfocus\b/.test(target.className);
        }
    };
    /**
     * Send a click event to the specified element.
     *
     * @param {EventTarget|Element} targetElement
     * @param {Event} event
     */
    FastClick.prototype.sendClick = function(targetElement, event) {
        "use strict";
        var clickEvent, touch;
        // On some Android devices activeElement needs to be blurred otherwise the synthetic click will have no effect (#24)
        if (document.activeElement && document.activeElement !== targetElement) {
            document.activeElement.blur();
        }
        touch = event.changedTouches[0];
        // Synthesise a click event, with an extra attribute so it can be tracked
        clickEvent = document.createEvent("MouseEvents");
        clickEvent.initMouseEvent(this.determineEventType(targetElement), true, true, window, 1, touch.screenX, touch.screenY, touch.clientX, touch.clientY, false, false, false, false, 0, null);
        clickEvent.forwardedTouchEvent = true;
        targetElement.dispatchEvent(clickEvent);
    };
    FastClick.prototype.determineEventType = function(targetElement) {
        "use strict";
        //Issue #159: Android Chrome Select Box does not open with a synthetic click event
        if (deviceIsAndroid && targetElement.tagName.toLowerCase() === "select") {
            return "mousedown";
        }
        return "click";
    };
    /**
     * @param {EventTarget|Element} targetElement
     */
    FastClick.prototype.focus = function(targetElement) {
        "use strict";
        var length;
        // Issue #160: on iOS 7, some input elements (e.g. date datetime) throw a vague TypeError on setSelectionRange. These elements don't have an integer value for the selectionStart and selectionEnd properties, but unfortunately that can't be used for detection because accessing the properties also throws a TypeError. Just check the type instead. Filed as Apple bug #15122724.
        if (deviceIsIOS && targetElement.setSelectionRange && targetElement.type.indexOf("date") !== 0 && targetElement.type !== "time") {
            length = targetElement.value.length;
            targetElement.setSelectionRange(length, length);
        } else {
            targetElement.focus();
        }
    };
    /**
     * Check whether the given target element is a child of a scrollable layer and if so, set a flag on it.
     *
     * @param {EventTarget|Element} targetElement
     */
    FastClick.prototype.updateScrollParent = function(targetElement) {
        "use strict";
        var scrollParent, parentElement;
        scrollParent = targetElement.fastClickScrollParent;
        // Attempt to discover whether the target element is contained within a scrollable layer. Re-check if the
        // target element was moved to another parent.
        if (!scrollParent || !scrollParent.contains(targetElement)) {
            parentElement = targetElement;
            do {
                if (parentElement.scrollHeight > parentElement.offsetHeight) {
                    scrollParent = parentElement;
                    targetElement.fastClickScrollParent = parentElement;
                    break;
                }
                parentElement = parentElement.parentElement;
            } while (parentElement);
        }
        // Always update the scroll top tracker if possible.
        if (scrollParent) {
            scrollParent.fastClickLastScrollTop = scrollParent.scrollTop;
        }
    };
    /**
     * @param {EventTarget} targetElement
     * @returns {Element|EventTarget}
     */
    FastClick.prototype.getTargetElementFromEventTarget = function(eventTarget) {
        "use strict";
        // On some older browsers (notably Safari on iOS 4.1 - see issue #56) the event target may be a text node.
        if (eventTarget.nodeType === Node.TEXT_NODE) {
            return eventTarget.parentNode;
        }
        return eventTarget;
    };
    /**
     * On touch start, record the position and scroll offset.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.onTouchStart = function(event) {
        "use strict";
        var targetElement, touch, selection;
        // Ignore multiple touches, otherwise pinch-to-zoom is prevented if both fingers are on the FastClick element (issue #111).
        if (event.targetTouches.length > 1) {
            return true;
        }
        targetElement = this.getTargetElementFromEventTarget(event.target);
        touch = event.targetTouches[0];
        if (deviceIsIOS) {
            // Only trusted events will deselect text on iOS (issue #49)
            selection = window.getSelection();
            if (selection.rangeCount && !selection.isCollapsed) {
                return true;
            }
            if (!deviceIsIOS4) {
                // Weird things happen on iOS when an alert or confirm dialog is opened from a click event callback (issue #23):
                // when the user next taps anywhere else on the page, new touchstart and touchend events are dispatched
                // with the same identifier as the touch event that previously triggered the click that triggered the alert.
                // Sadly, there is an issue on iOS 4 that causes some normal touch events to have the same identifier as an
                // immediately preceeding touch event (issue #52), so this fix is unavailable on that platform.
                if (touch.identifier === this.lastTouchIdentifier) {
                    event.preventDefault();
                    return false;
                }
                this.lastTouchIdentifier = touch.identifier;
                // If the target element is a child of a scrollable layer (using -webkit-overflow-scrolling: touch) and:
                // 1) the user does a fling scroll on the scrollable layer
                // 2) the user stops the fling scroll with another tap
                // then the event.target of the last 'touchend' event will be the element that was under the user's finger
                // when the fling scroll was started, causing FastClick to send a click event to that layer - unless a check
                // is made to ensure that a parent layer was not scrolled before sending a synthetic click (issue #42).
                this.updateScrollParent(targetElement);
            }
        }
        this.trackingClick = true;
        this.trackingClickStart = event.timeStamp;
        this.targetElement = targetElement;
        this.touchStartX = touch.pageX;
        this.touchStartY = touch.pageY;
        // Prevent phantom clicks on fast double-tap (issue #36)
        if (event.timeStamp - this.lastClickTime < this.tapDelay) {
            event.preventDefault();
        }
        return true;
    };
    /**
     * Based on a touchmove event object, check whether the touch has moved past a boundary since it started.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.touchHasMoved = function(event) {
        "use strict";
        var touch = event.changedTouches[0], boundary = this.touchBoundary;
        if (Math.abs(touch.pageX - this.touchStartX) > boundary || Math.abs(touch.pageY - this.touchStartY) > boundary) {
            return true;
        }
        return false;
    };
    /**
     * Update the last position.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.onTouchMove = function(event) {
        "use strict";
        if (!this.trackingClick) {
            return true;
        }
        // If the touch has moved, cancel the click tracking
        if (this.targetElement !== this.getTargetElementFromEventTarget(event.target) || this.touchHasMoved(event)) {
            this.trackingClick = false;
            this.targetElement = null;
        }
        return true;
    };
    /**
     * Attempt to find the labelled control for the given label element.
     *
     * @param {EventTarget|HTMLLabelElement} labelElement
     * @returns {Element|null}
     */
    FastClick.prototype.findControl = function(labelElement) {
        "use strict";
        // Fast path for newer browsers supporting the HTML5 control attribute
        if (labelElement.control !== undefined) {
            return labelElement.control;
        }
        // All browsers under test that support touch events also support the HTML5 htmlFor attribute
        if (labelElement.htmlFor) {
            return document.getElementById(labelElement.htmlFor);
        }
        // If no for attribute exists, attempt to retrieve the first labellable descendant element
        // the list of which is defined here: http://www.w3.org/TR/html5/forms.html#category-label
        return labelElement.querySelector("button, input:not([type=hidden]), keygen, meter, output, progress, select, textarea");
    };
    /**
     * On touch end, determine whether to send a click event at once.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.onTouchEnd = function(event) {
        "use strict";
        var forElement, trackingClickStart, targetTagName, scrollParent, touch, targetElement = this.targetElement;
        if (!this.trackingClick) {
            return true;
        }
        // Prevent phantom clicks on fast double-tap (issue #36)
        if (event.timeStamp - this.lastClickTime < this.tapDelay) {
            this.cancelNextClick = true;
            return true;
        }
        // Reset to prevent wrong click cancel on input (issue #156).
        this.cancelNextClick = false;
        this.lastClickTime = event.timeStamp;
        trackingClickStart = this.trackingClickStart;
        this.trackingClick = false;
        this.trackingClickStart = 0;
        // On some iOS devices, the targetElement supplied with the event is invalid if the layer
        // is performing a transition or scroll, and has to be re-detected manually. Note that
        // for this to function correctly, it must be called *after* the event target is checked!
        // See issue #57; also filed as rdar://13048589 .
        if (deviceIsIOSWithBadTarget) {
            touch = event.changedTouches[0];
            // In certain cases arguments of elementFromPoint can be negative, so prevent setting targetElement to null
            targetElement = document.elementFromPoint(touch.pageX - window.pageXOffset, touch.pageY - window.pageYOffset) || targetElement;
            targetElement.fastClickScrollParent = this.targetElement.fastClickScrollParent;
        }
        targetTagName = targetElement.tagName.toLowerCase();
        if (targetTagName === "label") {
            forElement = this.findControl(targetElement);
            if (forElement) {
                this.focus(targetElement);
                if (deviceIsAndroid) {
                    return false;
                }
                targetElement = forElement;
            }
        } else if (this.needsFocus(targetElement)) {
            // Case 1: If the touch started a while ago (best guess is 100ms based on tests for issue #36) then focus will be triggered anyway. Return early and unset the target element reference so that the subsequent click will be allowed through.
            // Case 2: Without this exception for input elements tapped when the document is contained in an iframe, then any inputted text won't be visible even though the value attribute is updated as the user types (issue #37).
            if (event.timeStamp - trackingClickStart > 100 || deviceIsIOS && window.top !== window && targetTagName === "input") {
                this.targetElement = null;
                return false;
            }
            this.focus(targetElement);
            this.sendClick(targetElement, event);
            // Select elements need the event to go through on iOS 4, otherwise the selector menu won't open.
            // Also this breaks opening selects when VoiceOver is active on iOS6, iOS7 (and possibly others)
            if (!deviceIsIOS || targetTagName !== "select") {
                this.targetElement = null;
                event.preventDefault();
            }
            return false;
        }
        if (deviceIsIOS && !deviceIsIOS4) {
            // Don't send a synthetic click event if the target element is contained within a parent layer that was scrolled
            // and this tap is being used to stop the scrolling (usually initiated by a fling - issue #42).
            scrollParent = targetElement.fastClickScrollParent;
            if (scrollParent && scrollParent.fastClickLastScrollTop !== scrollParent.scrollTop) {
                return true;
            }
        }
        // Prevent the actual click from going though - unless the target node is marked as requiring
        // real clicks or if it is in the whitelist in which case only non-programmatic clicks are permitted.
        if (!this.needsClick(targetElement)) {
            event.preventDefault();
            this.sendClick(targetElement, event);
        }
        return false;
    };
    /**
     * On touch cancel, stop tracking the click.
     *
     * @returns {void}
     */
    FastClick.prototype.onTouchCancel = function() {
        "use strict";
        this.trackingClick = false;
        this.targetElement = null;
    };
    /**
     * Determine mouse events which should be permitted.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.onMouse = function(event) {
        "use strict";
        // If a target element was never set (because a touch event was never fired) allow the event
        if (!this.targetElement) {
            return true;
        }
        if (event.forwardedTouchEvent) {
            return true;
        }
        // Programmatically generated events targeting a specific element should be permitted
        if (!event.cancelable) {
            return true;
        }
        // Derive and check the target element to see whether the mouse event needs to be permitted;
        // unless explicitly enabled, prevent non-touch click events from triggering actions,
        // to prevent ghost/doubleclicks.
        if (!this.needsClick(this.targetElement) || this.cancelNextClick) {
            // Prevent any user-added listeners declared on FastClick element from being fired.
            if (event.stopImmediatePropagation) {
                event.stopImmediatePropagation();
            } else {
                // Part of the hack for browsers that don't support Event#stopImmediatePropagation (e.g. Android 2)
                event.propagationStopped = true;
            }
            // Cancel the event
            event.stopPropagation();
            event.preventDefault();
            return false;
        }
        // If the mouse event is permitted, return true for the action to go through.
        return true;
    };
    /**
     * On actual clicks, determine whether this is a touch-generated click, a click action occurring
     * naturally after a delay after a touch (which needs to be cancelled to avoid duplication), or
     * an actual click which should be permitted.
     *
     * @param {Event} event
     * @returns {boolean}
     */
    FastClick.prototype.onClick = function(event) {
        "use strict";
        var permitted;
        // It's possible for another FastClick-like library delivered with third-party code to fire a click event before FastClick does (issue #44). In that case, set the click-tracking flag back to false and return early. This will cause onTouchEnd to return early.
        if (this.trackingClick) {
            this.targetElement = null;
            this.trackingClick = false;
            return true;
        }
        // Very odd behaviour on iOS (issue #18): if a submit element is present inside a form and the user hits enter in the iOS simulator or clicks the Go button on the pop-up OS keyboard the a kind of 'fake' click event will be triggered with the submit-type input element as the target.
        if (event.target.type === "submit" && event.detail === 0) {
            return true;
        }
        permitted = this.onMouse(event);
        // Only unset targetElement if the click is not permitted. This will ensure that the check for !targetElement in onMouse fails and the browser's click doesn't go through.
        if (!permitted) {
            this.targetElement = null;
        }
        // If clicks are permitted, return true for the action to go through.
        return permitted;
    };
    /**
     * Remove all FastClick's event listeners.
     *
     * @returns {void}
     */
    FastClick.prototype.destroy = function() {
        "use strict";
        var layer = this.layer;
        if (deviceIsAndroid) {
            layer.removeEventListener("mouseover", this.onMouse, true);
            layer.removeEventListener("mousedown", this.onMouse, true);
            layer.removeEventListener("mouseup", this.onMouse, true);
        }
        layer.removeEventListener("click", this.onClick, true);
        layer.removeEventListener("touchstart", this.onTouchStart, false);
        layer.removeEventListener("touchmove", this.onTouchMove, false);
        layer.removeEventListener("touchend", this.onTouchEnd, false);
        layer.removeEventListener("touchcancel", this.onTouchCancel, false);
    };
    /**
     * Check whether FastClick is needed.
     *
     * @param {Element} layer The layer to listen on
     */
    FastClick.notNeeded = function(layer) {
        "use strict";
        var metaViewport;
        var chromeVersion;
        // Devices that don't support touch don't need FastClick
        if (typeof window.ontouchstart === "undefined") {
            return true;
        }
        // Chrome version - zero for other browsers
        chromeVersion = +(/Chrome\/([0-9]+)/.exec(navigator.userAgent) || [ , 0 ])[1];
        if (chromeVersion) {
            if (deviceIsAndroid) {
                metaViewport = document.querySelector("meta[name=viewport]");
                if (metaViewport) {
                    // Chrome on Android with user-scalable="no" doesn't need FastClick (issue #89)
                    if (metaViewport.content.indexOf("user-scalable=no") !== -1) {
                        return true;
                    }
                    // Chrome 32 and above with width=device-width or less don't need FastClick
                    if (chromeVersion > 31 && document.documentElement.scrollWidth <= window.outerWidth) {
                        return true;
                    }
                }
            } else {
                return true;
            }
        }
        // IE10 with -ms-touch-action: none, which disables double-tap-to-zoom (issue #97)
        if (layer.style.msTouchAction === "none") {
            return true;
        }
        return false;
    };
    /**
     * Factory method for creating a FastClick object
     *
     * @param {Element} layer The layer to listen on
     * @param {Object} options The options to override the defaults
     */
    FastClick.attach = function(layer, options) {
        "use strict";
        return new FastClick(layer, options);
    };
    module.exports = FastClick;
});
define("util.fullscreen", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector" ], function(require, exports, module) {
    require("core");
    var $ = window.Zepto, UI = $.AMUI;
    /**
     * @via https://github.com/sindresorhus/screenfull.js
     * @license MIT © Sindre Sorhus
     * @version 1.2.1
     */
    "use strict";
    var keyboardAllowed = typeof Element !== "undefined" && "ALLOW_KEYBOARD_INPUT" in Element;
    var fn = function() {
        var val;
        var valLength;
        var fnMap = [ [ "requestFullscreen", "exitFullscreen", "fullscreenElement", "fullscreenEnabled", "fullscreenchange", "fullscreenerror" ], // new WebKit
        [ "webkitRequestFullscreen", "webkitExitFullscreen", "webkitFullscreenElement", "webkitFullscreenEnabled", "webkitfullscreenchange", "webkitfullscreenerror" ], // old WebKit (Safari 5.1)
        [ "webkitRequestFullScreen", "webkitCancelFullScreen", "webkitCurrentFullScreenElement", "webkitCancelFullScreen", "webkitfullscreenchange", "webkitfullscreenerror" ], [ "mozRequestFullScreen", "mozCancelFullScreen", "mozFullScreenElement", "mozFullScreenEnabled", "mozfullscreenchange", "mozfullscreenerror" ], [ "msRequestFullscreen", "msExitFullscreen", "msFullscreenElement", "msFullscreenEnabled", "MSFullscreenChange", "MSFullscreenError" ] ];
        var i = 0;
        var l = fnMap.length;
        var ret = {};
        for (;i < l; i++) {
            val = fnMap[i];
            if (val && val[1] in document) {
                for (i = 0, valLength = val.length; i < valLength; i++) {
                    ret[fnMap[0][i]] = val[i];
                }
                return ret;
            }
        }
        return false;
    }();
    var fullscreen = {
        request: function(elem) {
            var request = fn.requestFullscreen;
            elem = elem || document.documentElement;
            // Work around Safari 5.1 bug: reports support for
            // keyboard in fullscreen even though it doesn't.
            // Browser sniffing, since the alternative with
            // setTimeout is even worse.
            if (/5\.1[\.\d]* Safari/.test(navigator.userAgent)) {
                elem[request]();
            } else {
                elem[request](keyboardAllowed && Element.ALLOW_KEYBOARD_INPUT);
            }
        },
        exit: function() {
            document[fn.exitFullscreen]();
        },
        toggle: function(elem) {
            if (this.isFullscreen) {
                this.exit();
            } else {
                this.request(elem);
            }
        },
        onchange: function() {},
        onerror: function() {},
        raw: fn
    };
    if (!fn) {
        module.exports = false;
        return;
    }
    Object.defineProperties(fullscreen, {
        isFullscreen: {
            get: function() {
                return !!document[fn.fullscreenElement];
            }
        },
        element: {
            enumerable: true,
            get: function() {
                return document[fn.fullscreenElement];
            }
        },
        enabled: {
            enumerable: true,
            get: function() {
                // Coerce to boolean in case of old WebKit
                return !!document[fn.fullscreenEnabled];
            }
        }
    });
    document.addEventListener(fn.fullscreenchange, function(e) {
        fullscreen.onchange.call(fullscreen, e);
    });
    document.addEventListener(fn.fullscreenerror, function(e) {
        fullscreen.onerror.call(fullscreen, e);
    });
    //!window.fullscreen && (window.fullscreen = fullscreen);
    UI.fullscreen = fullscreen;
    module.exports = fullscreen;
});
define("util.qrcode", [], function(require, exports, module) {
    var $ = Zepto;
    /**
     * @ver 1.1.0
     * @via https://github.com/aralejs/qrcode/blob/master/src/qrcode.js
     * @license http://aralejs.org/
     */
    var qrcodeAlgObjCache = [];
    /**
     * 二维码构造函数，主要用于绘制
     * @param  {参数列表} opt 传递参数
     * @return {}
     */
    var qrcode = function(opt) {
        if (typeof opt === "string") {
            // 只编码ASCII字符串
            opt = {
                text: opt
            };
        }
        //设置默认参数
        this.options = $.extend({}, {
            text: "",
            render: "",
            width: 256,
            height: 256,
            correctLevel: 3,
            background: "#ffffff",
            foreground: "#000000"
        }, opt);
        //使用QRCodeAlg创建二维码结构
        var qrCodeAlg = null;
        for (var i = 0, l = qrcodeAlgObjCache.length; i < l; i++) {
            if (qrcodeAlgObjCache[i].text == this.options.text && qrcodeAlgObjCache[i].text.correctLevel == this.options.correctLevel) {
                qrCodeAlg = qrcodeAlgObjCache[i].obj;
                break;
            }
        }
        if (i == l) {
            qrCodeAlg = new QRCodeAlg(this.options.text, this.options.correctLevel);
            qrcodeAlgObjCache.push({
                text: this.options.text,
                correctLevel: this.options.correctLevel,
                obj: qrCodeAlg
            });
        }
        if (this.options.render) {
            switch (this.options.render) {
              case "canvas":
                return this.createCanvas(qrCodeAlg);

              case "table":
                return this.createTable(qrCodeAlg);

              case "svg":
                return this.createSVG(qrCodeAlg);

              default:
                return this.createDefault(qrCodeAlg);
            }
        }
        return this.createDefault(qrCodeAlg);
    };
    /**
     * 使用Canvas来画二维码
     * @return {}
     */
    qrcode.prototype.createDefault = function(qrCodeAlg) {
        var canvas = document.createElement("canvas");
        if (canvas.getContext) return this.createCanvas(qrCodeAlg);
        SVG_NS = "http://www.w3.org/2000/svg";
        if (!!document.createElementNS && !!document.createElementNS(SVG_NS, "svg").createSVGRect) return this.createSVG(qrCodeAlg);
        return this.createTable(qrCodeAlg);
    };
    qrcode.prototype.createCanvas = function(qrCodeAlg) {
        //创建canvas节点
        var canvas = document.createElement("canvas");
        canvas.width = this.options.width;
        canvas.height = this.options.height;
        var ctx = canvas.getContext("2d");
        //计算每个点的长宽
        var tileW = (this.options.width / qrCodeAlg.getModuleCount()).toPrecision(4);
        var tileH = this.options.height / qrCodeAlg.getModuleCount().toPrecision(4);
        //绘制
        for (var row = 0; row < qrCodeAlg.getModuleCount(); row++) {
            for (var col = 0; col < qrCodeAlg.getModuleCount(); col++) {
                ctx.fillStyle = qrCodeAlg.modules[row][col] ? this.options.foreground : this.options.background;
                var w = Math.ceil((col + 1) * tileW) - Math.floor(col * tileW);
                var h = Math.ceil((row + 1) * tileW) - Math.floor(row * tileW);
                ctx.fillRect(Math.round(col * tileW), Math.round(row * tileH), w, h);
            }
        }
        //返回绘制的节点
        return canvas;
    };
    /**
     * 使用table来绘制二维码
     * @return {}
     */
    qrcode.prototype.createTable = function(qrCodeAlg) {
        //创建table节点
        var s = [];
        s.push('<table style="border:0px; margin:0px; padding:0px; border-collapse:collapse; background-color: ' + this.options.background + ';">');
        // 计算每个节点的长宽；取整，防止点之间出现分离
        var tileW = -1, tileH = -1, caculateW = -1, caculateH = -1;
        tileW = caculateW = Math.floor(this.options.width / qrCodeAlg.getModuleCount());
        tileH = caculateH = Math.floor(this.options.height / qrCodeAlg.getModuleCount());
        if (caculateW <= 0) {
            if (qrCodeAlg.getModuleCount() < 80) {
                tileW = 2;
            } else {
                tileW = 1;
            }
        }
        if (caculateH <= 0) {
            if (qrCodeAlg.getModuleCount() < 80) {
                tileH = 2;
            } else {
                tileH = 1;
            }
        }
        // 绘制二维码
        foreTd = '<td style="border:0px; margin:0px; padding:0px; width:' + tileW + "px; background-color: " + this.options.foreground + '"></td>', 
        backTd = '<td style="border:0px; margin:0px; padding:0px; width:' + tileW + "px; background-color: " + this.options.background + '"></td>', 
        l = qrCodeAlg.getModuleCount();
        for (var row = 0; row < l; row++) {
            s.push('<tr style="border:0px; margin:0px; padding:0px; height: ' + tileH + 'px">');
            for (var col = 0; col < l; col++) {
                s.push(qrCodeAlg.modules[row][col] ? foreTd : backTd);
            }
            s.push("</tr>");
        }
        s.push("</table>");
        var span = document.createElement("span");
        span.innerHTML = s.join("");
        return span.firstChild;
    };
    /**
     * 使用SVG开绘制二维码
     * @return {}
     */
    qrcode.prototype.createSVG = function(qrCodeAlg) {
        var x, dx, y, dy, moduleCount = qrCodeAlg.getModuleCount(), scale = this.options.height / this.options.width, svg = '<svg xmlns="http://www.w3.org/2000/svg" ' + 'width="' + this.options.width + 'px" height="' + this.options.height + 'px" ' + 'viewbox="0 0 ' + moduleCount * 10 + " " + moduleCount * 10 * scale + '">', rectHead = "<path ", foreRect = ' style="stroke-width:0.5;stroke:' + this.options.foreground + ";fill:" + this.options.foreground + ';"></path>', backRect = ' style="stroke-width:0.5;stroke:' + this.options.background + ";fill:" + this.options.background + ';"></path>';
        // draw in the svg
        for (var row = 0; row < moduleCount; row++) {
            for (var col = 0; col < moduleCount; col++) {
                x = col * 10;
                y = row * 10 * scale;
                dx = (col + 1) * 10;
                dy = (row + 1) * 10 * scale;
                svg += rectHead + 'd="M ' + x + "," + y + " L " + dx + "," + y + " L " + dx + "," + dy + " L " + x + "," + dy + ' Z"';
                svg += qrCodeAlg.modules[row][col] ? foreRect : backRect;
            }
        }
        svg += "</svg>";
        // return just built svg
        return $(svg)[0];
    };
    module.exports = qrcode;
    /**
     * 获取单个字符的utf8编码
     * unicode BMP平面约65535个字符
     * @param {num} code
     * return {array}
     */
    function unicodeFormat8(code) {
        // 1 byte
        if (code < 128) {
            return [ code ];
        } else if (code < 2048) {
            c0 = 192 + (code >> 6);
            c1 = 128 + (code & 63);
            return [ c0, c1 ];
        } else {
            c0 = 224 + (code >> 12);
            c1 = 128 + (code >> 6 & 63);
            c2 = 128 + (code & 63);
            return [ c0, c1, c2 ];
        }
    }
    /**
     * 获取字符串的utf8编码字节串
     * @param {string} string
     * @return {array}
     */
    function getUTF8Bytes(string) {
        var utf8codes = [];
        for (var i = 0; i < string.length; i++) {
            var code = string.charCodeAt(i);
            var utf8 = unicodeFormat8(code);
            for (var j = 0; j < utf8.length; j++) {
                utf8codes.push(utf8[j]);
            }
        }
        return utf8codes;
    }
    /**
     * 二维码算法实现
     * @param {string} data              要编码的信息字符串
     * @param {num} errorCorrectLevel 纠错等级
     */
    function QRCodeAlg(data, errorCorrectLevel) {
        this.typeNumber = -1;
        //版本
        this.errorCorrectLevel = errorCorrectLevel;
        this.modules = null;
        //二维矩阵，存放最终结果
        this.moduleCount = 0;
        //矩阵大小
        this.dataCache = null;
        //数据缓存
        this.rsBlocks = null;
        //版本数据信息
        this.totalDataCount = -1;
        //可使用的数据量
        this.data = data;
        this.utf8bytes = getUTF8Bytes(data);
        this.make();
    }
    QRCodeAlg.prototype = {
        constructor: QRCodeAlg,
        /**
         * 获取二维码矩阵大小
         * @return {num} 矩阵大小
         */
        getModuleCount: function() {
            return this.moduleCount;
        },
        /**
         * 编码
         */
        make: function() {
            this.getRightType();
            this.dataCache = this.createData();
            this.createQrcode();
        },
        /**
         * 设置二位矩阵功能图形
         * @param  {bool} test 表示是否在寻找最好掩膜阶段
         * @param  {num} maskPattern 掩膜的版本
         */
        makeImpl: function(maskPattern) {
            this.moduleCount = this.typeNumber * 4 + 17;
            this.modules = new Array(this.moduleCount);
            for (var row = 0; row < this.moduleCount; row++) {
                this.modules[row] = new Array(this.moduleCount);
            }
            this.setupPositionProbePattern(0, 0);
            this.setupPositionProbePattern(this.moduleCount - 7, 0);
            this.setupPositionProbePattern(0, this.moduleCount - 7);
            this.setupPositionAdjustPattern();
            this.setupTimingPattern();
            this.setupTypeInfo(true, maskPattern);
            if (this.typeNumber >= 7) {
                this.setupTypeNumber(true);
            }
            this.mapData(this.dataCache, maskPattern);
        },
        /**
         * 设置二维码的位置探测图形
         * @param  {num} row 探测图形的中心横坐标
         * @param  {num} col 探测图形的中心纵坐标
         */
        setupPositionProbePattern: function(row, col) {
            for (var r = -1; r <= 7; r++) {
                if (row + r <= -1 || this.moduleCount <= row + r) continue;
                for (var c = -1; c <= 7; c++) {
                    if (col + c <= -1 || this.moduleCount <= col + c) continue;
                    if (0 <= r && r <= 6 && (c == 0 || c == 6) || 0 <= c && c <= 6 && (r == 0 || r == 6) || 2 <= r && r <= 4 && 2 <= c && c <= 4) {
                        this.modules[row + r][col + c] = true;
                    } else {
                        this.modules[row + r][col + c] = false;
                    }
                }
            }
        },
        /**
         * 创建二维码
         * @return {[type]} [description]
         */
        createQrcode: function() {
            var minLostPoint = 0;
            var pattern = 0;
            var bestModules = null;
            for (var i = 0; i < 8; i++) {
                this.makeImpl(i);
                var lostPoint = QRUtil.getLostPoint(this);
                if (i == 0 || minLostPoint > lostPoint) {
                    minLostPoint = lostPoint;
                    pattern = i;
                    bestModules = this.modules;
                }
            }
            this.modules = bestModules;
            this.setupTypeInfo(false, pattern);
            if (this.typeNumber >= 7) {
                this.setupTypeNumber(false);
            }
        },
        /**
         * 设置定位图形
         * @return {[type]} [description]
         */
        setupTimingPattern: function() {
            for (var r = 8; r < this.moduleCount - 8; r++) {
                if (this.modules[r][6] != null) {
                    continue;
                }
                this.modules[r][6] = r % 2 == 0;
                if (this.modules[6][r] != null) {
                    continue;
                }
                this.modules[6][r] = r % 2 == 0;
            }
        },
        /**
         * 设置矫正图形
         * @return {[type]} [description]
         */
        setupPositionAdjustPattern: function() {
            var pos = QRUtil.getPatternPosition(this.typeNumber);
            for (var i = 0; i < pos.length; i++) {
                for (var j = 0; j < pos.length; j++) {
                    var row = pos[i];
                    var col = pos[j];
                    if (this.modules[row][col] != null) {
                        continue;
                    }
                    for (var r = -2; r <= 2; r++) {
                        for (var c = -2; c <= 2; c++) {
                            if (r == -2 || r == 2 || c == -2 || c == 2 || r == 0 && c == 0) {
                                this.modules[row + r][col + c] = true;
                            } else {
                                this.modules[row + r][col + c] = false;
                            }
                        }
                    }
                }
            }
        },
        /**
         * 设置版本信息（7以上版本才有）
         * @param  {bool} test 是否处于判断最佳掩膜阶段
         * @return {[type]}      [description]
         */
        setupTypeNumber: function(test) {
            var bits = QRUtil.getBCHTypeNumber(this.typeNumber);
            for (var i = 0; i < 18; i++) {
                var mod = !test && (bits >> i & 1) == 1;
                this.modules[Math.floor(i / 3)][i % 3 + this.moduleCount - 8 - 3] = mod;
                this.modules[i % 3 + this.moduleCount - 8 - 3][Math.floor(i / 3)] = mod;
            }
        },
        /**
         * 设置格式信息（纠错等级和掩膜版本）
         * @param  {bool} test
         * @param  {num} maskPattern 掩膜版本
         * @return {}
         */
        setupTypeInfo: function(test, maskPattern) {
            var data = QRErrorCorrectLevel[this.errorCorrectLevel] << 3 | maskPattern;
            var bits = QRUtil.getBCHTypeInfo(data);
            // vertical
            for (var i = 0; i < 15; i++) {
                var mod = !test && (bits >> i & 1) == 1;
                if (i < 6) {
                    this.modules[i][8] = mod;
                } else if (i < 8) {
                    this.modules[i + 1][8] = mod;
                } else {
                    this.modules[this.moduleCount - 15 + i][8] = mod;
                }
                // horizontal
                var mod = !test && (bits >> i & 1) == 1;
                if (i < 8) {
                    this.modules[8][this.moduleCount - i - 1] = mod;
                } else if (i < 9) {
                    this.modules[8][15 - i - 1 + 1] = mod;
                } else {
                    this.modules[8][15 - i - 1] = mod;
                }
            }
            // fixed module
            this.modules[this.moduleCount - 8][8] = !test;
        },
        /**
         * 数据编码
         * @return {[type]} [description]
         */
        createData: function() {
            var buffer = new QRBitBuffer();
            var lengthBits = this.typeNumber > 9 ? 16 : 8;
            buffer.put(4, 4);
            //添加模式
            buffer.put(this.utf8bytes.length, lengthBits);
            for (var i = 0, l = this.utf8bytes.length; i < l; i++) {
                buffer.put(this.utf8bytes[i], 8);
            }
            if (buffer.length + 4 <= this.totalDataCount * 8) {
                buffer.put(0, 4);
            }
            // padding
            while (buffer.length % 8 != 0) {
                buffer.putBit(false);
            }
            // padding
            while (true) {
                if (buffer.length >= this.totalDataCount * 8) {
                    break;
                }
                buffer.put(QRCodeAlg.PAD0, 8);
                if (buffer.length >= this.totalDataCount * 8) {
                    break;
                }
                buffer.put(QRCodeAlg.PAD1, 8);
            }
            return this.createBytes(buffer);
        },
        /**
         * 纠错码编码
         * @param  {buffer} buffer 数据编码
         * @return {[type]}
         */
        createBytes: function(buffer) {
            var offset = 0;
            var maxDcCount = 0;
            var maxEcCount = 0;
            var length = this.rsBlock.length / 3;
            var rsBlocks = new Array();
            for (var i = 0; i < length; i++) {
                var count = this.rsBlock[i * 3 + 0];
                var totalCount = this.rsBlock[i * 3 + 1];
                var dataCount = this.rsBlock[i * 3 + 2];
                for (var j = 0; j < count; j++) {
                    rsBlocks.push([ dataCount, totalCount ]);
                }
            }
            var dcdata = new Array(rsBlocks.length);
            var ecdata = new Array(rsBlocks.length);
            for (var r = 0; r < rsBlocks.length; r++) {
                var dcCount = rsBlocks[r][0];
                var ecCount = rsBlocks[r][1] - dcCount;
                maxDcCount = Math.max(maxDcCount, dcCount);
                maxEcCount = Math.max(maxEcCount, ecCount);
                dcdata[r] = new Array(dcCount);
                for (var i = 0; i < dcdata[r].length; i++) {
                    dcdata[r][i] = 255 & buffer.buffer[i + offset];
                }
                offset += dcCount;
                var rsPoly = QRUtil.getErrorCorrectPolynomial(ecCount);
                var rawPoly = new QRPolynomial(dcdata[r], rsPoly.getLength() - 1);
                var modPoly = rawPoly.mod(rsPoly);
                ecdata[r] = new Array(rsPoly.getLength() - 1);
                for (var i = 0; i < ecdata[r].length; i++) {
                    var modIndex = i + modPoly.getLength() - ecdata[r].length;
                    ecdata[r][i] = modIndex >= 0 ? modPoly.get(modIndex) : 0;
                }
            }
            var data = new Array(this.totalDataCount);
            var index = 0;
            for (var i = 0; i < maxDcCount; i++) {
                for (var r = 0; r < rsBlocks.length; r++) {
                    if (i < dcdata[r].length) {
                        data[index++] = dcdata[r][i];
                    }
                }
            }
            for (var i = 0; i < maxEcCount; i++) {
                for (var r = 0; r < rsBlocks.length; r++) {
                    if (i < ecdata[r].length) {
                        data[index++] = ecdata[r][i];
                    }
                }
            }
            return data;
        },
        /**
         * 布置模块，构建最终信息
         * @param  {} data
         * @param  {} maskPattern
         * @return {}
         */
        mapData: function(data, maskPattern) {
            var inc = -1;
            var row = this.moduleCount - 1;
            var bitIndex = 7;
            var byteIndex = 0;
            for (var col = this.moduleCount - 1; col > 0; col -= 2) {
                if (col == 6) col--;
                while (true) {
                    for (var c = 0; c < 2; c++) {
                        if (this.modules[row][col - c] == null) {
                            var dark = false;
                            if (byteIndex < data.length) {
                                dark = (data[byteIndex] >>> bitIndex & 1) == 1;
                            }
                            var mask = QRUtil.getMask(maskPattern, row, col - c);
                            if (mask) {
                                dark = !dark;
                            }
                            this.modules[row][col - c] = dark;
                            bitIndex--;
                            if (bitIndex == -1) {
                                byteIndex++;
                                bitIndex = 7;
                            }
                        }
                    }
                    row += inc;
                    if (row < 0 || this.moduleCount <= row) {
                        row -= inc;
                        inc = -inc;
                        break;
                    }
                }
            }
        }
    };
    /**
     * 填充字段
     */
    QRCodeAlg.PAD0 = 236;
    QRCodeAlg.PAD1 = 17;
    //---------------------------------------------------------------------
    // 纠错等级对应的编码
    //---------------------------------------------------------------------
    var QRErrorCorrectLevel = [ 1, 0, 3, 2 ];
    //---------------------------------------------------------------------
    // 掩膜版本
    //---------------------------------------------------------------------
    var QRMaskPattern = {
        PATTERN000: 0,
        PATTERN001: 1,
        PATTERN010: 2,
        PATTERN011: 3,
        PATTERN100: 4,
        PATTERN101: 5,
        PATTERN110: 6,
        PATTERN111: 7
    };
    //---------------------------------------------------------------------
    // 工具类
    //---------------------------------------------------------------------
    var QRUtil = {
        /*
         每个版本矫正图形的位置
         */
        PATTERN_POSITION_TABLE: [ [], [ 6, 18 ], [ 6, 22 ], [ 6, 26 ], [ 6, 30 ], [ 6, 34 ], [ 6, 22, 38 ], [ 6, 24, 42 ], [ 6, 26, 46 ], [ 6, 28, 50 ], [ 6, 30, 54 ], [ 6, 32, 58 ], [ 6, 34, 62 ], [ 6, 26, 46, 66 ], [ 6, 26, 48, 70 ], [ 6, 26, 50, 74 ], [ 6, 30, 54, 78 ], [ 6, 30, 56, 82 ], [ 6, 30, 58, 86 ], [ 6, 34, 62, 90 ], [ 6, 28, 50, 72, 94 ], [ 6, 26, 50, 74, 98 ], [ 6, 30, 54, 78, 102 ], [ 6, 28, 54, 80, 106 ], [ 6, 32, 58, 84, 110 ], [ 6, 30, 58, 86, 114 ], [ 6, 34, 62, 90, 118 ], [ 6, 26, 50, 74, 98, 122 ], [ 6, 30, 54, 78, 102, 126 ], [ 6, 26, 52, 78, 104, 130 ], [ 6, 30, 56, 82, 108, 134 ], [ 6, 34, 60, 86, 112, 138 ], [ 6, 30, 58, 86, 114, 142 ], [ 6, 34, 62, 90, 118, 146 ], [ 6, 30, 54, 78, 102, 126, 150 ], [ 6, 24, 50, 76, 102, 128, 154 ], [ 6, 28, 54, 80, 106, 132, 158 ], [ 6, 32, 58, 84, 110, 136, 162 ], [ 6, 26, 54, 82, 110, 138, 166 ], [ 6, 30, 58, 86, 114, 142, 170 ] ],
        G15: 1 << 10 | 1 << 8 | 1 << 5 | 1 << 4 | 1 << 2 | 1 << 1 | 1 << 0,
        G18: 1 << 12 | 1 << 11 | 1 << 10 | 1 << 9 | 1 << 8 | 1 << 5 | 1 << 2 | 1 << 0,
        G15_MASK: 1 << 14 | 1 << 12 | 1 << 10 | 1 << 4 | 1 << 1,
        /*
         BCH编码格式信息
         */
        getBCHTypeInfo: function(data) {
            var d = data << 10;
            while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15) >= 0) {
                d ^= QRUtil.G15 << QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15);
            }
            return (data << 10 | d) ^ QRUtil.G15_MASK;
        },
        /*
         BCH编码版本信息
         */
        getBCHTypeNumber: function(data) {
            var d = data << 12;
            while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18) >= 0) {
                d ^= QRUtil.G18 << QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18);
            }
            return data << 12 | d;
        },
        /*
         获取BCH位信息
         */
        getBCHDigit: function(data) {
            var digit = 0;
            while (data != 0) {
                digit++;
                data >>>= 1;
            }
            return digit;
        },
        /*
         获取版本对应的矫正图形位置
         */
        getPatternPosition: function(typeNumber) {
            return QRUtil.PATTERN_POSITION_TABLE[typeNumber - 1];
        },
        /*
         掩膜算法
         */
        getMask: function(maskPattern, i, j) {
            switch (maskPattern) {
              case QRMaskPattern.PATTERN000:
                return (i + j) % 2 == 0;

              case QRMaskPattern.PATTERN001:
                return i % 2 == 0;

              case QRMaskPattern.PATTERN010:
                return j % 3 == 0;

              case QRMaskPattern.PATTERN011:
                return (i + j) % 3 == 0;

              case QRMaskPattern.PATTERN100:
                return (Math.floor(i / 2) + Math.floor(j / 3)) % 2 == 0;

              case QRMaskPattern.PATTERN101:
                return i * j % 2 + i * j % 3 == 0;

              case QRMaskPattern.PATTERN110:
                return (i * j % 2 + i * j % 3) % 2 == 0;

              case QRMaskPattern.PATTERN111:
                return (i * j % 3 + (i + j) % 2) % 2 == 0;

              default:
                throw new Error("bad maskPattern:" + maskPattern);
            }
        },
        /*
         获取RS的纠错多项式
         */
        getErrorCorrectPolynomial: function(errorCorrectLength) {
            var a = new QRPolynomial([ 1 ], 0);
            for (var i = 0; i < errorCorrectLength; i++) {
                a = a.multiply(new QRPolynomial([ 1, QRMath.gexp(i) ], 0));
            }
            return a;
        },
        /*
         获取评价
         */
        getLostPoint: function(qrCode) {
            var moduleCount = qrCode.getModuleCount(), lostPoint = 0, darkCount = 0;
            for (var row = 0; row < moduleCount; row++) {
                var sameCount = 0;
                var head = qrCode.modules[row][0];
                for (var col = 0; col < moduleCount; col++) {
                    var current = qrCode.modules[row][col];
                    //level 3 评价
                    if (col < moduleCount - 6) {
                        if (current && !qrCode.modules[row][col + 1] && qrCode.modules[row][col + 2] && qrCode.modules[row][col + 3] && qrCode.modules[row][col + 4] && !qrCode.modules[row][col + 5] && qrCode.modules[row][col + 6]) {
                            if (col < moduleCount - 10) {
                                if (qrCode.modules[row][col + 7] && qrCode.modules[row][col + 8] && qrCode.modules[row][col + 9] && qrCode.modules[row][col + 10]) {
                                    lostPoint += 40;
                                }
                            } else if (col > 3) {
                                if (qrCode.modules[row][col - 1] && qrCode.modules[row][col - 2] && qrCode.modules[row][col - 3] && qrCode.modules[row][col - 4]) {
                                    lostPoint += 40;
                                }
                            }
                        }
                    }
                    //level 2 评价
                    if (row < moduleCount - 1 && col < moduleCount - 1) {
                        var count = 0;
                        if (current) count++;
                        if (qrCode.modules[row + 1][col]) count++;
                        if (qrCode.modules[row][col + 1]) count++;
                        if (qrCode.modules[row + 1][col + 1]) count++;
                        if (count == 0 || count == 4) {
                            lostPoint += 3;
                        }
                    }
                    //level 1 评价
                    if (head ^ current) {
                        sameCount++;
                    } else {
                        head = current;
                        if (sameCount >= 5) {
                            lostPoint += 3 + sameCount - 5;
                        }
                        sameCount = 1;
                    }
                    //level 4 评价
                    if (current) {
                        darkCount++;
                    }
                }
            }
            for (var col = 0; col < moduleCount; col++) {
                var sameCount = 0;
                var head = qrCode.modules[0][col];
                for (var row = 0; row < moduleCount; row++) {
                    var current = qrCode.modules[row][col];
                    //level 3 评价
                    if (row < moduleCount - 6) {
                        if (current && !qrCode.modules[row + 1][col] && qrCode.modules[row + 2][col] && qrCode.modules[row + 3][col] && qrCode.modules[row + 4][col] && !qrCode.modules[row + 5][col] && qrCode.modules[row + 6][col]) {
                            if (row < moduleCount - 10) {
                                if (qrCode.modules[row + 7][col] && qrCode.modules[row + 8][col] && qrCode.modules[row + 9][col] && qrCode.modules[row + 10][col]) {
                                    lostPoint += 40;
                                }
                            } else if (row > 3) {
                                if (qrCode.modules[row - 1][col] && qrCode.modules[row - 2][col] && qrCode.modules[row - 3][col] && qrCode.modules[row - 4][col]) {
                                    lostPoint += 40;
                                }
                            }
                        }
                    }
                    //level 1 评价
                    if (head ^ current) {
                        sameCount++;
                    } else {
                        head = current;
                        if (sameCount >= 5) {
                            lostPoint += 3 + sameCount - 5;
                        }
                        sameCount = 1;
                    }
                }
            }
            // LEVEL4
            var ratio = Math.abs(100 * darkCount / moduleCount / moduleCount - 50) / 5;
            lostPoint += ratio * 10;
            return lostPoint;
        }
    };
    //---------------------------------------------------------------------
    // QRMath使用的数学工具
    //---------------------------------------------------------------------
    var QRMath = {
        /*
         将n转化为a^m
         */
        glog: function(n) {
            if (n < 1) {
                throw new Error("glog(" + n + ")");
            }
            return QRMath.LOG_TABLE[n];
        },
        /*
         将a^m转化为n
         */
        gexp: function(n) {
            while (n < 0) {
                n += 255;
            }
            while (n >= 256) {
                n -= 255;
            }
            return QRMath.EXP_TABLE[n];
        },
        EXP_TABLE: new Array(256),
        LOG_TABLE: new Array(256)
    };
    for (var i = 0; i < 8; i++) {
        QRMath.EXP_TABLE[i] = 1 << i;
    }
    for (var i = 8; i < 256; i++) {
        QRMath.EXP_TABLE[i] = QRMath.EXP_TABLE[i - 4] ^ QRMath.EXP_TABLE[i - 5] ^ QRMath.EXP_TABLE[i - 6] ^ QRMath.EXP_TABLE[i - 8];
    }
    for (var i = 0; i < 255; i++) {
        QRMath.LOG_TABLE[QRMath.EXP_TABLE[i]] = i;
    }
    //---------------------------------------------------------------------
    // QRPolynomial 多项式
    //---------------------------------------------------------------------
    /**
     * 多项式类
     * @param {Array} num   系数
     * @param {num} shift a^shift
     */
    function QRPolynomial(num, shift) {
        if (num.length == undefined) {
            throw new Error(num.length + "/" + shift);
        }
        var offset = 0;
        while (offset < num.length && num[offset] == 0) {
            offset++;
        }
        this.num = new Array(num.length - offset + shift);
        for (var i = 0; i < num.length - offset; i++) {
            this.num[i] = num[i + offset];
        }
    }
    QRPolynomial.prototype = {
        get: function(index) {
            return this.num[index];
        },
        getLength: function() {
            return this.num.length;
        },
        /**
         * 多项式乘法
         * @param  {QRPolynomial} e 被乘多项式
         * @return {[type]}   [description]
         */
        multiply: function(e) {
            var num = new Array(this.getLength() + e.getLength() - 1);
            for (var i = 0; i < this.getLength(); i++) {
                for (var j = 0; j < e.getLength(); j++) {
                    num[i + j] ^= QRMath.gexp(QRMath.glog(this.get(i)) + QRMath.glog(e.get(j)));
                }
            }
            return new QRPolynomial(num, 0);
        },
        /**
         * 多项式模运算
         * @param  {QRPolynomial} e 模多项式
         * @return {}
         */
        mod: function(e) {
            var tl = this.getLength(), el = e.getLength();
            if (tl - el < 0) {
                return this;
            }
            var num = new Array(tl);
            for (var i = 0; i < tl; i++) {
                num[i] = this.get(i);
            }
            while (num.length >= el) {
                var ratio = QRMath.glog(num[0]) - QRMath.glog(e.get(0));
                for (var i = 0; i < e.getLength(); i++) {
                    num[i] ^= QRMath.gexp(QRMath.glog(e.get(i)) + ratio);
                }
                while (num[0] == 0) {
                    num.shift();
                }
            }
            return new QRPolynomial(num, 0);
        }
    };
    //---------------------------------------------------------------------
    // RS_BLOCK_TABLE
    //---------------------------------------------------------------------
    /*
     二维码各个版本信息[块数, 每块中的数据块数, 每块中的信息块数]
     */
    RS_BLOCK_TABLE = [ // L
    // M
    // Q
    // H
    // 1
    [ 1, 26, 19 ], [ 1, 26, 16 ], [ 1, 26, 13 ], [ 1, 26, 9 ], // 2
    [ 1, 44, 34 ], [ 1, 44, 28 ], [ 1, 44, 22 ], [ 1, 44, 16 ], // 3
    [ 1, 70, 55 ], [ 1, 70, 44 ], [ 2, 35, 17 ], [ 2, 35, 13 ], // 4
    [ 1, 100, 80 ], [ 2, 50, 32 ], [ 2, 50, 24 ], [ 4, 25, 9 ], // 5
    [ 1, 134, 108 ], [ 2, 67, 43 ], [ 2, 33, 15, 2, 34, 16 ], [ 2, 33, 11, 2, 34, 12 ], // 6
    [ 2, 86, 68 ], [ 4, 43, 27 ], [ 4, 43, 19 ], [ 4, 43, 15 ], // 7
    [ 2, 98, 78 ], [ 4, 49, 31 ], [ 2, 32, 14, 4, 33, 15 ], [ 4, 39, 13, 1, 40, 14 ], // 8
    [ 2, 121, 97 ], [ 2, 60, 38, 2, 61, 39 ], [ 4, 40, 18, 2, 41, 19 ], [ 4, 40, 14, 2, 41, 15 ], // 9
    [ 2, 146, 116 ], [ 3, 58, 36, 2, 59, 37 ], [ 4, 36, 16, 4, 37, 17 ], [ 4, 36, 12, 4, 37, 13 ], // 10
    [ 2, 86, 68, 2, 87, 69 ], [ 4, 69, 43, 1, 70, 44 ], [ 6, 43, 19, 2, 44, 20 ], [ 6, 43, 15, 2, 44, 16 ], // 11
    [ 4, 101, 81 ], [ 1, 80, 50, 4, 81, 51 ], [ 4, 50, 22, 4, 51, 23 ], [ 3, 36, 12, 8, 37, 13 ], // 12
    [ 2, 116, 92, 2, 117, 93 ], [ 6, 58, 36, 2, 59, 37 ], [ 4, 46, 20, 6, 47, 21 ], [ 7, 42, 14, 4, 43, 15 ], // 13
    [ 4, 133, 107 ], [ 8, 59, 37, 1, 60, 38 ], [ 8, 44, 20, 4, 45, 21 ], [ 12, 33, 11, 4, 34, 12 ], // 14
    [ 3, 145, 115, 1, 146, 116 ], [ 4, 64, 40, 5, 65, 41 ], [ 11, 36, 16, 5, 37, 17 ], [ 11, 36, 12, 5, 37, 13 ], // 15
    [ 5, 109, 87, 1, 110, 88 ], [ 5, 65, 41, 5, 66, 42 ], [ 5, 54, 24, 7, 55, 25 ], [ 11, 36, 12 ], // 16
    [ 5, 122, 98, 1, 123, 99 ], [ 7, 73, 45, 3, 74, 46 ], [ 15, 43, 19, 2, 44, 20 ], [ 3, 45, 15, 13, 46, 16 ], // 17
    [ 1, 135, 107, 5, 136, 108 ], [ 10, 74, 46, 1, 75, 47 ], [ 1, 50, 22, 15, 51, 23 ], [ 2, 42, 14, 17, 43, 15 ], // 18
    [ 5, 150, 120, 1, 151, 121 ], [ 9, 69, 43, 4, 70, 44 ], [ 17, 50, 22, 1, 51, 23 ], [ 2, 42, 14, 19, 43, 15 ], // 19
    [ 3, 141, 113, 4, 142, 114 ], [ 3, 70, 44, 11, 71, 45 ], [ 17, 47, 21, 4, 48, 22 ], [ 9, 39, 13, 16, 40, 14 ], // 20
    [ 3, 135, 107, 5, 136, 108 ], [ 3, 67, 41, 13, 68, 42 ], [ 15, 54, 24, 5, 55, 25 ], [ 15, 43, 15, 10, 44, 16 ], // 21
    [ 4, 144, 116, 4, 145, 117 ], [ 17, 68, 42 ], [ 17, 50, 22, 6, 51, 23 ], [ 19, 46, 16, 6, 47, 17 ], // 22
    [ 2, 139, 111, 7, 140, 112 ], [ 17, 74, 46 ], [ 7, 54, 24, 16, 55, 25 ], [ 34, 37, 13 ], // 23
    [ 4, 151, 121, 5, 152, 122 ], [ 4, 75, 47, 14, 76, 48 ], [ 11, 54, 24, 14, 55, 25 ], [ 16, 45, 15, 14, 46, 16 ], // 24
    [ 6, 147, 117, 4, 148, 118 ], [ 6, 73, 45, 14, 74, 46 ], [ 11, 54, 24, 16, 55, 25 ], [ 30, 46, 16, 2, 47, 17 ], // 25
    [ 8, 132, 106, 4, 133, 107 ], [ 8, 75, 47, 13, 76, 48 ], [ 7, 54, 24, 22, 55, 25 ], [ 22, 45, 15, 13, 46, 16 ], // 26
    [ 10, 142, 114, 2, 143, 115 ], [ 19, 74, 46, 4, 75, 47 ], [ 28, 50, 22, 6, 51, 23 ], [ 33, 46, 16, 4, 47, 17 ], // 27
    [ 8, 152, 122, 4, 153, 123 ], [ 22, 73, 45, 3, 74, 46 ], [ 8, 53, 23, 26, 54, 24 ], [ 12, 45, 15, 28, 46, 16 ], // 28
    [ 3, 147, 117, 10, 148, 118 ], [ 3, 73, 45, 23, 74, 46 ], [ 4, 54, 24, 31, 55, 25 ], [ 11, 45, 15, 31, 46, 16 ], // 29
    [ 7, 146, 116, 7, 147, 117 ], [ 21, 73, 45, 7, 74, 46 ], [ 1, 53, 23, 37, 54, 24 ], [ 19, 45, 15, 26, 46, 16 ], // 30
    [ 5, 145, 115, 10, 146, 116 ], [ 19, 75, 47, 10, 76, 48 ], [ 15, 54, 24, 25, 55, 25 ], [ 23, 45, 15, 25, 46, 16 ], // 31
    [ 13, 145, 115, 3, 146, 116 ], [ 2, 74, 46, 29, 75, 47 ], [ 42, 54, 24, 1, 55, 25 ], [ 23, 45, 15, 28, 46, 16 ], // 32
    [ 17, 145, 115 ], [ 10, 74, 46, 23, 75, 47 ], [ 10, 54, 24, 35, 55, 25 ], [ 19, 45, 15, 35, 46, 16 ], // 33
    [ 17, 145, 115, 1, 146, 116 ], [ 14, 74, 46, 21, 75, 47 ], [ 29, 54, 24, 19, 55, 25 ], [ 11, 45, 15, 46, 46, 16 ], // 34
    [ 13, 145, 115, 6, 146, 116 ], [ 14, 74, 46, 23, 75, 47 ], [ 44, 54, 24, 7, 55, 25 ], [ 59, 46, 16, 1, 47, 17 ], // 35
    [ 12, 151, 121, 7, 152, 122 ], [ 12, 75, 47, 26, 76, 48 ], [ 39, 54, 24, 14, 55, 25 ], [ 22, 45, 15, 41, 46, 16 ], // 36
    [ 6, 151, 121, 14, 152, 122 ], [ 6, 75, 47, 34, 76, 48 ], [ 46, 54, 24, 10, 55, 25 ], [ 2, 45, 15, 64, 46, 16 ], // 37
    [ 17, 152, 122, 4, 153, 123 ], [ 29, 74, 46, 14, 75, 47 ], [ 49, 54, 24, 10, 55, 25 ], [ 24, 45, 15, 46, 46, 16 ], // 38
    [ 4, 152, 122, 18, 153, 123 ], [ 13, 74, 46, 32, 75, 47 ], [ 48, 54, 24, 14, 55, 25 ], [ 42, 45, 15, 32, 46, 16 ], // 39
    [ 20, 147, 117, 4, 148, 118 ], [ 40, 75, 47, 7, 76, 48 ], [ 43, 54, 24, 22, 55, 25 ], [ 10, 45, 15, 67, 46, 16 ], // 40
    [ 19, 148, 118, 6, 149, 119 ], [ 18, 75, 47, 31, 76, 48 ], [ 34, 54, 24, 34, 55, 25 ], [ 20, 45, 15, 61, 46, 16 ] ];
    /**
     * 根据数据获取对应版本
     * @return {[type]} [description]
     */
    QRCodeAlg.prototype.getRightType = function() {
        for (var typeNumber = 1; typeNumber < 41; typeNumber++) {
            var rsBlock = RS_BLOCK_TABLE[(typeNumber - 1) * 4 + this.errorCorrectLevel];
            if (rsBlock == undefined) {
                throw new Error("bad rs block @ typeNumber:" + typeNumber + "/errorCorrectLevel:" + this.errorCorrectLevel);
            }
            var length = rsBlock.length / 3;
            var totalDataCount = 0;
            for (var i = 0; i < length; i++) {
                var count = rsBlock[i * 3 + 0];
                var dataCount = rsBlock[i * 3 + 2];
                totalDataCount += dataCount * count;
            }
            var lengthBytes = typeNumber > 9 ? 2 : 1;
            if (this.utf8bytes.length + lengthBytes < totalDataCount || typeNumber == 40) {
                this.typeNumber = typeNumber;
                this.rsBlock = rsBlock;
                this.totalDataCount = totalDataCount;
                break;
            }
        }
    };
    //---------------------------------------------------------------------
    // QRBitBuffer
    //---------------------------------------------------------------------
    function QRBitBuffer() {
        this.buffer = new Array();
        this.length = 0;
    }
    QRBitBuffer.prototype = {
        get: function(index) {
            var bufIndex = Math.floor(index / 8);
            return this.buffer[bufIndex] >>> 7 - index % 8 & 1;
        },
        put: function(num, length) {
            for (var i = 0; i < length; i++) {
                this.putBit(num >>> length - i - 1 & 1);
            }
        },
        putBit: function(bit) {
            var bufIndex = Math.floor(this.length / 8);
            if (this.buffer.length <= bufIndex) {
                this.buffer.push(0);
            }
            if (bit) {
                this.buffer[bufIndex] |= 128 >>> this.length % 8;
            }
            this.length++;
        }
    };
    /**
     * 获取单个字符的utf8编码
     * unicode BMP平面约65535个字符
     * @param {num} code
     * return {array}
     */
    function unicodeFormat8(code) {
        // 1 byte
        if (code < 128) {
            return [ code ];
        } else if (code < 2048) {
            c0 = 192 + (code >> 6);
            c1 = 128 + (code & 63);
            return [ c0, c1 ];
        } else {
            c0 = 224 + (code >> 12);
            c1 = 128 + (code >> 6 & 63);
            c2 = 128 + (code & 63);
            return [ c0, c1, c2 ];
        }
    }
    /**
     * 获取字符串的utf8编码字节串
     * @param {string} string
     * @return {array}
     */
    function getUTF8Bytes(string) {
        var utf8codes = [];
        for (var i = 0; i < string.length; i++) {
            var code = string.charCodeAt(i);
            var utf8 = unicodeFormat8(code);
            for (var j = 0; j < utf8.length; j++) {
                utf8codes.push(utf8[j]);
            }
        }
        return utf8codes;
    }
    /**
     * 二维码算法实现
     * @param {string} data              要编码的信息字符串
     * @param {num} errorCorrectLevel 纠错等级
     */
    function QRCodeAlg(data, errorCorrectLevel) {
        this.typeNumber = -1;
        //版本
        this.errorCorrectLevel = errorCorrectLevel;
        this.modules = null;
        //二维矩阵，存放最终结果
        this.moduleCount = 0;
        //矩阵大小
        this.dataCache = null;
        //数据缓存
        this.rsBlocks = null;
        //版本数据信息
        this.totalDataCount = -1;
        //可使用的数据量
        this.data = data;
        this.utf8bytes = getUTF8Bytes(data);
        this.make();
    }
    QRCodeAlg.prototype = {
        constructor: QRCodeAlg,
        /**
         * 获取二维码矩阵大小
         * @return {num} 矩阵大小
         */
        getModuleCount: function() {
            return this.moduleCount;
        },
        /**
         * 编码
         */
        make: function() {
            this.getRightType();
            this.dataCache = this.createData();
            this.createQrcode();
        },
        /**
         * 设置二位矩阵功能图形
         * @param  {bool} test 表示是否在寻找最好掩膜阶段
         * @param  {num} maskPattern 掩膜的版本
         */
        makeImpl: function(maskPattern) {
            this.moduleCount = this.typeNumber * 4 + 17;
            this.modules = new Array(this.moduleCount);
            for (var row = 0; row < this.moduleCount; row++) {
                this.modules[row] = new Array(this.moduleCount);
            }
            this.setupPositionProbePattern(0, 0);
            this.setupPositionProbePattern(this.moduleCount - 7, 0);
            this.setupPositionProbePattern(0, this.moduleCount - 7);
            this.setupPositionAdjustPattern();
            this.setupTimingPattern();
            this.setupTypeInfo(true, maskPattern);
            if (this.typeNumber >= 7) {
                this.setupTypeNumber(true);
            }
            this.mapData(this.dataCache, maskPattern);
        },
        /**
         * 设置二维码的位置探测图形
         * @param  {num} row 探测图形的中心横坐标
         * @param  {num} col 探测图形的中心纵坐标
         */
        setupPositionProbePattern: function(row, col) {
            for (var r = -1; r <= 7; r++) {
                if (row + r <= -1 || this.moduleCount <= row + r) continue;
                for (var c = -1; c <= 7; c++) {
                    if (col + c <= -1 || this.moduleCount <= col + c) continue;
                    if (0 <= r && r <= 6 && (c == 0 || c == 6) || 0 <= c && c <= 6 && (r == 0 || r == 6) || 2 <= r && r <= 4 && 2 <= c && c <= 4) {
                        this.modules[row + r][col + c] = true;
                    } else {
                        this.modules[row + r][col + c] = false;
                    }
                }
            }
        },
        /**
         * 创建二维码
         * @return {[type]} [description]
         */
        createQrcode: function() {
            var minLostPoint = 0;
            var pattern = 0;
            var bestModules = null;
            for (var i = 0; i < 8; i++) {
                this.makeImpl(i);
                var lostPoint = QRUtil.getLostPoint(this);
                if (i == 0 || minLostPoint > lostPoint) {
                    minLostPoint = lostPoint;
                    pattern = i;
                    bestModules = this.modules;
                }
            }
            this.modules = bestModules;
            this.setupTypeInfo(false, pattern);
            if (this.typeNumber >= 7) {
                this.setupTypeNumber(false);
            }
        },
        /**
         * 设置定位图形
         * @return {[type]} [description]
         */
        setupTimingPattern: function() {
            for (var r = 8; r < this.moduleCount - 8; r++) {
                if (this.modules[r][6] != null) {
                    continue;
                }
                this.modules[r][6] = r % 2 == 0;
                if (this.modules[6][r] != null) {
                    continue;
                }
                this.modules[6][r] = r % 2 == 0;
            }
        },
        /**
         * 设置矫正图形
         * @return {[type]} [description]
         */
        setupPositionAdjustPattern: function() {
            var pos = QRUtil.getPatternPosition(this.typeNumber);
            for (var i = 0; i < pos.length; i++) {
                for (var j = 0; j < pos.length; j++) {
                    var row = pos[i];
                    var col = pos[j];
                    if (this.modules[row][col] != null) {
                        continue;
                    }
                    for (var r = -2; r <= 2; r++) {
                        for (var c = -2; c <= 2; c++) {
                            if (r == -2 || r == 2 || c == -2 || c == 2 || r == 0 && c == 0) {
                                this.modules[row + r][col + c] = true;
                            } else {
                                this.modules[row + r][col + c] = false;
                            }
                        }
                    }
                }
            }
        },
        /**
         * 设置版本信息（7以上版本才有）
         * @param  {bool} test 是否处于判断最佳掩膜阶段
         * @return {[type]}      [description]
         */
        setupTypeNumber: function(test) {
            var bits = QRUtil.getBCHTypeNumber(this.typeNumber);
            for (var i = 0; i < 18; i++) {
                var mod = !test && (bits >> i & 1) == 1;
                this.modules[Math.floor(i / 3)][i % 3 + this.moduleCount - 8 - 3] = mod;
                this.modules[i % 3 + this.moduleCount - 8 - 3][Math.floor(i / 3)] = mod;
            }
        },
        /**
         * 设置格式信息（纠错等级和掩膜版本）
         * @param  {bool} test
         * @param  {num} maskPattern 掩膜版本
         * @return {}
         */
        setupTypeInfo: function(test, maskPattern) {
            var data = QRErrorCorrectLevel[this.errorCorrectLevel] << 3 | maskPattern;
            var bits = QRUtil.getBCHTypeInfo(data);
            // vertical
            for (var i = 0; i < 15; i++) {
                var mod = !test && (bits >> i & 1) == 1;
                if (i < 6) {
                    this.modules[i][8] = mod;
                } else if (i < 8) {
                    this.modules[i + 1][8] = mod;
                } else {
                    this.modules[this.moduleCount - 15 + i][8] = mod;
                }
                // horizontal
                var mod = !test && (bits >> i & 1) == 1;
                if (i < 8) {
                    this.modules[8][this.moduleCount - i - 1] = mod;
                } else if (i < 9) {
                    this.modules[8][15 - i - 1 + 1] = mod;
                } else {
                    this.modules[8][15 - i - 1] = mod;
                }
            }
            // fixed module
            this.modules[this.moduleCount - 8][8] = !test;
        },
        /**
         * 数据编码
         * @return {[type]} [description]
         */
        createData: function() {
            var buffer = new QRBitBuffer();
            var lengthBits = this.typeNumber > 9 ? 16 : 8;
            buffer.put(4, 4);
            //添加模式
            buffer.put(this.utf8bytes.length, lengthBits);
            for (var i = 0, l = this.utf8bytes.length; i < l; i++) {
                buffer.put(this.utf8bytes[i], 8);
            }
            if (buffer.length + 4 <= this.totalDataCount * 8) {
                buffer.put(0, 4);
            }
            // padding
            while (buffer.length % 8 != 0) {
                buffer.putBit(false);
            }
            // padding
            while (true) {
                if (buffer.length >= this.totalDataCount * 8) {
                    break;
                }
                buffer.put(QRCodeAlg.PAD0, 8);
                if (buffer.length >= this.totalDataCount * 8) {
                    break;
                }
                buffer.put(QRCodeAlg.PAD1, 8);
            }
            return this.createBytes(buffer);
        },
        /**
         * 纠错码编码
         * @param  {buffer} buffer 数据编码
         * @return {[type]}
         */
        createBytes: function(buffer) {
            var offset = 0;
            var maxDcCount = 0;
            var maxEcCount = 0;
            var length = this.rsBlock.length / 3;
            var rsBlocks = new Array();
            for (var i = 0; i < length; i++) {
                var count = this.rsBlock[i * 3 + 0];
                var totalCount = this.rsBlock[i * 3 + 1];
                var dataCount = this.rsBlock[i * 3 + 2];
                for (var j = 0; j < count; j++) {
                    rsBlocks.push([ dataCount, totalCount ]);
                }
            }
            var dcdata = new Array(rsBlocks.length);
            var ecdata = new Array(rsBlocks.length);
            for (var r = 0; r < rsBlocks.length; r++) {
                var dcCount = rsBlocks[r][0];
                var ecCount = rsBlocks[r][1] - dcCount;
                maxDcCount = Math.max(maxDcCount, dcCount);
                maxEcCount = Math.max(maxEcCount, ecCount);
                dcdata[r] = new Array(dcCount);
                for (var i = 0; i < dcdata[r].length; i++) {
                    dcdata[r][i] = 255 & buffer.buffer[i + offset];
                }
                offset += dcCount;
                var rsPoly = QRUtil.getErrorCorrectPolynomial(ecCount);
                var rawPoly = new QRPolynomial(dcdata[r], rsPoly.getLength() - 1);
                var modPoly = rawPoly.mod(rsPoly);
                ecdata[r] = new Array(rsPoly.getLength() - 1);
                for (var i = 0; i < ecdata[r].length; i++) {
                    var modIndex = i + modPoly.getLength() - ecdata[r].length;
                    ecdata[r][i] = modIndex >= 0 ? modPoly.get(modIndex) : 0;
                }
            }
            var data = new Array(this.totalDataCount);
            var index = 0;
            for (var i = 0; i < maxDcCount; i++) {
                for (var r = 0; r < rsBlocks.length; r++) {
                    if (i < dcdata[r].length) {
                        data[index++] = dcdata[r][i];
                    }
                }
            }
            for (var i = 0; i < maxEcCount; i++) {
                for (var r = 0; r < rsBlocks.length; r++) {
                    if (i < ecdata[r].length) {
                        data[index++] = ecdata[r][i];
                    }
                }
            }
            return data;
        },
        /**
         * 布置模块，构建最终信息
         * @param  {} data
         * @param  {} maskPattern
         * @return {}
         */
        mapData: function(data, maskPattern) {
            var inc = -1;
            var row = this.moduleCount - 1;
            var bitIndex = 7;
            var byteIndex = 0;
            for (var col = this.moduleCount - 1; col > 0; col -= 2) {
                if (col == 6) col--;
                while (true) {
                    for (var c = 0; c < 2; c++) {
                        if (this.modules[row][col - c] == null) {
                            var dark = false;
                            if (byteIndex < data.length) {
                                dark = (data[byteIndex] >>> bitIndex & 1) == 1;
                            }
                            var mask = QRUtil.getMask(maskPattern, row, col - c);
                            if (mask) {
                                dark = !dark;
                            }
                            this.modules[row][col - c] = dark;
                            bitIndex--;
                            if (bitIndex == -1) {
                                byteIndex++;
                                bitIndex = 7;
                            }
                        }
                    }
                    row += inc;
                    if (row < 0 || this.moduleCount <= row) {
                        row -= inc;
                        inc = -inc;
                        break;
                    }
                }
            }
        }
    };
    /**
     * 填充字段
     */
    QRCodeAlg.PAD0 = 236;
    QRCodeAlg.PAD1 = 17;
    //---------------------------------------------------------------------
    // 纠错等级对应的编码
    //---------------------------------------------------------------------
    var QRErrorCorrectLevel = [ 1, 0, 3, 2 ];
    //---------------------------------------------------------------------
    // 掩膜版本
    //---------------------------------------------------------------------
    var QRMaskPattern = {
        PATTERN000: 0,
        PATTERN001: 1,
        PATTERN010: 2,
        PATTERN011: 3,
        PATTERN100: 4,
        PATTERN101: 5,
        PATTERN110: 6,
        PATTERN111: 7
    };
    //---------------------------------------------------------------------
    // 工具类
    //---------------------------------------------------------------------
    var QRUtil = {
        /*
         每个版本矫正图形的位置
         */
        PATTERN_POSITION_TABLE: [ [], [ 6, 18 ], [ 6, 22 ], [ 6, 26 ], [ 6, 30 ], [ 6, 34 ], [ 6, 22, 38 ], [ 6, 24, 42 ], [ 6, 26, 46 ], [ 6, 28, 50 ], [ 6, 30, 54 ], [ 6, 32, 58 ], [ 6, 34, 62 ], [ 6, 26, 46, 66 ], [ 6, 26, 48, 70 ], [ 6, 26, 50, 74 ], [ 6, 30, 54, 78 ], [ 6, 30, 56, 82 ], [ 6, 30, 58, 86 ], [ 6, 34, 62, 90 ], [ 6, 28, 50, 72, 94 ], [ 6, 26, 50, 74, 98 ], [ 6, 30, 54, 78, 102 ], [ 6, 28, 54, 80, 106 ], [ 6, 32, 58, 84, 110 ], [ 6, 30, 58, 86, 114 ], [ 6, 34, 62, 90, 118 ], [ 6, 26, 50, 74, 98, 122 ], [ 6, 30, 54, 78, 102, 126 ], [ 6, 26, 52, 78, 104, 130 ], [ 6, 30, 56, 82, 108, 134 ], [ 6, 34, 60, 86, 112, 138 ], [ 6, 30, 58, 86, 114, 142 ], [ 6, 34, 62, 90, 118, 146 ], [ 6, 30, 54, 78, 102, 126, 150 ], [ 6, 24, 50, 76, 102, 128, 154 ], [ 6, 28, 54, 80, 106, 132, 158 ], [ 6, 32, 58, 84, 110, 136, 162 ], [ 6, 26, 54, 82, 110, 138, 166 ], [ 6, 30, 58, 86, 114, 142, 170 ] ],
        G15: 1 << 10 | 1 << 8 | 1 << 5 | 1 << 4 | 1 << 2 | 1 << 1 | 1 << 0,
        G18: 1 << 12 | 1 << 11 | 1 << 10 | 1 << 9 | 1 << 8 | 1 << 5 | 1 << 2 | 1 << 0,
        G15_MASK: 1 << 14 | 1 << 12 | 1 << 10 | 1 << 4 | 1 << 1,
        /*
         BCH编码格式信息
         */
        getBCHTypeInfo: function(data) {
            var d = data << 10;
            while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15) >= 0) {
                d ^= QRUtil.G15 << QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G15);
            }
            return (data << 10 | d) ^ QRUtil.G15_MASK;
        },
        /*
         BCH编码版本信息
         */
        getBCHTypeNumber: function(data) {
            var d = data << 12;
            while (QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18) >= 0) {
                d ^= QRUtil.G18 << QRUtil.getBCHDigit(d) - QRUtil.getBCHDigit(QRUtil.G18);
            }
            return data << 12 | d;
        },
        /*
         获取BCH位信息
         */
        getBCHDigit: function(data) {
            var digit = 0;
            while (data != 0) {
                digit++;
                data >>>= 1;
            }
            return digit;
        },
        /*
         获取版本对应的矫正图形位置
         */
        getPatternPosition: function(typeNumber) {
            return QRUtil.PATTERN_POSITION_TABLE[typeNumber - 1];
        },
        /*
         掩膜算法
         */
        getMask: function(maskPattern, i, j) {
            switch (maskPattern) {
              case QRMaskPattern.PATTERN000:
                return (i + j) % 2 == 0;

              case QRMaskPattern.PATTERN001:
                return i % 2 == 0;

              case QRMaskPattern.PATTERN010:
                return j % 3 == 0;

              case QRMaskPattern.PATTERN011:
                return (i + j) % 3 == 0;

              case QRMaskPattern.PATTERN100:
                return (Math.floor(i / 2) + Math.floor(j / 3)) % 2 == 0;

              case QRMaskPattern.PATTERN101:
                return i * j % 2 + i * j % 3 == 0;

              case QRMaskPattern.PATTERN110:
                return (i * j % 2 + i * j % 3) % 2 == 0;

              case QRMaskPattern.PATTERN111:
                return (i * j % 3 + (i + j) % 2) % 2 == 0;

              default:
                throw new Error("bad maskPattern:" + maskPattern);
            }
        },
        /*
         获取RS的纠错多项式
         */
        getErrorCorrectPolynomial: function(errorCorrectLength) {
            var a = new QRPolynomial([ 1 ], 0);
            for (var i = 0; i < errorCorrectLength; i++) {
                a = a.multiply(new QRPolynomial([ 1, QRMath.gexp(i) ], 0));
            }
            return a;
        },
        /*
         获取评价
         */
        getLostPoint: function(qrCode) {
            var moduleCount = qrCode.getModuleCount(), lostPoint = 0, darkCount = 0;
            for (var row = 0; row < moduleCount; row++) {
                var sameCount = 0;
                var head = qrCode.modules[row][0];
                for (var col = 0; col < moduleCount; col++) {
                    var current = qrCode.modules[row][col];
                    //level 3 评价
                    if (col < moduleCount - 6) {
                        if (current && !qrCode.modules[row][col + 1] && qrCode.modules[row][col + 2] && qrCode.modules[row][col + 3] && qrCode.modules[row][col + 4] && !qrCode.modules[row][col + 5] && qrCode.modules[row][col + 6]) {
                            if (col < moduleCount - 10) {
                                if (qrCode.modules[row][col + 7] && qrCode.modules[row][col + 8] && qrCode.modules[row][col + 9] && qrCode.modules[row][col + 10]) {
                                    lostPoint += 40;
                                }
                            } else if (col > 3) {
                                if (qrCode.modules[row][col - 1] && qrCode.modules[row][col - 2] && qrCode.modules[row][col - 3] && qrCode.modules[row][col - 4]) {
                                    lostPoint += 40;
                                }
                            }
                        }
                    }
                    //level 2 评价
                    if (row < moduleCount - 1 && col < moduleCount - 1) {
                        var count = 0;
                        if (current) count++;
                        if (qrCode.modules[row + 1][col]) count++;
                        if (qrCode.modules[row][col + 1]) count++;
                        if (qrCode.modules[row + 1][col + 1]) count++;
                        if (count == 0 || count == 4) {
                            lostPoint += 3;
                        }
                    }
                    //level 1 评价
                    if (head ^ current) {
                        sameCount++;
                    } else {
                        head = current;
                        if (sameCount >= 5) {
                            lostPoint += 3 + sameCount - 5;
                        }
                        sameCount = 1;
                    }
                    //level 4 评价
                    if (current) {
                        darkCount++;
                    }
                }
            }
            for (var col = 0; col < moduleCount; col++) {
                var sameCount = 0;
                var head = qrCode.modules[0][col];
                for (var row = 0; row < moduleCount; row++) {
                    var current = qrCode.modules[row][col];
                    //level 3 评价
                    if (row < moduleCount - 6) {
                        if (current && !qrCode.modules[row + 1][col] && qrCode.modules[row + 2][col] && qrCode.modules[row + 3][col] && qrCode.modules[row + 4][col] && !qrCode.modules[row + 5][col] && qrCode.modules[row + 6][col]) {
                            if (row < moduleCount - 10) {
                                if (qrCode.modules[row + 7][col] && qrCode.modules[row + 8][col] && qrCode.modules[row + 9][col] && qrCode.modules[row + 10][col]) {
                                    lostPoint += 40;
                                }
                            } else if (row > 3) {
                                if (qrCode.modules[row - 1][col] && qrCode.modules[row - 2][col] && qrCode.modules[row - 3][col] && qrCode.modules[row - 4][col]) {
                                    lostPoint += 40;
                                }
                            }
                        }
                    }
                    //level 1 评价
                    if (head ^ current) {
                        sameCount++;
                    } else {
                        head = current;
                        if (sameCount >= 5) {
                            lostPoint += 3 + sameCount - 5;
                        }
                        sameCount = 1;
                    }
                }
            }
            // LEVEL4
            var ratio = Math.abs(100 * darkCount / moduleCount / moduleCount - 50) / 5;
            lostPoint += ratio * 10;
            return lostPoint;
        }
    };
    //---------------------------------------------------------------------
    // QRMath使用的数学工具
    //---------------------------------------------------------------------
    var QRMath = {
        /*
         将n转化为a^m
         */
        glog: function(n) {
            if (n < 1) {
                throw new Error("glog(" + n + ")");
            }
            return QRMath.LOG_TABLE[n];
        },
        /*
         将a^m转化为n
         */
        gexp: function(n) {
            while (n < 0) {
                n += 255;
            }
            while (n >= 256) {
                n -= 255;
            }
            return QRMath.EXP_TABLE[n];
        },
        EXP_TABLE: new Array(256),
        LOG_TABLE: new Array(256)
    };
    for (var i = 0; i < 8; i++) {
        QRMath.EXP_TABLE[i] = 1 << i;
    }
    for (var i = 8; i < 256; i++) {
        QRMath.EXP_TABLE[i] = QRMath.EXP_TABLE[i - 4] ^ QRMath.EXP_TABLE[i - 5] ^ QRMath.EXP_TABLE[i - 6] ^ QRMath.EXP_TABLE[i - 8];
    }
    for (var i = 0; i < 255; i++) {
        QRMath.LOG_TABLE[QRMath.EXP_TABLE[i]] = i;
    }
    //---------------------------------------------------------------------
    // QRPolynomial 多项式
    //---------------------------------------------------------------------
    /**
     * 多项式类
     * @param {Array} num   系数
     * @param {num} shift a^shift
     */
    function QRPolynomial(num, shift) {
        if (num.length == undefined) {
            throw new Error(num.length + "/" + shift);
        }
        var offset = 0;
        while (offset < num.length && num[offset] == 0) {
            offset++;
        }
        this.num = new Array(num.length - offset + shift);
        for (var i = 0; i < num.length - offset; i++) {
            this.num[i] = num[i + offset];
        }
    }
    QRPolynomial.prototype = {
        get: function(index) {
            return this.num[index];
        },
        getLength: function() {
            return this.num.length;
        },
        /**
         * 多项式乘法
         * @param  {QRPolynomial} e 被乘多项式
         * @return {[type]}   [description]
         */
        multiply: function(e) {
            var num = new Array(this.getLength() + e.getLength() - 1);
            for (var i = 0; i < this.getLength(); i++) {
                for (var j = 0; j < e.getLength(); j++) {
                    num[i + j] ^= QRMath.gexp(QRMath.glog(this.get(i)) + QRMath.glog(e.get(j)));
                }
            }
            return new QRPolynomial(num, 0);
        },
        /**
         * 多项式模运算
         * @param  {QRPolynomial} e 模多项式
         * @return {}
         */
        mod: function(e) {
            var tl = this.getLength(), el = e.getLength();
            if (tl - el < 0) {
                return this;
            }
            var num = new Array(tl);
            for (var i = 0; i < tl; i++) {
                num[i] = this.get(i);
            }
            while (num.length >= el) {
                var ratio = QRMath.glog(num[0]) - QRMath.glog(e.get(0));
                for (var i = 0; i < e.getLength(); i++) {
                    num[i] ^= QRMath.gexp(QRMath.glog(e.get(i)) + ratio);
                }
                while (num[0] == 0) {
                    num.shift();
                }
            }
            return new QRPolynomial(num, 0);
        }
    };
    //---------------------------------------------------------------------
    // RS_BLOCK_TABLE
    //---------------------------------------------------------------------
    /*
     二维码各个版本信息[块数, 每块中的数据块数, 每块中的信息块数]
     */
    RS_BLOCK_TABLE = [ // L
    // M
    // Q
    // H
    // 1
    [ 1, 26, 19 ], [ 1, 26, 16 ], [ 1, 26, 13 ], [ 1, 26, 9 ], // 2
    [ 1, 44, 34 ], [ 1, 44, 28 ], [ 1, 44, 22 ], [ 1, 44, 16 ], // 3
    [ 1, 70, 55 ], [ 1, 70, 44 ], [ 2, 35, 17 ], [ 2, 35, 13 ], // 4
    [ 1, 100, 80 ], [ 2, 50, 32 ], [ 2, 50, 24 ], [ 4, 25, 9 ], // 5
    [ 1, 134, 108 ], [ 2, 67, 43 ], [ 2, 33, 15, 2, 34, 16 ], [ 2, 33, 11, 2, 34, 12 ], // 6
    [ 2, 86, 68 ], [ 4, 43, 27 ], [ 4, 43, 19 ], [ 4, 43, 15 ], // 7
    [ 2, 98, 78 ], [ 4, 49, 31 ], [ 2, 32, 14, 4, 33, 15 ], [ 4, 39, 13, 1, 40, 14 ], // 8
    [ 2, 121, 97 ], [ 2, 60, 38, 2, 61, 39 ], [ 4, 40, 18, 2, 41, 19 ], [ 4, 40, 14, 2, 41, 15 ], // 9
    [ 2, 146, 116 ], [ 3, 58, 36, 2, 59, 37 ], [ 4, 36, 16, 4, 37, 17 ], [ 4, 36, 12, 4, 37, 13 ], // 10
    [ 2, 86, 68, 2, 87, 69 ], [ 4, 69, 43, 1, 70, 44 ], [ 6, 43, 19, 2, 44, 20 ], [ 6, 43, 15, 2, 44, 16 ], // 11
    [ 4, 101, 81 ], [ 1, 80, 50, 4, 81, 51 ], [ 4, 50, 22, 4, 51, 23 ], [ 3, 36, 12, 8, 37, 13 ], // 12
    [ 2, 116, 92, 2, 117, 93 ], [ 6, 58, 36, 2, 59, 37 ], [ 4, 46, 20, 6, 47, 21 ], [ 7, 42, 14, 4, 43, 15 ], // 13
    [ 4, 133, 107 ], [ 8, 59, 37, 1, 60, 38 ], [ 8, 44, 20, 4, 45, 21 ], [ 12, 33, 11, 4, 34, 12 ], // 14
    [ 3, 145, 115, 1, 146, 116 ], [ 4, 64, 40, 5, 65, 41 ], [ 11, 36, 16, 5, 37, 17 ], [ 11, 36, 12, 5, 37, 13 ], // 15
    [ 5, 109, 87, 1, 110, 88 ], [ 5, 65, 41, 5, 66, 42 ], [ 5, 54, 24, 7, 55, 25 ], [ 11, 36, 12 ], // 16
    [ 5, 122, 98, 1, 123, 99 ], [ 7, 73, 45, 3, 74, 46 ], [ 15, 43, 19, 2, 44, 20 ], [ 3, 45, 15, 13, 46, 16 ], // 17
    [ 1, 135, 107, 5, 136, 108 ], [ 10, 74, 46, 1, 75, 47 ], [ 1, 50, 22, 15, 51, 23 ], [ 2, 42, 14, 17, 43, 15 ], // 18
    [ 5, 150, 120, 1, 151, 121 ], [ 9, 69, 43, 4, 70, 44 ], [ 17, 50, 22, 1, 51, 23 ], [ 2, 42, 14, 19, 43, 15 ], // 19
    [ 3, 141, 113, 4, 142, 114 ], [ 3, 70, 44, 11, 71, 45 ], [ 17, 47, 21, 4, 48, 22 ], [ 9, 39, 13, 16, 40, 14 ], // 20
    [ 3, 135, 107, 5, 136, 108 ], [ 3, 67, 41, 13, 68, 42 ], [ 15, 54, 24, 5, 55, 25 ], [ 15, 43, 15, 10, 44, 16 ], // 21
    [ 4, 144, 116, 4, 145, 117 ], [ 17, 68, 42 ], [ 17, 50, 22, 6, 51, 23 ], [ 19, 46, 16, 6, 47, 17 ], // 22
    [ 2, 139, 111, 7, 140, 112 ], [ 17, 74, 46 ], [ 7, 54, 24, 16, 55, 25 ], [ 34, 37, 13 ], // 23
    [ 4, 151, 121, 5, 152, 122 ], [ 4, 75, 47, 14, 76, 48 ], [ 11, 54, 24, 14, 55, 25 ], [ 16, 45, 15, 14, 46, 16 ], // 24
    [ 6, 147, 117, 4, 148, 118 ], [ 6, 73, 45, 14, 74, 46 ], [ 11, 54, 24, 16, 55, 25 ], [ 30, 46, 16, 2, 47, 17 ], // 25
    [ 8, 132, 106, 4, 133, 107 ], [ 8, 75, 47, 13, 76, 48 ], [ 7, 54, 24, 22, 55, 25 ], [ 22, 45, 15, 13, 46, 16 ], // 26
    [ 10, 142, 114, 2, 143, 115 ], [ 19, 74, 46, 4, 75, 47 ], [ 28, 50, 22, 6, 51, 23 ], [ 33, 46, 16, 4, 47, 17 ], // 27
    [ 8, 152, 122, 4, 153, 123 ], [ 22, 73, 45, 3, 74, 46 ], [ 8, 53, 23, 26, 54, 24 ], [ 12, 45, 15, 28, 46, 16 ], // 28
    [ 3, 147, 117, 10, 148, 118 ], [ 3, 73, 45, 23, 74, 46 ], [ 4, 54, 24, 31, 55, 25 ], [ 11, 45, 15, 31, 46, 16 ], // 29
    [ 7, 146, 116, 7, 147, 117 ], [ 21, 73, 45, 7, 74, 46 ], [ 1, 53, 23, 37, 54, 24 ], [ 19, 45, 15, 26, 46, 16 ], // 30
    [ 5, 145, 115, 10, 146, 116 ], [ 19, 75, 47, 10, 76, 48 ], [ 15, 54, 24, 25, 55, 25 ], [ 23, 45, 15, 25, 46, 16 ], // 31
    [ 13, 145, 115, 3, 146, 116 ], [ 2, 74, 46, 29, 75, 47 ], [ 42, 54, 24, 1, 55, 25 ], [ 23, 45, 15, 28, 46, 16 ], // 32
    [ 17, 145, 115 ], [ 10, 74, 46, 23, 75, 47 ], [ 10, 54, 24, 35, 55, 25 ], [ 19, 45, 15, 35, 46, 16 ], // 33
    [ 17, 145, 115, 1, 146, 116 ], [ 14, 74, 46, 21, 75, 47 ], [ 29, 54, 24, 19, 55, 25 ], [ 11, 45, 15, 46, 46, 16 ], // 34
    [ 13, 145, 115, 6, 146, 116 ], [ 14, 74, 46, 23, 75, 47 ], [ 44, 54, 24, 7, 55, 25 ], [ 59, 46, 16, 1, 47, 17 ], // 35
    [ 12, 151, 121, 7, 152, 122 ], [ 12, 75, 47, 26, 76, 48 ], [ 39, 54, 24, 14, 55, 25 ], [ 22, 45, 15, 41, 46, 16 ], // 36
    [ 6, 151, 121, 14, 152, 122 ], [ 6, 75, 47, 34, 76, 48 ], [ 46, 54, 24, 10, 55, 25 ], [ 2, 45, 15, 64, 46, 16 ], // 37
    [ 17, 152, 122, 4, 153, 123 ], [ 29, 74, 46, 14, 75, 47 ], [ 49, 54, 24, 10, 55, 25 ], [ 24, 45, 15, 46, 46, 16 ], // 38
    [ 4, 152, 122, 18, 153, 123 ], [ 13, 74, 46, 32, 75, 47 ], [ 48, 54, 24, 14, 55, 25 ], [ 42, 45, 15, 32, 46, 16 ], // 39
    [ 20, 147, 117, 4, 148, 118 ], [ 40, 75, 47, 7, 76, 48 ], [ 43, 54, 24, 22, 55, 25 ], [ 10, 45, 15, 67, 46, 16 ], // 40
    [ 19, 148, 118, 6, 149, 119 ], [ 18, 75, 47, 31, 76, 48 ], [ 34, 54, 24, 34, 55, 25 ], [ 20, 45, 15, 61, 46, 16 ] ];
    /**
     * 根据数据获取对应版本
     * @return {[type]} [description]
     */
    QRCodeAlg.prototype.getRightType = function() {
        for (var typeNumber = 1; typeNumber < 41; typeNumber++) {
            var rsBlock = RS_BLOCK_TABLE[(typeNumber - 1) * 4 + this.errorCorrectLevel];
            if (rsBlock == undefined) {
                throw new Error("bad rs block @ typeNumber:" + typeNumber + "/errorCorrectLevel:" + this.errorCorrectLevel);
            }
            var length = rsBlock.length / 3;
            var totalDataCount = 0;
            for (var i = 0; i < length; i++) {
                var count = rsBlock[i * 3 + 0];
                var dataCount = rsBlock[i * 3 + 2];
                totalDataCount += dataCount * count;
            }
            var lengthBytes = typeNumber > 9 ? 2 : 1;
            if (this.utf8bytes.length + lengthBytes < totalDataCount || typeNumber == 40) {
                this.typeNumber = typeNumber;
                this.rsBlock = rsBlock;
                this.totalDataCount = totalDataCount;
                break;
            }
        }
    };
    //---------------------------------------------------------------------
    // QRBitBuffer
    //---------------------------------------------------------------------
    function QRBitBuffer() {
        this.buffer = new Array();
        this.length = 0;
    }
    QRBitBuffer.prototype = {
        get: function(index) {
            var bufIndex = Math.floor(index / 8);
            return this.buffer[bufIndex] >>> 7 - index % 8 & 1;
        },
        put: function(num, length) {
            for (var i = 0; i < length; i++) {
                this.putBit(num >>> length - i - 1 & 1);
            }
        },
        putBit: function(bit) {
            var bufIndex = Math.floor(this.length / 8);
            if (this.buffer.length <= bufIndex) {
                this.buffer.push(0);
            }
            if (bit) {
                this.buffer[bufIndex] |= 128 >>> this.length % 8;
            }
            this.length++;
        }
    };
});
define("zepto.extend.data", [], function(require, exports, module) {
    var $ = window.Zepto;
    //     Zepto.js
    //     (c) 2010-2014 Thomas Fuchs
    //     Zepto.js may be freely distributed under the MIT license.
    // The following code is heavily inspired by jQuery's $.fn.data()
    var data = {}, dataAttr = $.fn.data, camelize = $.camelCase, exp = $.expando = "Zepto" + +new Date(), emptyArray = [];
    // Get value from node:
    // 1. first try key as given,
    // 2. then try camelized key,
    // 3. fall back to reading "data-*" attribute.
    function getData(node, name) {
        var id = node[exp], store = id && data[id];
        if (name === undefined) return store || setData(node); else {
            if (store) {
                if (name in store) return store[name];
                var camelName = camelize(name);
                if (camelName in store) return store[camelName];
            }
            return dataAttr.call($(node), name);
        }
    }
    // Store value under camelized key on node
    function setData(node, name, value) {
        var id = node[exp] || (node[exp] = ++$.uuid), store = data[id] || (data[id] = attributeData(node));
        if (name !== undefined) store[camelize(name)] = value;
        return store;
    }
    // Read all "data-*" attributes from a node
    function attributeData(node) {
        var store = {};
        $.each(node.attributes || emptyArray, function(i, attr) {
            if (attr.name.indexOf("data-") == 0) store[camelize(attr.name.replace("data-", ""))] = $.zepto.deserializeValue(attr.value);
        });
        return store;
    }
    $.fn.data = function(name, value) {
        return value === undefined ? // set multiple values via object
        $.isPlainObject(name) ? this.each(function(i, node) {
            $.each(name, function(key, value) {
                setData(node, key, value);
            });
        }) : // get value from first element
        0 in this ? getData(this[0], name) : undefined : // set value on all elements
        this.each(function() {
            setData(this, name, value);
        });
    };
    $.fn.removeData = function(names) {
        if (typeof names == "string") names = names.split(/\s+/);
        return this.each(function() {
            var id = this[exp], store = id && data[id];
            if (store) $.each(names || store, function(key) {
                delete store[names ? camelize(this) : key];
            });
        });
    };
    [ "remove", "empty" ].forEach(function(methodName) {
        var origFn = $.fn[methodName];
        $.fn[methodName] = function() {
            var elements = this.find("*");
            if (methodName === "remove") elements = elements.add(this);
            elements.removeData();
            return origFn.call(this);
        };
    });
});
define("zepto.extend.fx", [], function(require, exports, module) {
    var $ = window.Zepto;
    // Zepto.js
    // (c) 2010-2014 Thomas Fuchs
    // Zepto.js may be freely distributed under the MIT license.
    var prefix = "", eventPrefix, endEventName, endAnimationName, vendors = {
        Webkit: "webkit",
        Moz: "",
        O: "o"
    }, document = window.document, testEl = document.createElement("div"), supportedTransforms = /^((translate|rotate|scale)(X|Y|Z|3d)?|matrix(3d)?|perspective|skew(X|Y)?)$/i, transform, transitionProperty, transitionDuration, transitionTiming, transitionDelay, animationName, animationDuration, animationTiming, animationDelay, cssReset = {};
    function dasherize(str) {
        return str.replace(/([a-z])([A-Z])/, "$1-$2").toLowerCase();
    }
    function normalizeEvent(name) {
        return eventPrefix ? eventPrefix + name : name.toLowerCase();
    }
    $.each(vendors, function(vendor, event) {
        if (testEl.style[vendor + "TransitionProperty"] !== undefined) {
            prefix = "-" + vendor.toLowerCase() + "-";
            eventPrefix = event;
            return false;
        }
    });
    transform = prefix + "transform";
    cssReset[transitionProperty = prefix + "transition-property"] = cssReset[transitionDuration = prefix + "transition-duration"] = cssReset[transitionDelay = prefix + "transition-delay"] = cssReset[transitionTiming = prefix + "transition-timing-function"] = cssReset[animationName = prefix + "animation-name"] = cssReset[animationDuration = prefix + "animation-duration"] = cssReset[animationDelay = prefix + "animation-delay"] = cssReset[animationTiming = prefix + "animation-timing-function"] = "";
    $.fx = {
        off: eventPrefix === undefined && testEl.style.transitionProperty === undefined,
        speeds: {
            _default: 400,
            fast: 200,
            slow: 600
        },
        cssPrefix: prefix,
        transitionEnd: normalizeEvent("TransitionEnd"),
        animationEnd: normalizeEvent("AnimationEnd")
    };
    $.fn.animate = function(properties, duration, ease, callback, delay) {
        if ($.isFunction(duration)) callback = duration, ease = undefined, duration = undefined;
        if ($.isFunction(ease)) callback = ease, ease = undefined;
        if ($.isPlainObject(duration)) ease = duration.easing, callback = duration.complete, 
        delay = duration.delay, duration = duration.duration;
        if (duration) duration = (typeof duration == "number" ? duration : $.fx.speeds[duration] || $.fx.speeds._default) / 1e3;
        if (delay) delay = parseFloat(delay) / 1e3;
        return this.anim(properties, duration, ease, callback, delay);
    };
    $.fn.anim = function(properties, duration, ease, callback, delay) {
        var key, cssValues = {}, cssProperties, transforms = "", that = this, wrappedCallback, endEvent = $.fx.transitionEnd, fired = false;
        if (duration === undefined) duration = $.fx.speeds._default / 1e3;
        if (delay === undefined) delay = 0;
        if ($.fx.off) duration = 0;
        if (typeof properties == "string") {
            // keyframe animation
            cssValues[animationName] = properties;
            cssValues[animationDuration] = duration + "s";
            cssValues[animationDelay] = delay + "s";
            cssValues[animationTiming] = ease || "linear";
            endEvent = $.fx.animationEnd;
        } else {
            cssProperties = [];
            // CSS transitions
            for (key in properties) if (supportedTransforms.test(key)) transforms += key + "(" + properties[key] + ") "; else cssValues[key] = properties[key], 
            cssProperties.push(dasherize(key));
            if (transforms) cssValues[transform] = transforms, cssProperties.push(transform);
            if (duration > 0 && typeof properties === "object") {
                cssValues[transitionProperty] = cssProperties.join(", ");
                cssValues[transitionDuration] = duration + "s";
                cssValues[transitionDelay] = delay + "s";
                cssValues[transitionTiming] = ease || "linear";
            }
        }
        wrappedCallback = function(event) {
            if (typeof event !== "undefined") {
                if (event.target !== event.currentTarget) return;
                // makes sure the event didn't bubble from "below"
                $(event.target).unbind(endEvent, wrappedCallback);
            } else $(this).unbind(endEvent, wrappedCallback);
            // triggered by setTimeout
            fired = true;
            $(this).css(cssReset);
            callback && callback.call(this);
        };
        if (duration > 0) {
            this.bind(endEvent, wrappedCallback);
            // transitionEnd is not always firing on older Android phones
            // so make sure it gets fired
            setTimeout(function() {
                if (fired) return;
                wrappedCallback.call(that);
            }, duration * 1e3 + 25);
        }
        // trigger page reflow so new elements can animate
        this.size() && this.get(0).clientLeft;
        this.css(cssValues);
        if (duration <= 0) setTimeout(function() {
            that.each(function() {
                wrappedCallback.call(this);
            });
        }, 0);
        return this;
    };
    testEl = null;
});
define("zepto.extend.selector", [], function(require, exports, module) {
    var $ = window.Zepto;
    // Zepto.js
    // (c) 2010-2014 Thomas Fuchs
    // Zepto.js may be freely distributed under the MIT license.
    var zepto = $.zepto, oldQsa = zepto.qsa, oldMatches = zepto.matches;
    function visible(elem) {
        elem = $(elem);
        return !!(elem.width() || elem.height()) && elem.css("display") !== "none";
    }
    // Implements a subset from:
    // http://api.jquery.com/category/selectors/jquery-selector-extensions/
    //
    // Each filter function receives the current index, all nodes in the
    // considered set, and a value if there were parentheses. The value
    // of `this` is the node currently being considered. The function returns the
    // resulting node(s), null, or undefined.
    //
    // Complex selectors are not supported:
    // li:has(label:contains("foo")) + li:has(label:contains("bar"))
    // ul.inner:first > li
    var filters = $.expr[":"] = {
        visible: function() {
            if (visible(this)) return this;
        },
        hidden: function() {
            if (!visible(this)) return this;
        },
        selected: function() {
            if (this.selected) return this;
        },
        checked: function() {
            if (this.checked) return this;
        },
        parent: function() {
            return this.parentNode;
        },
        first: function(idx) {
            if (idx === 0) return this;
        },
        last: function(idx, nodes) {
            if (idx === nodes.length - 1) return this;
        },
        eq: function(idx, _, value) {
            if (idx === value) return this;
        },
        contains: function(idx, _, text) {
            if ($(this).text().indexOf(text) > -1) return this;
        },
        has: function(idx, _, sel) {
            if (zepto.qsa(this, sel).length) return this;
        }
    };
    var filterRe = new RegExp("(.*):(\\w+)(?:\\(([^)]+)\\))?$\\s*"), childRe = /^\s*>/, classTag = "Zepto" + +new Date();
    function process(sel, fn) {
        // quote the hash in `a[href^=#]` expression
        sel = sel.replace(/=#\]/g, '="#"]');
        var filter, arg, match = filterRe.exec(sel);
        if (match && match[2] in filters) {
            filter = filters[match[2]], arg = match[3];
            sel = match[1];
            if (arg) {
                var num = Number(arg);
                if (isNaN(num)) arg = arg.replace(/^["']|["']$/g, ""); else arg = num;
            }
        }
        return fn(sel, filter, arg);
    }
    zepto.qsa = function(node, selector) {
        return process(selector, function(sel, filter, arg) {
            try {
                var taggedParent;
                if (!sel && filter) sel = "*"; else if (childRe.test(sel)) // support "> *" child queries by tagging the parent node with a
                // unique class and prepending that classname onto the selector
                taggedParent = $(node).addClass(classTag), sel = "." + classTag + " " + sel;
                var nodes = oldQsa(node, sel);
            } catch (e) {
                console.error("error performing selector: %o", selector);
                throw e;
            } finally {
                if (taggedParent) taggedParent.removeClass(classTag);
            }
            return !filter ? nodes : zepto.uniq($.map(nodes, function(n, i) {
                return filter.call(n, i, nodes, arg);
            }));
        });
    };
    zepto.matches = function(node, selector) {
        return process(selector, function(sel, filter, arg) {
            return (!sel || oldMatches(node, sel)) && (!filter || filter.call(node, null, arg) === node);
        });
    };
});
define("zepto.extend.touch", [], function(require, exports, module) {
    // 2014.01.22  Support IE11 touch events.
    var $ = window.Zepto;
    // Zepto.js
    // (c) 2010-2014 Thomas Fuchs
    // Zepto.js may be freely distributed under the MIT license.
    var touch = {}, touchTimeout, tapTimeout, swipeTimeout, longTapTimeout, longTapDelay = 750, gesture;
    function swipeDirection(x1, x2, y1, y2) {
        return Math.abs(x1 - x2) >= Math.abs(y1 - y2) ? x1 - x2 > 0 ? "Left" : "Right" : y1 - y2 > 0 ? "Up" : "Down";
    }
    function longTap() {
        longTapTimeout = null;
        if (touch.last) {
            touch.el.trigger("longTap");
            touch = {};
        }
    }
    function cancelLongTap() {
        if (longTapTimeout) clearTimeout(longTapTimeout);
        longTapTimeout = null;
    }
    function cancelAll() {
        if (touchTimeout) clearTimeout(touchTimeout);
        if (tapTimeout) clearTimeout(tapTimeout);
        if (swipeTimeout) clearTimeout(swipeTimeout);
        if (longTapTimeout) clearTimeout(longTapTimeout);
        touchTimeout = tapTimeout = swipeTimeout = longTapTimeout = null;
        touch = {};
    }
    function isPrimaryTouch(event) {
        return (event.pointerType == "touch" || event.pointerType == event.MSPOINTER_TYPE_TOUCH) && event.isPrimary;
    }
    function isPointerEventType(e, type) {
        return e.type == "pointer" + type || e.type.toLowerCase() == "mspointer" + type;
    }
    $(document).ready(function() {
        var now, delta, deltaX = 0, deltaY = 0, firstTouch, _isPointerType;
        if ("MSGesture" in window) {
            gesture = new MSGesture();
            gesture.target = document.body;
        }
        $(document).bind("MSGestureEnd", function(e) {
            var swipeDirectionFromVelocity = e.velocityX > 1 ? "Right" : e.velocityX < -1 ? "Left" : e.velocityY > 1 ? "Down" : e.velocityY < -1 ? "Up" : null;
            if (swipeDirectionFromVelocity) {
                touch.el.trigger("swipe");
                touch.el.trigger("swipe" + swipeDirectionFromVelocity);
            }
        }).on("touchstart MSPointerDown pointerdown", function(e) {
            if ((_isPointerType = isPointerEventType(e, "down")) && !isPrimaryTouch(e)) return;
            firstTouch = _isPointerType ? e : e.touches[0];
            if (e.touches && e.touches.length === 1 && touch.x2) {
                // Clear out touch movement data if we have it sticking around
                // This can occur if touchcancel doesn't fire due to preventDefault, etc.
                touch.x2 = undefined;
                touch.y2 = undefined;
            }
            now = Date.now();
            delta = now - (touch.last || now);
            touch.el = $("tagName" in firstTouch.target ? firstTouch.target : firstTouch.target.parentNode);
            touchTimeout && clearTimeout(touchTimeout);
            touch.x1 = firstTouch.pageX;
            touch.y1 = firstTouch.pageY;
            if (delta > 0 && delta <= 250) touch.isDoubleTap = true;
            touch.last = now;
            longTapTimeout = setTimeout(longTap, longTapDelay);
            // adds the current touch contact for IE gesture recognition
            if (gesture && _isPointerType) gesture.addPointer(e.pointerId);
        }).on("touchmove MSPointerMove pointermove", function(e) {
            if ((_isPointerType = isPointerEventType(e, "move")) && !isPrimaryTouch(e)) return;
            firstTouch = _isPointerType ? e : e.touches[0];
            cancelLongTap();
            touch.x2 = firstTouch.pageX;
            touch.y2 = firstTouch.pageY;
            deltaX += Math.abs(touch.x1 - touch.x2);
            deltaY += Math.abs(touch.y1 - touch.y2);
        }).on("touchend MSPointerUp pointerup", function(e) {
            if ((_isPointerType = isPointerEventType(e, "up")) && !isPrimaryTouch(e)) return;
            cancelLongTap();
            // swipe
            if (touch.x2 && Math.abs(touch.x1 - touch.x2) > 30 || touch.y2 && Math.abs(touch.y1 - touch.y2) > 30) swipeTimeout = setTimeout(function() {
                touch.el.trigger("swipe");
                touch.el.trigger("swipe" + swipeDirection(touch.x1, touch.x2, touch.y1, touch.y2));
                touch = {};
            }, 0); else if ("last" in touch) // don't fire tap when delta position changed by more than 30 pixels,
            // for instance when moving to a point and back to origin
            if (deltaX < 30 && deltaY < 30) {
                // delay by one tick so we can cancel the 'tap' event if 'scroll' fires
                // ('tap' fires before 'scroll')
                tapTimeout = setTimeout(function() {
                    // trigger universal 'tap' with the option to cancelTouch()
                    // (cancelTouch cancels processing of single vs double taps for faster 'tap' response)
                    var event = $.Event("tap");
                    event.cancelTouch = cancelAll;
                    touch.el.trigger(event);
                    // trigger double tap immediately
                    if (touch.isDoubleTap) {
                        if (touch.el) touch.el.trigger("doubleTap");
                        touch = {};
                    } else {
                        touchTimeout = setTimeout(function() {
                            touchTimeout = null;
                            if (touch.el) touch.el.trigger("singleTap");
                            touch = {};
                        }, 250);
                    }
                }, 0);
            } else {
                touch = {};
            }
            deltaX = deltaY = 0;
        }).on("touchcancel MSPointerCancel pointercancel", cancelAll);
        // scrolling the window indicates intention of the user
        // to scroll, not tap or swipe, so cancel all ongoing events
        $(window).on("scroll", cancelAll);
    });
    [ "swipe", "swipeLeft", "swipeRight", "swipeUp", "swipeDown", "doubleTap", "tap", "singleTap", "longTap" ].forEach(function(eventName) {
        $.fn[eventName] = function(callback) {
            return this.on(eventName, callback);
        };
    });
});
define("zepto.flexslider", [ "core", "zepto.extend.fx", "zepto.extend.data", "zepto.extend.selector", "zepto.extend.data" ], function(require, exports, module) {
    var $ = window.Zepto;
    require("core");
    require("zepto.extend.data");
    // Zepto $.data patch
    $.data = function(elem, key, value) {
        return $(elem).data(key, value);
    };
    //scrollLeft
    [ "Left", "Top" ].forEach(function(name, i) {
        var method = "scroll" + name;
        function isWindow(obj) {
            return obj && typeof obj === "object" && "setInterval" in obj;
        }
        function getWindow(elem) {
            return isWindow(elem) ? elem : elem.nodeType === 9 ? elem.defaultView || elem.parentWindow : false;
        }
        $.fn[method] = function(val) {
            var elem, win;
            if (val === undefined) {
                elem = this[0];
                if (!elem) {
                    return null;
                }
                win = getWindow(elem);
                // Return the scroll offset
                return win ? "pageXOffset" in win ? win[i ? "pageYOffset" : "pageXOffset"] : win.document.documentElement[method] || win.document.body[method] : elem[method];
            }
            // Set the scroll offset
            this.each(function() {
                win = getWindow(this);
                if (win) {
                    var xCoord = !i ? val : $(win).scrollLeft();
                    var yCoord = i ? val : $(win).scrollTop();
                    win.scrollTo(xCoord, yCoord);
                } else {
                    this[method] = val;
                }
            });
        };
    });
    // Create outerHeight and outerWidth methods
    [ "width", "height" ].forEach(function(dimension) {
        var offset, Dimension = dimension.replace(/./, function(m) {
            return m[0].toUpperCase();
        });
        $.fn["outer" + Dimension] = function(margin) {
            var elem = this;
            if (elem) {
                var size = elem[dimension]();
                var sides = {
                    width: [ "left", "right" ],
                    height: [ "top", "bottom" ]
                };
                sides[dimension].forEach(function(side) {
                    if (margin) size += parseInt(elem.css("margin-" + side), 10);
                });
                return size;
            } else {
                return null;
            }
        };
    });
    /*
     * Zepto FlexSlider v2.3
     * @desc Porting from jQuery FlexSlider v2.3, Contributing Author: Tyler Smith
     * @license Copyright 2012 WooThemes GPLv2
     */
    //FlexSlider: Object Instance
    $.flexslider = function(el, options) {
        var slider = $(el);
        // making variables public
        slider.vars = $.extend({}, $.flexslider.defaults, options);
        var namespace = slider.vars.namespace, msGesture = window.navigator && window.navigator.msPointerEnabled && window.MSGesture, touch = ("ontouchstart" in window || msGesture || window.DocumentTouch && document instanceof DocumentTouch) && slider.vars.touch, // depricating this idea, as devices are being released with both of these events
        //eventType = (touch) ? "touchend" : "click",
        eventType = "click touchend MSPointerUp keyup", watchedEvent = "", watchedEventClearTimer, vertical = slider.vars.direction === "vertical", reverse = slider.vars.reverse, carousel = slider.vars.itemWidth > 0, fade = slider.vars.animation === "fade", asNav = slider.vars.asNavFor !== "", methods = {}, focused = true;
        // Store a reference to the slider object
        $.data(el, "flexslider", slider);
        // Private slider methods
        methods = {
            init: function() {
                slider.animating = false;
                // Get current slide and make sure it is a number
                slider.currentSlide = parseInt(slider.vars.startAt ? slider.vars.startAt : 0, 10);
                if (isNaN(slider.currentSlide)) slider.currentSlide = 0;
                slider.animatingTo = slider.currentSlide;
                slider.atEnd = slider.currentSlide === 0 || slider.currentSlide === slider.last;
                slider.containerSelector = slider.vars.selector.substr(0, slider.vars.selector.search(" "));
                slider.slides = $(slider.vars.selector, slider);
                slider.container = $(slider.containerSelector, slider);
                slider.count = slider.slides.length;
                // SYNC:
                slider.syncExists = $(slider.vars.sync).length > 0;
                // SLIDE:
                if (slider.vars.animation === "slide") slider.vars.animation = "swing";
                slider.prop = vertical ? "top" : "marginLeft";
                slider.args = {};
                // SLIDESHOW:
                slider.manualPause = false;
                slider.stopped = false;
                //PAUSE WHEN INVISIBLE
                slider.started = false;
                slider.startTimeout = null;
                // TOUCH/USECSS:
                slider.transitions = !slider.vars.video && !fade && slider.vars.useCSS && function() {
                    var obj = document.createElement("div"), props = [ "perspectiveProperty", "WebkitPerspective", "MozPerspective", "OPerspective", "msPerspective" ];
                    for (var i in props) {
                        if (obj.style[props[i]] !== undefined) {
                            slider.pfx = props[i].replace("Perspective", "").toLowerCase();
                            slider.prop = "-" + slider.pfx + "-transform";
                            return true;
                        }
                    }
                    return false;
                }();
                slider.ensureAnimationEnd = "";
                // CONTROLSCONTAINER:
                if (slider.vars.controlsContainer !== "") slider.controlsContainer = $(slider.vars.controlsContainer).length > 0 && $(slider.vars.controlsContainer);
                // MANUAL:
                if (slider.vars.manualControls !== "") slider.manualControls = $(slider.vars.manualControls).length > 0 && $(slider.vars.manualControls);
                // RANDOMIZE:
                if (slider.vars.randomize) {
                    slider.slides.sort(function() {
                        return Math.round(Math.random()) - .5;
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
                if (slider.vars.keyboard && ($(slider.containerSelector).length === 1 || slider.vars.multipleKeyboard)) {
                    $(document).bind("keyup", function(event) {
                        var keycode = event.keyCode;
                        if (!slider.animating && (keycode === 39 || keycode === 37)) {
                            var target = keycode === 39 ? slider.getTarget("next") : keycode === 37 ? slider.getTarget("prev") : false;
                            slider.flexAnimate(target, slider.vars.pauseOnAction);
                        }
                    });
                }
                // MOUSEWHEEL:
                if (slider.vars.mousewheel) {
                    slider.bind("mousewheel", function(event, delta, deltaX, deltaY) {
                        event.preventDefault();
                        var target = delta < 0 ? slider.getTarget("next") : slider.getTarget("prev");
                        slider.flexAnimate(target, slider.vars.pauseOnAction);
                    });
                }
                // PAUSEPLAY
                if (slider.vars.pausePlay) methods.pausePlay.setup();
                //PAUSE WHEN INVISIBLE
                if (slider.vars.slideshow && slider.vars.pauseInvisible) methods.pauseInvisible.init();
                // SLIDSESHOW
                if (slider.vars.slideshow) {
                    if (slider.vars.pauseOnHover) {
                        /*slider.hover(function() {
                            if (!slider.manualPlay && !slider.manualPause) slider.pause();
                        }, function() {
                            if (!slider.manualPause && !slider.manualPlay && !slider.stopped) slider.play();
                        });*/
                        slider.on("mouseover", function() {
                            if (!slider.manualPlay && !slider.manualPause) slider.pause();
                        });
                        slider.on("mouseout", function() {
                            if (!slider.manualPause && !slider.manualPlay && !slider.stopped) slider.play();
                        });
                    }
                    // initialize animation
                    //If we're visible, or we don't use PageVisibility API
                    if (!slider.vars.pauseInvisible || !methods.pauseInvisible.isHidden()) {
                        slider.vars.initDelay > 0 ? slider.startTimeout = setTimeout(slider.play, slider.vars.initDelay) : slider.play();
                    }
                }
                // ASNAV:
                if (asNav) methods.asNav.setup();
                // TOUCH
                if (touch && slider.vars.touch) methods.touch();
                // FADE&&SMOOTHHEIGHT || SLIDE:
                if (!fade || fade && slider.vars.smoothHeight) $(window).bind("resize orientationchange focus", methods.resize);
                slider.find("img").attr("draggable", "false");
                // API: start() Callback
                setTimeout(function() {
                    slider.vars.start(slider);
                }, 200);
            },
            asNav: {
                setup: function() {
                    slider.asNav = true;
                    slider.animatingTo = Math.floor(slider.currentSlide / slider.move);
                    slider.currentItem = slider.currentSlide;
                    slider.slides.removeClass(namespace + "active-slide").eq(slider.currentItem).addClass(namespace + "active-slide");
                    if (!msGesture) {
                        slider.slides.on(eventType, function(e) {
                            e.preventDefault();
                            var $slide = $(this), target = $slide.index();
                            var posFromLeft = $slide.offset().left - $(slider).scrollLeft();
                            // Find position of slide relative to left of slider container
                            if (posFromLeft <= 0 && $slide.hasClass(namespace + "active-slide")) {
                                slider.flexAnimate(slider.getTarget("prev"), true);
                            } else if (!$(slider.vars.asNavFor).data("flexslider").animating && !$slide.hasClass(namespace + "active-slide")) {
                                slider.direction = slider.currentItem < target ? "next" : "prev";
                                slider.flexAnimate(target, slider.vars.pauseOnAction, false, true, true);
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
                                if (e.currentTarget._gesture) e.currentTarget._gesture.addPointer(e.pointerId);
                            }, false);
                            that.addEventListener("MSGestureTap", function(e) {
                                e.preventDefault();
                                var $slide = $(this), target = $slide.index();
                                if (!$(slider.vars.asNavFor).data("flexslider").animating && !$slide.hasClass("active")) {
                                    slider.direction = slider.currentItem < target ? "next" : "prev";
                                    slider.flexAnimate(target, slider.vars.pauseOnAction, false, true, true);
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
                    } else {
                        // MANUALCONTROLS:
                        methods.controlNav.setupManual();
                    }
                },
                setupPaging: function() {
                    var type = slider.vars.controlNav === "thumbnails" ? "control-thumbs" : "control-paging", j = 1, item, slide;
                    slider.controlNavScaffold = $('<ol class="' + namespace + "control-nav " + namespace + type + '"></ol>');
                    if (slider.pagingCount > 1) {
                        for (var i = 0; i < slider.pagingCount; i++) {
                            slide = slider.slides.eq(i);
                            item = slider.vars.controlNav === "thumbnails" ? '<img src="' + slide.attr("data-thumb") + '"/>' : "<a>" + j + "</a>";
                            if ("thumbnails" === slider.vars.controlNav && true === slider.vars.thumbCaptions) {
                                var captn = slide.attr("data-thumbcaption");
                                if ("" != captn && undefined != captn) item += '<span class="' + namespace + 'caption">' + captn + "</span>";
                            }
                            //slider.controlNavScaffold.append('<li>' + item + '</li>');
                            slider.controlNavScaffold.append("<li>" + item + "<i></i></li>");
                            j++;
                        }
                    }
                    // CONTROLSCONTAINER:
                    slider.controlsContainer ? $(slider.controlsContainer).append(slider.controlNavScaffold) : slider.append(slider.controlNavScaffold);
                    methods.controlNav.set();
                    methods.controlNav.active();
                    slider.controlNavScaffold.delegate("a, img", eventType, function(event) {
                        event.preventDefault();
                        if (watchedEvent === "" || watchedEvent === event.type) {
                            var $this = $(this), target = slider.controlNav.index($this);
                            if (!$this.hasClass(namespace + "active")) {
                                slider.direction = target > slider.currentSlide ? "next" : "prev";
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
                            var $this = $(this), target = slider.controlNav.index($this);
                            if (!$this.hasClass(namespace + "active")) {
                                target > slider.currentSlide ? slider.direction = "next" : slider.direction = "prev";
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
                    var selector = slider.vars.controlNav === "thumbnails" ? "img" : "a";
                    slider.controlNav = $("." + namespace + "control-nav li " + selector, slider.controlsContainer ? slider.controlsContainer : slider);
                },
                active: function() {
                    slider.controlNav.removeClass(namespace + "active").eq(slider.animatingTo).addClass(namespace + "active");
                },
                update: function(action, pos) {
                    if (slider.pagingCount > 1 && action === "add") {
                        slider.controlNavScaffold.append($("<li><a>" + slider.count + "</a></li>"));
                    } else if (slider.pagingCount === 1) {
                        slider.controlNavScaffold.find("li").remove();
                    } else {
                        slider.controlNav.eq(pos).closest("li").remove();
                    }
                    methods.controlNav.set();
                    slider.pagingCount > 1 && slider.pagingCount !== slider.controlNav.length ? slider.update(pos, action) : methods.controlNav.active();
                }
            },
            directionNav: {
                setup: function() {
                    var directionNavScaffold = $('<ul class="' + namespace + 'direction-nav"><li><a class="' + namespace + 'prev" href="#">' + slider.vars.prevText + '</a></li><li><a class="' + namespace + 'next" href="#">' + slider.vars.nextText + "</a></li></ul>");
                    // CONTROLSCONTAINER:
                    if (slider.controlsContainer) {
                        $(slider.controlsContainer).append(directionNavScaffold);
                        slider.directionNav = $("." + namespace + "direction-nav li a", slider.controlsContainer);
                    } else {
                        slider.append(directionNavScaffold);
                        slider.directionNav = $("." + namespace + "direction-nav li a", slider);
                    }
                    methods.directionNav.update();
                    slider.directionNav.bind(eventType, function(event) {
                        event.preventDefault();
                        var target;
                        if (watchedEvent === "" || watchedEvent === event.type) {
                            target = $(this).hasClass(namespace + "next") ? slider.getTarget("next") : slider.getTarget("prev");
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
                    var disabledClass = namespace + "disabled";
                    if (slider.pagingCount === 1) {
                        slider.directionNav.addClass(disabledClass).attr("tabindex", "-1");
                    } else if (!slider.vars.animationLoop) {
                        if (slider.animatingTo === 0) {
                            slider.directionNav.removeClass(disabledClass).filter("." + namespace + "prev").addClass(disabledClass).attr("tabindex", "-1");
                        } else if (slider.animatingTo === slider.last) {
                            slider.directionNav.removeClass(disabledClass).filter("." + namespace + "next").addClass(disabledClass).attr("tabindex", "-1");
                        } else {
                            slider.directionNav.removeClass(disabledClass).removeAttr("tabindex");
                        }
                    } else {
                        slider.directionNav.removeClass(disabledClass).removeAttr("tabindex");
                    }
                }
            },
            pausePlay: {
                setup: function() {
                    var pausePlayScaffold = $('<div class="' + namespace + 'pauseplay"><a></a></div>');
                    // CONTROLSCONTAINER:
                    if (slider.controlsContainer) {
                        slider.controlsContainer.append(pausePlayScaffold);
                        slider.pausePlay = $("." + namespace + "pauseplay a", slider.controlsContainer);
                    } else {
                        slider.append(pausePlayScaffold);
                        slider.pausePlay = $("." + namespace + "pauseplay a", slider);
                    }
                    methods.pausePlay.update(slider.vars.slideshow ? namespace + "pause" : namespace + "play");
                    slider.pausePlay.bind(eventType, function(event) {
                        event.preventDefault();
                        if (watchedEvent === "" || watchedEvent === event.type) {
                            if ($(this).hasClass(namespace + "pause")) {
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
                    state === "play" ? slider.pausePlay.removeClass(namespace + "pause").addClass(namespace + "play").html(slider.vars.playText) : slider.pausePlay.removeClass(namespace + "play").addClass(namespace + "pause").html(slider.vars.pauseText);
                }
            },
            touch: function() {
                var startX, startY, offset, cwidth, dx, startT, scrolling = false, localX = 0, localY = 0, accDx = 0;
                if (!msGesture) {
                    el.addEventListener("touchstart", onTouchStart, false);
                    function onTouchStart(e) {
                        if (slider.animating) {
                            e.preventDefault();
                        } else if (window.navigator.msPointerEnabled || e.touches.length === 1) {
                            slider.pause();
                            // CAROUSEL:
                            cwidth = vertical ? slider.h : slider.w;
                            startT = Number(new Date());
                            // CAROUSEL:
                            // Local vars for X and Y points.
                            localX = e.touches[0].pageX;
                            localY = e.touches[0].pageY;
                            offset = carousel && reverse && slider.animatingTo === slider.last ? 0 : carousel && reverse ? slider.limit - (slider.itemW + slider.vars.itemMargin) * slider.move * slider.animatingTo : carousel && slider.currentSlide === slider.last ? slider.limit : carousel ? (slider.itemW + slider.vars.itemMargin) * slider.move * slider.currentSlide : reverse ? (slider.last - slider.currentSlide + slider.cloneOffset) * cwidth : (slider.currentSlide + slider.cloneOffset) * cwidth;
                            startX = vertical ? localY : localX;
                            startY = vertical ? localX : localY;
                            el.addEventListener("touchmove", onTouchMove, false);
                            el.addEventListener("touchend", onTouchEnd, false);
                        }
                    }
                    function onTouchMove(e) {
                        // Local vars for X and Y points.
                        localX = e.touches[0].pageX;
                        localY = e.touches[0].pageY;
                        dx = vertical ? startX - localY : startX - localX;
                        scrolling = vertical ? Math.abs(dx) < Math.abs(localX - startY) : Math.abs(dx) < Math.abs(localY - startY);
                        var fxms = 500;
                        if (!scrolling || Number(new Date()) - startT > fxms) {
                            e.preventDefault();
                            if (!fade && slider.transitions) {
                                if (!slider.vars.animationLoop) {
                                    dx = dx / (slider.currentSlide === 0 && dx < 0 || slider.currentSlide === slider.last && dx > 0 ? Math.abs(dx) / cwidth + 2 : 1);
                                }
                                slider.setProps(offset + dx, "setTouch");
                            }
                        }
                    }
                    function onTouchEnd(e) {
                        // finish the touch by undoing the touch session
                        el.removeEventListener("touchmove", onTouchMove, false);
                        if (slider.animatingTo === slider.currentSlide && !scrolling && !(dx === null)) {
                            var updateDx = reverse ? -dx : dx, target = updateDx > 0 ? slider.getTarget("next") : slider.getTarget("prev");
                            if (slider.canAdvance(target) && (Number(new Date()) - startT < 550 && Math.abs(updateDx) > 50 || Math.abs(updateDx) > cwidth / 2)) {
                                slider.flexAnimate(target, slider.vars.pauseOnAction);
                            } else {
                                if (!fade) slider.flexAnimate(slider.currentSlide, slider.vars.pauseOnAction, true);
                            }
                        }
                        el.removeEventListener("touchend", onTouchEnd, false);
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
                    el.addEventListener("MSGestureChange", onMSGestureChange, false);
                    el.addEventListener("MSGestureEnd", onMSGestureEnd, false);
                    function onMSPointerDown(e) {
                        e.stopPropagation();
                        if (slider.animating) {
                            e.preventDefault();
                        } else {
                            slider.pause();
                            el._gesture.addPointer(e.pointerId);
                            accDx = 0;
                            cwidth = vertical ? slider.h : slider.w;
                            startT = Number(new Date());
                            // CAROUSEL:
                            offset = carousel && reverse && slider.animatingTo === slider.last ? 0 : carousel && reverse ? slider.limit - (slider.itemW + slider.vars.itemMargin) * slider.move * slider.animatingTo : carousel && slider.currentSlide === slider.last ? slider.limit : carousel ? (slider.itemW + slider.vars.itemMargin) * slider.move * slider.currentSlide : reverse ? (slider.last - slider.currentSlide + slider.cloneOffset) * cwidth : (slider.currentSlide + slider.cloneOffset) * cwidth;
                        }
                    }
                    function onMSGestureChange(e) {
                        e.stopPropagation();
                        var slider = e.target._slider;
                        if (!slider) {
                            return;
                        }
                        var transX = -e.translationX, transY = -e.translationY;
                        //Accumulate translations.
                        accDx = accDx + (vertical ? transY : transX);
                        dx = accDx;
                        scrolling = vertical ? Math.abs(accDx) < Math.abs(-transX) : Math.abs(accDx) < Math.abs(-transY);
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
                                    dx = accDx / (slider.currentSlide === 0 && accDx < 0 || slider.currentSlide === slider.last && accDx > 0 ? Math.abs(accDx) / cwidth + 2 : 1);
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
                        if (slider.animatingTo === slider.currentSlide && !scrolling && !(dx === null)) {
                            var updateDx = reverse ? -dx : dx, target = updateDx > 0 ? slider.getTarget("next") : slider.getTarget("prev");
                            if (slider.canAdvance(target) && (Number(new Date()) - startT < 550 && Math.abs(updateDx) > 50 || Math.abs(updateDx) > cwidth / 2)) {
                                slider.flexAnimate(target, slider.vars.pauseOnAction);
                            } else {
                                if (!fade) slider.flexAnimate(slider.currentSlide, slider.vars.pauseOnAction, true);
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
                if (!slider.animating && slider.is(":visible")) {
                    if (!carousel) slider.doMath();
                    if (fade) {
                        // SMOOTH HEIGHT:
                        methods.smoothHeight();
                    } else if (carousel) {
                        //CAROUSEL:
                        slider.slides.width(slider.computedW);
                        slider.update(slider.pagingCount);
                        slider.setProps();
                    } else if (vertical) {
                        //VERTICAL:
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
                    var $obj = fade ? slider : slider.viewport;
                    dur ? $obj.animate({
                        height: slider.slides.eq(slider.animatingTo).height()
                    }, dur) : $obj.height(slider.slides.eq(slider.animatingTo).height());
                }
            },
            sync: function(action) {
                var $obj = $(slider.vars.sync).data("flexslider"), target = slider.animatingTo;
                switch (action) {
                  case "animate":
                    $obj.flexAnimate(target, slider.vars.pauseOnAction, false, true);
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
                $clone.filter("[id]").add($clone.find("[id]")).each(function() {
                    var $this = $(this);
                    $this.attr("id", $this.attr("id") + "_clone");
                });
                return $clone;
            },
            pauseInvisible: {
                visProp: null,
                init: function() {
                    var prefixes = [ "webkit", "moz", "ms", "o" ];
                    if ("hidden" in document) return "hidden";
                    for (var i = 0; i < prefixes.length; i++) {
                        if (prefixes[i] + "Hidden" in document) methods.pauseInvisible.visProp = prefixes[i] + "Hidden";
                    }
                    if (methods.pauseInvisible.visProp) {
                        var evtname = methods.pauseInvisible.visProp.replace(/[H|h]idden/, "") + "visibilitychange";
                        document.addEventListener(evtname, function() {
                            if (methods.pauseInvisible.isHidden()) {
                                if (slider.startTimeout) clearTimeout(slider.startTimeout); else slider.pause();
                            } else {
                                if (slider.started) slider.play(); else slider.vars.initDelay > 0 ? setTimeout(slider.play, slider.vars.initDelay) : slider.play();
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
                }, 3e3);
            }
        };
        // public methods
        slider.flexAnimate = function(target, pause, override, withSync, fromNav) {
            if (!slider.vars.animationLoop && target !== slider.currentSlide) {
                slider.direction = target > slider.currentSlide ? "next" : "prev";
            }
            if (asNav && slider.pagingCount === 1) slider.direction = slider.currentItem < target ? "next" : "prev";
            if (!slider.animating && (slider.canAdvance(target, fromNav) || override) && slider.is(":visible")) {
                if (asNav && withSync) {
                    var master = $(slider.vars.asNavFor).data("flexslider");
                    slider.atEnd = target === 0 || target === slider.count - 1;
                    master.flexAnimate(target, true, false, true, fromNav);
                    slider.direction = slider.currentItem < target ? "next" : "prev";
                    master.direction = slider.direction;
                    if (Math.ceil((target + 1) / slider.visible) - 1 !== slider.currentSlide && target !== 0) {
                        slider.currentItem = target;
                        slider.slides.removeClass(namespace + "active-slide").eq(target).addClass(namespace + "active-slide");
                        target = Math.floor(target / slider.visible);
                    } else {
                        slider.currentItem = target;
                        slider.slides.removeClass(namespace + "active-slide").eq(target).addClass(namespace + "active-slide");
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
                if (!carousel) slider.slides.removeClass(namespace + "active-slide").eq(target).addClass(namespace + "active-slide");
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
                    var dimension = vertical ? slider.slides.filter(":first").height() : slider.computedW, margin, slideString, calcNext;
                    // INFINITE LOOP / REVERSE:
                    if (carousel) {
                        //margin = (slider.vars.itemWidth > slider.w) ? slider.vars.itemMargin * 2 : slider.vars.itemMargin;
                        margin = slider.vars.itemMargin;
                        calcNext = (slider.itemW + margin) * slider.move * slider.animatingTo;
                        slideString = calcNext > slider.limit && slider.visible !== 1 ? slider.limit : calcNext;
                    } else if (slider.currentSlide === 0 && target === slider.count - 1 && slider.vars.animationLoop && slider.direction !== "next") {
                        slideString = reverse ? (slider.count + slider.cloneOffset) * dimension : 0;
                    } else if (slider.currentSlide === slider.last && target === 0 && slider.vars.animationLoop && slider.direction !== "prev") {
                        slideString = reverse ? 0 : (slider.count + 1) * dimension;
                    } else {
                        slideString = reverse ? (slider.count - 1 - target + slider.cloneOffset) * dimension : (target + slider.cloneOffset) * dimension;
                    }
                    slider.setProps(slideString, "", slider.vars.animationSpeed);
                    if (slider.transitions) {
                        if (!slider.vars.animationLoop || !slider.atEnd) {
                            slider.animating = false;
                            slider.currentSlide = slider.animatingTo;
                        }
                        // Unbind previous transitionEnd events and re-bind new transitionEnd event
                        slider.container.unbind("webkitTransitionEnd transitionend");
                        slider.container.bind("webkitTransitionEnd transitionend", function() {
                            clearTimeout(slider.ensureAnimationEnd);
                            slider.wrapup(dimension);
                        });
                        // Insurance for the ever-so-fickle transitionEnd event
                        clearTimeout(slider.ensureAnimationEnd);
                        slider.ensureAnimationEnd = setTimeout(function() {
                            slider.wrapup(dimension);
                        }, slider.vars.animationSpeed + 100);
                    } else {
                        slider.container.animate(slider.args, slider.vars.animationSpeed, slider.vars.easing, function() {
                            slider.wrapup(dimension);
                        });
                    }
                } else {
                    // FADE:
                    if (!touch) {
                        //slider.slides.eq(slider.currentSlide).fadeOut(slider.vars.animationSpeed, slider.vars.easing);
                        //slider.slides.eq(target).fadeIn(slider.vars.animationSpeed, slider.vars.easing, slider.wrapup);
                        slider.slides.eq(slider.currentSlide).css({
                            zIndex: 1
                        }).animate({
                            opacity: 0
                        }, slider.vars.animationSpeed, slider.vars.easing);
                        slider.slides.eq(target).css({
                            zIndex: 2
                        }).animate({
                            opacity: 1
                        }, slider.vars.animationSpeed, slider.vars.easing, slider.wrapup);
                    } else {
                        slider.slides.eq(slider.currentSlide).css({
                            opacity: 0,
                            zIndex: 1
                        });
                        slider.slides.eq(target).css({
                            opacity: 1,
                            zIndex: 2
                        });
                        slider.wrapup(dimension);
                    }
                }
                // SMOOTH HEIGHT:
                if (slider.vars.smoothHeight) methods.smoothHeight(slider.vars.animationSpeed);
            }
        };
        slider.wrapup = function(dimension) {
            // SLIDE:
            if (!fade && !carousel) {
                if (slider.currentSlide === 0 && slider.animatingTo === slider.last && slider.vars.animationLoop) {
                    slider.setProps(dimension, "jumpEnd");
                } else if (slider.currentSlide === slider.last && slider.animatingTo === 0 && slider.vars.animationLoop) {
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
            if (!slider.animating && focused) slider.flexAnimate(slider.getTarget("next"));
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
            slider.animatedSlides = slider.animatedSlides || setInterval(slider.animateSlides, slider.vars.slideshowSpeed);
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
            var last = asNav ? slider.pagingCount - 1 : slider.last;
            return fromNav ? true : asNav && slider.currentItem === slider.count - 1 && target === 0 && slider.direction === "prev" ? true : asNav && slider.currentItem === 0 && target === slider.pagingCount - 1 && slider.direction !== "next" ? false : target === slider.currentSlide && !asNav ? false : slider.vars.animationLoop ? true : slider.atEnd && slider.currentSlide === 0 && target === last && slider.direction !== "next" ? false : slider.atEnd && slider.currentSlide === last && target === 0 && slider.direction === "next" ? false : true;
        };
        slider.getTarget = function(dir) {
            slider.direction = dir;
            if (dir === "next") {
                return slider.currentSlide === slider.last ? 0 : slider.currentSlide + 1;
            } else {
                return slider.currentSlide === 0 ? slider.last : slider.currentSlide - 1;
            }
        };
        // SLIDE:
        slider.setProps = function(pos, special, dur) {
            var target = function() {
                var posCheck = pos ? pos : (slider.itemW + slider.vars.itemMargin) * slider.move * slider.animatingTo, posCalc = function() {
                    if (carousel) {
                        return special === "setTouch" ? pos : reverse && slider.animatingTo === slider.last ? 0 : reverse ? slider.limit - (slider.itemW + slider.vars.itemMargin) * slider.move * slider.animatingTo : slider.animatingTo === slider.last ? slider.limit : posCheck;
                    } else {
                        switch (special) {
                          case "setTotal":
                            return reverse ? (slider.count - 1 - slider.currentSlide + slider.cloneOffset) * pos : (slider.currentSlide + slider.cloneOffset) * pos;

                          case "setTouch":
                            return reverse ? pos : pos;

                          case "jumpEnd":
                            return reverse ? pos : slider.count * pos;

                          case "jumpStart":
                            return reverse ? slider.count * pos : pos;

                          default:
                            return pos;
                        }
                    }
                }();
                return posCalc * -1 + "px";
            }();
            if (slider.transitions) {
                target = vertical ? "translate3d(0," + target + ",0)" : "translate3d(" + target + ",0,0)";
                dur = dur !== undefined ? dur / 1e3 + "s" : "0s";
                slider.container.css("-" + slider.pfx + "-transition-duration", dur);
                slider.container.css("transition-duration", dur);
            }
            slider.args[slider.prop] = target;
            if (slider.transitions || dur === undefined) slider.container.css(slider.args);
            slider.container.css("transform", target);
        };
        slider.setup = function(type) {
            // SLIDE:
            if (!fade) {
                var sliderOffset, arr;
                if (type === "init") {
                    slider.viewport = $('<div class="' + namespace + 'viewport"></div>').css({
                        overflow: "hidden",
                        position: "relative"
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
                    if (type !== "init") slider.container.find(".clone").remove();
                    slider.container.append(methods.uniqueID(slider.slides.first().clone().addClass("clone")).attr("aria-hidden", "true")).prepend(methods.uniqueID(slider.slides.last().clone().addClass("clone")).attr("aria-hidden", "true"));
                }
                slider.newSlides = $(slider.vars.selector, slider);
                sliderOffset = reverse ? slider.count - 1 - slider.currentSlide + slider.cloneOffset : slider.currentSlide + slider.cloneOffset;
                // VERTICAL:
                if (vertical && !carousel) {
                    slider.container.height((slider.count + slider.cloneCount) * 200 + "%").css("position", "absolute").width("100%");
                    setTimeout(function() {
                        slider.newSlides.css({
                            display: "block"
                        });
                        slider.doMath();
                        slider.viewport.height(slider.h);
                        slider.setProps(sliderOffset * slider.h, "init");
                    }, type === "init" ? 100 : 0);
                } else {
                    slider.container.width((slider.count + slider.cloneCount) * 200 + "%");
                    slider.setProps(sliderOffset * slider.computedW, "init");
                    setTimeout(function() {
                        slider.doMath();
                        slider.newSlides.css({
                            width: slider.computedW,
                            "float": "left",
                            display: "block"
                        });
                        // SMOOTH HEIGHT:
                        if (slider.vars.smoothHeight) methods.smoothHeight();
                    }, type === "init" ? 100 : 0);
                }
            } else {
                // FADE:
                slider.slides.css({
                    width: "100%",
                    "float": "left",
                    marginRight: "-100%",
                    position: "relative"
                });
                if (type === "init") {
                    if (!touch) {
                        //slider.slides.eq(slider.currentSlide).fadeIn(slider.vars.animationSpeed, slider.vars.easing);
                        if (slider.vars.fadeFirstSlide == false) {
                            slider.slides.css({
                                opacity: 0,
                                display: "block",
                                zIndex: 1
                            }).eq(slider.currentSlide).css({
                                zIndex: 2
                            }).css({
                                opacity: 1
                            });
                        } else {
                            slider.slides.css({
                                opacity: 0,
                                display: "block",
                                zIndex: 1
                            }).eq(slider.currentSlide).css({
                                zIndex: 2
                            }).animate({
                                opacity: 1
                            }, slider.vars.animationSpeed, slider.vars.easing);
                        }
                    } else {
                        slider.slides.css({
                            opacity: 0,
                            display: "block",
                            webkitTransition: "opacity " + slider.vars.animationSpeed / 1e3 + "s ease",
                            zIndex: 1
                        }).eq(slider.currentSlide).css({
                            opacity: 1,
                            zIndex: 2
                        });
                    }
                }
                // SMOOTH HEIGHT:
                if (slider.vars.smoothHeight) methods.smoothHeight();
            }
            // !CAROUSEL:
            // CANDIDATE: active slide
            if (!carousel) slider.slides.removeClass(namespace + "active-slide").eq(slider.currentSlide).addClass(namespace + "active-slide");
            //FlexSlider: init() Callback
            slider.vars.init(slider);
        };
        slider.doMath = function() {
            var slide = slider.slides.first(), slideMargin = slider.vars.itemMargin, minItems = slider.vars.minItems, maxItems = slider.vars.maxItems;
            slider.w = slider.viewport === undefined ? slider.width() : slider.viewport.width();
            slider.h = slide.height();
            slider.boxPadding = slide.outerWidth() - slide.width();
            // CAROUSEL:
            if (carousel) {
                slider.itemT = slider.vars.itemWidth + slideMargin;
                slider.minW = minItems ? minItems * slider.itemT : slider.w;
                slider.maxW = maxItems ? maxItems * slider.itemT - slideMargin : slider.w;
                slider.itemW = slider.minW > slider.w ? (slider.w - slideMargin * (minItems - 1)) / minItems : slider.maxW < slider.w ? (slider.w - slideMargin * (maxItems - 1)) / maxItems : slider.vars.itemWidth > slider.w ? slider.w : slider.vars.itemWidth;
                slider.visible = Math.floor(slider.w / slider.itemW);
                slider.move = slider.vars.move > 0 && slider.vars.move < slider.visible ? slider.vars.move : slider.visible;
                slider.pagingCount = Math.ceil((slider.count - slider.visible) / slider.move + 1);
                slider.last = slider.pagingCount - 1;
                slider.limit = slider.pagingCount === 1 ? 0 : slider.vars.itemWidth > slider.w ? slider.itemW * (slider.count - 1) + slideMargin * (slider.count - 1) : (slider.itemW + slideMargin) * slider.count - slider.w - slideMargin;
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
                if (action === "add" && !carousel || slider.pagingCount > slider.controlNav.length) {
                    methods.controlNav.update("add");
                } else if (action === "remove" && !carousel || slider.pagingCount < slider.controlNav.length) {
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
                pos !== undefined ? slider.slides.eq(slider.count - pos).after($obj) : slider.container.prepend($obj);
            } else {
                pos !== undefined ? slider.slides.eq(pos).before($obj) : slider.container.append($obj);
            }
            // update currentSlide, animatingTo, controlNav, and directionNav
            slider.update(pos, "add");
            // update slider.slides
            slider.slides = $(slider.vars.selector + ":not(.clone)", slider);
            // re-setup the slider to accomdate new slide
            slider.setup();
            //FlexSlider: added() Callback
            slider.vars.added(slider);
        };
        slider.removeSlide = function(obj) {
            var pos = isNaN(obj) ? slider.slides.index($(obj)) : obj;
            // update count
            slider.count -= 1;
            slider.last = slider.count - 1;
            // remove slide
            if (isNaN(obj)) {
                $(obj, slider.slides).remove();
            } else {
                vertical && reverse ? slider.slides.eq(slider.last).remove() : slider.slides.eq(obj).remove();
            }
            // update currentSlide, animatingTo, controlNav, and directionNav
            slider.doMath();
            slider.update(pos, "remove");
            // update slider.slides
            slider.slides = $(slider.vars.selector + ":not(.clone)", slider);
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
    //FlexSlider: Default Settings
    $.flexslider.defaults = {
        namespace: "am-",
        //{NEW} String: Prefix string attached to the class of every element generated by the plugin
        selector: ".am-slides > li",
        //{NEW} Selector: Must match a simple pattern. '{container} > {slide}' -- Ignore pattern at your own peril
        animation: "slide",
        //String: Select your animation type, "fade" or "slide"
        easing: "swing",
        //{NEW} String: Determines the easing method used in jQuery transitions. jQuery easing plugin is supported!
        direction: "horizontal",
        //String: Select the sliding direction, "horizontal" or "vertical"
        reverse: false,
        //{NEW} Boolean: Reverse the animation direction
        animationLoop: true,
        //Boolean: Should the animation loop? If false, directionNav will received "disable" classes at either end
        smoothHeight: false,
        //{NEW} Boolean: Allow height of the slider to animate smoothly in horizontal mode
        startAt: 0,
        //Integer: The slide that the slider should start on. Array notation (0 = first slide)
        slideshow: true,
        //Boolean: Animate slider automatically
        slideshowSpeed: 5e3,
        //Integer: Set the speed of the slideshow cycling, in milliseconds
        animationSpeed: 600,
        //Integer: Set the speed of animations, in milliseconds
        initDelay: 0,
        //{NEW} Integer: Set an initialization delay, in milliseconds
        randomize: false,
        //Boolean: Randomize slide order
        thumbCaptions: false,
        //Boolean: Whether or not to put captions on thumbnails when using the "thumbnails" controlNav.
        // Usability features
        pauseOnAction: true,
        //Boolean: Pause the slideshow when interacting with control elements, highly recommended.
        pauseOnHover: false,
        //Boolean: Pause the slideshow when hovering over slider, then resume when no longer hovering
        pauseInvisible: true,
        //{NEW} Boolean: Pause the slideshow when tab is invisible, resume when visible. Provides better UX, lower CPU usage.
        useCSS: true,
        //{NEW} Boolean: Slider will use CSS3 transitions if available
        touch: true,
        //{NEW} Boolean: Allow touch swipe navigation of the slider on touch-enabled devices
        video: false,
        //{NEW} Boolean: If using video in the slider, will prevent CSS3 3D Transforms to avoid graphical glitches
        // Primary Controls
        controlNav: true,
        //Boolean: Create navigation for paging control of each clide? Note: Leave true for manualControls usage
        directionNav: true,
        //Boolean: Create navigation for previous/next navigation? (true/false)
        prevText: "Previous",
        //String: Set the text for the "previous" directionNav item
        nextText: "Next",
        //String: Set the text for the "next" directionNav item
        // Secondary Navigation
        keyboard: true,
        //Boolean: Allow slider navigating via keyboard left/right keys
        multipleKeyboard: false,
        //{NEW} Boolean: Allow keyboard navigation to affect multiple sliders. Default behavior cuts out keyboard navigation with more than one slider present.
        mousewheel: false,
        //{UPDATED} Boolean: Requires jquery.mousewheel.js (https://github.com/brandonaaron/jquery-mousewheel) - Allows slider navigating via mousewheel
        pausePlay: false,
        //Boolean: Create pause/play dynamic element
        pauseText: "Pause",
        //String: Set the text for the "pause" pausePlay item
        playText: "Play",
        //String: Set the text for the "play" pausePlay item
        // Special properties
        controlsContainer: "",
        //{UPDATED} jQuery Object/Selector: Declare which container the navigation elements should be appended too. Default container is the FlexSlider element. Example use would be $(".flexslider-container"). Property is ignored if given element is not found.
        manualControls: "",
        //{UPDATED} jQuery Object/Selector: Declare custom control navigation. Examples would be $(".flex-control-nav li") or "#tabs-nav li img", etc. The number of elements in your controlNav should match the number of slides/tabs.
        sync: "",
        //{NEW} Selector: Mirror the actions performed on this slider with another slider. Use with care.
        asNavFor: "",
        //{NEW} Selector: Internal property exposed for turning the slider into a thumbnail navigation for another slider
        // Carousel Options
        itemWidth: 0,
        //{NEW} Integer: Box-model width of individual carousel items, including horizontal borders and padding.
        itemMargin: 0,
        //{NEW} Integer: Margin between carousel items.
        minItems: 1,
        //{NEW} Integer: Minimum number of carousel items that should be visible. Items will resize fluidly when below this.
        maxItems: 0,
        //{NEW} Integer: Maxmimum number of carousel items that should be visible. Items will resize fluidly when above this limit.
        move: 0,
        //{NEW} Integer: Number of carousel items that should move on animation. If 0, slider will move all visible items.
        allowOneSlide: true,
        //{NEW} Boolean: Whether or not to allow a slider comprised of a single slide
        // Callback API
        start: function() {},
        //Callback: function(slider) - Fires when the slider loads the first slide
        before: function() {},
        //Callback: function(slider) - Fires asynchronously with each slider animation
        after: function() {},
        //Callback: function(slider) - Fires after each slider animation completes
        end: function() {},
        //Callback: function(slider) - Fires when the slider reaches the last slide (asynchronous)
        added: function() {},
        //{NEW} Callback: function(slider) - Fires after a slide is added
        removed: function() {},
        //{NEW} Callback: function(slider) - Fires after a slide is removed
        init: function() {}
    };
    //FlexSlider: Plugin Function
    $.fn.flexslider = function(options) {
        if (options === undefined) options = {};
        if (typeof options === "object") {
            return this.each(function() {
                var $this = $(this), selector = options.selector ? options.selector : ".am-slides > li", $slides = $this.find(selector);
                if ($slides.length === 1 && options.allowOneSlide === true || $slides.length === 0) {
                    $slides.animate({
                        opacity: 1
                    }, 400);
                    if (options.start) options.start($this);
                } else if ($this.data("flexslider") === undefined) {
                    new $.flexslider(this, options);
                }
            });
        } else {
            // Helper strings to quickly perform functions on the slider
            var $slider = $(this).data("flexslider");
            switch (options) {
              case "play":
                $slider.play();
                break;

              case "pause":
                $slider.pause();
                break;

              case "stop":
                $slider.stop();
                break;

              case "next":
                $slider.flexAnimate($slider.getTarget("next"), true);
                break;

              case "prev":
              case "previous":
                $slider.flexAnimate($slider.getTarget("prev"), true);
                break;

              default:
                if (typeof options === "number") $slider.flexAnimate(options, true);
            }
        }
    };
});
define("zepto.outerdemension", [], function(require, exports, module) {
    var $ = window.Zepto;
    // Create outerHeight and outerWidth methods
    [ "width", "height" ].forEach(function(dimension) {
        var offset, Dimension = dimension.replace(/./, function(m) {
            return m[0].toUpperCase();
        });
        $.fn["outer" + Dimension] = function(margin) {
            var elem = this;
            if (elem) {
                var size = elem[dimension]();
                var sides = {
                    width: [ "left", "right" ],
                    height: [ "top", "bottom" ]
                };
                sides[dimension].forEach(function(side) {
                    if (margin) size += parseInt(elem.css("margin-" + side), 10);
                });
                return size;
            } else {
                return null;
            }
        };
    });
});
define("zepto.pinchzoom", [], function(require, exports, module) {
    /**
     * @via http://rtp-ch.github.io/pinchzoom/
     * @license GNU General Public License,
     */
    var definePinchZoom = function(d) {
        var PinchZoom = function(h, g) {
            this.el = d(h);
            this.zoomFactor = 1;
            this.lastScale = 1;
            this.offset = {
                x: 0,
                y: 0
            };
            this.options = d.extend({}, this.defaults, g);
            this.setupMarkup();
            this.bindEvents();
            this.update();
        }, b = function(h, g) {
            return h + g;
        }, e = function(h, g) {
            return h > g - .01 && h < g + .01;
        };
        PinchZoom.prototype = {
            defaults: {
                tapZoomFactor: 2,
                zoomOutFactor: 1.3,
                animationDuration: 300,
                animationInterval: 5,
                maxZoom: 4,
                minZoom: .5,
                use2d: true
            },
            handleDragStart: function(g) {
                this.stopAnimation();
                this.lastDragPosition = false;
                this.hasInteraction = true;
                this.handleDrag(g);
            },
            handleDrag: function(g) {
                if (this.zoomFactor > 1) {
                    var h = this.getTouches(g)[0];
                    this.drag(h, this.lastDragPosition);
                    this.offset = this.sanitizeOffset(this.offset);
                    this.lastDragPosition = h;
                }
            },
            handleDragEnd: function() {
                this.end();
            },
            handleZoomStart: function(g) {
                this.stopAnimation();
                this.lastScale = 1;
                this.nthZoom = 0;
                this.lastZoomCenter = false;
                this.hasInteraction = true;
            },
            handleZoom: function(h, j) {
                var g = this.getTouchCenter(this.getTouches(h)), i = j / this.lastScale;
                this.lastScale = j;
                this.nthZoom += 1;
                if (this.nthZoom > 3) {
                    this.scale(i, g);
                    this.drag(g, this.lastZoomCenter);
                }
                this.lastZoomCenter = g;
            },
            handleZoomEnd: function() {
                this.end();
            },
            handleDoubleTap: function(i) {
                var g = this.getTouches(i)[0], h = this.zoomFactor > 1 ? 1 : this.options.tapZoomFactor, j = this.zoomFactor, k = function(l) {
                    this.scaleTo(j + l * (h - j), g);
                }.bind(this);
                if (this.hasInteraction) {
                    return;
                }
                if (j > h) {
                    g = this.getCurrentZoomCenter();
                }
                if (h > 1) {
                    this.options.doubleTapOutCallback && this.options.doubleTapOutCallback();
                } else {
                    this.options.doubleTapInCallback && this.options.doubleTapInCallback();
                }
                this.animate(this.options.animationDuration, this.options.animationInterval, k, this.swing);
            },
            sanitizeOffset: function(m) {
                var l = (this.zoomFactor - 1) * this.getContainerX(), k = (this.zoomFactor - 1) * this.getContainerY(), j = Math.max(l, 0), i = Math.max(k, 0), h = Math.min(l, 0), g = Math.min(k, 0);
                return {
                    x: Math.min(Math.max(m.x, h), j),
                    y: Math.min(Math.max(m.y, g), i)
                };
            },
            scaleTo: function(h, g) {
                this.scale(h / this.zoomFactor, g);
            },
            scale: function(h, g) {
                h = this.scaleZoomFactor(h);
                this.addOffset({
                    x: (h - 1) * (g.x + this.offset.x),
                    y: (h - 1) * (g.y + this.offset.y)
                });
            },
            scaleZoomFactor: function(g) {
                var h = this.zoomFactor;
                this.zoomFactor *= g;
                this.zoomFactor = Math.min(this.options.maxZoom, Math.max(this.zoomFactor, this.options.minZoom));
                return this.zoomFactor / h;
            },
            drag: function(g, h) {
                if (h) {
                    this.addOffset({
                        x: -(g.x - h.x),
                        y: -(g.y - h.y)
                    });
                }
            },
            getTouchCenter: function(g) {
                return this.getVectorAvg(g);
            },
            getVectorAvg: function(g) {
                return {
                    x: g.map(function(h) {
                        return h.x;
                    }).reduce(b) / g.length,
                    y: g.map(function(h) {
                        return h.y;
                    }).reduce(b) / g.length
                };
            },
            addOffset: function(g) {
                this.offset = {
                    x: this.offset.x + g.x,
                    y: this.offset.y + g.y
                };
            },
            sanitize: function() {
                if (this.zoomFactor < this.options.zoomOutFactor) {
                    this.zoomOutAnimation();
                } else {
                    if (this.isInsaneOffset(this.offset)) {
                        this.sanitizeOffsetAnimation();
                    }
                }
            },
            isInsaneOffset: function(h) {
                var g = this.sanitizeOffset(h);
                return g.x !== h.x || g.y !== h.y;
            },
            sanitizeOffsetAnimation: function() {
                var h = this.sanitizeOffset(this.offset), g = {
                    x: this.offset.x,
                    y: this.offset.y
                }, i = function(j) {
                    this.offset.x = g.x + j * (h.x - g.x);
                    this.offset.y = g.y + j * (h.y - g.y);
                    this.update();
                }.bind(this);
                this.animate(this.options.animationDuration, this.options.animationInterval, i, this.swing);
            },
            zoomOutAnimation: function() {
                var i = this.zoomFactor, h = 1, g = this.getCurrentZoomCenter(), j = function(k) {
                    this.scaleTo(i + k * (h - i), g);
                }.bind(this);
                this.animate(this.options.animationDuration, this.options.animationInterval, j, this.swing);
            },
            updateAspectRatio: function() {
                this.setContainerY(window.innerHeight);
            },
            getInitialZoomFactor: function() {
                return 1;
            },
            getAspectRatio: function() {
                return this.el.width() / this.el.height();
            },
            getCurrentZoomCenter: function() {
                var o = this.getContainerX(), h = o * this.zoomFactor, i = this.offset.x, l = h - i - o, q = i / l, n = q * o / (q + 1), m = this.getContainerY(), r = m * this.zoomFactor, g = this.offset.y, j = r - g - m, p = g / j, k = p * m / (p + 1);
                if (l === 0) {
                    n = o;
                }
                if (j === 0) {
                    k = m;
                }
                return {
                    x: n,
                    y: k
                };
            },
            canDrag: function() {
                return !e(this.zoomFactor, 1);
            },
            getTouches: function(h) {
                var g = this.container.offset();
                return Array.prototype.slice.call(h.touches).map(function(i) {
                    return {
                        x: i.pageX - g.left,
                        y: i.pageY - g.top
                    };
                });
            },
            animate: function(i, g, m, l, k) {
                var h = new Date().getTime(), j = function() {
                    if (!this.inAnimation) {
                        return;
                    }
                    var o = new Date().getTime() - h, n = o / i;
                    if (o >= i) {
                        m(1);
                        if (k) {
                            k();
                        }
                        this.update();
                        this.stopAnimation();
                        this.update();
                    } else {
                        if (l) {
                            n = l(n);
                        }
                        m(n);
                        this.update();
                        setTimeout(j, g);
                    }
                }.bind(this);
                this.inAnimation = true;
                j();
            },
            stopAnimation: function() {
                this.inAnimation = false;
            },
            swing: function(g) {
                return -Math.cos(g * Math.PI) / 2 + .5;
            },
            getContainerX: function() {
                return window.innerWidth;
            },
            getContainerY: function() {
                return window.innerHeight;
            },
            setContainerY: function(g) {
                this.el.width(window.innerWidth);
                this.el.height(window.innerHeight);
                return this.container.height(g);
            },
            setupMarkup: function() {
                this.container = d('<div class="pinch-zoom-container"></div>');
                this.el.before(this.container);
                this.container.append(this.el);
                this.container.css({
                    overflow: "hidden",
                    position: "relative"
                });
                this.el.css({
                    "-webkit-transform-origin": "0% 0%",
                    transformOrigin: "0% 0%",
                    position: "absolute"
                });
            },
            end: function() {
                this.hasInteraction = false;
                this.sanitize();
                this.update();
            },
            bindEvents: function() {
                c(this.container.get(0), this);
                d(window).bind("ortchange", this.ortHandle.bind(this));
            },
            isCached: function(h) {
                var g = document.createElement("img");
                g.src = h;
                var i = g.complete || g.width + g.height > 0;
                g = null;
                return i;
            },
            ortHandle: function() {
                this.zoomFactor = 1;
                this.offset = {
                    x: 0,
                    y: 0
                };
                this.update();
            },
            update: function() {
                if (this.updatePlaned) {
                    return;
                }
                this.updatePlaned = true;
                setTimeout(function() {
                    this.updatePlaned = false;
                    this.updateAspectRatio();
                    var k = this.getInitialZoomFactor() * this.zoomFactor, h = parseFloat(-this.offset.x / k).toFixed(4), m = parseFloat(-this.offset.y / k).toFixed(4), j = "scale3d(" + k + ", " + k + ",1) translate3d(" + h + "px," + m + "px,0px)", g = "scale(" + k + ", " + k + ") translate(" + h + "px," + m + "px)", i = function() {
                        if (this.clone) {
                            this.clone.remove();
                            delete this.clone;
                        }
                    }.bind(this);
                    if (!this.options.use2d || this.hasInteraction || this.inAnimation) {
                        this.is3d = true;
                        i();
                        this.el.css({
                            "-webkit-transform": j,
                            background: "rgba(0,0,0,0.9)",
                            transform: j
                        }).addClass("zooming");
                    } else {
                        if (this.is3d) {
                            var l = this.el.find("img").attr("src");
                            if (this.isCached(l)) {
                                this.clone = this.el.clone();
                                this.clone.css({
                                    "pointer-events": "none"
                                });
                                this.clone.appendTo(this.container);
                                setTimeout(i, 200);
                            }
                        }
                        this.el.css({
                            "-webkit-transform": g,
                            transform: g
                        }).removeClass("zooming");
                        this.is3d = false;
                    }
                }.bind(this), 0);
            }
        };
        var c = function(h, q) {
            var s = null, l = 0, j = null, u = null, i = 1, o = function(v, w) {
                if (s !== v) {
                    if (s && !v) {
                        switch (s) {
                          case "zoom":
                            q.handleZoomEnd(w);
                            break;

                          case "drag":
                            q.handleDragEnd(w);
                            break;
                        }
                    }
                    switch (v) {
                      case "zoom":
                        q.handleZoomStart(w);
                        break;

                      case "drag":
                        q.handleDragStart(w);
                        break;
                    }
                }
                s = v;
            }, n = function(v) {
                if (l === 2) {
                    o("zoom");
                } else {
                    if (l === 1 && q.canDrag()) {
                        o("drag", v);
                    } else {
                        o(null, v);
                    }
                }
            }, r = function(v) {
                return Array.prototype.slice.call(v).map(function(w) {
                    return {
                        x: w.pageX,
                        y: w.pageY
                    };
                });
            }, m = function(z, w) {
                var v, A;
                v = z.x - w.x;
                A = z.y - w.y;
                return Math.sqrt(v * v + A * A);
            }, k = function(y, x) {
                var v = m(y[0], y[1]), w = m(x[0], x[1]);
                return w / v;
            }, p = function(v) {
                v.stopPropagation();
                v.preventDefault();
            }, t = function(v) {
                var w = new Date().getTime();
                if (l > 1) {
                    j = null;
                }
                if (w - j < 400) {
                    p(v);
                    q.handleDoubleTap(v);
                    switch (s) {
                      case "zoom":
                        q.handleZoomEnd(v);
                        break;

                      case "drag":
                        q.handleDragEnd(v);
                        break;
                    }
                }
                if (l === 1) {
                    j = w;
                }
            }, g = true;
            h.addEventListener("touchstart", function(v) {
                g = true;
                i = q.zoomFactor, l = v.touches.length;
                t(v);
            });
            h.addEventListener("touchmove", function(v) {
                if (g) {
                    n(v);
                    if (s) {
                        p(v);
                    }
                    u = r(v.touches);
                } else {
                    switch (s) {
                      case "zoom":
                        q.handleZoom(v, k(u, r(v.touches)));
                        break;

                      case "drag":
                        q.handleDrag(v);
                        break;
                    }
                    if (s) {
                        p(v);
                        q.update();
                    }
                }
                g = false;
            });
            h.addEventListener("touchend", function(v) {
                if (s) {
                    p(v);
                }
                if (s == "zoom") {
                    if (q.zoomFactor >= i) {
                        q.options.zoomOutCallback && q.options.zoomOutCallback();
                    } else {
                        q.options.zoomInCallback && q.options.zoomInCallback();
                    }
                }
                l = v.touches.length;
                n(v);
            });
        };
        return PinchZoom;
    };
    module.exports = definePinchZoom(window.Zepto);
});
define("zepto.touchgallery", [ "./zepto.extend.touch", "zepto.pinchzoom" ], function(require, exports, module) {
    require("./zepto.extend.touch");
    // PinchZoom Plugin
    var PinchZoom = require("zepto.pinchzoom");
    var $ = window.Zepto;
    /**
     * @name        jQuery touchTouch plugin
     * @author        Martin Angelov
     * @version    1.0
     * @url            http://tutorialzine.com/2012/04/mobile-touch-gallery/
     * @license        MIT License
     */
    /* Private variables */
    var overlay = $('<div id="galleryOverlay">'), slider = $('<div id="gallerySlider">'), prevArrow = $('<a id="prevArrow"></a>'), nextArrow = $('<a id="nextArrow"></a>'), navControl = $('<ol class="nav-control"></ol>'), overlayVisible = false, msie = navigator.userAgent.indexOf("MSIE") > -1;
    /* Creating the plugin */
    $.fn.touchTouch = function() {
        var placeholders = $([]), index = 0, allitems = this, items = allitems, navControlItems = $([]);
        // Appending the markup to the page
        if ($("[data-am-gallery]").length) {
            overlay.hide().appendTo("body");
            slider.appendTo(overlay);
        }
        // Creating a placeholder for each image
        items.each(function(i) {
            placeholders = placeholders.add($('<div class="placeholder">'));
            navControlItems = navControlItems.add($("<li>" + (i + 1) + "</li>"));
        });
        navControl.append(navControlItems);
        overlay.append(navControl);
        // Hide the gallery if the background is touched / clicked
        slider.append(placeholders).on("click", function(e) {
            if (!$(e.target).is("img")) {
                hideOverlay();
            }
        });
        // Listen for touch events on the body and check if they
        // originated in #gallerySlider img - the images in the slider.
        $("body").on("touchstart", "#gallerySlider img", function(e) {
            var touch = e.originalEvent ? e.originalEvent : e, startX = touch.changedTouches[0].pageX;
            slider.on("touchmove", function(e) {
                e.preventDefault();
                touch = e.touches[0] || e.changedTouches[0];
                if (touch.pageX - startX > 10) {
                    slider.off("touchmove");
                    showPrevious();
                } else if (touch.pageX - startX < -10) {
                    slider.off("touchmove");
                    showNext();
                }
            });
            // Return false to prevent image
            // highlighting on Android
            return false;
        }).on("touchend", function() {
            slider.off("touchmove");
        });
        // for IE 10+
        if (window.PointerEvent || window.MSPointerEvent) {
            $("body").on("swipe", "#gallerySlider img", function(e) {
                e.preventDefault();
            }).on("swipeRight", "#gallerySlider img", function(e) {
                showPrevious();
            }).on("swipeLeft", "#gallerySlider img", function(e) {
                showNext();
            });
        }
        // Listening for clicks on the thumbnails
        items.on("click", function(e) {
            e.preventDefault();
            var $this = $(this), galleryName, selectorType, $closestGallery = $this.parent().closest("[data-gallery]");
            // Find gallery name and change items object to only have
            // that gallery
            //If gallery name given to each item
            if ($this.attr("data-gallery")) {
                galleryName = $this.attr("data-gallery");
                selectorType = "item";
            } else if ($closestGallery.length) {
                galleryName = $closestGallery.attr("data-gallery");
                selectorType = "ancestor";
            }
            //These statements kept seperate in case elements have data-gallery on both
            //items and ancestor. Ancestor will always win because of above statments.
            if (galleryName && selectorType == "item") {
                items = $("[data-gallery=" + galleryName + "]");
            } else if (galleryName && selectorType == "ancestor") {
                //Filter to check if item has an ancestory with data-gallery attribute
                items = items.filter(function() {
                    return $(this).parent().closest("[data-gallery]").length;
                });
            }
            // Find the position of this image
            // in the collection
            index = items.index(this);
            showOverlay(index);
            showImage(index);
            activeNavControl(index);
            // Preload the next image
            preload(index + 1);
            // Preload the previous
            preload(index - 1);
        });
        // If the browser does not have support
        // for touch, display the arrows
        if (!("ontouchstart" in window)) {
            overlay.append(prevArrow).append(nextArrow);
            prevArrow.click(function(e) {
                e.preventDefault();
                showPrevious();
            });
            nextArrow.click(function(e) {
                e.preventDefault();
                showNext();
            });
        }
        // Listen for arrow keys
        $(window).on("keydown", function(e) {
            var keyCode = e.keyCode;
            if (keyCode == 37) {
                showPrevious();
            } else if (keyCode == 39) {
                showNext();
            } else if (keyCode == 27) {
                hideOverlay();
            }
        });
        /* Private functions */
        function showOverlay(index) {
            // If the overlay is already shown, exit
            if (overlayVisible) {
                return false;
            }
            // Show the overlay
            overlay.show();
            setTimeout(function() {
                // Trigger the opacity CSS transition
                overlay.addClass("visible");
            }, 100);
            // Move the slider to the correct image
            offsetSlider(index);
            // Raise the visible flag
            overlayVisible = true;
        }
        function hideOverlay() {
            // If the overlay is not shown, exit
            if (!overlayVisible) {
                return false;
            }
            // Hide the overlay
            overlay.animate({
                opacity: 0,
                display: "none"
            }, 300).removeClass("visible");
            overlayVisible = false;
            //Clear preloaded items
            $(".placeholder").empty();
            //Reset possibly filtered items
            items = allitems;
        }
        function offsetSlider(index) {
            if (msie) {
                // windows phone 8 IE 显示有问题，单独处理
                slider.find(".placeholder").css({
                    display: "none"
                }).eq(index).css({
                    display: "inline-block"
                });
            } else {
                // This will trigger a smooth css transition
                slider.css("left", -index * 100 + "%");
            }
        }
        // Preload an image by its index in the items array
        function preload(index) {
            setTimeout(function() {
                showImage(index);
            }, 1e3);
        }
        // active nav control
        function activeNavControl(index) {
            var navItems = navControl.children("li");
            navItems.removeClass().eq(index).addClass("nav-active");
        }
        // Show image in the slider
        function showImage(index) {
            // If the index is outside the bonds of the array
            if (index < 0 || index >= items.length) {
                return false;
            }
            // Call the load function with the href attribute of the item
            loadImage(items.eq(index).attr("href"), function() {
                placeholders.eq(index).html(this).wrapInner('<div class="pinch-zoom"></div>');
                new PinchZoom(placeholders.eq(index).find(".pinch-zoom"), {});
            });
        }
        // Load the image and execute a callback function.
        // Returns a jQuery object
        function loadImage(src, callback) {
            var img = $("<img>").on("load", function() {
                callback.call(img);
            });
            img.attr("src", src);
        }
        function showNext() {
            // If this is not the last image
            if (index + 1 < items.length) {
                index++;
                offsetSlider(index);
                preload(index + 1);
                activeNavControl(index);
            } else {
                // Trigger the spring animation
                slider.addClass("rightSpring");
                setTimeout(function() {
                    slider.removeClass("rightSpring");
                }, 500);
            }
        }
        function showPrevious() {
            // If this is not the first image
            if (index > 0) {
                index--;
                offsetSlider(index);
                preload(index - 1);
                activeNavControl(index);
            } else {
                // Trigger the spring animation
                slider.addClass("leftSpring");
                setTimeout(function() {
                    slider.removeClass("leftSpring");
                }, 500);
            }
        }
    };
});
seajs.use(["accordion","core","divider","figure","footer","gallery","gotop","intro","list_news","map","menu","nav","navbar","pagination","paragraph","slider","sohucs","tabs","titlebar","ui.accordion","ui.add2home","ui.alert","ui.button","ui.collapse","ui.dimmer","ui.dropdown","ui.modal","ui.offcanvas","ui.popover","ui.progress","ui.scrollspy","ui.scrollspynav","ui.smooth-scroll","ui.sticky","util.cookie","util.fastclick","util.fullscreen","util.qrcode","zepto.extend.data","zepto.extend.fx","zepto.extend.selector","zepto.extend.touch","zepto.flexslider","zepto.outerdemension","zepto.pinchzoom","zepto.touchgallery"]);