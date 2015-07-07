---
id: collapse
title: 折叠面板
titleEn: Collapse
prev: javascript/button.html
next: javascript/dropdown.html
source: js/ui.collapse.js
doc: docs/javascript/collapse.md
---

# Collapse
---

The collapse component can be used to make dropdown and accordion.

## Examples

### Accordion

To create an accordion, use [Panel](/css/panel) with following classes in collapse:

* Use `.am-collapse` to hide contents;
* Use `.am-collapse.am-in` to show contents;

Use Data API after adding above classes:

```html
<h4 data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}"></h4>
```

Where

* `parent` is the container ID
* `target` is the content container ID

If the trigger element is `<a>`, `target` can be set in `href`.

```html
<a data-am-collapse="{parent: '#accordion'}" href="#do-not-say-1">
  ...
</a>
```

`````html
<div class="am-panel-group" id="accordion">
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}">
        Sonnet 19 Part 1 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-1" class="am-panel-collapse am-collapse am-in">
      <div class="am-panel-bd">
        Devouring Time,blunt thou the lion'paws, <br/>
        And make the earth devour her own sweet brood; <br/>
        Pluck the keen teeth from the fierce tiger's jaws, <br/>
        And burn the long--liv'd phoenix in her blood; <br/>
        
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-2'}">
        Sonnet 19 Part 2 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-2" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        Make glad and sorry seasons as thou fleets, <br/>
        And do whate'er thou wilt,swift--footed Time, <br/>
        To the wide world and all her fading sweets; <br/>
        But I forbid thee one most heinous crime, <br/>
        O carve not with thy hours my love's fair brow, <br/>

      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-3'}">
        Sonnet 19 Part 3 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-3" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        Nor draw no lines there with thine antique pen.<br/>
        Him in thy course untainted do allow,<br/>
        For beauty's pattern to succeeding men. <br/>
        &nbsp;&nbsp;Yet do thy worst,old Time;despite thy wrong,<br/>
        &nbsp;&nbsp;My love shall in my verse ever live young.<br/>
      </div>
    </div>
  </div>
</div>
`````

```html
<a data-am-collapse="{parent: '#accordion'}" href="#do-not-say-1">
  ...
</a>
```

`````html
<div class="am-panel-group" id="accordion">
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-1'}">
        Sonnet 19 Part 1 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-1" class="am-panel-collapse am-collapse am-in">
      <div class="am-panel-bd">
        Devouring Time,blunt thou the lion'paws, <br/>
        And make the earth devour her own sweet brood; <br/>
        Pluck the keen teeth from the fierce tiger's jaws, <br/>
        And burn the long--liv'd phoenix in her blood; <br/>
        
      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-2'}">
        Sonnet 19 Part 2 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-2" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        Make glad and sorry seasons as thou fleets, <br/>
        And do whate'er thou wilt,swift--footed Time, <br/>
        To the wide world and all her fading sweets; <br/>
        But I forbid thee one most heinous crime, <br/>
        O carve not with thy hours my love's fair brow, <br/>

      </div>
    </div>
  </div>
  <div class="am-panel am-panel-default">
    <div class="am-panel-hd">
      <h4 class="am-panel-title" data-am-collapse="{parent: '#accordion', target: '#do-not-say-3'}">
        Sonnet 19 Part 3 - By William Shakespeare
      </h4>
    </div>
    <div id="do-not-say-3" class="am-panel-collapse am-collapse">
      <div class="am-panel-bd">
        Nor draw no lines there with thine antique pen.<br/>
        Him in thy course untainted do allow,<br/>
        For beauty's pattern to succeeding men. <br/>
        &nbsp;&nbsp;Yet do thy worst,old Time;despite thy wrong,<br/>
        &nbsp;&nbsp;My love shall in my verse ever live young.<br/>
      </div>
    </div>
  </div>
</div>
`````

### Dropdown Menu

Please be aware of there is a nav container outside the target element, which is used to calculate the height of menu.

`````html
<button class="am-btn am-btn-primary" data-am-collapse="{target: '#collapse-nav'}">Menu <i class="am-icon-bars"></i></button>
<nav>
  <ul id="collapse-nav" class="am-nav am-collapse">
    <li><a href="">Getting Started</a></li>
    <li><a href="">CSS</a></li>
    <li class="am-active"><a href="">JS</a></li>
    <li><a href="">Customize</a></li>
  </ul>
</nav>
`````
```html
<button class="am-btn am-btn-primary" data-am-collapse="{target: '#collapse-nav'}">Menu <i class="am-icon-bars"></i></button>
<nav>
  <ul id="collapse-nav" class="am-nav am-collapse">
    <li><a href="">Getting Started</a></li>
    <li><a href="">CSS</a></li>
    <li class="am-active"><a href="">JS</a></li>
    <li><a href="">Customize</a></li>
  </ul>
</nav>
```

## Usage

### Using Data API

Add `data-am-collapse` attribute to element and set the value of `target` to the ID of collapse element: 

```html
<button data-am-collapse="{target: '#my-collapse'}"></button>
```

### Using JS

```js
$('#myCollapse').collapse()
```

#### Methods

- `$().collapse(options)` - Expand/Collapse target element

```javascript
$('#myCollapse').collapse({
  toggle: false
})
```

- `$().collapse('toggle')` - Switch the panel state
- `$().collapse('open')` - Expand the panel
- `$().collapse('close')` - Collapse the panel

#### Options

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">Parameter</th>
    <th style="width: 70px;">Type</th>
    <th style="width: 50px;">Default</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>parent</code></td>
    <td>Selector</td>
    <td><code>false</code></td>
    <td>If a <code>parent</code> selector is provided, then all collapsible elements under the specified parent will be closed when this collapsible item is shown.</td>
  </tr>
  <tr>
    <td><code>toggle</code></td>
    <td>boolean</td>
    <td><code>true</code></td>
    <td>Toggles the collapsible element on invocation</td>
  </tr>
  </tbody>
</table>


#### Events

Some events are defined in collapse and can be triggered on collapse element. Using the dropdown menu as example:

<script>
$(function() {
  $('#collapse-nav').on('open.collapse.amui', function() {
    console.log('Menu is opened!');
  }).on('close.collapse.amui', function() {
    console.log('Menu is closed!');
  });
});
</script>

```js
$(function() {
  $('#collapse-nav').on('open.collapse.amui', function() {
    console.log('Menu is opened!');
  }).on('close.collapse.amui', function() {
    console.log('Menu is closed!');
  });
});
```

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.collapse.amui</code></td>
    <td>This event fires when <code>open</code> method get called</td>
  </tr>
  <tr>
    <td><code>opened.collapse.amui</code></td>
    <td>This event fires when a collapse element is fully opend (will wait for CSS transitions to complete).</td>
  </tr>
  <tr>
    <td><code>close.collapse.amui</code></td>
    <td>This event fires when <code>close</code> method get called
    </td>
  </tr>
  <tr>
    <td><code>closed.collapse.amui</code></td>
    <td>This event fires when a collapse element is fully closed (will wait for CSS transitions to complete).</td>
  </tr>
  </tbody>
</table>

## Attention

**Don't use vertical `margin`/`padding`/`border` in container of collapse element.**

Because of the way jQuery calculate the heights of elements, styles can go wrong if `margin`/`padding`/`border` is added to container.