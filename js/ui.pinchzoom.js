'use strict';

var $ = require('jquery');
var UI = require('./core');

/**
 * @via https://github.com/manuelstofer/pinchzoom/blob/master/src/pinchzoom.js
 * @license the MIT License.
 */

var definePinchZoom = function($) {
  /**
   * Pinch zoom using jQuery
   * @version 0.0.2
   * @author Manuel Stofer <mst@rtp.ch>
   * @param el
   * @param options
   * @constructor
   */
  var PinchZoom = function(el, options) {
      this.el = $(el);
      this.zoomFactor = 1;
      this.lastScale = 1;
      this.offset = {
        x: 0,
        y: 0
      };
      this.options = $.extend({}, this.defaults, options);
      this.setupMarkup();
      this.bindEvents();
      this.update();
      // default enable.
      this.enable();

    },
    sum = function(a, b) {
      return a + b;
    },
    isCloseTo = function(value, expected) {
      return value > expected - 0.01 && value < expected + 0.01;
    };

  PinchZoom.prototype = {

    defaults: {
      tapZoomFactor: 2,
      zoomOutFactor: 1.3,
      animationDuration: 300,
      animationInterval: 5,
      maxZoom: 5,
      minZoom: 0.5,
      lockDragAxis: false,
      use2d: false,
      zoomStartEventName: 'pz_zoomstart',
      zoomEndEventName: 'pz_zoomend',
      dragStartEventName: 'pz_dragstart',
      dragEndEventName: 'pz_dragend',
      doubleTapEventName: 'pz_doubletap'
    },

    /**
     * Event handler for 'dragstart'
     * @param event
     */
    handleDragStart: function(event) {
      this.el.trigger(this.options.dragStartEventName);
      this.stopAnimation();
      this.lastDragPosition = false;
      this.hasInteraction = true;
      this.handleDrag(event);
    },

    /**
     * Event handler for 'drag'
     * @param event
     */
    handleDrag: function(event) {

      if (this.zoomFactor > 1.0) {
        var touch = this.getTouches(event)[0];
        this.drag(touch, this.lastDragPosition);
        this.offset = this.sanitizeOffset(this.offset);
        this.lastDragPosition = touch;
      }
    },

    handleDragEnd: function() {
      this.el.trigger(this.options.dragEndEventName);
      this.end();
    },

    /**
     * Event handler for 'zoomstart'
     * @param event
     */
    handleZoomStart: function(event) {
      this.el.trigger(this.options.zoomStartEventName);
      this.stopAnimation();
      this.lastScale = 1;
      this.nthZoom = 0;
      this.lastZoomCenter = false;
      this.hasInteraction = true;
    },

    /**
     * Event handler for 'zoom'
     * @param event
     */
    handleZoom: function(event, newScale) {

      // a relative scale factor is used
      var touchCenter = this.getTouchCenter(this.getTouches(event)),
        scale = newScale / this.lastScale;
      this.lastScale = newScale;

      // the first touch events are thrown away since they are not precise
      this.nthZoom += 1;
      if (this.nthZoom > 3) {

        this.scale(scale, touchCenter);
        this.drag(touchCenter, this.lastZoomCenter);
      }
      this.lastZoomCenter = touchCenter;
    },

    handleZoomEnd: function() {
      this.el.trigger(this.options.zoomEndEventName);
      this.end();
    },

    /**
     * Event handler for 'doubletap'
     * @param event
     */
    handleDoubleTap: function(event) {
      var center = this.getTouches(event)[0],
        zoomFactor = this.zoomFactor > 1 ? 1 : this.options.tapZoomFactor,
        startZoomFactor = this.zoomFactor,
        updateProgress = (function(progress) {
          this.scaleTo(startZoomFactor + progress * (zoomFactor - startZoomFactor), center);
        }).bind(this);

      if (this.hasInteraction) {
        return;
      }
      if (startZoomFactor > zoomFactor) {
        center = this.getCurrentZoomCenter();
      }

      this.animate(this.options.animationDuration, this.options.animationInterval, updateProgress, this.swing);
      this.el.trigger(this.options.doubleTapEventName);
    },

    /**
     * Max / min values for the offset
     * @param offset
     * @return {Object} the sanitized offset
     */
    sanitizeOffset: function(offset) {
      var maxX = (this.zoomFactor - 1) * this.getContainerX(),
        maxY = (this.zoomFactor - 1) * this.getContainerY(),
        maxOffsetX = Math.max(maxX, 0),
        maxOffsetY = Math.max(maxY, 0),
        minOffsetX = Math.min(maxX, 0),
        minOffsetY = Math.min(maxY, 0);

      return {
        x: Math.min(Math.max(offset.x, minOffsetX), maxOffsetX),
        y: Math.min(Math.max(offset.y, minOffsetY), maxOffsetY)
      };
    },

    /**
     * Scale to a specific zoom factor (not relative)
     * @param zoomFactor
     * @param center
     */
    scaleTo: function(zoomFactor, center) {
      this.scale(zoomFactor / this.zoomFactor, center);
    },

    /**
     * Scales the element from specified center
     * @param scale
     * @param center
     */
    scale: function(scale, center) {
      scale = this.scaleZoomFactor(scale);
      this.addOffset({
        x: (scale - 1) * (center.x + this.offset.x),
        y: (scale - 1) * (center.y + this.offset.y)
      });
    },

    /**
     * Scales the zoom factor relative to current state
     * @param scale
     * @return the actual scale (can differ because of max min zoom factor)
     */
    scaleZoomFactor: function(scale) {
      var originalZoomFactor = this.zoomFactor;
      this.zoomFactor *= scale;
      this.zoomFactor = Math.min(this.options.maxZoom, Math.max(this.zoomFactor, this.options.minZoom));
      return this.zoomFactor / originalZoomFactor;
    },

    /**
     * Drags the element
     * @param center
     * @param lastCenter
     */
    drag: function(center, lastCenter) {
      if (lastCenter) {
        if (this.options.lockDragAxis) {
          // lock scroll to position that was changed the most
          if (Math.abs(center.x - lastCenter.x) > Math.abs(center.y - lastCenter.y)) {
            this.addOffset({
              x: -(center.x - lastCenter.x),
              y: 0
            });
          }
          else {
            this.addOffset({
              y: -(center.y - lastCenter.y),
              x: 0
            });
          }
        }
        else {
          this.addOffset({
            y: -(center.y - lastCenter.y),
            x: -(center.x - lastCenter.x)
          });
        }
      }
    },

    /**
     * Calculates the touch center of multiple touches
     * @param touches
     * @return {Object}
     */
    getTouchCenter: function(touches) {
      return this.getVectorAvg(touches);
    },

    /**
     * Calculates the average of multiple vectors (x, y values)
     */
    getVectorAvg: function(vectors) {
      return {
        x: vectors.map(function(v) {
          return v.x;
        }).reduce(sum) / vectors.length,
        y: vectors.map(function(v) {
          return v.y;
        }).reduce(sum) / vectors.length
      };
    },

    /**
     * Adds an offset
     * @param offset the offset to add
     * @return return true when the offset change was accepted
     */
    addOffset: function(offset) {
      this.offset = {
        x: this.offset.x + offset.x,
        y: this.offset.y + offset.y
      };
    },

    sanitize: function() {
      if (this.zoomFactor < this.options.zoomOutFactor) {
        this.zoomOutAnimation();
      } else if (this.isInsaneOffset(this.offset)) {
        this.sanitizeOffsetAnimation();
      }
    },

    /**
     * Checks if the offset is ok with the current zoom factor
     * @param offset
     * @return {Boolean}
     */
    isInsaneOffset: function(offset) {
      var sanitizedOffset = this.sanitizeOffset(offset);
      return sanitizedOffset.x !== offset.x ||
        sanitizedOffset.y !== offset.y;
    },

    /**
     * Creates an animation moving to a sane offset
     */
    sanitizeOffsetAnimation: function() {
      var targetOffset = this.sanitizeOffset(this.offset),
        startOffset = {
          x: this.offset.x,
          y: this.offset.y
        },
        updateProgress = (function(progress) {
          this.offset.x = startOffset.x + progress * (targetOffset.x - startOffset.x);
          this.offset.y = startOffset.y + progress * (targetOffset.y - startOffset.y);
          this.update();
        }).bind(this);

      this.animate(
        this.options.animationDuration,
        this.options.animationInterval,
        updateProgress,
        this.swing
      );
    },

    /**
     * Zooms back to the original position,
     * (no offset and zoom factor 1)
     */
    zoomOutAnimation: function() {
      var startZoomFactor = this.zoomFactor,
        zoomFactor = 1,
        center = this.getCurrentZoomCenter(),
        updateProgress = (function(progress) {
          this.scaleTo(startZoomFactor + progress * (zoomFactor - startZoomFactor), center);
        }).bind(this);

      this.animate(
        this.options.animationDuration,
        this.options.animationInterval,
        updateProgress,
        this.swing
      );
    },

    /**
     * Updates the aspect ratio
     */
    updateAspectRatio: function() {
      // this.setContainerY(this.getContainerX() / this.getAspectRatio());
      // @modified
      this.setContainerY()
    },

    /**
     * Calculates the initial zoom factor (for the element to fit into the container)
     * @return the initial zoom factor
     */
    getInitialZoomFactor: function() {
      // use .offsetWidth instead of width()
      // because jQuery-width() return the original width but Zepto-width() will calculate width with transform.
      // the same as .height()
      return this.container[0].offsetWidth / this.el[0].offsetWidth;
    },

    /**
     * Calculates the aspect ratio of the element
     * @return the aspect ratio
     */
    getAspectRatio: function() {
      return this.el[0].offsetWidth / this.el[0].offsetHeight;
    },

    /**
     * Calculates the virtual zoom center for the current offset and zoom factor
     * (used for reverse zoom)
     * @return {Object} the current zoom center
     */
    getCurrentZoomCenter: function() {

      // uses following formula to calculate the zoom center x value
      // offset_left / offset_right = zoomcenter_x / (container_x - zoomcenter_x)
      var length = this.container[0].offsetWidth * this.zoomFactor,
        offsetLeft = this.offset.x,
        offsetRight = length - offsetLeft - this.container[0].offsetWidth,
        widthOffsetRatio = offsetLeft / offsetRight,
        centerX = widthOffsetRatio * this.container[0].offsetWidth / (widthOffsetRatio + 1),

      // the same for the zoomcenter y
        height = this.container[0].offsetHeight * this.zoomFactor,
        offsetTop = this.offset.y,
        offsetBottom = height - offsetTop - this.container[0].offsetHeight,
        heightOffsetRatio = offsetTop / offsetBottom,
        centerY = heightOffsetRatio * this.container[0].offsetHeight / (heightOffsetRatio + 1);

      // prevents division by zero
      if (offsetRight === 0) {
        centerX = this.container[0].offsetWidth;
      }
      if (offsetBottom === 0) {
        centerY = this.container[0].offsetHeight;
      }

      return {
        x: centerX,
        y: centerY
      };
    },

    canDrag: function() {
      return !isCloseTo(this.zoomFactor, 1);
    },

    /**
     * Returns the touches of an event relative to the container offset
     * @param event
     * @return array touches
     */
    getTouches: function(event) {
      var position = this.container.offset();
      return Array.prototype.slice.call(event.touches).map(function(touch) {
        return {
          x: touch.pageX - position.left,
          y: touch.pageY - position.top
        };
      });
    },

    /**
     * Animation loop
     * does not support simultaneous animations
     * @param duration
     * @param interval
     * @param framefn
     * @param timefn
     * @param callback
     */
    animate: function(duration, interval, framefn, timefn, callback) {
      var startTime = new Date().getTime(),
        renderFrame = (function() {
          if (!this.inAnimation) {
            return;
          }
          var frameTime = new Date().getTime() - startTime,
            progress = frameTime / duration;
          if (frameTime >= duration) {
            framefn(1);
            if (callback) {
              callback();
            }
            this.update();
            this.stopAnimation();
            this.update();
          } else {
            if (timefn) {
              progress = timefn(progress);
            }
            framefn(progress);
            this.update();
            setTimeout(renderFrame, interval);
          }
        }).bind(this);
      this.inAnimation = true;
      renderFrame();
    },

    /**
     * Stops the animation
     */
    stopAnimation: function() {
      this.inAnimation = false;
    },

    /**
     * Swing timing function for animations
     * @param p
     * @return {Number}
     */
    swing: function(p) {
      return -Math.cos(p * Math.PI) / 2 + 0.5;
    },

    getContainerX: function() {
      // return this.container[0].offsetWidth;
      // @modified
      return window.innerWidth
    },

    getContainerY: function() {
      // return this.container[0].offsetHeight;
      // @modified
      return window.innerHeight
    },

    setContainerY: function(y) {
      // return this.container.height(y);
      // @modified
      var t = window.innerHeight;
      return this.el.css({height: t}), this.container.height(t);
    },

    /**
     * Creates the expected html structure
     */
    setupMarkup: function() {
      this.container = $('<div class="pinch-zoom-container"></div>');
      this.el.before(this.container);
      this.container.append(this.el);

      this.container.css({
        'overflow': 'hidden',
        'position': 'relative'
      });

      // Zepto doesn't recognize `webkitTransform..` style
      this.el.css({
        '-webkit-transform-origin': '0% 0%',
        '-moz-transform-origin': '0% 0%',
        '-ms-transform-origin': '0% 0%',
        '-o-transform-origin': '0% 0%',
        'transform-origin': '0% 0%',
        'position': 'absolute'
      });
    },

    end: function() {
      this.hasInteraction = false;
      this.sanitize();
      this.update();
    },

    /**
     * Binds all required event listeners
     */
    bindEvents: function() {
      detectGestures(this.container.get(0), this);
      // Zepto and jQuery both know about `on`
      $(window).on('resize', this.update.bind(this));
      $(this.el).find('img').on('load', this.update.bind(this));
    },

    /**
     * Updates the css values according to the current zoom factor and offset
     */
    update: function() {

      if (this.updatePlaned) {
        return;
      }
      this.updatePlaned = true;

      setTimeout((function() {
        this.updatePlaned = false;
        this.updateAspectRatio();

        var zoomFactor = this.getInitialZoomFactor() * this.zoomFactor,
          offsetX = -this.offset.x / zoomFactor,
          offsetY = -this.offset.y / zoomFactor,
          transform3d = 'scale3d(' + zoomFactor + ', ' + zoomFactor + ',1) ' +
            'translate3d(' + offsetX + 'px,' + offsetY + 'px,0px)',
          transform2d = 'scale(' + zoomFactor + ', ' + zoomFactor + ') ' +
            'translate(' + offsetX + 'px,' + offsetY + 'px)',
          removeClone = (function() {
            if (this.clone) {
              this.clone.remove();
              delete this.clone;
            }
          }).bind(this);

        // Scale 3d and translate3d are faster (at least on ios)
        // but they also reduce the quality.
        // PinchZoom uses the 3d transformations during interactions
        // after interactions it falls back to 2d transformations
        if (!this.options.use2d || this.hasInteraction || this.inAnimation) {
          this.is3d = true;
          removeClone();
          this.el.css({
            '-webkit-transform': transform3d,
            '-o-transform': transform2d,
            '-ms-transform': transform2d,
            '-moz-transform': transform2d,
            'transform': transform3d
          });
        } else {

          // When changing from 3d to 2d transform webkit has some glitches.
          // To avoid this, a copy of the 3d transformed element is displayed in the
          // foreground while the element is converted from 3d to 2d transform
          if (this.is3d) {
            this.clone = this.el.clone();
            this.clone.css('pointer-events', 'none');
            this.clone.appendTo(this.container);
            setTimeout(removeClone, 200);
          }
          this.el.css({
            '-webkit-transform': transform2d,
            '-o-transform': transform2d,
            '-ms-transform': transform2d,
            '-moz-transform': transform2d,
            'transform': transform2d
          });
          this.is3d = false;
        }
      }).bind(this), 0);
    },

    /**
     * Enables event handling for gestures
     */
    enable: function() {
      this.enabled = true;
    },

    /**
     * Disables event handling for gestures
     */
    disable: function() {
      this.enabled = false;
    }
  };

  var detectGestures = function(el, target) {
    var interaction = null,
      fingers = 0,
      lastTouchStart = null,
      startTouches = null,

      setInteraction = function(newInteraction, event) {
        if (interaction !== newInteraction) {

          if (interaction && !newInteraction) {
            switch (interaction) {
              case "zoom":
                target.handleZoomEnd(event);
                break;
              case 'drag':
                target.handleDragEnd(event);
                break;
            }
          }

          switch (newInteraction) {
            case 'zoom':
              target.handleZoomStart(event);
              break;
            case 'drag':
              target.handleDragStart(event);
              break;
          }
        }
        interaction = newInteraction;
      },

      updateInteraction = function(event) {
        if (fingers === 2) {
          setInteraction('zoom');
        } else if (fingers === 1 && target.canDrag()) {
          setInteraction('drag', event);
        } else {
          setInteraction(null, event);
        }
      },

      targetTouches = function(touches) {
        return Array.prototype.slice.call(touches).map(function(touch) {
          return {
            x: touch.pageX,
            y: touch.pageY
          };
        });
      },

      getDistance = function(a, b) {
        var x, y;
        x = a.x - b.x;
        y = a.y - b.y;
        return Math.sqrt(x * x + y * y);
      },

      calculateScale = function(startTouches, endTouches) {
        var startDistance = getDistance(startTouches[0], startTouches[1]),
          endDistance = getDistance(endTouches[0], endTouches[1]);
        return endDistance / startDistance;
      },

      cancelEvent = function(event) {
        event.stopPropagation();
        event.preventDefault();
      },

      detectDoubleTap = function(event) {
        var time = (new Date()).getTime();

        if (fingers > 1) {
          lastTouchStart = null;
        }

        if (time - lastTouchStart < 300) {
          cancelEvent(event);

          target.handleDoubleTap(event);
          switch (interaction) {
            case "zoom":
              target.handleZoomEnd(event);
              break;
            case 'drag':
              target.handleDragEnd(event);
              break;
          }
        }

        if (fingers === 1) {
          lastTouchStart = time;
        }
      },
      firstMove = true;

    el.addEventListener('touchstart', function(event) {
      if (target.enabled) {
        firstMove = true;
        fingers = event.touches.length;
        detectDoubleTap(event);
      }
    });

    el.addEventListener('touchmove', function(event) {
      if (target.enabled) {
        if (firstMove) {
          updateInteraction(event);
          if (interaction) {
            cancelEvent(event);
          }
          startTouches = targetTouches(event.touches);
        } else {
          switch (interaction) {
            case 'zoom':
              target.handleZoom(event, calculateScale(startTouches, targetTouches(event.touches)));
              break;
            case 'drag':
              target.handleDrag(event);
              break;
          }
          if (interaction) {
            cancelEvent(event);
            target.update();
          }
        }

        firstMove = false;
      }
    });

    el.addEventListener('touchend', function(event) {
      if (target.enabled) {
        fingers = event.touches.length;
        updateInteraction(event);
      }
    });
  };

  return PinchZoom;
};

module.exports = UI.pichzoom = definePinchZoom($);
