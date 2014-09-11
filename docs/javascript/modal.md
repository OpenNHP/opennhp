# Modal
---

Modal 交互窗口，可以用来模拟浏览器的 `alert`、`confirm`、`prompt` 窗口。

## 使用演示

使用过程中需按示例代码组织好 HTML。

### 基本形式

`````html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#doc-modal-1'}">Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal 内容。
    </div>
  </div>
</div>
`````

```html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#doc-modal-1'}">Modal</button>

<div class="am-modal am-modal-no-btn" tabindex="-1" id="doc-modal-1">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Modal 标题
      <a href="" class="am-close am-close-spin" data-am-modal-close>&times;</a>
    </div>
    <div class="am-modal-bd">
      Modal 内容。
    </div>
  </div>
</div>
```

### 模拟 Alert

`````html
<button class="am-btn am-btn-primary" data-am-modal="{target: '#my-alert'}">Alert</button>

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
<button class="am-btn am-btn-primary" data-am-modal="{target: '#my-alert'}">Alert</button>

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

`````html
<button class="am-btn am-btn-warning" id="doc-confirm-toggle">Confirm</button>

<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      你是猴子派来的逗比吗?
    </div>
    <div class="am-modal-footer">
      <span class="am-modal-btn" data-am-modal-cancel>取消</span>
      <span class="am-modal-btn" data-am-modal-confirm>确定</span>
    </div>
  </div>
</div>

<script>
$(function() {
  $('#doc-confirm-toggle').on('click', function() {
    $('#my-confirm').modal({
      relatedElement: this,
      onConfirm: function() {
        alert('你是猴子派来的逗比!')
      },
      onCancel: function() {
        alert('你不确定是不是猴子派来的逗比!')
      }
    });
  });
});
</script>
`````
```html
<button class="am-btn am-btn-warning" id="doc-confirm-toggle">Confirm</button>

<div class="am-modal am-modal-confirm" tabindex="-1" id="my-confirm">
  <div class="am-modal-dialog">
    <div class="am-modal-hd">Amaze UI</div>
    <div class="am-modal-bd">
      你是猴子派来的逗比吗?
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
  $('#doc-confirm-toggle').on('click', function() {
    $('#my-confirm').modal({
      relatedElement: this,
      onConfirm: function() {
        alert('你是猴子派来的逗比!')
      },
      onCancel: function() {
        alert('你不确定是不是猴子派来的逗比!')
      }
    });
  });
});
```
    
### 模拟 Prompt

`````html
<button class="am-btn am-btn-success" id="doc-prompt-toggle">Prompt</button>

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
      relatedElement: this,
      onConfirm: function(data) {
        alert('你输入的是：' + data)
      },
      onCancel: function() {
        alert('不想说!');
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
      relatedElement: this,
      onConfirm: function(data) {
        alert('你输入的是：' + data)
      },
      onCancel: function() {
        alert('不想说!');
      }
    });
  });
});
```

### Modal Loading

`````html
<button class="am-btn am-btn-success" data-am-modal="{target: '#my-modal-loading'}">Modal Loading</button>

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
<button class="am-btn am-btn-success" data-am-modal="{target: '#my-modal-loading'}">Modal Loading</button>

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
<button class="am-btn am-btn-secondary" data-am-modal="{target: '#my-actions'}">Actions</button>

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
    <button class="am-btn am-btn-secondary am-btn-block" data-am-modal-close>取消</button>
  </div>
</div>
```

### Popup


`````html
<button class="am-btn am-btn-danger" data-am-modal="{target: '#my-popup'}">Popup</button>

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

## 使用方式

### 通过 Data API

在 `<button>`、`<a>` 等元素上添加 `data-am-modal="{target: '#my-modal'}"`，其中 `#my-modal` 为 Modal 窗口容器 ID。

```html
<button data-am-modal="{target: '#my-modal'}">My Modal</button>
```

### 通过 JS

组织好 Modal HTML 以后，可以通过 Javascript 调用。

```javascript
$('#myModal').modal(options);
```

#### 参数说明

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>参数</th>
    <th>类型</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>onConfirm</code></td>
    <td><code>function</code></td>
    <td>具有 <code>data-am-modal-confirm</code> 属性的按钮关闭时触发的函数</td>
  </tr>
  <tr>
    <td><code>onCancel</code></td>
    <td><code>function</code></td>
    <td>具有 <code>data-am-modal-cancel</code> 属性的按钮关闭时触发的函数</td>
  </tr>
  </tbody>
</table>

#### 方法

- `.modal(options)` - 激活元素的 Modal 窗口交互，`options` 为对象
- `.modal('toggle')` - 交替 Modal 窗口状态
- `.modal('open')` - 显示 Modal 窗口
- `.modal('close')` - 关闭 Modal 窗口

#### 自定义事件

自定义事件在弹窗上触发，可以监听弹窗元素来执行其他操作。

```javascript
$('#doc-modal-1').on('open:modal:amui', function(){
  console.log('第一个演示弹窗打开了');
});
```

拷贝上面的代码粘贴到控制台执行，然后每次打开第一个演示弹窗（标题 `1.1` 下面的），控制台都会输出那行文字。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open:modal:amui</code></td>
    <td><code>open</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>opened:modal:amui</code></td>
    <td>Modal 窗口被关闭以后触发（CSS 动画执行完成）</td>
  </tr>
  <tr>
    <td><code>close:modal:amui</code></td>
    <td><code>close</code> 方法被调用是立即触发</td>
  </tr>
  <tr>
    <td><code>closed:modal:amui</code></td>
    <td>Modal 窗口被关闭以后触发（CSS 动画执行完成）</td>
  </tr>
  </tbody>
</table>

