---
id: smooth-scroll
title: 平滑滚动
titleEn: Smooth Scroll
prev: javascript/scrollspynav.html
next: javascript/sticky.html
source: js/ui.smooth-scroll.js
doc: docs/javascript/smooth-scroll.md
---

# Smooth Scroll
---

This plugin provide buttons that help user scroll smoothly to top and bottom. Via [founder of Zepto](https://gist.github.com/madrobby/8507960#file-scrolltotop-annotated-js)。

<div class="am-alert am-alert-danger">This plugin is not supported by IE 9 and older version. Please find other replacement if required.</div>

Use following JS code to scroll top in old IE.

```js
$('html, body').animate({scrollTop: 0}, '500');
```

## Examples

### Scroll top

`````html
<button data-am-smooth-scroll class="am-btn am-btn-success">Scroll top</button>
`````

```html
<button data-am-smooth-scroll class="am-btn am-btn-success">Scroll top</button>
```

### Scroll bottom

`````html
<button id="doc-scroll-to-btm" class="am-btn am-btn-primary">Scroll bottom</button>
<script>
  $('#doc-scroll-to-btm').on('click', function() {
    var $w = $(window);
    $w.smoothScroll({position: $(document).height() - $w.height()});
  });
</script>
`````

```html
<button id="doc-scroll-to-btm" class="am-btn am-btn-primary">Scroll bottom</button>
<script>
  $('#doc-scroll-to-btm').on('click', function() {
    var $w = $(window);
    $w.smoothScroll({position: $(document).height() - $w.height()});
  });
</script>
```

### Options

`````html
<button data-am-smooth-scroll="{position: 57, speed: 5000}" class="am-btn am-btn-danger">Slowly scroll to 57px away from top</button>
`````

```html
<button data-am-smooth-scroll="{position: 57, speed: 5000}" class="am-btn am-btn-danger">Slowly scroll to 57px away from top</button>
```

## Usage

### Using Data API

Add `data-am-smooth-scroll` attribute to element.

```html
<button data-am-smooth-scroll class="am-btn am-btn-success">Scroll top</button>
```

Use position option to scroll to specific position in the page.

`````html
<button data-am-smooth-scroll="{position: 189}" class="am-btn am-btn-secondary">Scroll to 189px away from top</button>
`````
```html
<button data-am-smooth-scroll="{position: 189}" class="am-btn am-btn-secondary">...</button>
```
### Using Javascript

#### Methods

In order to work correctly on different browsers, please call `$().smoothScroll()` method in `$(window)`.

```javascript
$(window).smoothScroll([options])
```

```javascript
// Scroll to bottom
$('#my-button').on('click', function() {
  var $w = $(window);
  $w.smoothScroll({position: $(document).height() - $w.height()});
});
```

#### Options

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">Option</th>
    <th style="width: 70px;">Type</th>
    <th style="width: 110px;">Default</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>position</code></td>
    <td>number</td>
    <td><code>0</code></td>
    <td>The position of view port when stop scrolling. Default value is `0`, which means scroll to top</td>
  </tr>
  <tr>
    <td><code>speed</code></td>
    <td>number</td>
    <td><code>750 ~ 1500</code></td>
    <td>Scrolling speed in `ms`. Default value is `750 - 1500`, adapted to the distance.</td>
  </tr>
  </tbody>
</table>
