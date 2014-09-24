# Button JS 交互
---

Button 及 Button group 与 JS 交互。

## 按钮 loading 状态

### 默认文字

默认的文字为 `loading...`。

`````html
<button type="button" class="am-btn am-btn-primary btn-loading-example">按钮 - button 元素</button> &nbsp;
<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="按钮 - input 元素" />
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

### 自定义选项

可以在元素上添加 `data-am-loading` 来设置选项：

- `spinner` 加载动画图标，适用于支持 CSS3 动画、非 `input` 元素，写图标名称即可；
- `loadingText` 加载时显示的文字， 默认为 `loading`；
- `resetText` 重置以后的显示的文字，默认为原来的内容。

`````html
<button type="button" class="am-btn am-btn-primary btn-loading-example" data-am-loading="{spinner: 'circle-o-notch', loadingText: '加载中...', resetText: '加载过了'}">按钮 - button 元素</button> &nbsp;
<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="按钮 - input 元素" data-am-loading="{loadingText: '努力加载中...'}" />
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
<button type="button" class="am-btn am-btn-primary btn-loading-example" data-am-loading="{spinner: 'circle-o-notch', loadingText: '加载中...', resetText: '加载过了'}">按钮 - button 元素</button>

<input type="button" class="am-btn am-btn-secondary btn-loading-example" value="按钮 - input 元素" data-am-loading="{loadingText: '努力加载中...'}" />
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

## 单按钮状态切换

`````html
<button id="doc-single-toggle" type="button" class="am-btn am-btn-primary" data-am-button>切换状态</button>

<p>按钮状态：<span id="doc-single-toggle-status" class="am-text-danger">未激活</span></p>

<script>
  $(function() {
    var $toggleButton = $('#doc-single-toggle');
    $toggleButton.on('click', function() {
      setButtonStatus();
    });

    function setButtonStatus() {
      var status = $toggleButton.hasClass('am-active') ? '未激活' : '激活';
      $('#doc-single-toggle-status').text(status);
    }
  })
</script>
`````
```html
<button type="button" class="am-btn am-btn-primary" data-am-button>切换状态</button>
```

## 复选框

`````html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 3
  </label>
</div>
`````
```html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> 选项 3
  </label>
</div>
```

## 单选框

`````html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option1"> 选项 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option2"> 选项 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option3"> 选项 3
  </label>
</div>
`````

```html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option1"> 选项 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option2"> 选项 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option3"> 选项 3
  </label>
</div>
```