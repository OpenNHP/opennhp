define(function(require, exports, module) {

    // via http://rtp-ch.github.io/pinchzoom/

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
                this.update()
            },
            b = function(h, g) {
                return h + g
            },
            e = function(h, g) {
                return h > g - 0.01 && h < g + 0.01
            };
        PinchZoom.prototype = {
            defaults: {
                tapZoomFactor: 2,
                zoomOutFactor: 1.3,
                animationDuration: 300,
                animationInterval: 5,
                maxZoom: 4,
                minZoom: 0.5,
                use2d: true
            },
            handleDragStart: function(g) {
                this.stopAnimation();
                this.lastDragPosition = false;
                this.hasInteraction = true;
                this.handleDrag(g)
            },
            handleDrag: function(g) {
                if (this.zoomFactor > 1) {
                    var h = this.getTouches(g)[0];
                    this.drag(h, this.lastDragPosition);
                    this.offset = this.sanitizeOffset(this.offset);
                    this.lastDragPosition = h
                }
            },
            handleDragEnd: function() {
                this.end()
            },
            handleZoomStart: function(g) {
                this.stopAnimation();
                this.lastScale = 1;
                this.nthZoom = 0;
                this.lastZoomCenter = false;
                this.hasInteraction = true
            },
            handleZoom: function(h, j) {
                var g = this.getTouchCenter(this.getTouches(h)),
                    i = j / this.lastScale;
                this.lastScale = j;
                this.nthZoom += 1;
                if (this.nthZoom > 3) {
                    this.scale(i, g);
                    this.drag(g, this.lastZoomCenter)
                }
                this.lastZoomCenter = g
            },
            handleZoomEnd: function() {
                this.end()
            },
            handleDoubleTap: function(i) {
                var g = this.getTouches(i)[0],
                    h = this.zoomFactor > 1 ? 1 : this.options.tapZoomFactor,
                    j = this.zoomFactor,
                    k = (function(l) {
                        this.scaleTo(j + l * (h - j), g)
                    }).bind(this);
                if (this.hasInteraction) {
                    return
                }
                if (j > h) {
                    g = this.getCurrentZoomCenter()
                }
                if (h > 1) {
                    this.options.doubleTapOutCallback && this.options.doubleTapOutCallback()
                } else {
                    this.options.doubleTapInCallback && this.options.doubleTapInCallback()
                }
                this.animate(this.options.animationDuration, this.options.animationInterval, k, this.swing)
            },
            sanitizeOffset: function(m) {
                var l = (this.zoomFactor - 1) * this.getContainerX(),
                    k = (this.zoomFactor - 1) * this.getContainerY(),
                    j = Math.max(l, 0),
                    i = Math.max(k, 0),
                    h = Math.min(l, 0),
                    g = Math.min(k, 0);
                return {
                    x: Math.min(Math.max(m.x, h), j),
                    y: Math.min(Math.max(m.y, g), i)
                }
            },
            scaleTo: function(h, g) {
                this.scale(h / this.zoomFactor, g)
            },
            scale: function(h, g) {
                h = this.scaleZoomFactor(h);
                this.addOffset({
                    x: (h - 1) * (g.x + this.offset.x),
                    y: (h - 1) * (g.y + this.offset.y)
                })
            },
            scaleZoomFactor: function(g) {
                var h = this.zoomFactor;
                this.zoomFactor *= g;
                this.zoomFactor = Math.min(this.options.maxZoom, Math.max(this.zoomFactor, this.options.minZoom));
                return this.zoomFactor / h
            },
            drag: function(g, h) {
                if (h) {
                    this.addOffset({
                        x: -(g.x - h.x),
                        y: -(g.y - h.y)
                    })
                }
            },
            getTouchCenter: function(g) {
                return this.getVectorAvg(g)
            },
            getVectorAvg: function(g) {
                return {
                    x: g.map(function(h) {
                        return h.x
                    }).reduce(b) / g.length,
                    y: g.map(function(h) {
                        return h.y
                    }).reduce(b) / g.length
                }
            },
            addOffset: function(g) {
                this.offset = {
                    x: this.offset.x + g.x,
                    y: this.offset.y + g.y
                }
            },
            sanitize: function() {
                if (this.zoomFactor < this.options.zoomOutFactor) {
                    this.zoomOutAnimation()
                } else {
                    if (this.isInsaneOffset(this.offset)) {
                        this.sanitizeOffsetAnimation()
                    }
                }
            },
            isInsaneOffset: function(h) {
                var g = this.sanitizeOffset(h);
                return g.x !== h.x || g.y !== h.y
            },
            sanitizeOffsetAnimation: function() {
                var h = this.sanitizeOffset(this.offset),
                    g = {
                        x: this.offset.x,
                        y: this.offset.y
                    },
                    i = (function(j) {
                        this.offset.x = g.x + j * (h.x - g.x);
                        this.offset.y = g.y + j * (h.y - g.y);
                        this.update()
                    }).bind(this);
                this.animate(this.options.animationDuration, this.options.animationInterval, i, this.swing)
            },
            zoomOutAnimation: function() {
                var i = this.zoomFactor,
                    h = 1,
                    g = this.getCurrentZoomCenter(),
                    j = (function(k) {
                        this.scaleTo(i + k * (h - i), g)
                    }).bind(this);
                this.animate(this.options.animationDuration, this.options.animationInterval, j, this.swing)
            },
            updateAspectRatio: function() {
                this.setContainerY(window.innerHeight)
            },
            getInitialZoomFactor: function() {
                return 1
            },
            getAspectRatio: function() {
                return this.el.width() / this.el.height()
            },
            getCurrentZoomCenter: function() {
                var o = this.getContainerX(),
                    h = o * this.zoomFactor,
                    i = this.offset.x,
                    l = h - i - o,
                    q = i / l,
                    n = q * o / (q + 1),
                    m = this.getContainerY(),
                    r = m * this.zoomFactor,
                    g = this.offset.y,
                    j = r - g - m,
                    p = g / j,
                    k = p * m / (p + 1);
                if (l === 0) {
                    n = o
                }
                if (j === 0) {
                    k = m
                }
                return {
                    x: n,
                    y: k
                }
            },
            canDrag: function() {
                return !e(this.zoomFactor, 1)
            },
            getTouches: function(h) {
                var g = this.container.offset();
                return Array.prototype.slice.call(h.touches).map(function(i) {
                    return {
                        x: i.pageX - g.left,
                        y: i.pageY - g.top
                    }
                })
            },
            animate: function(i, g, m, l, k) {
                var h = new Date().getTime(),
                    j = (function() {
                        if (!this.inAnimation) {
                            return
                        }
                        var o = new Date().getTime() - h,
                            n = o / i;
                        if (o >= i) {
                            m(1);
                            if (k) {
                                k()
                            }
                            this.update();
                            this.stopAnimation();
                            this.update()
                        } else {
                            if (l) {
                                n = l(n)
                            }
                            m(n);
                            this.update();
                            setTimeout(j, g)
                        }
                    }).bind(this);
                this.inAnimation = true;
                j()
            },
            stopAnimation: function() {
                this.inAnimation = false
            },
            swing: function(g) {
                return -Math.cos(g * Math.PI) / 2 + 0.5
            },
            getContainerX: function() {
                return window.innerWidth
            },
            getContainerY: function() {
                return window.innerHeight
            },
            setContainerY: function(g) {
                this.el.width(window.innerWidth);
                this.el.height(window.innerHeight);
                return this.container.height(g)
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
                })
            },
            end: function() {
                this.hasInteraction = false;
                this.sanitize();
                this.update()
            },
            bindEvents: function() {
                c(this.container.get(0), this);
                d(window).bind("ortchange", this.ortHandle.bind(this))
            },
            isCached: function(h) {
                var g = document.createElement("img");
                g.src = h;
                var i = g.complete || g.width + g.height > 0;
                g = null;
                return i
            },
            ortHandle: function() {
                this.zoomFactor = 1;
                this.offset = {
                    x: 0,
                    y: 0
                };
                this.update()
            },
            update: function() {
                if (this.updatePlaned) {
                    return
                }
                this.updatePlaned = true;
                setTimeout((function() {
                    this.updatePlaned = false;
                    this.updateAspectRatio();
                    var k = this.getInitialZoomFactor() * this.zoomFactor,
                        h = parseFloat(-this.offset.x / k).toFixed(4),
                        m = parseFloat(-this.offset.y / k).toFixed(4),
                        j = "scale3d(" + k + ", " + k + ",1) translate3d(" + h + "px," + m + "px,0px)",
                        g = "scale(" + k + ", " + k + ") translate(" + h + "px," + m + "px)",
                        i = (function() {
                            if (this.clone) {
                                this.clone.remove();
                                delete this.clone
                            }
                        }).bind(this);
                    if (!this.options.use2d || this.hasInteraction || this.inAnimation) {
                        this.is3d = true;
                        i();
                        this.el.css({
                            "-webkit-transform": j,
                            background: "rgba(0,0,0,0.9)",
                            transform: j
                        }).addClass("zooming")
                    } else {
                        if (this.is3d) {
                            var l = this.el.find("img").attr("src");
                            if (this.isCached(l)) {
                                this.clone = this.el.clone();
                                this.clone.css({
                                    "pointer-events": "none"
                                });
                                this.clone.appendTo(this.container);
                                setTimeout(i, 200)
                            }
                        }
                        this.el.css({
                            "-webkit-transform": g,
                            transform: g
                        }).removeClass("zooming");
                        this.is3d = false
                    }
                }).bind(this), 0)
            }
        };
        var c = function(h, q) {
            var s = null,
                l = 0,
                j = null,
                u = null,
                i = 1,
                o = function(v, w) {
                    if (s !== v) {
                        if (s && !v) {
                            switch (s) {
                                case "zoom":
                                    q.handleZoomEnd(w);
                                    break;
                                case "drag":
                                    q.handleDragEnd(w);
                                    break
                            }
                        }
                        switch (v) {
                            case "zoom":
                                q.handleZoomStart(w);
                                break;
                            case "drag":
                                q.handleDragStart(w);
                                break
                        }
                    }
                    s = v
                },
                n = function(v) {
                    if (l === 2) {
                        o("zoom")
                    } else {
                        if (l === 1 && q.canDrag()) {
                            o("drag", v)
                        } else {
                            o(null, v)
                        }
                    }
                },
                r = function(v) {
                    return Array.prototype.slice.call(v).map(function(w) {
                        return {
                            x: w.pageX,
                            y: w.pageY
                        }
                    })
                },
                m = function(z, w) {
                    var v, A;
                    v = z.x - w.x;
                    A = z.y - w.y;
                    return Math.sqrt(v * v + A * A)
                },
                k = function(y, x) {
                    var v = m(y[0], y[1]),
                        w = m(x[0], x[1]);
                    return w / v
                },
                p = function(v) {
                    v.stopPropagation();
                    v.preventDefault()
                },
                t = function(v) {
                    var w = (new Date()).getTime();
                    if (l > 1) {
                        j = null
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
                                break
                        }
                    }
                    if (l === 1) {
                        j = w
                    }
                },
                g = true;
            h.addEventListener("touchstart", function(v) {
                g = true;
                i = q.zoomFactor, l = v.touches.length;
                t(v)
            });
            h.addEventListener("touchmove", function(v) {
                if (g) {
                    n(v);
                    if (s) {
                        p(v)
                    }
                    u = r(v.touches)
                } else {
                    switch (s) {
                        case "zoom":
                            q.handleZoom(v, k(u, r(v.touches)));
                            break;
                        case "drag":
                            q.handleDrag(v);
                            break
                    }
                    if (s) {
                        p(v);
                        q.update()
                    }
                }
                g = false
            });
            h.addEventListener("touchend", function(v) {
                if (s) {
                    p(v)
                }
                if (s == "zoom") {
                    if (q.zoomFactor >= i) {
                        q.options.zoomOutCallback && q.options.zoomOutCallback()
                    } else {
                        q.options.zoomInCallback && q.options.zoomInCallback()
                    }
                }
                l = v.touches.length;
                n(v)
            })
        };
        return PinchZoom
    };

    module.exports = definePinchZoom(window.Zepto);
});