---
id: scrollspy
title: 滚动侦测
titleEn: ScrollSpy
prev: javascript/offcanvas.html
next: javascript/scrollspynav.html
source: js/ui.scrollspy.js
doc: docs/javascript/scrollspy.md
---

# ScrollSpy
---

Add animations to elements when scrolling. (require support to CSS3 animation).

## Examples

The following example contains different animations.

`````html
<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade</div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-4 am-u-lg-offset-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'scale-up'}">
      <div class="am-panel-hd">Scale-up</div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-6">
    <div class="am-panel am-panel-success" data-am-scrollspy="{animation: 'scale-down'}">
      <div class="am-panel-hd">Scale-down
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-9">
    <div class="am-panel am-panel-warning" data-am-scrollspy="{animation: 'slide-top'}">
      <div class="am-panel-hd">Slide Top
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-8 am-u-lg-offset-6">
    <div class="am-panel am-panel-danger" data-am-scrollspy="{animation: 'slide-bottom'}">
      <div class="am-panel-hd">Slide Bottom
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3 am-u-md-offset-4 am-u-lg-offset-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'slide-right'}">
      <div class="am-panel-hd">Slide Right
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-secondary" data-am-scrollspy="{animation: 'slide-left'}">
      <div class="am-panel-hd">Slide Left</div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>

<div class="am-g">
  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-success" data-am-scrollspy="{animation: 'fade'}">
      <div class="am-panel-hd">Fade</div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-warning" data-am-scrollspy="{animation: 'fade', delay: 300}">
      <div class="am-panel-hd">Fade delay: 300</div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-danger" data-am-scrollspy="{animation: 'fade', delay: 600}">
      <div class="am-panel-hd">Fade delay: 600
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>

  <div class="am-u-sm-12 am-u-md-4 am-u-lg-3">
    <div class="am-panel am-panel-primary" data-am-scrollspy="{animation: 'fade', delay: 900}">
      <div class="am-panel-hd">Fade delay: 900
      </div>
      <div class="am-panel-bd">
        The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
      </div>
    </div>
  </div>
</div>
`````
```html
<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade'}">...</div>

<div class="am-panel am-panel-default" data-am-scrollspy="{animation: 'fade', delay: 300}">...</div>
```

## Usage

### Using Data API

Add `data-am-scrollspy` attribute to element.

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Attribute</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade'}"</code></td>
    <td>Animation. See more in <a href="/css/animation">Amaze UI Animation</a>. Default animation is <code>fade</code></td>
  </tr>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade', delay: 300}"</code></td>
    <td>Delay(ms). Default value is <code>0</code></td>
  </tr>
  <tr>
    <td><code>data-am-scrollspy="{animation:'fade', repeat: false}"</code></td>
    <td>Whether repeat the animation. Default value is <code>true</code></td>
  </tr>
  </tbody>
</table>

### Using JS

Initialize through `$().scrollspy(options)`.


`````html
<div class="am-panel am-panel-primary" id="my-scrollspy">
  <div class="am-panel-hd">ScrollSpy via JS
  </div>
  <div class="am-panel-bd">
    The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
  </div>
</div>
<script>
$(function() {
  $('#my-scrollspy').scrollspy({
    animation: 'slide-left',
    delay: 500
  })
});
</script>
`````
```html
<div class="am-panel am-panel-primary" id="my-scrollspy">
  <div class="am-panel-hd">ScrollSpy via JS
  </div>
  <div class="am-panel-bd">
    The purpose of his life is to procure for himself everything that contributes to bodily welfare. He is happy enough when this causes him a lot of trouble. For if those good things are heaped on him in advance, he will inevitably lapse into boredom —— Arthur Schopenhauer
  </div>
</div>
<script>
$(function() {
  $('#my-scrollspy').scrollspy({
    animation: 'slide-left',
    delay: 500
  })
});
</script>
```
#### Events

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>inview.scrollspy.amui</code></td>
    <td>Fired when the element comes into the viewport.</td>
  </tr>
  <tr>
    <td><code>outview.scrollspy.amui</code></td>
    <td>Fired when the element get out of the viewport.</td>
  </tr>
  </tbody>
</table>

<script>
  $(function() {
    $('#my-scrollspy').on('inview.scrollspy.amui', function() {
      console.log('inview');
    }).on('outview.scrollspy.amui', function() {
      console.log('outview');
    });
  });
</script>

```javascript
$(function() {
  $('#my-scrollspy').on('inview.scrollspy.amui', function() {
    console.log('inview');
  }).on('outview.scrollspy.amui', function() {
    console.log('outview');
  });
});
```

#### MutationObserver

Use [Mutation Observer](https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver) to add animation to dynamiclly loaded elements.

`````html
<p><button class="am-btn am-btn-primary" id="doc-scrollspy-insert">Insert</button></p>
<div id="doc-scrollspy-wrapper" data-am-observe>
<p>Insert elements here: </p>
</div>
<script>
  $(function() {
    var i = 1;
    var $wrapper = $('#doc-scrollspy-wrapper');
    var appendPanel = function(index) {
      var panel = '<div class="am-panel am-panel-primary" ' +
        'data-am-scrollspy="{animation: \'scale-up\'}">' +
        '<div class="am-panel-bd">I am the No.' + index + ' element。</div></div>';
      $wrapper.append(panel);
    };

    $('#doc-scrollspy-insert').on('click', function() {
      appendPanel(i);
      i++;
    });
  });
</script>
`````
```html
<p><button class="am-btn am-btn-primary" id="doc-scrollspy-insert">Insert</button></p>
<div id="doc-scrollspy-wrapper" data-am-observe>
<p>Insert elements here: </p>
</div>
<script>
  $(function() {
    var i = 1;
    var $wrapper = $('#doc-scrollspy-wrapper');
    var appendPanel = function(index) {
      var panel = '<div class="am-panel am-panel-primary" ' +
        'data-am-scrollspy="{animation: \'scale-up\'}">' +
        '<div class="am-panel-bd">I am the No.' + index + ' element。</div></div>';
      $wrapper.append(panel);
    };

    $('#doc-scrollspy-insert').on('click', function() {
      appendPanel(i);
      i++;
    });
  });
</script>
```

