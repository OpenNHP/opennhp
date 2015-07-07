---
title: 按钮 JS 交互
titleEn: Button JS
prev: javascript/alert.html
next: javascript/button.html
source: js/ui.button.js
doc: docs/javascript/button.md
---


# Button JS
---

Button and Button group can interact with JS.

## Button Loading Status

### Default Text

The default text is `loading...`.

`````html
<button type="button" class="am-btn am-btn-primary btn-loading-example">Button - Button Element</button> &nbsp;
<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="Button - Input Element" />
`````

```html
<button type="button" class="am-btn am-btn-primary btn-loading-example">Submit - Button</button>
<input type="button" class="am-btn am-btn-primary btn-loading-example" value="Submit - Input" />
```
```js
$('.btn-loading-example').click(function () {
  var $btn = $(this)
  $btn.button('loading');
    setTimeout(function(){
      $btn.button('reset');
  }, 5000);
});
```

### Customize

Use the `data-am-loading` class to set the options.

- `spinner`: Loading Animation. Need support of CSS3 animation. Don't works in input;
- `loadingText` Loading text. Default text is `loading`;
- `resetText` Text after reset. Default text is the original text.

`````html
<button type="button" class="am-btn am-btn-primary btn-loading-example" data-am-loading="{spinner: 'circle-o-notch', loadingText: 'I'm Loading...', resetText: 'Loaded'}">Button - button element</button> &nbsp;
<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="Button - input element" data-am-loading="{loadingText: 'Still Loading...'}" />
<script>
$(function() {
  $('.btn-loading-example').click(function() {
    var $btn = $(this);
    $btn.button('loading');
    setTimeout(function() {
      $btn.button('reset');
    }, 5000);
  });
});
</script>
`````

```html
<button type="button" class="am-btn am-btn-primary btn-loading-example" data-am-loading="{spinner: 'circle-o-notch', loadingText: 'I'm Loading...', resetText: 'Loaded'}">Button - button element</button> &nbsp;
<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="Button - input element" data-am-loading="{loadingText: 'Still Loading...'}" />
<script>
$(function() {
  $('.btn-loading-example').click(function() {
    var $btn = $(this);
    $btn.button('loading');
    setTimeout(function() {
      $btn.button('reset');
    }, 5000);
  });
});
</script>
```

## Toggle Button

`````html
<button id="doc-single-toggle" type="button" class="am-btn am-btn-primary" data-am-button>Switch</button>

<p>Status：<span id="doc-single-toggle-status" class="am-text-danger">Inactivated</span></p>

<script>
  $(function() {
    var $toggleButton = $('#doc-single-toggle');
    $toggleButton.on('click', function() {
      setButtonStatus();
    });

    function setButtonStatus() {
      var status = $toggleButton.hasClass('am-active') ? 'Inactivated' : 'Activated';
      $('#doc-single-toggle-status').text(status);
    }
  })
</script>
`````
```html
<button id="doc-single-toggle" type="button" class="am-btn am-btn-primary" data-am-button>Switch</button>

<p>Status：<span id="doc-single-toggle-status" class="am-text-danger">Inactivated</span></p>

<script>
  $(function() {
    var $toggleButton = $('#doc-single-toggle');
    $toggleButton.on('click', function() {
      setButtonStatus();
    });

    function setButtonStatus() {
      var status = $toggleButton.hasClass('am-active') ? 'Inactivated' : 'Activated';
      $('#doc-single-toggle-status').text(status);
    }
  })
</script>
```

## Checkbox

**Attention**: Because of the FastClick, please add the `.needsclick` class to `input` element to allow getting value from checkbox on touch screen devices.

`````html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Apple"> Apple
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Orange"> Orange
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Banana"> Banana
  </label>
</div>
<script>
  $(function() {
    var $cb = $('[name="doc-js-btn"]');
    $cb.on('change', function() {
      var checked = [];
      $cb.filter(':checked').each(function() {
        checked.push(this.value);
      });

      console.log('The checked checkboxes are：', checked.join(' | '));
    });
  });
</script>
`````
```html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Apple"> Apple
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Orange"> Orange
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox" class="needsclick" name="doc-js-btn" value="Banana"> Banana
  </label>
</div>
<script>
  $(function() {
    var $cb = $('[name="doc-js-btn"]');
    $cb.on('change', function() {
      var checked = [];
      $cb.filter(':checked').each(function() {
        checked.push(this.value);
      });

      console.log('The checked checkboxes are：', checked.join(' | '));
    });
  });
</script>
```

## Radio

`````html
<div class="am-btn-group doc-js-btn-1" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 1" id="option1"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 2" id="option2"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 3" id="option3"> Option 3
  </label>
  <label class="am-btn am-btn-primary am-disabled">
    <input type="radio" name="options" value="Option 4" id="option4"> Option 4
  </label>
</div>
<script>
  // Get selected value
  $(function() {
    var $radios = $('[name="options"]');
    $radios.on('change',function() {
      console.log('The selected radio is: ', $radios.filter(':checked').val());
    });
  });
</script>
`````

```html
<div class="am-btn-group doc-js-btn-1" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 1" id="option1"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 2" id="option2"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" value="Option 3" id="option3"> Option 3
  </label>
  <label class="am-btn am-btn-primary am-disabled">
    <input type="radio" name="options" value="Option 4" id="option4"> Option 4
  </label>
</div>
<script>
  // Get selected value
  $(function() {
    var $radios = $('[name="options"]');
    $radios.on('change',function() {
      console.log('The selected radio is: ', $radios.filter(':checked').val());
    });
  });
</script>
```
