# AllMobilize HTML/CSS Style Guide

---

## HTML/CSS 通用规则

### 编码规则

#### 省略外链资源 URL 协议部分

省略外链资源（图片及其它媒体资源）URL 中的 `http` / `https` 协议，使 URL 成为相对地址，避免 [Mixed Content](https://developer.mozilla.org/en-US/docs/Security/MixedContent) 问题，减小文件字节数。

**其它协议（`ftp` 等）的 URL 不省略。**

```html
<!-- Recommended -->
<script src="//www.google.com/js/gweb/analytics/autotrack.js"></script>
```

```html
<!-- Not recommended -->
<script src="http://www.google.com/js/gweb/analytics/autotrack.js"></script>
```

```css
/* Recommended */
.example {
  background: url(//www.google.com/images/example);
}
```

```css
/* Not recommended */
.example {
  background: url(http://www.google.com/images/example);
}
```

**在同一网站中，推荐使用相对路径链接资源，并避免使用层级过深的文件夹，以提高文件查找效率、减小文件体积**。

### 代码格式

#### 缩进

CSS（包括 Less 等扩展语言）中缩进 **2 个空格**（因模板、Less 等嵌套问题，将代码缩进定为两个空格，避免代码过长）。

**不要**使用 `Tab` 或者 `Tab`、空格混搭（大多编辑器都支持把 `Tab` 定义为希望的空格数）。

```html
<ul>
  <li>Fantastic</li>
  <li>Great</li>
</ul>
```

```css
.example {
color: blue;
}
```

#### 一律使用小写字母

```html
<!-- Recommended -->
<img src="google.png" alt="Google">

<!-- Not recommended -->
<A HREF="/">Home</A>
```

```css
/* Recommended */
color: #e5e5e5;

/* Not recommended */
color: #E5E5E5;
```

#### 删除行尾空格

### 元数据规则

#### 文件编码

**使用不带 BOM 的 UTF-8 编码。**

- 在 HTML中指定编码 `<meta charset="utf-8">`；
- 无需使用 `@charset` 指定样式表的编码，它默认为 UTF-8 （参考 [@charset](https://developer.mozilla.org/en-US/docs/Web/CSS/@charset)）；
- 引用第三方 JS 文件时 `<script charset="">` 标签应指明外链文件编码，以避免不同编码文件混合使用时页面乱码。

```html
<meta charset="UTF-8">
```

#### 注释

**根据需要尽可能解释代码。**

用注释来解释代码：它包括什么，它的目的是什么，它能做什么，为什么使用这个解决方案，还是说只是因为个人偏好？

（本规则根据实际项目而定，主要取决于项目的复杂程度。）

#### 未完成条目

**用 `TODO` 标记待办事项和活动的条目。**

- 只用 TODO 来强调代办事项， 不要用其他的常见格式，例如 @@ 。
- 附加联系人（用户名或电子邮件），用括号括起来，例如 `TODO(contact)` 。
- 可在冒号之后附加活动条目说明等信息，例如 `TODO: 活动条目说明` 。

```html
<!-- TODO(john.doe): revisit centering -->
<center>Test</center>

<!-- TODO: remove optional tags -->
<ul>
  <li>Apples</li>
  <li>Oranges</li>
</ul>
```

```css
/* TODO: remove IE hacks */
.nav {
  _width: 1%;
}
```

## HTML 风格指南

### 编码规则

#### 文档类型（DTD）使用 HTML5

```
<!DOCTYPE html>
```

#### 语言属性

根据 HTML5 规范：

> 强烈建议为 `html` 元素指定 `lang` 属性，从而为文档设置正确的语言。这将有助于语音合成工具确定其所应该采用的发音，有助于翻译工具确定其翻译时所应遵守的规则等等。

更多关于 `lang` 属性的知识可以从[此规范](http://www.w3.org/html/wg/drafts/html/master/semantics.html#the-html-element)中了解。

这里列出了[语言代码表](http://reference.sitepoint.com/html/lang-codes)。

#### IE 兼容模式设置为 `Edge`

IE 支持通过 `<meta>` 标签来设置渲染前页面所应该采用的 IE 版本。除非有特殊需求，否则最好是设置为 `Edge`。

```html
<meta http-equiv="X-UA-Compatible" content="IE=Edge">
```

更多细节参考[此文](http://stackoverflow.com/questions/6771258/whats-the-difference-if-meta-http-equiv-x-ua-compatible-content-ie-edge-e)。

#### HTML 有效性

编写有效的 HTML 代码，能通过[代码校验工具](http://validator.w3.org/nu/)验证。使用符合 HTML5 规范的标签，不允许使用废弃的标签，如 `<font>`、`<center>`等。

- [HTML5 新增标签及废弃标签列表](http://www.w3schools.com/tags/default.asp)
- [HTML5 element list](https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/HTML5/HTML5_element_list)
- [HTML5 Differences from HTML4](http://www.w3.org/TR/html5-diff/)

```html
<!-- Recommended -->
<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Document</title>
</head>
<body>
  <article>This is only a test.</article>
</body>
</html>

<!-- Not recommended -->
<title>Test</title>
<article>This is only a test.\
```

#### 语义化

- 根据 HTML 各元素设计的用途使用他们。这事关可访性、重用以及代码效率，例如 `h1-h6`用于标题，`<p>` 用于段落，`<a>`用于链接。
- 使用 HTML5 语义化标签。

```html
<!-- Recommended -->
<a href="recommendations/">All recommendations</a>

<!-- Not recommended -->
<div onclick="goToRecommendations();">All recommendations</div>
```

#### 多媒体内容提供后备方案

- 图片添加 `alt` 属性；视频、音频添加标题和文案说明。
- 备选内容是保障可访性的重要途径。
- 无法立即在 CSS 里设置的纯装饰图片将 `alt` 属性设置为空：`alt=""`。

```html
<!-- Recommended -->
<img src="spreadsheet.png" alt="Spreadsheet screenshot." />

<!-- Not recommended -->
<img src="spreadsheet.png" />
```

#### 关注分离：结构、表现、行为分离。

严格保持结构 （标记），表现 （样式），和行为 （脚本）分离, 并尽量让三者之间的交互保持在最低限度。

- 确保文档和模板只包含 HTML 结构， 所有表现都放到样式表里，所有行为都放到脚本里；
- 尽量将脚本和样式合并到一个文件，减少外链。

关注分离对于可维护性至关重要，修改 HTML（结构）要比修改样式、脚本花费更多成本。

```html
<!-- Recommended -->
<!DOCTYPE html>
<title>My first CSS-only redesign</title>
<link rel="stylesheet" href="default.css">
<h1>My first CSS-only redesign</h1>
<p>I’ve read about this on a few sites but today I’m actually
doing it: separating concerns and avoiding anything in the HTML of
my website that is presentational.</p>
<p>It’s awesome!</p>

<!-- Not recommended -->
<!DOCTYPE html>
<title>HTML sucks</title>
<link rel="stylesheet" href="base.css" media="screen">
<link rel="stylesheet" href="grid.css" media="screen">
<link rel="stylesheet" href="print.css" media="print">
<h1 style="font-size: 1em;">HTML sucks</h1>
<p>I’ve read about this on a few sites but now I’m sure:
    <u>HTML is stupid!!1</u>
<center>I can’t believe there’s no way to control the styling of
my website without doing everything all over again!</center>
```

#### 不要用 HTML 实体

使用 UTF-8 作为文件编码时，不需要使用 HTML 实体引用，如`&mdash;`、 `&rdquo;`、`&#x263a;`。

在 HTML 文档中具有特殊含义的字符（如 `<` 和 `&`）例外，控制或不可见字符也例外（如不换行空格）。

```html
<!-- Recommended -->
The currency symbol for the Euro is “€”.

<!-- Not recommended -->
The currency symbol for the Euro is &ldquo;&eur;&rdquo;.
```

#### 保留可选标签，保持结构完整性

[HTML5 规范](http://www.whatwg.org/specs/web-apps/current-work/multipage/syntax.html#syntax-tag-omission)定义了一些标签可以省略，以保持代码简洁性。但我们要求保持结构的完整，不省略可省标签，闭合所有元素，以避免不同浏览器环境下不可预知的问题出现。

- `<br>`, `<hr>`, `<img>`, `<input>` 等自闭合（self-closing）标签都应该在结尾添加 `/`；
- 不要省略可选的结束标签（closing tag）（例如 `</li>` 或 `</body>`）。

```html
<!-- Recommended -->
<!DOCTYPE html>
<html>
<head>
  <title>Spending money, spending bytes</title>
</head>
<body>
  <p>Sic.</p>
  <img src="xxx.jpg" alt="xxx" />
</body>
</html>

<!-- Not Recommended -->
<!DOCTYPE html>
<title>Saving money, saving bytes</title>
<p>Qed.
<img src="xxx" >
```

#### 省略样式表和脚本的 `type` 属性

使用 CSS 的样式表、使用 JavaScript 的脚本都不需要添加 `type` 属性。HTML5 默认按照 `text/css` 及 `text/javascript` 解析，兼容较老的浏览器。

```html
<!-- Recommended -->
<link rel="stylesheet" href="//www.google.com/css/maia.css">

<!-- Not recommended -->
<link rel="stylesheet" href="//www.google.com/css/maia.css"
type="text/css">
```

```html
<!-- Recommended -->
<script src="//www.google.com/js/gweb/analytics/autotrack.js"></script>

<!-- Not recommended -->
<script src="//www.google.com/js/gweb/analytics/autotrack.js"
        type="text/javascript"></script>
```

- [The link element](http://www.w3.org/TR/2011/WD-html5-20110525/semantics.html#the-link-element)
- [The style element](http://www.w3.org/TR/2011/WD-html5-20110525/semantics.html#the-style-element)
- [The script element](http://www.w3.org/TR/2011/WD-html5-20110525/scripting-1.html#the-script-element)

#### HTML 属性顺序

HTML 属性应当按照以下给出的顺序依次排列，确保代码的易读性。

- `class`
- `id`, `name`
- `data-*`
- `src`, `for`, `type`, `href`
- `title`, `alt`
- `aria-*`, `role`

Class 用于标识高度可复用组件，因此应该排在首位。id 用于标识具体组件，排在第二位。

#### 布尔值属性

HTML5 规范中 `disabled`、`checked`、`selected` 等属性不用设置值（[via](https://html.spec.whatwg.org/multipage/infrastructure.html#boolean-attributes)）。

```html
<input type="text" disabled>

<input type="checkbox" value="1" checked>

<select>
  <option value="1" selected>1</option>
</select>
```

如果非要赋值，不要使用 `true`、`false`，值必须是空字符串或属性的规范名称，且不要在末尾添加空格。

```html
<input type='checkbox' checked name=cheese disabled="disabled">

/* or */

<input type='checkbox' checked name=cheese disabled="">
```

#### JavaScript 生成的标签

通过 JavaScript 生成的标签让内容变得不易查找、编辑，并且降低性能。能避免时尽量避免。

#### 其他细节

- 使用 `<ul>`、`<ol>`、`<dl>` 组织列表，不要使用一堆 `<div>` 或者 `<p>`；
- 每个包含附加文字的表单输入框都应该利用 `<label>` 标签，特别是 `radio`、`checkbox`元素；
- 使用 `<label>` 标签包裹 `radio` / `checkbox`，不需要设置 `for` 属性；
- 避免写关闭标签注释，如 `<!-- /.element -->`，大多编辑器都支持开闭标签高亮；
- 不要手动设置 `tabindex`，浏览器会自动设置顺序。

### HTML 代码格式

#### 基本格式

每个块级、列表、表格元素单独占一行，每个子元素都相对父元素缩进。

```html
<blockquote>
  <p><em>Space</em>, the final frontier.</p>
</blockquote>

<ul>
  <li>Moe</li>
  <li>Larry</li>
  <li>Curly</li>
</ul>

<table>
  <thead>
  <tr>
    <th scope="col">Income</th>
    <th scope="col">Taxes</th>
  </thead>
  <tbody>
  <tr>
    <td>$ 5.00</td>
    <td>$ 4.50</td>
  </tbody>
</table>
```

**纯文本在 HTML 标签结束之前不要换行**。

#### 表格

恰当地使用 `<thead>`, `<tfoot>`, `<tbody>`, `<th>` 标签（注意：出于速度考虑，`<tfoot>` 应放在 `  <tbody>`，以便在所有表格数据渲染完成之前显示 `<tfoot>`）。

```html
<table summary="This is a chart of invoices for 2011.">
  <thead>
  <tr>
    <th scope="col">Table header 1</th>
    <th scope="col">Table header 2</th>
  </tr>
  </thead>
  <tfoot>
  <tr>
    <td>Table footer 1</td>
    <td>Table footer 2</td>
  </tr>
  </tfoot>
  <tbody>
  <tr>
    <td>Table data 1</td>
    <td>Table data 2</td>
  </tr>
  </tbody>
</table>
```

#### HTML 属性值使用双引号

不要省略属性值的引号，应始终添加双引号。

```html
<!-- Recommended -->
<a class="maia-button maia-button-secondary">Sign in</a>

<!-- Not recommended -->
<a class='maia-button maia-button-secondary'>Sign in</a>
```

## CSS 风格指南

### 编码规则

#### CSS 代码有效性。

**尽量标准定义的、有效的 CSS 代码**，最终的结果应该能通过 CSS 校验器校验（具体项目中视目标平台而定，不盲目追求通过代码校验）。

#### 不要使用 `@import`

与 `<link>` 相比，`@import` 要慢很多，不光增加额外的请求数，还会导致不可预料的问题。

替代办法：

- 使用多个 `<link>` 元素；
- 通过 Sass 或 Less 类似的 CSS 预处理器将多个 CSS 文件编译为一个文件；
- 其他 CSS 文件合并工具。

参考 [don’t use @import](http://www.stevesouders.com/blog/2009/04/09/dont-use-import/)。

#### ID、Class 命名

**ID、Class 使用语义化、通用的命名方式。**

- 只允许使用的小写字母、连字符、数字；
- 应该从 ID 和 Class 的名字上就能看出这元素是干嘛用的（角色、功能、状态），而不是表象（颜色、位置等）或模糊不清的命名；
- 表明、反映元素作用的名字易于理解，且后期修改的可能性小；
- 对与同级元素相比没有特殊的意义的元素使用通用的命名；
- 使用功能性或通用的名字可以减少不必要的文件修改。

<!--
- 使用 .js-* class 来标识行为关联的钩子（与样式相对），并且不要将这些 class 包含到 CSS 文件中。-->

```css
/* Recommended: specific */
#gallery {}
#login {}
.video {}

/* Recommended: generic */
.aux {}
.alt {}


/* Not recommended: meaningless */
#yee-1901 {}

/* Not recommended: presentational */
.button-green {}
.red {}
.left {}
```

#### ID、Class 命名风格

**ID 和 Class 的名称应尽量简短**，但应注意保留可读、可辨识性。

```css
/* Recommended */
#nav {} /* 简洁、可辨识的简写 */
.author {} /* 本身比较短，可辨识 */

/* Not recommended */
#navigation {} /* 太长，可以精简 */
.atr {} /* 不可读，无法辨识 */
```

#### 避免元素选择器和 ID、Class 叠加使用

出于[性能考量](http://www.stevesouders.com/blog/2009/06/18/simplifying-css-selectors/)，在没有必要的情况下避免元素选择器叠加 ID、Class 使用。

元素选择器和 ID、Class 混合使用也违反 **关注分离** 原则。如果HTML标签修改了，就要再去修改 CSS 代码，不利于后期维护。

```css
/* Recommended */
#example {}
.error {}

/* Not recommended */
ul#example {}
div.error {}
```

#### 使用属性简写

使用简写可以提高代码执行效率和易读性。

```css
/* Recommended */
.el {
  border-top: 0;
  font: 100%/1.6 palatino, georgia, serif;
  padding: 0 1em 2em;
}

/* Not recommended */
.el {
  border-top-style: none;
  font-family: palatino, georgia, serif;
  font-size: 100%;
  line-height: 1.6;
  padding-bottom: 2em;
  padding-left: 1em;
  padding-right: 1em;
  padding-top: 0;
}
```
大部分情况下，不需要为简写形式的属性声明指定所有值。例如，`h1-h6` 元素只需要设置上、下边距（margin）的值，因此，一般只需覆盖这两个值就可以。
过度使用简写形式会对属性值带来不必要的覆盖从而引起意外的副作用。

常见的滥用简写属性声明的情况如下：

- `padding`
- `margin`
- `font`
- `background`
- `border`
- `border-radius`

```css
/* Recommended */
.element {
  margin-bottom: 10px;
  background-color: red;
  border-top-left-radius: 3px;
  border-top-right-radius: 3px;
}

/* Not recommended */
.element {
  margin: 0 0 10px; /* 只需要设置下边距而已 */
  background: red; /* 只需要设置背景颜色而已 */
  border-radius: 3px 3px 0 0;
}
```

**参考链接**：

- [Shorthand properties](https://developer.mozilla.org/en-US/docs/Web/CSS/Shorthand_properties)

#### 属性值为 `0` 时省略单位

```css
margin: 0;
padding: 0;
```

#### 省略小数点前面的 `0`

```css
font-size: .8rem;
```

#### 使用 `rem` 作为字号、长度单位

使用 `px `对可访性会造成一定的问题，`em` 则随着上下文不断变化，计算较为繁杂。

- 推荐使用 `rem`：[Font sizing with rem](http://snook.ca/archives/html_and_css/font-size-with-rem)
- 需要 `1px` 级别精准定位的，仍然使用 `px`；
- 需要根据字号变化的（如 `padding`、`margin` 等）场景使用可以使用 `em`，较少不必要的代码。

#### `line-height` 不加单位

需要精确控制的场景除外。

#### 使用十六进制颜色编码（`rgba()`除外）

#### 简写可简写的十六进制颜色值

```css
/* Recommended */
color: #ebc;

/* Not recommended */
color: #eebbcc;
```

#### 根据项目在 ID、Class 前面加上特定前缀（命名空间）

命名空间可以防止命名冲突，方便维护（查找和替换）。

```css
.adw-help {} /* AdWords */
#maia-note {} /* Maia */
```

#### 使用连字符 `-` 作为 ID、Class 名称界定符

CSS 中不要驼峰命名法和下划线。

```css
/* Recommended */
#video-id {}
.ads-sample {}

/* Not recommended */
.demoimage {} /* 没有分割 `demo` 和 `image` */
.error_status {} /* 使用下划线链接 */
.errorStatus {} /* 使用驼峰式命名 */

```

元素在页面中仅仅出现一次的，应该使用 ID，否则使用 Class。

#### 尽量避免 Hacks

面向现代浏览器编写样式，针对过时浏览器的 Hack 可以放在单独的样式表中并使用条件注释引入。

移动开发针对 IE10+ 及其他现代浏览器，移除针对老版本 IE 的 Hack。

#### Less / SCSS 编写

- 使用 `;` 分割 Mixin 参数；
- 避免非必要的嵌套。


### CSS 代码格式

#### 属性声明顺序

推荐的样式编写顺序：

1. Positioning
2. Box model
3. Typographic
4. Visual

由于定位（positioning）可以从正常的文档流中移除元素，并且还能覆盖盒模型（box model）相关的样式，因此排在首位。盒模型决定了组件的尺寸和位置，因此排在第二位。

其他属性只是影响组件的内部（`inside`）或者是不影响前两组属性，因此排在后面。

```css
.declaration-order {
  /* Positioning */
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 100;

  /* Box-model */
  display: block;
  float: right;
  width: 100px;
  height: 100px;

  /* Typography */
  font: normal 13px "Helvetica Neue", sans-serif;
  line-height: 1.5;
  color: #333;
  text-align: center;

  /* Visual */
  background-color: #f5f5f5;
  border: 1px solid #e5e5e5;
  border-radius: 3px;

  /* Misc */
  opacity: 1;
}
```

<!--
1. 显示属性 `display/list-style/position/float/clear`
2. 自身属性（盒模型）`width/height/margin/padding/border`
3. 背景 `background`
4. 行高 `line-height`
5. 文本属性 `color/font/text-decoration/text-align/text-indent/vertical-align/white-space/content`
6. 其他 `cursor/z-index`
7. CSS3 属性 `transform/transition/animation/box-shadow/border-radius`-->

**链接的样式请严格按照如下顺序添加：**

`a:link -> a:visited -> a:hover -> a:active（LoVeHAte）`

**参考：**

- [Mozilla官方推荐CSS书写顺序](http://www.mozilla.org/css/base/content.css) 布局 - 具体元素样式
- [Bootstrap CSS Comb 顺序](https://github.com/twbs/bootstrap/blob/master/less/.csscomb.json)
- [RECESS - Twitter's CSS Hinter](http://twitter.github.io/recess/)


#### 代码块缩进

缩进花括号（`{}`）之间的[代码块](http://www.w3.org/TR/CSS21/syndata.html#block)。

```css
@media screen {

  html {
   background: #fff;
   color: #444;
  }

}
```

#### 声明的最后一行仍然添加分号

代码压缩由部署工具完成，编写源代码时应该保持每一行代码的完整性。

```css
/* Recommended */
.test {
  display: block;
  height: 100px;
}

/* Not recommended */
.test {
  display: block;
  height: 100px /* 这一行没有分号 */
}
```

#### 空格使用

- 属性和值之间（`:`后面）用一个空格分隔。这样可以提高代码的可读性。
- 对于以逗号分隔的属性值，每个逗号后面都空一个格（如 `box-shadow`）。
- 不要在 `rgb()`、`rgba()`、`hsl()`、`hsla()` 或 `rect()` 值的内部的逗号后面添加空格。这样利于从多个属性值（既加逗号也加空格）中区分多个颜色值（只加逗号，不加空格）。

```css
/* Recommended */
.selector,
.selector-secondary,
.selector[type="text"] {
  padding: 15px;
  margin-bottom: 15px;
  background-color: rgba(0,0,0,.5);
  box-shadow: 0 1px 2px #ccc, inset 0 1px 0 #fff;
}

/* Not recommended */

.selector{ /* 选择符和左花括号之间没有添加空格*/
  padding:15px; /* 冒号和值之间没有空格 */
  margin:0px 0px 15px;
  background-color:rgba(0, 0, 0, 0.5); /* 颜色值不加空格 */
  box-shadow:0px 1px 2px #CCC,inset 0 1px 0 #FFFFFF; /* 逗号分隔的属性值应该添加空格 */
}
```

#### 带前缀的属性

当使用特定厂商的带有前缀的属性时，通过缩进的方式，让每个属性的值在垂直方向对齐，这样便于多行编辑。

- 在 Textmate 中，使用 `Text` → `Edit Each Line in Selection (⌃⌘A)`；
- 在 Sublime Text 2 中，使用 `Selection` → `Add Previous Line (⌃⇧↑)` 和 `Selection` → `Add Next Line (⌃⇧↓)`。

```css
/* Prefixed properties */
.selector {
  -webkit-box-shadow: 0 1px 2px rgba(0,0,0,.15);
          box-shadow: 0 1px 2px rgba(0,0,0,.15);
}
```

#### 声明块分隔

最后一个选择器和起始花括号在一行，并用一个空格分隔。

```css
/* Recommended */
#video {
 margin-top: 1em;
}

/* Not recommended: missing space */
#video{ /* 选择器和花括号中间没有空格 */
  margin-top: 1em;
}

/* Not recommended: unnecessary line break */
#video
{ /* 最后一个选择器和花括号之间换行了 */
  margin-top: 1em;
}
```

#### 选择符、声明单独成行

```css
/* Recommended */
h1,
h2,
h3 {
  font-weight: normal;
  line-height: 1.2;
}

a:focus,
a:active {
  position: relative;
  top: 1px;
}

/* Not recommended */
a:focus, a:active {
  position: relative; top: 1px;
}
```

#### 规则之间使用一空行分隔

```css
html {
  background: #fff;
}

body {
  margin: auto;
  width: 50%;
}
```

#### 闭合花括号单独成行

```css
/* Recommended */
h2 {
  font-size: 1.8rem;
}

/* Not recommended */
h2 {
  font-size: 1.8rem;}
```

#### 引号使用

`url()` 添加双引号。属性选择符、属性值使用**双引号**。

```css
/* Recommended */
@import url("//www.google.com/css/maia.css");

html {
  font-family: "open sans", arial, sans-serif;
}

.selector[type="text"] {

}

/* Not recommended */
@import url(//www.google.com/css/maia.css);

html {
  font-family: 'open sans', arial, sans-serif;
}
```

**参考链接**

- [Is quoting the value of url() really necessary?](http://stackoverflow.com/questions/2168855/is-quoting-the-value-of-url-really-necessary)

#### 字体名称

字体名称请映射成对应的英文名，例如：黑体(SimHei) 、宋体(SimSun) 、微软雅黑(Microsoft Yahei)。

**如果字体名称中有空格，则必须加引号**。


### CSS 元数据规则

#### 按组编写注释

组与组之间空一行，注释与对应组之间不空行（视具体情况约定）。

```css
/* Header */
#adw-header {
}


/* Footer */
#adw-footer {
}


/* Gallery */
.adw-gallery {
}
```

## 参考 EditorConfig

```
# editorconfig.org

root = true

[*]
charset = utf-8
end_of_line = lf
indent_style = space
indent_size = 2
trim_trailing_whitespace = true
insert_final_newline = true
```

## 参考链接

- http://google-styleguide.googlecode.com/svn/trunk/htmlcssguide.xml
- http://css-tricks.com/css-style-guides/
- http://24ways.org/2011/front-end-style-guides/
- [Mozilla官方推荐CSS书写顺序](http://www.mozilla.org/css/base/content.css)
- [Use efficient CSS selectors](https://developers.google.com/speed/docs/best-practices/rendering#UseEfficientCSSSelectors)
- [Writing efficient CSS](https://developer.mozilla.org/en-US/docs/Web/Guide/CSS/Writing_efficient_CSS)
- [GitHub CSS Style Guide](https://github.com/styleguide/css)
- [GitHub Markup and templates Style Guide](https://github.com/styleguide/templates)
- [Code Guide by @mdo](http://mdo.github.io/code-guide/)


*Revision: 2014.10.29*
