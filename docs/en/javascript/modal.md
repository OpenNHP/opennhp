---
id: modal
title: 模态窗口
titleEn: Modal
prev: javascript/dropdown.html
next: javascript/popover.html
source: js/ui.modal.js
doc: docs/javascript/modal.md
---

# Modal
---

Modal can be used to simulate the `alert`, `confirm` and `prompt`  window in browser.

## Examples

### Default Style

`closeViaDimmer`, `width` and `height` are set in this Demo.

`````html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#doc-modal-1', closeViaDimmer: 0, width: 400, height: 225}">Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal Title
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal Content. This Modal can't be closed via dimmer.
    </div>
  </div>
</div>
`````

```html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#doc-modal-1', closeViaDimmer: 0, width: 400, height: 225}">Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal Title
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal Content. This Modal can't be closed via dimmer.
    </div>
  </div>
</div>
```

### Simulate Alert

`````html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#my-alert'}">Alert</button>

<div class="am-modal am-modal-alert" tabindex="-1" id="my-alert">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Hello world！
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn am-modal-btn-bold">OK</span>
    </div>
  </div>
</div>
`````
```html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#my-alert'}">Alert</button>

<div class="am-modal am-modal-alert" tabindex="-1" id="my-alert">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Hello world！
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn am-modal-btn-bold">OK</span>
    </div>
  </div>
</div>
```

### Simulate Confirm

Click on the `x` on the right of the list to check effect.

`````html
<button class="am-btn am-btn-warning" id="doc-confirm-toggle">Confirm</button>

<style>
  .confirm-list i {
    position: absolute;
    right: 0;
    top: 12px;
    color: #888;
    width: 32px;
    text-align: center;
    cursor: pointer;
  }

  .confirm-list i:hover {
    color: #555;
  }
</style>

<ul class="am-list confirm-list" id="doc-modal-list">
  <li><a data-id="1" href="#">Devouring Time,blunt thou the lion'paws,</a><i class="am-icon-close"></i></li>
  <li><a data-id="2" href="#">And make the earth devour her own sweet brood;</a><i class="am-icon-close"></i></li>
  <li><a data-id="3" href="#">Pluck the keen teeth from the fierce tiger's jaws,</a><i class="am-icon-close"></i></li>
  <li><a data-id="4" href="#">And burn the long--liv'd phoenix in her blood;</a><i class="am-icon-close"></i></li>
</ul>


<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Are you sure you want to delete this recored?
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>Cancel</span>
      <span class="am-modal-btn" data-am-modal-confirm>OK</span>
    </div>
  </div>
</div>

<script>
$(function() {
  $('#doc-modal-list').find('.am-icon-close').add('#doc-confirm-toggle').
    on('click', function() {
      $('#my-confirm').modal({
        relatedTarget: this,
        onConfirm: function(e) {
          var $link = $(this.relatedTarget).prev('a');
          var msg = $link.length ? 'The ID of deleted link is ' + $link.data('id') :
            'OK, so what?';
          alert(msg);
        },
        onCancel: function(e) {
          alert('Fine, fine, fine.');
        }
      });
    });
});
</script>
`````
```html
<button class="am-btn am-btn-warning" id="doc-confirm-toggle">Confirm</button>

<style>
  .confirm-list i {
    position: absolute;
    right: 0;
    top: 12px;
    color: #888;
    width: 32px;
    text-align: center;
    cursor: pointer;
  }

  .confirm-list i:hover {
    color: #555;
  }
</style>

<ul class="am-list confirm-list" id="doc-modal-list">
  <li><a data-id="1" href="#">Devouring Time,blunt thou the lion'paws,</a><i class="am-icon-close"></i></li>
  <li><a data-id="2" href="#">And make the earth devour her own sweet brood;</a><i class="am-icon-close"></i></li>
  <li><a data-id="3" href="#">Pluck the keen teeth from the fierce tiger's jaws,</a><i class="am-icon-close"></i></li>
  <li><a data-id="4" href="#">And burn the long--liv'd phoenix in her blood;</a><i class="am-icon-close"></i></li>
</ul>


<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Are you sure you want to delete this recored?
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>Cancel</span>
      <span class="am-modal-btn" data-am-modal-confirm>OK</span>
    </div>
  </div>
</div>

<script>
$(function() {
  $('#doc-modal-list').find('.am-icon-close').add('#doc-confirm-toggle').
    on('click', function() {
      $('#my-confirm').modal({
        relatedTarget: this,
        onConfirm: function(e) {
          var $link = $(this.relatedTarget).prev('a');
          var msg = $link.length ? 'The ID of deleted link is ' + $link.data('id') :
            'OK, so what?';
          alert(msg);
        },
        onCancel: function(e) {
          alert('Fine, fine, fine.');
        }
      });
    });
});
</script>
```

**Problems:**

Considering effeciency, instances of models are stored in the `$('.am-modal').data('am.modal')` attribute of corresponding element. `onConfirm`/`onCancel` will be set when Modal is ran first time, so Model may not work as expect in some scenario.([#274](https://github.com/allmobilize/amazeui/issues/274#issuecomment-65182344)). We tried to fix this bug in version `2.1`, but our solution didn't work perfectly. If you have better solution, please contact us.

Optional Solutions:

- **1**: Using `relatedTarget` to get data. As shown above, use this element as a bridge to get the data you want. (**Supported from version 2.1 **); 
- 2: Use [**this way**](http://jsbin.com/fahawe/1/edit?html,output) to assign the value to these two parameters;
- 3: Remove the cached instance when Confirm is closed, and initialize it again when it is called next time.

```javascript
$('#your-confirm').on('closed.modal.amui', function() {
  $(this).removeData('amui.modal');
});
```

- 4: Rewrite the event handler of `cancel`/`confirm` button.

```javascript
$(function() {
  var $confirm = $('#your-confirm');
  var $confirmBtn = $confirm.find('[data-am-modal-confirm]');
  var $cancelBtn = $confirm.find('[data-am-modal-cancel]');
  $confirmBtn.off('click.confirm.modal.amui').on('click', function() {
    // do something
  });

  $cancelBtn.off('click.cancel.modal.amui').on('click', function() {
    // do something
  });
});
```

Solution 2 and 3 solve this problem in expense of performance; if use solution 4, we don't need to use `data-am-modal-xx` at the beginning, and then we don't need to unbind the default event.

### Simulate Prompt

Prompt support multiple input box since version `2.1`. The input value can be get using `options.data`: 

- When there is only one input box, `options.data` is a string；
- When there are multiple input boxes, `options.data` is an array.

`````html
<button class="am-btn am-btn-success" id="doc-prompt-toggle">Prompt</button>

<div class="am-modal am-modal-prompt" tabindex="-1" id="my-prompt">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Just say something.
      <input type="text" class="am-modal-prompt-input">
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>Cancel</span>
      <span class="am-modal-btn" data-am-modal-confirm>Submit</span>
    </div>
  </div>
</div>
<script>
$(function() {
  $('#doc-prompt-toggle').on('click', function() {
    $('#my-prompt').modal({
      relatedTarget: this,
      onConfirm: function(e) {
        alert('Your input is: ' + e.data || '');
      },
      onCancel: function(e) {
        alert('You said nothing!');
      }
    });
  });
});
</script>
`````
```html
<button class="am-btn am-btn-success" id="doc-prompt-toggle">Prompt</button>

<div class="am-modal am-modal-prompt" tabindex="-1" id="my-prompt">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Just say something.
      <input type="text" class="am-modal-prompt-input">
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>Cancel</span>
      <span class="am-modal-btn" data-am-modal-confirm>Submit</span>
    </div>
  </div>
</div>
<script>
$(function() {
  $('#doc-prompt-toggle').on('click', function() {
    $('#my-prompt').modal({
      relatedTarget: this,
      onConfirm: function(e) {
        alert('Your input is: ' + e.data || '');
      },
      onCancel: function(e) {
        alert('You said nothing!');
      }
    });
  });
});
</script>
```

### Modal Loading

Suggested by Developers, Loading window can only be closed through JS.

`````html
<button class="am-btn am-btn-success" data-am-modal="{target: '#my-modal-loading'}">Modal Loading</button>

<div class="am-modal am-modal-loading am-modal-no-btn" tabindex="-1" id="my-modal-loading">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Loading...</div>
    <div class="am-modal-bd">
      <span class="am-icon-spinner am-icon-spin"></span>
    </div>
  </div>
</div>
`````
```html
<button class="am-btn am-btn-success" data-am-modal="{target: '#my-modal-loading'}">Modal Loading</button>

<div class="am-modal am-modal-loading am-modal-no-btn" tabindex="-1" id="my-modal-loading">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Loading...</div>
    <div class="am-modal-bd">
      <span class="am-icon-spinner am-icon-spin"></span>
    </div>
  </div>
</div>
```

### Actions

Use actions with [List](/css/list) to create iOS style operation list.

`````html
<button class="am-btn am-btn-secondary" data-am-modal="{target: '#my-actions'}">Actions</button>

<div class="am-modal-actions" id="my-actions">
  <div class="am-modal-actions-group">
    <ul class="am-list">
      <li class="am-modal-actions-header">What do you want?What do you want?What do you want?What do you want?What do you want?What do you want?What do you want?</li>
      <li><a href="#"><span class="am-icon-wechat"></span> Share to Wechat</a></li>
      <li><a href="#"><i class="am-icon-mobile"></i> Share to message</a></li>
      <li class="am-modal-actions-danger"><a href="#"><i class="am-icon-twitter"></i> Share to XX 萎跛</a></li>
    </ul>
  </div>
  <div class="am-modal-actions-group">
    <button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>Cancel</button>
  </div>
</div>
`````
```html
<button class="am-btn am-btn-secondary" data-am-modal="{target: '#my-actions'}">Actions</button>

<div class="am-modal-actions" id="my-actions">
  <div class="am-modal-actions-group">
    <ul class="am-list">
      <li class="am-modal-actions-header">...</li>
      <li><a href="#"><span class="am-icon-wechat"></span> ...</a></li>
      <li class="am-modal-actions-danger">
        <a href="#"><i class="am-icon-twitter"></i> ...</a>
      </li>
    </ul>
  </div>
  <div class="am-modal-actions-group">
    <button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>Cancel</button>
  </div>
</div>
```

### Popup


`````html
<button class="am-btn am-btn-danger" data-am-modal="{target: '#my-popup'}">Popup</button>

<div class="am-popup" id="my-popup">
  <div class="am-popup-inner">
    <div class="am-popup-hd">
      <h4 class="am-popup-title">Title - Rolling in the Deep</h4>
      <span data-am-modal-close
            class="am-close">&times;</span>
    </div>
    <div class="am-popup-bd">
      <p>There's a fire starting in my heart<br/>Reaching a fever pitch and it's bringing me out the dark<br/>Finally I can see you crystal clear<br/>Go ahead and sell me out and I'll lay your ship bare<br/>See how I leave with every piece of you<br/>Don't underestimate the things that I will do<br/>There's a fire starting in my heart<br/>Reaching a fever pitch and it's bringing me out the dark<br/>The scars of your love remind me of us<br/>They keep me thinking that we almost had it all<br/>The scars of your love they leave me breathless<br/>I can't help feeling<br/>We could have had it all<br/>Rolling in the Deep<br/>
      </p>

      <p>Your had my heart Inside of your hand<br/>And you played it To the beat （Rolling in the deep）<br/>Baby I have no story to be told<br/>But I've heard one of you and I'm gonna make your head burn<br/>Think of me in the depths of your despair<br/>Making a home down there as mine sure won't be shared<br/>The scars of your love remind you of us<br/>They keep me thinking that we almost had it all<br/>The scars of your love they leave me breathless<br/>I can't help feeling<br/>We could have had it all<br/>Rolling in the Deep<br/>Your had my heart inside of your hand<br/>And you played it<br/>To the beat<br/>Could have had it all<br/>Rolling in the deep<br/>You had my heart inside of your hand<br/>But you played it with your beating<br/>Throw yourself through ever open door<br/>Count your blessings to find what look for<br/>Turn my sorrow into treasured gold<br/>And pay me back in kind- You reap just what you sow<br/>We could have had it all<br/>We could have had it all yeah<br/>It all<br/>It all<br/>We could have had it all<br/>Rolling in the deep<br/>You had my heart inside of your hand<br/>And you played it to the beat<br/>We could have had it all<br/>Rolling in the deep<br/>You had my heart<br/>Inside of your hand<br/>But you played it<br/>You played it<br/>You played it<br/>You played it to the beat
      </p>
    </div>
  </div>
</div>
`````

```html
<button class="am-btn am-btn-danger" data-am-modal="{target: '#my-popup'}">Popup</button>

<div class="am-popup" id="my-popup">
  <div class="am-popup-inner">
    <div class="am-popup-hd">
      <h4 class="am-popup-title">...</h4>
      <span data-am-modal-close
            class="am-close">&times;</span>
    </div>
    <div class="am-popup-bd">
      ...
    </div>
  </div>
</div>
```

## Usage

### Using Data API

Add `data-am-modal="{target: '#my-modal'}"` to elements like `<button>` and `<a>` to create an modal, where `#my-modal` is the ID of Modal container.

```html
<button data-am-modal="{target: '#my-modal'}">My Modal</button>
```

### Using JS

When Modal HTML is organized, Model can be controlled using Javascript.

```javascript
$('#your-modal').modal(options);
```
`````html
<button class="am-btn am-btn-primary js-modal-open">Show Modal</button>
  <button class="am-btn am-btn-secondary js-modal-close">Close Modal</button>
  <button class="am-btn am-btn-danger js-modal-toggle">Toggle Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="your-modal">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal Title
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal Contents.
    </div>
  </div>
</div>
<script>
$(function() {
  var $modal = $('#your-modal');

  $modal.siblings('.am-btn').on('click', function(e) {
    var $target = $(e.target);
    if (($target).hasClass('js-modal-open')) {
      $modal.modal();
    } else if (($target).hasClass('js-modal-close')) {
      $modal.modal('close');
    } else {
      $modal.modal('toggle');
    }
  });
});
</script>
`````

```html
<button class="am-btn am-btn-primary js-modal-open">Show Modal</button>
  <button class="am-btn am-btn-secondary js-modal-close">Close Modal</button>
  <button class="am-btn am-btn-danger js-modal-toggle">Toggle Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="your-modal">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal Title
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal Contents.
    </div>
  </div>
</div>
<script>
$(function() {
  var $modal = $('#your-modal');

  $modal.siblings('.am-btn').on('click', function(e) {
    var $target = $(e.target);
    if (($target).hasClass('js-modal-open')) {
      $modal.modal();
    } else if (($target).hasClass('js-modal-close')) {
      $modal.modal('close');
    } else {
      $modal.modal('toggle');
    }
  });
});
</script>
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
    <td><code>onConfirm</code></td>
    <td><code>function</code></td>
    <td>This function is called when button with <code>data-am-modal-confirm</code> attribute is clicked.</td>
  </tr>
  <tr>
    <td><code>closeOnConfirm</code></td>
    <td><code>bool</code></td>
    <td>Close the modal when the button with <code>data-am-modal-confirm</code> attribute is clicked. Default value is <code>true</code>
     (<strong>Added in v2.4.1</strong>)
    </td>
  </tr>
  <tr>
    <td><code>onCancel</code></td>
    <td><code>function</code></td>
    <td>This function is called when button with <code>data-am-modal-cancel</code> attribute is clicked.</td>
  </tr>
  <tr>
    <td><code>closeOnCancel</code></td>
    <td><code>bool</code></td>
    <td>Close the modal when the button with <code>data-am-modal-cancel</code> attribute is clicked. Default value is <code>true</code>
      （<strong>Added in v2.4.1</strong>）
    </td>
  </tr>
  <tr>
    <td><code>closeViaDimmer</code></td>
    <td><code>boolean</code></td>
    <td>Close the modal when dimmer get clicked. Default value is <code>true</code></td>
  </tr>
  <tr>
    <td><code>width</code></td>
    <td><code>number</code></td>
    <td>Modal width. Invalid for Popup and Actions.</td>
  </tr>
  <tr>
    <td><code>height</code></td>
    <td><code>number</code></td>
    <td>Modal height. Invalid for Popup and Actions.</td>
  </tr>
  </tbody>
</table>

**Attension:**

- **Please don't change `width` and `height` if not necessary. Any change may influence the responsive style.**
- The this pointer in `onConfirm`/`onCanel` points to the Modal instance. The methods and attributes of instance can be accessed through `this.`.


#### Methods

- `.modal(options)` - Activate the Modal window on element. `options` is an object.
- `.modal('toggle')` - Switch between show and hide.
- `.modal('open')` - Show the Modal window.
- `.modal('close')` - Hide the Modal window.

#### Events

Event will be emitted from the popup window.

```javascript
$('#doc-modal-1').on('open.modal.amui', function(){
  console.log('第一个演示弹窗打开了');
});
```

Copy the code above to the console, and then open the first modal window (in `1.1`). Message will be shown in console.

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Event</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.modal.amui</code></td>
    <td>Fired immediately when <code>open</code> method get called</td>
  </tr>
  <tr>
    <td><code>opened.modal.amui</code></td>
    <td>Fired after Modal window is opend (After CSS animation is finished).</td>
  </tr>
  <tr>
    <td><code>close.modal.amui</code></td>
    <td>Fired immediately when <code>close</code> method get called</td>
  </tr>
  <tr>
    <td><code>closed.modal.amui</code></td>
    <td>Fired after Modal window is closed (After CSS animation is finished).</td>
  </tr>
  </tbody>
</table>

<script>
$(function() {
  $(document).on('open.modal.amui opened.modal.amui close.modal.amui closed.modal.amui', function(e) {
    console.log('#' + $(e.target).attr('id') + ' triggered ' + e.type + ' event');
  });
});
</script>
