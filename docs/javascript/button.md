# Button JS 交互
---

Button 及 Button group 与 JS 交互。

## 按钮 loading 状态

支持 CSS3 动画的非 `input` 元素会自动添加一个图标。

`````html
<button type="button" class="am-btn am-btn-primary btn-loading-example">Submit - Button</button> &nbsp;
<input type="button" class="am-btn am-btn-primary btn-loading-example" value="Submit - Input" />
<script>
  seajs.use(['ui.button'], function(){
    $('.btn-loading-example').click(function () {
      var $btn = $(this);
      $btn.button('loading');
      setTimeout(function(){
        $btn.button('reset');
      }, 3000);
    });
  });
</script>
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
  }, 3000);
});
```

## 单按钮状态切换

`````html
<button type="button" class="am-btn am-btn-primary" data-am-button>Single toggle</button>
`````
```html
<button type="button" class="btn btn-primary" data-toggle="button">Single toggle</button>
```

## 复选框

`````html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 3
  </label>
</div>
<script>
  Zepto(function($) {

  });
</script>
`````
```html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="checkbox"> Option 3
  </label>
</div>
```

## 单选框

`````html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option1"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option2"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option3"> Option 3
  </label>
</div>
`````

```html
<div class="am-btn-group" data-am-button>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option1"> Option 1
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option2"> Option 2
  </label>
  <label class="am-btn am-btn-primary">
    <input type="radio" name="options" id="option3"> Option 3
  </label>
</div>
```