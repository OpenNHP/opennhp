---
id: offcanvas
title: 侧边栏
titleEn: OffCanvas
prev: javascript/slider.html
next: javascript/scrollspy.html
source: js/ui.offcanvas.js
doc: docs/javascript/offcanvas.md
---

# OffCanvas
---

Create a smooth off-canvas sidebar that slides in and out of the page. See more details in menu widget([Example 1](/widgets/menu/offcanvas1/0), [Example 2](/widgets/menu/offcanvas1/1)).

## Examples

The Off-canvas component consists of an overlay and an off-canvas bar. Add `data-am-offcanvas` attribute to the overlay, and use specified HTML stucture for off-canvas bar.

### Default Style

`````html
<!-- Link Trigger, whose href attribute is the ID of target element -->
<a href="#doc-oc-demo1" data-am-offcanvas>Show offcanvas</a>

<!-- Offcanvas contents -->
<div id="doc-oc-demo1" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒
      </p>
    </div>
  </div>
</div>
`````

```html
<!-- Link Trigger, whose href attribute is the ID of target element -->
<a href="#doc-oc-demo1" data-am-offcanvas>Show offcanvas</a>

<!-- Offcanvas contents -->
<div id="doc-oc-demo1" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒
      </p>
    </div>
  </div>
</div>
```

### Push Effect

Set `effect: 'push'`.

`````html
<!-- Button trigger. Need to specify the target -->
<button class="am-btn am-btn-primary" data-am-offcanvas="{target: '#doc-oc-demo2', effect: 'push'}">Click to show the offcanvas</button>

<!-- Offcanvas contents -->
<div id="doc-oc-demo2" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒
      </p>
    </div>
  </div>
</div>
`````
```html
<!-- Button trigger. Need to specify the target -->
<button class="am-btn am-btn-primary" data-am-offcanvas="{target: '#doc-oc-demo2', effect: 'push'}">Click to show the offcanvas</button>

<!-- Offcanvas contents -->
<div id="doc-oc-demo2" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒
      </p>
    </div>
  </div>
</div>
```

### Right Offcanvas

Offcanvas is on the left by default. Add the `.am-offcanvas-bar-flip` class to offcanvas content to move it to the right.

`````html
<!-- Button trigger. Need to specify the target -->
<button class="am-btn am-btn-success" data-am-offcanvas="{target: '#doc-oc-demo3'}">右侧显示侧边栏</button>

<!-- Offcanvas contents -->
<div id="doc-oc-demo3" class="am-offcanvas">
  <div class="am-offcanvas-bar am-offcanvas-bar-flip">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒 <br/>
      </p>
      <p><a href="http://music.163.com/#/song?id=385554" target="_blank">网易音乐</a>      </p>
    </div>
  </div>
</div>
`````
```html
<!-- Button trigger. Need to specify the target -->
<button class="am-btn am-btn-success" data-am-offcanvas="{target: '#doc-oc-demo3'}">右侧显示侧边栏</button>

<!-- Offcanvas contents -->
<div id="doc-oc-demo3" class="am-offcanvas">
  <div class="am-offcanvas-bar am-offcanvas-bar-flip">
    <div class="am-offcanvas-content">
      <p>
        我不愿让你一个人 <br/>
        承受这世界的残忍 <br/>
        我不愿眼泪陪你到 永恒 <br/>
      </p>
      <p><a href="http://music.163.com/#/song?id=385554" target="_blank">网易音乐</a>      </p>
    </div>
  </div>
</div>
```

```html
<div id="my-id" class="am-offcanvas">
	<div class="am-offcanvas-bar am-offcanvas-bar-flip">...</div>
</div>
```

## Usage

First, create your offcanvas content using following structure:

```html
<div id="your-id" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <!-- Your content -->
  </div>
</div>
```

### Using Data API

Add `data-am-offcanvas` attribute to the overlay element:

- If this element is an `<a>`, set `href` to be the ID of offcanvas：`href="#your-id"`;
- Otherwise, specify the ID of offcanvas in `data-am-offcanvas` :

```html
<button data-am-offcanvas="{target: '#your-id'}"></button>
```

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Attribute</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>{target: '#your-id'}</code></td>
    <td>Specify the offcanvas. If overlay is an <code>a</code> element, specify it in <code>href</code>.</td>
  </tr>
  <tr>
    <td><code>{effect: 'push'}</code></td>
    <td>Offcanvas animation. Optional values include <code>overlay | push</code>. Default value is <code>overlay</code></td>
  </tr>
  </tbody>
</table>

### Using JS

Use Javascript to control the Offcanvas.

```javascript
$('#my-offcanvas').offCanvas(options);
```

__Attension:__ The `#my-offcanvas` here points directly to the offcanvas instead of the overlay.

#### Options

- `options.effect`. Optional values include `overlay | push`. Default value is `overlay`.

```javascript
$('#my-offcanvas').offCanvas({effect: 'push'});
```

Set through `$().offCanvas(options)`.

#### Methods

- `$().offCanvas(options)` - Set the offcanvas attribute and show it.
- `$().offCanvas('open')` - Show the offcanvas.
- `$().offCanvas('close')` - Hide the offcanvas.

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
    <td><code>open.offcanvas.amui</code></td>
    <td>Fired immediately when offcanvas is opened.</td>
  </tr>
  <tr>
    <td><code>close.offcanvas.amui</code></td>
    <td>Fired immediately when offcanvas is closed.</td>
  </tr>
  </tbody>
</table>

#### JS Control Examples

The following example shows how to use JS to open/close the offcanvas. The close button can't be directly clicked when offcanvas is opened. Please use following code to simulate the click event:

```js
$('[data-rel="close"]').click();
```

`````html
<!-- Offcanvas contents -->
<div id="my-offcanvas" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        你那张略带着 <br/>
        一点点颓废的脸孔 <br/>
        轻薄的嘴唇 <br/>
        含着一千个谎言
      </p>
    </div>
  </div>
</div>

<button class="am-btn am-btn-primary doc-oc-js" data-rel="open">Open the offcanvas</button>
<button class="am-btn am-btn-primary doc-oc-js" data-rel="close">Close the offcanvas</button>

<script>
  $(function() {
    var id = '#my-offcanvas';
    var $myOc = $(id);
    $('.doc-oc-js').on('click', function() {
      $myOc.offCanvas($(this).data('rel'));
    });

    $myOc.on('open.offcanvas.amui', function() {
      console.log(id + ' is opened.');
    }).on('close.offcanvas.amui', function() {
      console.log(id + ' is closed.');
    });
  });
</script>
`````

```html
<!-- Offcanvas contents -->
<div id="my-offcanvas" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <div class="am-offcanvas-content">
      <p>
        你那张略带着 <br/>
        一点点颓废的脸孔 <br/>
        轻薄的嘴唇 <br/>
        含着一千个谎言
      </p>
    </div>
  </div>
</div>

<button class="am-btn am-btn-primary doc-oc-js" data-rel="open">Open the offcanvas</button>
<button class="am-btn am-btn-primary doc-oc-js" data-rel="close">Close the offcanvas</button>

<script>
  $(function() {
    var id = '#my-offcanvas';
    var $myOc = $(id);
    $('.doc-oc-js').on('click', function() {
      $myOc.offCanvas($(this).data('rel'));
    });

    $myOc.on('open.offcanvas.amui', function() {
      console.log(id + ' is opened.');
    }).on('close.offcanvas.amui', function() {
      console.log(id + ' is closed.');
    });
  });
</script>
```
