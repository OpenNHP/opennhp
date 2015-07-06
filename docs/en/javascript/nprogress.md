---
id: nprogress
title: 加载进度条
titleEn: NProgress
prev: javascript/popover.html
next: javascript/slider.html
source: js/ui.progress.js
doc: docs/javascript/nprogress.md
---

# Progress
---

Like progress bar in Google, Youtube and Medium. Via [NProgress](http://ricostacruz.com/nprogress)。

## Usage
----------

Call `start()` and `done()` to control the progress bar.

```javascript
$.AMUI.progress.start();
$.AMUI.progress.done();
```

`````html
<button id="np-s" class="am-btn am-btn-primary">$.AMUI.progress.start();</button> <button id="np-d" class="am-btn am-btn-success">$.AMUI.progress.done();</button>
`````
```html
<button id="np-s" class="am-btn am-btn-primary">$.AMUI.progress.start();</button>
<button id="np-d" class="am-btn am-btn-success">$.AMUI.progress.done();</button>
```
```js
$(function(){
  var progress = $.AMUI.progress;

  $('#np-s').on('click', function() {
    progress.start();
  });

  $('#np-d').on('click', function() {
    progress.done();
  });
});
```

Progress can be called in event callback, if using libraries like [Turbolinks] 1.3.0+.

~~~ js
$(document).on('page:fetch',   function() { $.AMUI.progress.start(); });
$(document).on('page:change',  function() { $.AMUI.progress.done(); });
$(document).on('page:restore', function() { $.AMUI.progress.remove(); });
~~~

Where to use?
----------

 * Use progress bar in Ajax applications. Bind to `ajaxStart` and
 `ajaxStop` events in Zepto(jQuery).

 * You can use progress bar even if not using Turbolinks/Pjax. Just bind progress bar to `$(document).ready` and `$(window).load`.

Advanced Usage
--------------

**Percentage:** Call `.set(n)` to set the percentage. *n* is in `0..1`.

~~~ js
$.AMUI.progress.set(0.0);     // Sorta same as .start()
$.AMUI.progress.set(0.4);
$.AMUI.progress.set(1.0);     // Sorta same as .done()
~~~

**Increase**: Call `.inc()` to increase a random value to progress bar, but never to 100%.

~~~ js
$.AMUI.progress.inc();
~~~

Or increase a fixed value by passing a parameter to `.inc()`: 

~~~ js
$.AMUI.progress.inc(0.2);    // This will get the current status value and adds 0.2 until status is 0.994
~~~

`.inc()` method will get the current progress value and add 0.2 to it, but it will never get over 0.994.

**Forced done**: Pass `true` to `done()` to force to show the progress bar (The *.done()* won't do anything by default if there is no *.start()*).

~~~ js
$.AMUI.progress.done(true);
~~~

**Get progress**: Use `.status` attribute.

`````html
<button id="np-set" class="am-btn am-btn-primary">$.AMUI.progress.set(0.4);</button>
<button id="np-inc" class="am-btn am-btn-warning">$.AMUI.progress.inc();</button>
<button id="np-fd" class="am-btn am-btn-success">$.AMUI.progress.done(true);</button>
<button id="np-status" class="am-btn am-btn-danger">$.AMUI.progress.status;</button>
`````
```html
<button id="np-set" class="am-btn am-btn-primary">$.AMUI.progress.set(0.4);</button>
<button id="np-inc" class="am-btn am-btn-warning">$.AMUI.progress.inc();</button>
<button id="np-fd" class="am-btn am-btn-success">$.AMUI.progress.done(true);</button>
<button id="np-status" class="am-btn am-btn-danger">$.AMUI.progress.status;</button>
```
```js
$(function(){
  var progress = $.AMUI.progress;

  $('#np-set').on('click', function() {
    progress.set(0.4);
  });

  $('#np-inc').on('click', function() {
    progress.inc();
  });

  $('#np-fd').on('click', function() {
    progress.done(true);
  });

  $('#np-status').on('click', function() {
    $(this).text('Status: ' + progress.status);
  });
});
```

Options
-------

### Default Options

```js
  {
    minimum: 0.08,
    easing: 'ease',
    positionUsing: '',
    speed: 200,
    trickle: true,
    trickleRate: 0.02,
    trickleSpeed: 800,
    showSpinner: true,
    barSelector: '[role="nprogress-bar"]',
    spinnerSelector: '[role="nprogress-spinner"]',
    parent: 'body',
    template: '<div class="nprogress-bar" role="nprogress-bar">' +
        '<div class="nprogress-peg"></div></div>' +
        '<div class="nprogress-spinner" role="nprogress-spinner">' +
        '<div class="nprogress-spinner-icon"></div></div>'
  }
```

### Explanation
`minimum`: Set the minimum percentage.

~~~ js
$.AMUI.progress.configure({ minimum: 0.1 });
~~~

`template`: Set the template. Please remember to modify `barSelector` and `spinnerSelector` accordingly.

~~~ js
$.AMUI.progress.configure({
  template: "<div class='....'>...</div>"
});
~~~

`ease`、`speed`: Set the Easing function and speed(`ms`).

~~~ js
$.AMUI.progress.configure({ ease: 'ease', speed: 500 });
~~~

`trickle`、`trickleRate`、`trickleSpeed`:

~~~ js
$.AMUI.progress.configure({ trickle: false });
~~~

~~~ js
$.AMUI.progress.configure({ trickleRate: 0.02, trickleSpeed: 800 });
~~~

`showSpinner`:

~~~ js
NProgress.configure({ showSpinner: false });
~~~

`parent`:

Set the parent container of progress. Default value is `body`.

```js
$.AMUI.progress.configure({ parent: '#container' });
```

Customize
-----

Change progress bar's style by editing the `ui.progress.less` in `less` folder.


Reference:
-------

 * [New UI Pattern: Website Loading
 Bars](http://www.usabilitypost.com/2013/08/19/new-ui-pattern-website-loading-bars/)


[Turbolinks]: https://github.com/rails/turbolinks
[nprogress.js]: http://ricostacruz.com/nprogress/nprogress.js
[nprogress.css]: http://ricostacruz.com/nprogress/nprogress.css


<script>
$(function(){
  var progress = $.AMUI.progress;

  $('#np-s').on('click', function() {
    progress.start();
  });

  $('#np-d').on('click', function() {
    progress.done();
  });

  $('#np-set').on('click', function() {
    progress.set(0.4);
  });

  $('#np-inc').on('click', function() {
    progress.inc();
  });

  $('#np-fd').on('click', function() {
    progress.done(true);
  });

  $('#np-status').on('click', function() {
    $(this).text('Status: ' + progress.status);
  });
});
</script>


