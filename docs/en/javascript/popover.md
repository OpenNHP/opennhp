---
id: popover
title: 弹出框
titleEn: Popover
prev: javascript/modal.html
next: javascript/nprogress.html
source: js/ui.popover.js
doc: docs/javascript/popover.md
---

# Popover
---

Add small overlays of content, like those on the iPad, to any element for housing secondary information.


## Examples

### Click to show

`````html
  <button class="am-btn am-btn-primary" data-am-popover="{content: 'I'm the click to show Popover'}">Click to show Popover</button>
`````
```html
<button class="am-btn am-btn-primary" data-am-popover="{content: 'I'm the click to show Popover'}">Click to show Popover</button>
```

### Hover/Focus to show

Tooltip effect.

`````html
<button class="am-btn am-btn-success" data-am-popover="{content: 'I'm the Hover/Focus to show Popover', trigger: 'hover focus'}">Hover/Focus to show Popover</button>
`````
```html
<button class="am-btn am-btn-success" 
        data-am-popover="{content: 'I'm the Hover/Focus to show Popover', trigger: 'hover focus'}">
  Hover/Focus to show Popover
</button>
```

## Usage

### Using Data API

Add `data-am-popover` attribute to elements and set options. All examples above use Data API.

```html
<button data-am-popover="{content: 'What do you want to show?', trigger: 'hover'}">Popover
</button>
```

### Using JS

Use `$().popover(options)` to add popover to elements.

`````html
<button class="am-btn am-btn-danger" id="my-popover">Popover via JS</button>
<script>
$(function() {
  $('#my-popover').popover({
    content: 'Popover via JavaScript'
  })
})
</script>
`````
```html
<button class="am-btn am-btn-danger" id="my-popover">Popover via JS</button>
```
```javascript
$(function() {
  $('#my-popover').popover({
    content: 'Popover via JavaScript'
  })
});
```

#### Options

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>Option</th>
    <th>Type</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>content</code></td>
    <td><code>string</code></td>
    <td>Contents in Popover</td>
  </tr>
  <tr>
    <td><code>trigger</code></td>
    <td><code>string</code></td>
    <td>Trigger type. Available options include <code>click|hover|focus</code>, Default to be <code>click</code></td>
  </tr>
  </tbody>
</table>

#### Method

- `.popover(options)` - Activate the Popover on element. `options` is an object.
- `.popover('toggle')` - Switch between show and hide.
- `.popover('open')` - Show Popover
- `.popover('close')` - Hide Popover
- `.popover('setContent', content)` - Set the popover contents. <span class="am-badge am-badge-danger">v2.4.1+</span>

#### Event

Events are defined on element that trigger the popover.

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.popover.amui</code></td>
    <td>Fired immediately when <code>open</code> is called. </td>
  </tr>
  <tr>
    <td><code>close.popover.amui</code></td>
    <td><code>Fired immediately when <code>close</code> is called.</td>
  </tr>
  </tbody>
</table>
