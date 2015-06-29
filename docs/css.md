---
id: css
title: CSS
titleEn: CSS
permalink: css.html
next: css/normalize.html
---

# CSS
---

## CSS 概述

Amaze UI CSS 大致分为四部分。

<div class="am-g">
  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">基础（默认）样式</div>
      <div class="am-panel-bd">
        使用 normalize.css 统一浏览器差异， 以及一些基础的元素样式。
      </div>
    </div>
  </div>
  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">布局样式</div>
      <div class="am-panel-bd">
        包含用于布局的 Grid、AVG Grid，以及一些辅助 Class。
      </div>
    </div>
  </div>

  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">元素样式</div>
      <div class="am-panel-bd">
        对 <code>code</code>、<code>form</code>、<code>table</code> 等 HTML 元素定义更多的样式。
      </div>
    </div>
  </div>

  <div class="am-u-md-6">
    <div class="am-panel am-panel-default">
      <div class="am-panel-hd">页面组件</div>
      <div class="am-panel-bd">
        定义网页中常用的、多个元素组合在一起的组件样式，如分页、面包屑导航等。
      </div>
    </div>
  </div>
</div>

### 浏览器前缀

Amaze UI 2.0 开始移除了所有标准属性的浏览器前缀，构建时通过 [AutoPrefixer](https://github.com/postcss/autoprefixer) 自动添加。

当前的 AutoPrefixer 浏览器支持设置为：

```javascript
[
  'ie >= 8',
  'ie_mob >= 10',
  'ff >= 30',
  'chrome >= 34',
  'safari >= 7',
  'opera >= 23',
  'ios >= 7',
  'android >= 2.3',
  'bb >= 10'
]
```
## 响应式断点

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>尺寸</th>
    <th class="am-text-nowrap">class 简写</th>
    <th>断点区间</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
    <tr>
      <td>small</td>
      <td><code>sm</code></td>
      <td>0 - 640px</td>
      <td>处理绝大数手机的横竖屏模式（<a href="http://viewportsizes.com/?filter=Galaxy%20Note%20N7" target="_blank">Galaxy Note 2 的视口为 360 * 640</a>）</td>
    </tr>
    <tr>
      <td>medium</td>
      <td><code>md</code></td>
      <td>641px - 1024px</td>
      <td>平板的横竖屏模式</td>
    </tr>
    <tr>
      <td>large</td>
      <td><code>lg</code></td>
      <td>1025px + </td>
      <td>桌面设备</td>
    </tr>
  </tbody>
</table>

目前 Amaze UI 对大于 1025px 的屏幕并没有做划分，虽然现在大屏显示器越来越多，但是设计一个过宽的网页对用户来说并不友好，用户眼睛左右移动的区间太大，浏览起来比较累。当然，一些特殊类型（购物、视频等）的网站例外。

LESS 变量中定义了一些 Media Query 变量，使用 LESS 的用户可以引入该文件，直接使用这些变量。

```css
@screen:        ~"only screen";

@landscape:     ~"@{screen} and (orientation: landscape)";
@portrait:      ~"@{screen} and (orientation: portrait)";

@small-up:      ~"@{screen}";
@small-only:    ~"@{screen} and (max-width: @{breakpoint-small-max})";

@medium-up:     ~"@{screen} and (min-width:@{breakpoint-medium-min})";
@medium-only:   ~"@{screen} and (min-width:@{breakpoint-medium-min}) and (max-width:@{breakpoint-medium-max})";

@large-up:      ~"@{screen} and (min-width:@{breakpoint-large-min})";
@large-only:    ~"@{screen} and (min-width:@{breakpoint-large-min}) and (max-width:@{breakpoint-large-max})";
```


## 使用 HTML5

Amaze UI 在 HTML5 下开发，没有测试其他 DOCTYPE，使用之前确保你的 HTML 第一行是 `<!doctype html>`。

建议使用的 HTML head:

```html
<!doctype html>
<html class="no-js">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="description" content="">
  <meta name="keywords" content="">
  <meta name="viewport"
        content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
  <title>Amaze UI Examples</title>

  <!-- Set render engine for 360 browser -->
  <meta name="renderer" content="webkit">

  <!-- No Baidu Siteapp-->
  <meta http-equiv="Cache-Control" content="no-siteapp"/>

  <link rel="icon" type="image/png" href="{{assets}}i/favicon.png">

  <!-- Add to homescreen for Chrome on Android -->
  <meta name="mobile-web-app-capable" content="yes">
  <link rel="icon" sizes="192x192" href="{{assets}}i/app-icon72x72@2x.png">

  <!-- Add to homescreen for Safari on iOS -->
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black">
  <meta name="apple-mobile-web-app-title" content="Amaze UI"/>
  <link rel="apple-touch-icon-precomposed" href="assets/i/app-icon72x72@2x.png">

  <!-- Tile icon for Win8 (144x144 + tile color) -->
  <meta name="msapplication-TileImage" content="assets/i/app-icon72x72@2x.png">
  <meta name="msapplication-TileColor" content="#0e90d2">

  <!-- SEO: If your mobile URL is different from the desktop URL, add a canonical link to the desktop page https://developers.google.com/webmasters/smartphone-sites/feature-phones -->
  <!--
  <link rel="canonical" href="http://www.example.com/">
  -->

  <link rel="stylesheet" href="assets/css/amazeui.min.css">
  <link rel="stylesheet" href="assets/css/app.css">
</head>
<body>
...

<!--[if (gte IE 9)|!(IE)]><!-->
<script src="assets/js/jquery.min.js"></script>
<!--<![endif]-->
<!--[if lte IE 8 ]>
<script src="http://libs.baidu.com/jquery/1.11.3/jquery.min.js"></script>
<script src="http://cdn.staticfile.org/modernizr/2.8.3/modernizr.js"></script>
<script src="assets/js/amazeui.ie8polyfill.min.js"></script>
<![endif]-->
<script src="assets/js/amazeui.min.js"></script>
</body>
</html>
```

**参考链接**：

- [HTML5 Boilerplate](https://github.com/h5bp/html5-boilerplate)
- [Google Web Starter Kit](https://developers.google.com/web/starter-kit/)

## 移动优先

Amaze UI 以移动优先的理念开发，需要在 `meta` 中设置相关 `viewport` 属性。

`width=device-width, initial-scale=1` 是必须的，而且我们认为好的设计是不需要用户去操作窗口缩放的，所以加上了 `maximum-scale=1, user-scalable=no`。

```html
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
```

## 国情国情

不可否认，这是一个神奇的国度，一切合理与不合理都可以用国情来解释。前端开发也不例外。

### 渲染引擎

国内不少 `X 核` 浏览器，对于前端开发者，他们更多时候可能充当了 Troublemaker 的角色，不过也会有让人眼前一亮的一刻。

```html
<meta name="renderer" content="webkit">
```

这行代码可以指定网页使用 `webkit` 引擎渲染，当然，只对 [360 浏览器 6.5+ 有效](http://se.360.cn/v6/help/meta.html)。就这点而言，我希望所有所有的小白都去用 360 浏览器，那该有多好...

### 防（zhēn）狼（cāo）术（dài）

如果你的网站不想被剥去外衣、往赤裸的身体上贴广告，就加上[下面的代码](http://m.baidu.com/pub/help.php?pn=22&ssid=0&from=844b&bd_page_type=1)。

```html
<meta http-equiv="Cache-Control" content="no-siteapp" />
```

## Class 命名说明

### 关注分离

Amaze UI CSS class 命名遵循关注分离、松耦合的原则，同时注重易于理解、理解，在参考
[BEM 命名法](http://bem.info/method/definitions/) 的基础上，采用更优雅的书写方式。

下面的代码直观展示了 Amaze UI CSS class 命名规范。

```css
.am-post           {} /* Block */
.am-post-title     {} /* Element */
.am-post-meta      {} /* Element */
.am-post-sticky    {} /* Generic Modifier - status */
.am-post-active	   {} /* Generic Modifier - status */
.am-post-title-highlight {}  /* Element Modifier */
```

#### 关注分离示例

```html
<article class="am-post">
  <h2 class="am-post-title"></h2>
  <p class="am-post-meta"></p>
  <main class="am-post-content">
    ...
  </main>
</article>
```

上面的代码中，可以直接使用下面的样式控制元素：

```css
.am-post > h2 {
    ...
}

.am-post > p {
  ...
}
```

乍看是没什么问题，这两个选择符也不会影响到 `<main>` 里面的元素，但是如果更改了 HTML 标签，
就必须同时修改 CSS 选择符，无疑加大了维护的工作量。所以，给相应元素加上 class 是关注分离一个不错的选择。

```html
<article class="am-post">
  <h1 class="am-post-title"></h1>
  <div class="am-post-meta"></div>
  <main class="am-post-content">
    ...
  </main>
</article>
```

#### 关注分离反模式

```html
<ul class="am-nav">
  <li class="am-nav-item"></li>
  <li class="am-nav-item"></li>
  <li class="am-nav-item"></li>
</ul>
```

上面是一个导航代码片段，我们给 `<li>` 都加上了 `.am-nav-item` class，表面是遵循关注分离，其实是一种反模式，因为 `<ul>` 里面肯定是要放 `<li>` 的，在没有其它更复杂的元素在里面的情况下，给 `<li>` 加 class 显然是多余的。

所以， __关注分离并不是简单地给每个元素都加上 class__，还需结合实际情况区别对待。


相关阅读：

- [Decoupling HTML From CSS](http://coding.smashingmagazine.com/2012/04/20/decoupling-html-from-css/)

### “多类症”（Classitis）

当 HTML 源代码满眼望去都是 class 时，是不是很抓狂？

不过为了实现代码复用，减少重复冗余，难免要把代码拆分在不同的 class 下面。我们只能寻找一个平衡点，避免过细的拆分，减少不必要的 class。

### 选择符书写

虽然使用 LESS 编写样式可以很方便的嵌套，但是我们不建议过度嵌套选择符，有些嵌套是没有必要的。

同时，我们也不建议把多个选择器堆叠在一起。

```css
.ui.form .fields.error .field .ui.selection.dropdown .menu .item:hover {
    ...
}
```

看看上面来自 [Semantic UI](http://semantic-ui.com/) 的选择符，威武霸气吧，整行都是选择符，class 加 class，n 层嵌套，我只想呵呵...

__选择符嵌套在必要的情况下一般不超过三层；选择符叠加一般不多于两个__。

## 一点禁忌

我们崇尚自由，但并不是百无禁忌。

Amaze UI 中有两个表示状态的 class：

- `.am-active` - 激活
- `.am-disabled` - 禁用

__不要单独使用、直接在里面编写样式__！！！

```css
/* 可以嵌套用 */
.am-nav .am-active {
  ...
}

/* 可以堆叠用 */
.am-btn.am-active {
  ...
}

/* 绝不要单独用！！！ */
.am-active {
  color: red;
}

/* 当然，如果你想给自己找点乐，那就随便了 */
```

## 禁用响应式

不喜欢响应式？可以尝试禁用：

- 删除 `head` 里的视口设置 `meta` 标签；

```html
<!--<meta name="viewport"
    content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">-->
```

- 固定容器 `.am-container` 宽度（可以自己添加一个 class，不一定要使用内置的）：

```css
.am-container {
  width: 980px !important;
  max-width: none;
}
```

- 使用网格系统时，只添加 `.am-u-sm-*` class，移除其他断点的 class。

至此，布局层的响应式被禁用了（[参考示例](/examples/non-responsive.html)）。

不过，这仅仅是个开始，一些组件的样式细节可能还需要调整，只能陪你到这了……

## 关于命名空间

<div class="am-alert am-alert-danger">
  似乎有些人看着 <code>.am</code> 有些不顺眼，在这里专门做一下说明。
</div>

可能有人不知道命名空间是什么东西，和 [YUI](https://github.com/yui/yui3/blob/master/build/app-transitions-css/app-transitions-css.css) 中的 `yui`、[Pure](http://purecss.io/) 中的 `pure` 一样，Amaze UI 里的 `am` 就是命名空间。

命名空间使类名变得冗长，可为什么还要加呢？

### 目的：防御与无侵入！

更直白的话说，__我不犯人，也不让人犯我__。

CSS 多基于 Class 应用样式，我们不愿看到：

- __多个框架共存__时，按照我们的 CSS 编写的 HTML 结构应用了其他框架的样式；
- __从第三方抓取的 HTML__ 存在 class 相同的元素，意外地应用了 Amaze UI 的样式；
- __用户编写自己的代码时，意外的覆盖了框架中的样式；__
- __多人协作开发__时，发生命名冲突，样式相互影响；__
- __第三方服务__（如分享按钮、评论组件）会向页面中插入一些样式，可能会意外的应用到我们编写的结构；
- ……

Amaze UI 内部在用，平台上的开发者也在用，命名空间能够有效地减少这些问题。

仅限于此，与品牌宣传什么的扯不上关系。

### 何去何从？

<dl>
  <dt>命名空间会被删除吗？</dt>
  <dd>一般不会，不过也说不定，不过那也是很久很久以后的事。</dd>
  <dt>我真的不喜欢命名空间这个东西，怎么办？</dt>
  <dd>未来版本会尝试自定义命名空间的可能性。你也可以尝试自己在编译的过程中把命名空间去掉，前端编译工具那么多，何不试试？不然葱油拌？还是飘香拌?</dd>
</dl>

妹子只能说这么多了，再往下就只能说：你不懂我，我不怪你。
