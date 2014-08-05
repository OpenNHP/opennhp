define(function(require, exports, module) {
    require('./zepto.extend.touch');

    // PinchZoom Plugin
    var PinchZoom = require('zepto.pinchzoom');

    var $ = window.Zepto;


    /**
     * @name        jQuery touchTouch plugin
     * @author        Martin Angelov
     * @version    1.0
     * @url            http://tutorialzine.com/2012/04/mobile-touch-gallery/
     * @license        MIT License
     */

    /* Private variables */
    var overlay = $('<div id="galleryOverlay">'),
        slider = $('<div id="gallerySlider">'),
        prevArrow = $('<a id="prevArrow"></a>'),
        nextArrow = $('<a id="nextArrow"></a>'),
        navControl = $('<ol class="nav-control"></ol>'),
        overlayVisible = false,
        msie = (navigator.userAgent.indexOf('MSIE') > -1);


    /* Creating the plugin */

    $.fn.touchTouch = function() {

        var placeholders = $([]),
            index = 0,
            allitems = this,
            items = allitems,
            navControlItems = $([]);

        // Appending the markup to the page
        if ($('[data-am-gallery]').length) {
            overlay.hide().appendTo('body');
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
        slider.append(placeholders).on('click', function(e) {
            if (!$(e.target).is('img')) {
                hideOverlay();
            }
        });


        // Listen for touch events on the body and check if they
        // originated in #gallerySlider img - the images in the slider.
        $('body').on('touchstart', '#gallerySlider img',function(e) {
            var touch = e.originalEvent ? e.originalEvent : e,
                startX = touch.changedTouches[0].pageX;

            slider.on('touchmove', function(e) {

                e.preventDefault();

                touch = e.touches[0] || e.changedTouches[0];


                if (touch.pageX - startX > 10) {

                    slider.off('touchmove');
                    showPrevious();
                }

                else if (touch.pageX - startX < -10) {

                    slider.off('touchmove');
                    showNext();
                }
            });

            // Return false to prevent image
            // highlighting on Android
            return false;

        }).on('touchend', function() {
                slider.off('touchmove');
            });

        // for IE 10+
        if (window.PointerEvent || window.MSPointerEvent) {
            $('body').on('swipe', '#gallerySlider img', function(e) {
                e.preventDefault();
            }).on('swipeRight', '#gallerySlider img', function(e) {
                showPrevious();
            }).on('swipeLeft', '#gallerySlider img', function(e) {
                    showNext();
                });
        }

        // Listening for clicks on the thumbnails
        items.on('click', function(e) {

            e.preventDefault();

            var $this = $(this),
                galleryName,
                selectorType,
                $closestGallery = $this.parent().closest('[data-gallery]');

            // Find gallery name and change items object to only have
            // that gallery

            //If gallery name given to each item
            if ($this.attr('data-gallery')) {

                galleryName = $this.attr('data-gallery');
                selectorType = 'item';

                //If gallery name given to some ancestor
            } else if ($closestGallery.length) {

                galleryName = $closestGallery.attr('data-gallery');
                selectorType = 'ancestor';

            }

            //These statements kept seperate in case elements have data-gallery on both
            //items and ancestor. Ancestor will always win because of above statments.
            if (galleryName && selectorType == 'item') {

                items = $('[data-gallery=' + galleryName + ']');

            } else if (galleryName && selectorType == 'ancestor') {

                //Filter to check if item has an ancestory with data-gallery attribute
                items = items.filter(function() {

                    return $(this).parent().closest('[data-gallery]').length;

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
        if (!('ontouchstart' in window)) {
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
        $(window).on('keydown', function(e) {
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
                overlay.addClass('visible');
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
            overlay.animate({opacity: 0, display: 'none'}, 300).removeClass('visible');
            overlayVisible = false;

            //Clear preloaded items
            $('.placeholder').empty();

            //Reset possibly filtered items
            items = allitems;
        }

        function offsetSlider(index) {
            if (msie) {
                // windows phone 8 IE 显示有问题，单独处理
                slider.find('.placeholder').css({display: 'none'}).eq(index).css({display: 'inline-block'});
            } else {
                // This will trigger a smooth css transition
                slider.css('left', (-index * 100) + '%');
                /*
                 css({
                 "webkitTransform":"",
                 "MozTransform":"",
                 "msTransform":"",
                 "OTransform":"",
                 "transform":""
                 });*/
            }
        }

        // Preload an image by its index in the items array
        function preload(index) {
            setTimeout(function() {
                showImage(index);
            }, 1000);
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
            loadImage(items.eq(index).attr('href'), function() {
                placeholders.eq(index).html(this).wrapInner('<div class="pinch-zoom"></div>');
                new PinchZoom(placeholders.eq(index).find('.pinch-zoom'), {});
            });
        }

        // Load the image and execute a callback function.
        // Returns a jQuery object
        function loadImage(src, callback) {
            var img = $('<img>').on('load', function() {
                callback.call(img);
            });

            img.attr('src', src);
        }

        function showNext() {
            // If this is not the last image
            if (index + 1 < items.length) {
                index++;
                offsetSlider(index);
                preload(index + 1);
                activeNavControl(index)
            }

            else {
                // Trigger the spring animation
                slider.addClass('rightSpring');
                setTimeout(function() {
                    slider.removeClass('rightSpring');
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
            }

            else {
                // Trigger the spring animation
                slider.addClass('leftSpring');
                setTimeout(function() {
                    slider.removeClass('leftSpring');
                }, 500);
            }
        }
    };
});