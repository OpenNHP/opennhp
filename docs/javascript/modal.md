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

Modal 交互窗口，可以用来模拟浏览器的 `alert`、`confirm`、`prompt` 窗口。

## 使用演示

使用过程中需按示例代码组织好 HTML。

### 基本形式

此 Demo 设置了 `closeViaDimmer`、`width`、`height` 参数。

`````html
<button
  type="button"
  class="am-btn am-btn-primary"
  data-am-modal="{target: '#doc-modal-1', closeViaDimmer: 0, width: 400, height: 225}">
  Modal
</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      <button class="am-btn am-btn-primary am-fr am-margin-left" data-am-popover="{content: '谁愿压抑心中怒愤冲动，咒骂这虚与伪与假'}">点击显示 Popover</button>
      Modal 内容。本 Modal 无法通过遮罩层关闭。
    </div>
  </div>
</div>
`````

```html
<button
  type="button"
  class="am-btn am-btn-primary"
  data-am-modal="{target: '#doc-modal-1', closeViaDimmer: 0, width: 400, height: 225}">
  Modal
</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal 内容。本 Modal 无法通过遮罩层关闭。
    </div>
  </div>
</div>
```

### 模拟 Alert

`````html
<button
  type="button"
  class="am-btn am-btn-primary"
  data-am-modal="{target: '#my-alert'}">
  Alert
</button>

<div class="am-modal am-modal-alert" tabindex="-1" id="my-alert">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Hello world！
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn am-modal-btn-bold">确定</span>
    </div>
  </div>
</div>
`````
```html
<button
  type="button"
  class="am-btn am-btn-primary"
  data-am-modal="{target: '#my-alert'}">
  Alert
</button>

<div class="am-modal am-modal-alert" tabindex="-1" id="my-alert">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      Hello world！
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn">确定</span>
    </div>
  </div>
</div>
```

### 模拟 Confirm

点击列表右边的 `x` 查看效果。

`````html
<button
  type="button"
  class="am-btn am-btn-warning"
  id="doc-confirm-toggle">
  Confirm
</button>

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
  <li><a data-id="1" href="#">每个人都有一个死角， 自己走不出来，别人也闯不进去。</a><i class="am-icon-close"></i></li>
  <li><a data-id="2" href="#">我把最深沉的秘密放在那里。</a><i class="am-icon-close"></i></li>
  <li><a data-id="3" href="#">你不懂我，我不怪你。</a><i class="am-icon-close"></i></li>
  <li><a data-id="4" href="#">每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a><i class="am-icon-close"></i></li>
</ul>


<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      你，确定要删除这条记录吗？
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>取消</span>
      <span class="am-modal-btn" data-am-modal-confirm>确定</span>
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
          var msg = $link.length ? '你要删除的链接 ID 为 ' + $link.data('id') :
            '确定了，但不知道要整哪样';
          alert(msg);
        },
        onCancel: function(e) {
          alert('算求，不弄了');
        }
      });
    });
});
</script>
`````
```html
<button
  type="button"
  class="am-btn am-btn-warning"
  id="doc-confirm-toggle">
  Confirm
</button>

<ul class="am-list confirm-list" id="doc-modal-list">
  <li><a data-id="1" href="#">每个人都有一个死角， 自己走不出来，别人也闯不进去。</a><i class="am-icon-close"></i></li>
  <li><a data-id="2" href="#">我把最深沉的秘密放在那里。</a><i class="am-icon-close"></i></li>
  <li><a data-id="3" href="#">你不懂我，我不怪你。</a><i class="am-icon-close"></i></li>
  <li><a data-id="4" href="#">每个人都有一道伤口， 或深或浅，盖上布，以为不存在。</a><i class="am-icon-close"></i></li>
</ul>

<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      你，确定要删除这条记录吗？
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>取消</span>
      <span class="am-modal-btn" data-am-modal-confirm>确定</span>
    </div>
  </div>
</div>
```
```javascript
$(function() {
  $('#doc-modal-list').find('.am-icon-close').add('#doc-confirm-toggle').
    on('click', function() {
      $('#my-confirm').modal({
        relatedTarget: this,
        onConfirm: function(options) {
          var $link = $(this.relatedTarget).prev('a');
          var msg = $link.length ? '你要删除的链接 ID 为 ' + $link.data('id') :
            '确定了，但不知道要整哪样';
          alert(msg);
        },
        // closeOnConfirm: false,
        onCancel: function() {
          alert('算求，不弄了');
        }
      });
    });
});
```

**存在问题：**

出于性能考虑，每个 Modal 实例都存储在对应元素的 `$('.am-modal').data('amui.modal')` 属性中，`onConfirm`/`onCancel` 会保存第一次运行 Modal 时候的数据，导致在某些场景不能按照预期工作（[#274](https://github.com/allmobilize/amazeui/issues/274#issuecomment-65182344)）。`2.1` 中做了一些处理，但并不是很如意，大家有更好的方案可以提供给我们。

可以选择的处理方式：

- **法一**：通过 `relatedTarget` 这个钩子获取数据，如上面的演示，以该元素为桥梁获取想要的数据（**2.1 开始支持**）；
- 法二：按照[**这种方式**](http://jsbin.com/fahawe/edit?html,output) 每次都重新给这两个参数赋值；
- 法三：Confirm 关闭后移除暂存的实例，再次调用时重新初始化；

```javascript
$('#your-confirm').on('closed.modal.amui', function() {
  $(this).removeData('amui.modal');
});
```

- 法四： 重写 `cancel`/`confirm` 按钮的事件处理函数。

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

法二、三以牺牲性能为代价，**不推荐**；如果要使用法四，可以不添加两个按钮的 `data-am-modal-xx` 属性，免去解绑默认事件的步奏。

### 模拟 Prompt

Prompt 从 `2.1` 开始支持多个输入框，输入框的值通过参数 `options.data` 获取：

- 只有一个输入框时，`options.data` 为字符串；
- 多个输入框时，`options.data` 为数组。

`````html
<button
  type="button"
  class="am-btn am-btn-success"
  id="doc-prompt-toggle">
  Prompt
</button>

<div class="am-modal am-modal-prompt" tabindex="-1" id="my-prompt">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      来来来，吐槽点啥吧
      <input type="text" class="am-modal-prompt-input">
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>取消</span>
      <span class="am-modal-btn" data-am-modal-confirm>提交</span>
    </div>
  </div>
</div>
<script>
$(function() {
  $('#doc-prompt-toggle').on('click', function() {
    $('#my-prompt').modal({
      relatedTarget: this,
      onConfirm: function(e) {
        alert('你输入的是：' + e.data || '');
      },
      onCancel: function(e) {
        alert('不想说!');
      }
    });
  });
});
</script>
`````
```html
<button
  type="button"
  class="am-btn am-btn-success"
  id="doc-prompt-toggle">
  Prompt
</button>

<div class="am-modal am-modal-prompt" tabindex="-1" id="my-prompt">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      来来来，吐槽点啥吧
      <input type="text" class="am-modal-prompt-input">
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>取消</span>
      <span class="am-modal-btn" data-am-modal-confirm>提交</span>
    </div>
  </div>
</div>
```
```javascript
$(function() {
  $('#doc-prompt-toggle').on('click', function() {
    $('#my-prompt').modal({
      relatedTarget: this,
      onConfirm: function(e) {
        alert('你输入的是：' + e.data || '')
      },
      onCancel: function(e) {
        alert('不想说!');
      }
    });
  });
});
```

### Modal Loading

采纳网友意见，Loading 窗口只能通过 JS 关闭。

`````html
<button
  type="button"
  class="am-btn am-btn-success"
  data-am-modal="{target: '#my-modal-loading'}">
  Modal Loading
</button>

<div class="am-modal am-modal-loading am-modal-no-btn" tabindex="-1" id="my-modal-loading">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">正在载入...</div>
    <div class="am-modal-bd">
      <span class="am-icon-spinner am-icon-spin"></span>
    </div>
  </div>
</div>
`````
```html
<button
  type="button"
  class="am-btn am-btn-success"
  data-am-modal="{target: '#my-modal-loading'}">
  Modal Loading
</button>

<div class="am-modal am-modal-loading am-modal-no-btn" tabindex="-1" id="my-modal-loading">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">正在载入...</div>
    <div class="am-modal-bd">
      <span class="am-icon-spinner am-icon-spin"></span>
    </div>
  </div>
</div>
```

### Actions

结合 [List 组件](/css/list)使用，创建类似 iOS 的操作列表。

`````html
<button
  type="button"
  class="am-btn am-btn-secondary"
  data-am-modal="{target: '#my-actions'}">
  Actions
</button>

<div class="am-modal-actions" id="my-actions">
  <div class="am-modal-actions-group">
    <ul class="am-list">
      <li class="am-modal-actions-header">你想整哪样？你想整哪样？你想整哪样？你想整哪样？你想整哪样？你想整哪样？你想整哪样？</li>
      <li><a href="#"><span class="am-icon-wechat"></span> 分享到微信</a></li>
      <li><a href="#"><i class="am-icon-mobile"></i> 短信分享</a></li>
      <li class="am-modal-actions-danger"><a href="#"><i class="am-icon-twitter"></i> 分享到 XX 萎跛</a></li>
    </ul>
  </div>
  <div class="am-modal-actions-group">
    <button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button>
  </div>
</div>
`````
```html
<button
  type="button"
  class="am-btn am-btn-secondary"
  data-am-modal="{target: '#my-actions'}">
  Actions
</button>

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
    <button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button>
  </div>
</div>
```

### Popup


`````html
<button
  type="button"
  class="am-btn am-btn-danger"
  data-am-modal="{target: '#my-popup'}">
  Popup
</button>

<div class="am-popup" id="my-popup">
  <div class="am-popup-inner">
    <div class="am-popup-hd">
      <h4 class="am-popup-title">标题 - 女爵</h4>
      <span data-am-modal-close
            class="am-close">&times;</span>
    </div>
    <div class="am-popup-bd"><p>为你封了国境<br/>为你赦了罪<br/>为你撤了历史记载<br/>为你涂了装扮<br/>为你喝了醉<br/>为你建了城池围墙<br/>一颗热的心穿着冰冷外衣<br/>一张白的脸漆上多少褪色的情节<br/>在我的空虚身体里面<br/>爱上哪个肤浅的王位<br/>在你的空虚宝座里面<br/>爱过什麽女爵的滋味<br/>为你封了国境
    </p>

      <p>为你赦了罪<br/>为你撤了历史记载<br/>一颗热的心<br/>穿着冰冷外衣<br/>一张白的脸<br/>漆上多少褪色的情节<br/>在我的空虚身体里面<br/>爱上哪个肤浅的王位<br/>在你的空虚宝座里面<br/>爱过什麽女爵的滋味<br/>在我的空虚身体里面<br/>爱上哪个肤浅的王位<br/>在你的空虚宝座里面<br/>爱过什麽女爵的滋味
      </p></div>
  </div>
</div>
`````

```html
<button
  type="button"
  class="am-btn am-btn-danger"
  data-am-modal="{target: '#my-popup'}">
  Popup
</button>

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

## 使用方式

### 通过 Data API

在 `<button>`、`<a>` 等元素上添加 `data-am-modal="{target: '#my-modal'}"`，其中 `#my-modal` 为 Modal 窗口容器 ID。

```html
<button
  type="button"
  data-am-modal="{target: '#my-modal'}">
  My Modal
</button>
```

### 通过 JS

组织好 Modal HTML 以后，可以通过 Javascript 调用。

```javascript
$('#your-modal').modal(options);
```
`````html
<button type="button" class="am-btn am-btn-primary js-modal-open">打开 Modal</button>
  <button type="button" class="am-btn am-btn-secondary js-modal-close">关闭 Modal</button>
  <button type="button" class="am-btn am-btn-danger js-modal-toggle">Toggle Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="your-modal">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal 内容。
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
  <button type="button" class="am-btn am-btn-primary js-modal-open">打开 Modal</button>
  <button type="button" class="am-btn am-btn-secondary js-modal-close">关闭 Modal</button>
  <button type="button" class="am-btn am-btn-danger js-modal-toggle">Toggle Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="your-modal">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="javascript: void(0)" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal 内容。
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

#### 参数说明

| 参数 | 类型 | 描述 |
| ----| --- | --- |
| `onConfirm` | `function` | 具有 <code>data-am-modal-confirm</code> 属性的按钮关闭时触发的函数 |
| `closeOnConfirm` | `bool` | 具有 <code>data-am-modal-confirm</code> 属性的按钮点击时是否关闭 Modal，默认为 <code>true</code>（<strong>v2.4.1 新增</strong>）|
| `onCancel` | `function` | 具有 <code>data-am-modal-cancel</code> 属性的按钮关闭时触发的函数 |
| `closeOnCancel` | `bool` | 具有 <code>data-am-modal-cancel</code> 属性的按钮点击时是否关闭 Modal，默认为 <code>true</code>（<strong>v2.4.1 新增</strong>）|
| `closeViaDimmer` | `bool` | 点击遮罩层时关闭 Modal，默认为 `true` |
| `width` | `number` | Modal 宽度，对 Popup 和 Actions 无效 |
| `height`| `number` | Modal 高度，对 Popup 和 Actions 无效 |
| `dimmer` | `bool` | 是否显示 Modal 遮罩背景，默认为 `true` (**v2.5**)|

**注意：**

- **如无必要，请不要设置 `width`、`height`，以免破坏响应式样式。**
- `onConfirm`/`onCanel` 函数中 `this` 指向 Modal 实例，可以通过 `this.` 访问实例的方法和属性。


#### 方法

- `.modal(options)` - 激活元素的 Modal 窗口交互，`options` 为对象
- `.modal('toggle')` - 交替 Modal 窗口状态
- `.modal('open')` - 显示 Modal 窗口
- `.modal('close')` - 关闭 Modal 窗口

#### 自定义事件

自定义事件在弹窗上触发，可以监听弹窗元素来执行其他操作。

```javascript
$('#doc-modal-1').on('open.modal.amui', function(){
  console.log('第一个演示弹窗打开了');
});
```

拷贝上面的代码粘贴到控制台执行，然后每次打开第一个演示弹窗（标题 `1.1` 下面的），控制台都会输出那行文字。

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.modal.amui</code></td>
    <td><code>open</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>opened.modal.amui</code></td>
    <td>Modal 窗口打开完成以后触发（CSS 动画执行完成）</td>
  </tr>
  <tr>
    <td><code>close.modal.amui</code></td>
    <td><code>close</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>closed.modal.amui</code></td>
    <td>Modal 窗口完全关闭以后触发（CSS 动画执行完成）</td>
  </tr>
  </tbody>
</table>

<script>
$(function() {
  $(document).on('open.modal.amui opened.modal.amui close.modal.amui closed.modal.amui', function(e) {
    console.log('#' + $(e.target).attr('id') + ' 触发了 ' + e.type + ' 事件');
  });
});
</script>
