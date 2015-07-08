---
id: sticky
title: 固定元素
titleEn: Sticky
prev: javascript/smooth-scroll.html
next: javascript/tabs.html
source: js/ui.sticky.js
doc: docs/javascript/sticky.md
---

# Sticky
---

Make elements remain at the top of the viewport, like a sticky navbar.

## Examples

### Default Style

Add `data-am-sticky` attribute to element.

`````html
<div data-am-sticky>
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
`````
```html
<div data-am-sticky>
  <button class="am-btn am-btn-primary am-btn-block">Stick to top</button>
</div>
```

### Define an offset

The offset is default to be 0, which means element will stick at the top of viewport. Offset can be changed by using `data-am-sticky="{top:100}"`.

`````html
<div data-am-sticky="{top:80}">
  <button class="am-btn am-btn-primary">Stick 80px below the top</button>
</div>
`````
```html
<div data-am-sticky="{top:80}">
  <button class="am-btn am-btn-primary">Stick 80px below the top</button>
</div>
```

### Animation

Use [CSS3 Animation](http://amazeui.org/css/animation).

`````html
<div data-am-sticky="{animation: 'slide-top'}">
  <button class="am-btn am-btn-success am-btn-block">Stick to top with animation</button>
</div>
`````
```html
<div data-am-sticky="{animation: 'slide-top'}">
  <button class="am-btn am-btn-success am-btn-block">Stick to top with animation</button>
</div>
```

## Usage

### Using Data API

As shown above, add `data-am-sticky` attribute to element.

### Using JS

Use `$.sticky(options)` to initialize the sticky.

`````html
<div id="my-sticky">
  <button class="am-btn am-btn-danger">Stick via JavaScript & 150px below the top</button>
</div>
<script>
$(function() {
$('#my-sticky').sticky({
    top: 150,
    bottom: function() {
      return $('.amz-footer').height();
    }
  });
});
</script>
`````
```html
<div id="my-sticky">
  <button class="am-btn am-btn-danger">Stick via JavaScript</button>
</div>
<script>
$(function() {
  $('#my-sticky').sticky({
    top: 150
  })
});
</script>
```

### Options

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">Option</th>
    <th style="width: 70px;">Type</th>
    <th style="width: 50px;">Default</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>top</code></td>
    <td>number</td>
    <td><code>0</code></td>
    <td>Distance to the top</td>
  </tr>
  <tr>
    <td><code>animation</code></td>
    <td>String</td>
    <td><code>''</code></td>
    <td>Name of animatino</td>
  </tr>
  <tr>
    <td><code>bottom</code></td>
    <td>number <br/> or a function that return a number</td>
    <td><code>0</code></td>
    <td>Stop moving when the distance between element and bottom is smaller than this value.</td>
  </tr>
  </tbody>
</table>

## Attention

- Sticky watch the scrolling event of window. When the scrolling distance is larger than the distance between top of element and top, the `.am-sticky` class will be added to the element. The `position` of element will be changed to `fixed` and the `top` value will be given(Default value is 0).
- The sticky element will be wrapped with a `.am-sticky-placeholder` after initialize, which can possiblly influence some selectors.
- __Problem:__ When using animation, if the window is `resized` repidly, the animation will be played multiple times.


<div style="height: 400px"></div>
