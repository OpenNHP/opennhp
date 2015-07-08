---
id: dropdown
title: 下拉组件
titleEn: Dropdown
prev: javascript/collapse.html
next: javascript/modal.html
source: js/ui.dropdown.js
doc: docs/javascript/dropdown.md
---

# Dropdown
---

## Example

### Dropdown

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>Dropdown <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Title</li>
        <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
        <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
        <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
        <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
        <li class="am-divider"></li>
        <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
  </ul>
</div>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-primary am-dropdown-toggle" data-am-dropdown-toggle>Dropdown <span class="am-icon-caret-down"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Title</li>
        <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
        <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
        <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
        <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
        <li class="am-divider"></li>
        <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
  </ul>
</div>
```

### Dropdown(up)

Add the `.am-dropdown-up` class to `.am-dropdown` to pop the list on the top of the button.

`````html
<div class="am-dropdown am-dropdown-up" data-am-dropdown>
  <button class="am-btn am-btn-danger am-dropdown-toggle" data-am-dropdown-toggle>Pullup <span class="am-icon-caret-up"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Title</li>
        <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
        <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
        <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
        <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
        <li class="am-divider"></li>
        <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
  </ul>
</div>
`````
```html
<div class="am-dropdown am-dropdown-up" data-am-dropdown>
  <button class="am-btn am-btn-danger am-dropdown-toggle" data-am-dropdown-toggle>Pullup <span class="am-icon-caret-up"></span></button>
  <ul class="am-dropdown-content">
    <li class="am-dropdown-header">Title</li>
        <li><a href="#">Devouring Time,blunt thou the lion'paws,</a></li>
        <li class="am-active"><a href="#">And make the earth devour her own sweet brood;</a></li>
        <li><a href="#">Pluck the keen teeth from the fierce tiger's jaws,</a></li>
        <li class="am-disabled"><a href="#">And burn the long--liv'd phoenix in her blood;</a></li>
        <li class="am-divider"></li>
        <li><a href="#">Make glad and sorry seasons as thou fleets,</a></li>
  </ul>
</div>
```

### Dropdown contents

`````html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">Dropdown contents <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>Hope Is the Thing with Feathers</h2>
    <p>
      Hope is the thing with feathers
      That perches in the soul,
      And sings the tune without the words,
      And never stops at all.
      And sweetest in the gale is heard;
      And sore must be the storm
      That could abash the little bird
      That kept so many warm.
      I’ve heard it in the chillest land,
      And on the strangest sea;
      Yet, never, in extremity,
      It asked a crumb of me.
    </p>
  </div>
</div>
`````
```html
<div class="am-dropdown" data-am-dropdown>
  <button class="am-btn am-btn-success am-dropdown-toggle">Dropdown contents <span class="am-icon-caret-down"></span></button>
  <div class="am-dropdown-content">
    <h2>Hope Is the Thing with Feathers</h2>
    <p>
      Hope is the thing with feathers
      That perches in the soul,
      And sings the tune without the words,
      And never stops at all.
      And sweetest in the gale is heard;
      And sore must be the storm
      That could abash the little bird
      That kept so many warm.
      I’ve heard it in the chillest land,
      And on the strangest sea;
      Yet, never, in extremity,
      It asked a crumb of me.
    </p>
  </div>
</div>
```

### Width Justifying

The content `.am-dropdown-content` use absolute positioning, and the width will be justified according to the contents. (min-width is `160px`)

In the following example, the dropdown is justitied to the `<div>` element `doc-dropdown-justify`. The width of the dropdown content is set to be same as the `<div>` element.

`````html
<div id="doc-dropdown-justify">
  <div class="am-dropdown" data-am-dropdown="{justify: '#doc-dropdown-justify'}">
    <button class="am-btn am-btn-success am-dropdown-toggle">Width Justifying <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>Hope Is the Thing with Feathers</h2>
      <p>
        Hope is the thing with feathers
        That perches in the soul,
        And sings the tune without the words,
        And never stops at all.
        And sweetest in the gale is heard;
        And sore must be the storm
        That could abash the little bird
        That kept so many warm.
        I’ve heard it in the chillest land,
        And on the strangest sea;
        Yet, never, in extremity,
        It asked a crumb of me.
      </p>
    </div>
  </div>
</div>
`````
```html
<div id="doc-dropdown-justify">
  <div class="am-dropdown" data-am-dropdown="{justify: '#doc-dropdown-justify'}">
    <button class="am-btn am-btn-success am-dropdown-toggle">Width Justifying <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>Hope Is the Thing with Feathers</h2>
      <p>
        Hope is the thing with feathers
        That perches in the soul,
        And sings the tune without the words,
        And never stops at all.
        And sweetest in the gale is heard;
        And sore must be the storm
        That could abash the little bird
        That kept so many warm.
        I’ve heard it in the chillest land,
        And on the strangest sea;
        Yet, never, in extremity,
        It asked a crumb of me.
      </p>
    </div>
  </div>
</div>
```


## Usage

### Using Data API

As shown above, add the `data-am-dropdown` attribute to `.am-dropdown`, and set values of options in this attribute.

### Using JS

Create an dropdown without `data-am-dropdown` attribute, and then modify it using JS.

`````html
<div id="doc-dropdown-justify-js" style="width: 400px">
  <div class="am-dropdown" id="doc-dropdown-js">
    <button class="am-btn am-btn-danger am-dropdown-toggle">Using JS <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">
      <h2>Hope Is the Thing with Feathers</h2>
      <p>
        Hope is the thing with feathers
        That perches in the soul,
        And sings the tune without the words,
        And never stops at all.
        And sweetest in the gale is heard;
        And sore must be the storm
        That could abash the little bird
        That kept so many warm.
        I’ve heard it in the chillest land,
        And on the strangest sea;
        Yet, never, in extremity,
        It asked a crumb of me.
      </p>
    </div>
  </div>
</div>
<script>
  $(function() {
    $('#doc-dropdown-js').dropdown({justify: '#doc-dropdown-justify-js'});
  });
</script>
`````
```html
<div id="doc-dropdown-justify-js" style="width: 400px">
  <div class="am-dropdown" id="doc-dropdown-js">
    <button class="am-btn am-btn-danger am-dropdown-toggle">Using JS <span class="am-icon-caret-down"></span></button>
    <div class="am-dropdown-content">...</div>
  </div>
</div>
<script>
  $(function() {
    $('#doc-dropdown-js').dropdown({justify: '#doc-dropdown-justify-js'});
  });
</script>
```

#### Methods

- `$(element).dropdown(options)` Activate dropdown;
- `$(element).dropdown('toggle')` Switch between show and hide;
- `$(element).dropdown('close')` Hide the dropdown contents;
- `$(element).dropdown('open')` Show the dropdown contents.

`````html
<button class="am-btn am-btn-secondary" id="doc-dropdown-toggle">Toggle</button>
<button class="am-btn am-btn-success" id="doc-dropdown-open">Open</button>
<button class="am-btn am-btn-warning" id="doc-dropdown-close">Close</button>
<script>
  $(function() {
    var $dropdown = $('#doc-dropdown-js'),
        data = $dropdown.data('amui.dropdown');

    function scrollToDropdown() {
      $(window).smoothScroll({position: $dropdown.offset().top});
    }

    $('#doc-dropdown-toggle').on('click', function(e) {
      scrollToDropdown();
      $dropdown.dropdown('toggle');
      return false;
    });

    $('#doc-dropdown-open').on('click', function(e) {
      scrollToDropdown();
      data.active ? alert('Dropdown has been opened') : $dropdown.dropdown('open');
      return false;
    });

    $('#doc-dropdown-close').on('click', function(e) {
      scrollToDropdown();
      data.active ? $dropdown.dropdown('close') : alert('Dropdown has been closed');
      return false;
    });

    $dropdown.on('open.dropdown.amui', function (e) {
      console.log('open event triggered');
    });
  });
</script>
`````
```html
<button class="am-btn am-btn-secondary" id="doc-dropdown-toggle">Toggle</button>
<button class="am-btn am-btn-success" id="doc-dropdown-open">Open</button>
<button class="am-btn am-btn-warning" id="doc-dropdown-close">Close</button>
<script>
  $(function() {
    var $dropdown = $('#doc-dropdown-js'),
        data = $dropdown.data('amui.dropdown');
    $('#doc-dropdown-toggle').on('click', function(e) {
      scrollToDropdown();
      $dropdown.dropdown('toggle');
      return false;
    });

    $('#doc-dropdown-open').on('click', function(e) {
      scrollToDropdown();
      data.active ? alert('Dropdown has been opened') : $dropdown.dropdown('open');
      return false;
    });

    $('#doc-dropdown-close').on('click', function(e) {
      scrollToDropdown();
      data.active ? $dropdown.dropdown('close') : alert('Dropdown has been closed');
      return false;
    });

    $dropdown.on('open.dropdown.amui', function (e) {
      console.log('open event triggered');
    });
  });
</script>
```

#### Options

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="width: 60px;">Parameter</th>
    <th style="width: 70px;">Type</th>
    <th style="width: 110px;">Default</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>boundary</code></td>
    <td>Selector</td>
    <td><code>window</code></td>
    <td>boundary of dropdown contents. Avoid cutting off the overflow contents.</td>
  </tr>
  <tr>
    <td><code>justify</code></td>
    <td>Selector</td>
    <td><code>undefined</code></td>
    <td>The element that width of content justifies to</td>
  </tr>
  </tbody>
</table>

#### Events

The events of dropdown is triggered on element with `.am-dropdown` class.

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.dropdown.amui</code></td>
    <td>This event fires immediately when open method get called.</td>
  </tr>
  <tr>
    <td><code>opened.dropdown.amui</code></td>
    <td>This event fires when dropdown has been opened. (After CSS animation ends)</td>
  </tr>
  <tr>
    <td><code>close.dropdown.amui</code></td>
    <td>This event fires immediately when close method get called.</td>
  </tr>
  <tr>
    <td><code>closed.dropdown.amui</code></td>
    <td>This event fires when dropdown has been closed. (After CSS animation ends)</td>
  </tr>
  </tbody>
</table>

```js
$(function() {
  $dropdown.on('open.dropdown.amui', function (e) {
    console.log('open event triggered');
  });
});
```
