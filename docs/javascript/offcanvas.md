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

侧滑插件。实际使用案例参见菜单组件[演示一](/widgets/menu/offcanvas1/0)、[演示二](/widgets/menu/offcanvas1/1)。

## 使用示例

该组件由触发器和侧滑元素两部分组成。触发器上添加 `data-am-offcanvas` 属性，侧滑元素使用固定的 HTML 结构。

### 默认样式

`````html
<!-- 链接触发器， href 属性为目标元素 ID -->
<a href="#doc-oc-demo1" data-am-offcanvas>点击显示侧边栏</a>

<!-- 侧边栏内容 -->
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
<!-- 链接触发器， href 属性为目标元素 ID -->
<a href="#doc-oc-demo1" data-am-offcanvas>点击显示侧边栏</a>

<!-- 侧边栏内容 -->
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

### Push 效果

设置 `effect: 'push'`。

`````html
<!-- 按钮触发器， 需要指定 target -->
<button class="am-btn am-btn-primary" data-am-offcanvas="{target: '#doc-oc-demo2', effect: 'push'}">点击显示侧边栏</button>

<!-- 侧边栏内容 -->
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
<!-- 按钮触发器， 需要指定 target -->
<button class="am-btn am-btn-primary" data-am-offcanvas="{target: '#doc-oc-demo2', effect: 'push'}">点击显示侧边栏</button>

<!-- 侧边栏内容 -->
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

### 右侧显示

边栏默认显示在左侧，在内容容器上添加 `.am-offcanvas-bar-flip` class 调整为右侧。

`````html
<!-- 按钮触发器， 需要指定 target -->
<button class="am-btn am-btn-success" data-am-offcanvas="{target: '#doc-oc-demo3'}">右侧显示侧边栏</button>

<!-- 侧边栏内容 -->
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
<!-- 按钮触发器， 需要指定 target -->
<button class="am-btn am-btn-success" data-am-offcanvas="{target: '#doc-oc-demo3'}">右侧显示边栏</button>

<!-- 侧边栏内容 -->
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

## 调用方式

首先，按照以下结构组织好侧栏内容：

```html
<div id="your-id" class="am-offcanvas">
  <div class="am-offcanvas-bar">
    <!-- 你的内容 -->
  </div>
</div>
```

### 通过 Data API

在要触发侧栏的元素上添加 `data-am-offcanvas` 属性：

- 如果是 `<a>` 元素，则把 `href` 的值设置为边栏的 ID：`href="#your-id"`；
- 如果是其他元素，则在 `data-am-offcanvas` 的值里面指定侧边栏 ID：

```html
<button data-am-offcanvas="{target: '#your-id'}"></button>
```

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>属性</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>{target: '#your-id'}</code></td>
    <td>指定侧边栏，如果是 <code>a</code> 元素则写在 <code>href</code> 里</td>
  </tr>
  <tr>
    <td><code>{effect: 'push'}</code></td>
    <td>边栏动画效果，可选的值为 <code>overlay | push</code> 默认为 <code>overlay</code></td>
  </tr>
  </tbody>
</table>

### 通过 JS

组织好 OffCanvas HTML 以后，可以通过 Javascript 调用。

```javascript
$('#my-offcanvas').offCanvas(options);
```

__注意：__ 这里 `#my-offcanvas` 直接指向侧边栏元素，而不是触发器。

#### 参数说明

- `options.effect`，值为 `overlay | push`，默认为 `overlay`。

```javascript
$('#my-offcanvas').offCanvas({effect: 'push'});
```

通过 `$().offCanvas(options)` 设置，参数同上。

#### 方法

- `$().offCanvas(options)` - 设置边栏参数并打开边栏
- `$().offCanvas('open')` - 打开边栏
- `$().offCanvas('close')` - 关闭边栏

#### 自定义事件

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>事件名称</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>open.offcanvas.amui</code></td>
    <td>打开边栏时立即触发</td>
  </tr>
  <tr>
    <td><code>close.offcanvas.amui</code></td>
    <td>关闭边栏时立即触发</td>
  </tr>
  </tbody>
</table>

#### JS 控制示例

下面的示例演示了使用 JS 打开/关闭侧边栏。侧边栏打开以后，关闭按钮无法点击到，可以在控制台输入以下代码模拟点击事件：

```js
$('[data-rel="close"]').click();
```

`````html
<!-- 侧边栏内容 -->
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

<button class="am-btn am-btn-primary doc-oc-js" data-rel="open">打开侧边栏</button>
<button class="am-btn am-btn-primary doc-oc-js" data-rel="close">关闭侧边栏</button>

<script>
  $(function() {
    var id = '#my-offcanvas';
    var $myOc = $(id);
    $('.doc-oc-js').on('click', function() {
      $myOc.offCanvas($(this).data('rel'));
    });

    $myOc.on('open.offcanvas.amui', function() {
      console.log(id + ' 打开了。');
    }).on('close.offcanvas.amui', function() {
      console.log(id + ' 关闭了。');
    });
  });
</script>
`````

```html
<!-- 侧边栏内容 -->
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

<button class="am-btn am-btn-primary doc-oc-js" data-rel="open">打开侧边栏</button>
<button class="am-btn am-btn-primary doc-oc-js" data-rel="close">关闭侧边栏</button>

<script>
  $(function() {
    var id = '#my-offcanvas';
    var $myOc = $(id);
    $('.doc-oc-js').on('click', function() {
      $myOc.offCanvas($(this).data('rel'));
    });

    $myOc.on('open.offcanvas.amui', function() {
      console.log(id + ' 打开了。');
    }).on('close.offcanvas.amui', function() {
      console.log(id + ' 关闭了。');
    });
  });
</script>
```
