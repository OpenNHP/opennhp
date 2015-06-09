# Utility
---

一些常用样式的 class，与 LESS minxins 的区别在于：mixins 在样式中调用，而 utility 直接在 HTML 中引用。比如要对一个元素清除浮动，在元素上添加 `am-cf` 这个 class 即可。

## 布局相关

### 容器

#### 基本容器

~~`.am-container`，盒模型为 `border-box`，水平居中对齐，清除浮动。~~

`.am-container` 放到了[网格](/css/grid?_ver=2.x)里面。

#### 水平滚动

`.am-scrollable-horizontal` 内容超出容器宽度时显示水平滚动条。

`````html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    <thead>
    <tr>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    </tbody>
  </table>
</div>
`````
```html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    ...
  </table>
</div>
```

#### 垂直滚动

`.am-scrollable-vertical` 内容超过设置的高度以后显示滚动条，默认设置的高度为 `240px`。

`````html
<div class="am-scrollable-vertical">
  <p>《你不懂我，我不怪你》<br>
    作者：莫言 </p>
    <p>每个人都有一个死角， 自己走不出来，别人也闯不进去。<br>
      我把最深沉的秘密放在那里。<br>
      你不懂我，我不怪你。 </p>
    <p>每个人都有一道伤口， 或深或浅，盖上布，以为不存在。<br>
      我把最殷红的鲜血涂在那里。<br>
      你不懂我，我不怪你。 </p>
    <p>每个人都有一场爱恋， 用心、用情、用力，感动也感伤。<br>
      我把最炙热的心情 藏在那里。<br>
      你不懂我，我不怪你。 </p>
    <p>每个人都有 一行眼泪， 喝下的冰冷的水，酝酿成的热泪。<br>
      我把最心酸的委屈汇在那里。<br>
      你不懂我，我不怪你。 </p>
    <p>每个人都有一段告白， 忐忑、不安，却饱含真心和勇气。<br>
      我把最抒情的语言用在那里。<br>
      你不懂我，我不怪你。 </p>
</div>
`````

### 浮动相关

- `.am-cf` - 清除浮动

- `.am-nbfc` - 使用 `overflow: hidden;` 创建新的 <abbr title="Block formatting context">BFC</abbr> 清除浮动（[参考](https://developer.mozilla.org/en-US/docs/Web/Guide/CSS/Block_formatting_context)）

```css
.am-cf {
  .clearfix();
}
```

- `.am-fl` - 左浮动
- `.am-fr` - 右浮动
- `.am-center` - 水平居中

```css
.am-center {
  display: block;
  margin-left: auto;
  margin-right: auto;
}
```

示例：

`````html
<div class="am-cf" style="padding: 10px; border: 1px dashed #ddd">
  <button class="am-btn am-btn-success am-fl">向左浮动</button>
  <button class="am-btn am-btn-success am-fr">向右浮动</button>
</div>
`````
```html
<div class="am-cf">
  <button class="am-btn am-btn-success am-fl">向左浮动</button>
  <button class="am-btn am-btn-success am-fr">向右浮动</button>
</div>
```

### 垂直对齐

垂直对齐的原理为把父容器内的 “幽灵”元素（使用伪元素）高度设置为 <code>100%</code>，再通过设置需要对齐的元素 `vertical-align` 属性实现（[参考](https://developer.mozilla.org/zh-CN/docs/Web/CSS/vertical-align)）。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Class</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-vertical-align</code></td>
    <td>将这个 class 添加到父容器，父容器需要指定高度。</td>
  </tr>
  <tr>
    <td><code>.am-vertical-align-middle</code></td>
    <td>需要垂直居中的元素</td>
  </tr>
  <tr>
    <td><code>.am-vertical-align-bottom</code></td>
    <td>添加到需要底部对齐的元素</td>
  </tr>
  </tbody>
</table>

#### 垂直居中对齐

`````html
<div class="am-vertical-align" style="height: 150px; border: 1px dashed #ddd;">
  <div class="am-vertical-align-middle">
    飘在半空中的 XX
  </div>
</div>
`````

```html
<div class="am-vertical-align" style="height: 150px;">
  <div class="am-vertical-align-middle">
    飘在半空中的 XX
  </div>
</div>
`````

#### 底部对齐

`````html
<div class="am-vertical-align" style="height: 150px; border: 1px dashed #ddd;">
  <div class="am-vertical-align-bottom">
    DOWN 到了谷底...降到零下几度 C
  </div>
</div>
`````

```html
<div class="am-vertical-align" style="height: 150px;">
  <div class="am-vertical-align-bottom">
    DOWN 到了谷底...降到零下几度 C
  </div>
</div>
`````

#### 参考链接

- [CSS 实现水平、垂直居中](http://coding.smashingmagazine.com/2013/08/09/absolute-horizontal-vertical-centering-css/)
- [Centering in the Unknown
](http://css-tricks.com/centering-in-the-unknown/)
- [Cube Layout.css](http://thx.alibaba-inc.com/cube/doc/layout/)

### 元素显示

#### 显示属性

- `.am-block` - `display` 设置为 `block`
- `.am-inline` - `display` 设置为 `inline`
- `.am-inline-block` - `display` 设置为 `inline-block`

#### 隐藏元素

添加 `.am-hide` class。

```css
.am-hide {
  display: none !important;
  visibility: hidden !important;
}
```

`````html
<button class="am-btn am-btn-danger am-hide">I'm hidden.....</button>
`````
```html
<!-- 隐藏了，Demo 里看不到按钮 -->
<button class="am-btn am-btn-danger am-hide">I'm hidden.....</button>
```

### 内外边距

#### 尺寸

- `xs` - 5px
- `sm` - 10px
- default - 16px
- `lg` - 24px
- `xl` - 32px

#### class 列表

不加尺寸为默认大小（16px），`{size}` 可以为 `0, xs, sm, lg, xl` 中之一。

- **v2.4:** 增加 `0` 值的内外边距辅助类。

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr>
      <th>Margin</th>
      <th>Padding</th>
    </tr>
  </thead>
  <tr>
    <td>
      <code>.am-margin</code> <br/>
      <code>.am-margin-{size}</code>
    </td>
    <td>
      <code>.am-padding</code> <br/>
      <code>.am-padding-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      水平方向外边距 <br/>
      <code>.am-margin-horizontal</code> <br/>
      <code>.am-margin-horizontal-{size}</code>
    </td>
    <td>
      水平方向内边距 <br/>
      <code>.am-padding-horizontal</code> <br/>
      <code>.am-padding-horizontal-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      垂直方向外边距 <br/>
      <code>.am-margin-vertical</code> <br/>
      <code>.am-margin-vertical-{size}</code>
    </td>
    <td>
      垂直方向内边距 <br/>
      <code>.am-padding-vertical</code> <br/>
      <code>.am-padding-vertical-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      左外边距 <br/>
      <code>.am-margin-left</code> <br/>
      <code>.am-margin-left-{size}</code>
    </td>
    <td>
      左内边距 <br/>
      <code>.am-padding-left</code> <br/>
      <code>.am-padding-left-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      右外边距 <br/>
      <code>.am-margin-right</code> <br/>
      <code>.am-margin-right-{size}</code>
    </td>
    <td>
      右内边距 <br/>
      <code>.am-padding-right</code> <br/>
      <code>.am-padding-right-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      上外边距 <br/>
      <code>.am-margin-top</code> <br/>
      <code>.am-margin-top-{size}</code>
    </td>
    <td>
      上内边距 <br/>
      <code>.am-padding-top</code> <br/>
      <code>.am-padding-top-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      下外边距 <br/>
      <code>.am-margin-bottom</code> <br/>
      <code>.am-margin-bottom-{size}</code>
    </td>
    <td>
      下内边距 <br/>
      <code>.am-padding-bottom</code> <br/>
      <code>.am-padding-bottom-{size}</code>
    </td>
  </tr>
</table>


## 文本工具

### 字体

- `.am-sans-serif` 非衬线，Amaze UI 主要使用的。
- `.am-serif` 衬线字体，中文为宋体，Amaze UI 中未使用。
- `.am-kai` 应为衬线字体，中文为楷体，Amaze UI `<blockquote>` 使用此。
- `.am-monospace` 等宽字体，Amaze UI 源代码中使用。

下面为几种字体系列演示：

`````html
<p>
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走得太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>

<p class="am-serif">
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走得太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>

<p class="am-kai">
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走得太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>

<p class="am-monospace">
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走得太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>
`````

```html
<p>...</p>

<p class="am-serif">...</p>

<p class="am-kai">...</p>

<p class="am-monospace">...</p>
```

### 文本颜色

`````html
<p>千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-primary">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-secondary">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-success">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-warning">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-danger">千万不要因为走得太久，而忘记了我们为什么出发。</p>
`````
```html
<p>...</p>
<p class="am-text-primary">...</p>
<p class="am-text-secondary">...</p>
<p class="am-text-success">...</p>
<p class="am-text-warning">...</p>
<p class="am-text-danger">...</p>
```

### 链接颜色减淡

超链接颜色默认为主色（蓝色），添加 `.am-link-muted` class 将链接颜色设置为灰色。

`````html
<a href="##" class="am-link-muted">超级链接</a>

<h3 class="am-link-muted"><a href="##">超级链接</a></h3>

<ul class="am-link-muted">
  <li><a href="##">超级链接</a></li>
  <li><a href="##">超级链接</a></li>
</ul>
`````
```html
<a href="" class="am-link-muted">...</a>

<h3 class="am-link-muted"><a href="">...</a></h3>

<ul class="am-link-muted">
  <li><a href="">...</a></li>
</ul>
`````

### 文字大小

- `.am-text-xs` - 12px
- `.am-text-sm` - 14px
- `.am-text-default` - 16px
- `.am-text-lg` - 18px
- `.am-text-xl` - 24px
- `.am-text-xxl` - 32px
- `.am-text-xxxl` - 42px


`````html
<p class="am-text-xs am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-sm am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-default am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-lg am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-xl am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-xxl am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
<p class="am-text-xxxl am-text-truncate">千万不要因为走得太久，而忘记了我们为什么出发。</p>
`````
```html
<p class="am-text-xs">...</p>
<p class="am-text-sm">...</p>
<p class="am-text-default">...</p>
<p class="am-text-lg">...</p>
<p class="am-text-xl">...</p>
<p class="am-text-xxl">...</p>
<p class="am-text-xxxl">...</p>
```


常用字号参考：

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th class="text-center">REMs</th>
    <th class="text-center">Pixels</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>6.8rem</td>
    <td>68px</td>
  </tr>
  <tr>
    <td>5rem</td>
    <td>50px</td>
  </tr>
  <tr>
    <td>3.8rem</td>
    <td>38px</td>
  </tr>
  <tr>
    <td>3.2rem</td>
    <td>32px</td>
  </tr>
  <tr>
    <td>2.8rem</td>
    <td>28px</td>
  </tr>
  <tr>
    <td>2.4rem</td>
    <td>24px</td>
  </tr>
  <tr>
    <td>2.2rem</td>
    <td>22px</td>
  </tr>
  <tr>
    <td>1.8rem</td>
    <td>18px</td>
  </tr>
  <tr>
    <td><strong>1.6rem (base)</strong></td>
    <td><strong>16px (base)</strong></td>
  </tr>
  <tr>
    <td>1.4rem</td>
    <td>14px</td>
  </tr>
  <tr>
    <td>1.2rem</td>
    <td>12px</td>
  </tr>
  <tr>
    <td>1rem</td>
    <td>10px</td>
  </tr>
  <tr>
    <td>0.8rem</td>
    <td>8px</td>
  </tr>
  <tr>
    <td>0.5rem</td>
    <td>5px</td>
  </tr>
  </tbody>
</table>


### 文本左右对齐

文字对齐辅助 class，可设置为响应式。

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr><th>左对齐</th>
    <th>右对齐</th>
    <th>居中</th>
    <th>自适应</th></tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-text-left</code></td>
    <td><code>.am-text-right</code></td>
    <td><code>.am-text-center</code></td>
    <td><code>.am-text-justify</code></td>
  </tr>
  <tr>
    <td><code>.am-sm-text-left</code></td>
    <td><code>.am-sm-text-right</code></td>
    <td><code>.am-sm-text-center</code></td>
    <td><code>.am-sm-text-justify</code></td>
  </tr>
  <tr>
    <td><code>.am-sm-only-text-left</code></td>
    <td><code>.am-sm-only-text-right</code></td>
    <td><code>.am-sm-only-text-center</code></td>
    <td><code>.am-sm-only-text-justify</code></td>
  </tr>
  <tr>
    <td><code>.am-md-text-left</code></td>
    <td><code>.am-md-text-right</code></td>
    <td><code>.am-md-text-center</code></td>
    <td><code>.am-md-text-justify</code></td>
  </tr>
  <tr>
    <td><code>.am-md-only-text-left</code></td>
    <td><code>.am-md-only-text-right</code></td>
    <td><code>.am-md-only-text-center</code></td>
    <td><code>.am-md-only-text-justify</code></td>
  </tr>
  <tr>
    <td><code>.am-lg-text-left</code></td>
    <td><code>.am-lg-text-right</code></td>
    <td><code>.am-lg-text-center</code></td>
    <td><code>.am-lg-text-justify</code></td>
  </tr>
 </tbody>
</table>

### 文本垂直对齐

- `.am-text-top` - 顶对齐
- `.am-text-middle` - 居中对齐
- `.am-text-bottom` - 底对齐

`````html
<div class="am-g">
  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-top">顶部对齐</span>
  </div>

  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-middle">居中对齐</span>
  </div>

  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-bottom">底部对齐</span>
  </div>
</div>
`````

### 文字换行及截断

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="min-width: 130px">Class</th>
    <th>描述</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-text-truncate</code></td>
    <td>禁止换行，超出容器部分截断（以 <code>...</code> 结束）</td>
  </tr>
  <tr>
    <td><code>.am-text-break</code></td>
    <td>超出文字容器宽度时强制换行，主要用于英文排版</td>
  </tr>
  <tr>
    <td><code>.am-text-nowrap</code></td>
    <td>禁止换行，不截断超出容器宽度部分</td>
  </tr>
  </tbody>
</table>

#### 单行文字截断

`.am-text-truncate`，元素 `display` 属性需为 `block` 或 `inline-block`。

```css
.am-text-truncate {
  word-wrap: normal; /* for IE */
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
}
```

`````html
<div class="am-text-truncate" style="width: 250px; padding: 10px; border: 1px dashed #ddd;">千万不要因为走得太久，而忘记了我们为什么出发</div>
`````

```html
<!-- 超出 200px 的文字将会被截断 -->
<div class="am-text-truncate" style="width: 250px; padding: 10px;">千万不要因为走得太久，而忘记了我们为什么出发</div>
```

参考链接：

- https://developer.mozilla.org/en-US/docs/Web/CSS/white-space
- https://developer.mozilla.org/en-US/docs/Web/CSS/text-overflow
- [IE8 & 9 white-space nowrap 失效](http://www.99css.com/archives/811)

#### 多行文字截断

在只针对 PC 端开发的年代，可以通过后端控制输出文字的长度来实现固定行数的效果。

但在响应式页面，这可能不再适用，只能输出足够多的文字，然后通过前端截取需要的行数。

Webkit 内核的浏览器可以通过 `-webkit-line-clamp` 私有属性实现多行文字截取。其他浏览器没有这个属性，我的做法通常是把容器的高度限定为 `行高 * 显示的行数`，超出的部分隐藏，勉强达到目的。

```css
.line-clamp {
  overflow : hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2; // 这里修改为要显示的行数
  -webkit-box-orient: vertical;
}
```

如果需要考虑其他内核的浏览器，可以使用 Amaze UI 封装的 Mixin:

```css
line-clamp(@lines, @line-height: 1.3em) {
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: @lines; // number of lines to show
  overflow: hidden;
  line-height: @line-height;
  max-height: @line-height * @lines;
}
```

当然，也有一些 JS 插件可以跨浏览器实现，但个人并不推荐在这种场合使用 JS。

__参考链接__

- [-webkit-line-clamp](http://dropshado.ws/post/1015351370/webkit-line-clamp)
- [Line Clampin’ - Truncating Multiple Line Text](http://css-tricks.com/line-clampin/)
- [CSS Ellipsis: How to Manage Multi-Line Ellipsis in Pure CSS](http://www.mobify.com/blog/multiline-ellipsis-in-pure-css/)
- [Clamp.js](https://github.com/josephschmitt/Clamp.js)
- [TextTailor.JS](https://github.com/jpntex/TextTailor.js)


### 图片替换

CSS Image Replacement 是一个历史悠久的技术，也随着前端技术的发展不断进化。

图片替换技术兼顾显示效果、可访性、SEO，推荐开发者在网站 Logo 、设计特殊的栏目标题等场合使用。

使用 `.am-text-ir` class 结合背景图片实现图片替换。

`````html
<header class="doc-ir-demo">
  <h1><a href="/" class="am-text-ir">Amaze UI</a></h1>
</header>
`````
```html
<header class="doc-ir-demo">
  <h1><a href="/" class="am-text-ir">Amaze UI</a></h1>
</header>
```

```css
.doc-ir-demo {
  background: #3bb4f2;
}

.doc-ir-demo h1 {
  margin: 0;
  padding: 10px;
}

.doc-ir-demo a {
  display: block;
  height: 29px;
  width: 125px;
  background: url(/i/landing/logo.png) no-repeat left center;
  -webkit-background-size: 125px 24px;
  background-size: 125px 24px;
}
```

- [Update CSS image replacement technique](https://github.com/h5bp/html5-boilerplate/commit/aa0396eae757)
- [CSS-Tricks | Search Results for 'image replace'](http://css-tricks.com/search-results/?q=image++replace)

### 图文混排辅助

使用 `float` 实现的类似 HTML `align` 属性的效果，父容器要清除浮动。与 `.am-fl`、`.am-fr` 相比，浮动的元素加了 `margin`。

- `.am-align-left`
- `.am-align-right`

`````html
<div class="am-cf">
  <p class="am-align-left">
    <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt="" width="240"/>
  </p>
  <p style="margin-top: 0">那时候刚好下着雨，柏油路面湿冷冷的，还闪烁着青、黄、红颜色的灯火。我们就在骑楼下躲雨，看绿色的邮筒孤独地站在街的对面。我白色风衣的大口袋里有一封要寄给南部的母亲的信。樱子说她可以撑伞过去帮我寄信。我默默点头。</p>

  <p>“谁叫我们只带来一把小伞哪。”她微笑着说，一面撑起伞，准备过马路帮我寄信。从她伞骨渗下来的小雨点，溅在我的眼镜玻璃上。</p>

  <p>随着一阵拔尖的煞车声，樱子的一生轻轻地飞了起来。缓缓地，飘落在湿冷的街面上，好像一只夜晚的蝴蝶。</p>

  <p>虽然是春天，好像已是秋深了。</p>

  <div class="am-align-right">
    <img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" alt="" width="240"/>
  </div>

  <p>
    她只是过马路去帮我寄信。这简单的行动，却要叫我终身难忘了。我缓缓睁开眼，茫然站在骑楼下，眼里裹着滚烫的泪水。世上所有的车子都停了下来，人潮涌向马路中央。没有人知道那躺在街面的，就是我的，蝴蝶。这时她只离我五公尺，竟是那么遥远。更大的雨点溅在我的眼镜上，溅到我的生命里来。</p>

  <p>为什么呢？只带一把雨伞？</p>

  <p>
    然而我又看到樱子穿着白色的风衣，撑着伞，静静地过马路了。她是要帮我寄信的。那，那是一封写给南部母亲的信。我茫然站在骑楼下，我又看到永远的樱子走到街心。其实雨下得并不大，却是一生一世中最大的一场雨。而那封信是这样写的，年轻的樱子知不知道呢？</p>

  <p>妈：我打算在下个月和樱子结婚。</p>
</div>
`````
```html
<div class="am-cf">
  <p class="am-align-left">
    <img src="..." alt=""/>
  </p>
  ...
  <p class="am-align-right">
    <img src="..." alt=""/>
  </p>
  ...
</div>
```


## 响应式辅助

### 视口大小

`.am-[show|hide]-[sm|md|lg][-up|-down|-only]`，调整浏览器窗口大小查看元素的显隐。

class 释义：

- `show` 显示，`hide` 隐藏，这应该不难理解；
- `sm`、`md`、`lg` 是划分屏幕大小区间的缩写，对应 small、medium、large，网格里面做了说明；
- `down` 指小于区间，`up` 指大于区间， `only` 指仅在这个区间。

按照上面的翻译一下下面的 class:

- `.am-show-sm-only`: 只在 `sm` 区间显示
- `.am-show-sm-up`: 大于 `sm` 区间时显示
- `.am-show-sm`: 在 `sm` 区间显示，如果没有设置 `md`、`lg` 区间的显隐，则元素在所有区间都显示
- `.am-show-md-down`: 小于 `md` 区间时显示

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>显示辅助 class</th>
    <th>隐藏辅助 class</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>
      <code>.am-show-sm-only</code> <br/>
      <code>.am-show-sm-up</code> <br/>
      <code>.am-show-sm</code> <br/>
      <code>.am-show-sm-down</code>
    </td>
    <td>
      <code>.am-hide-sm-only</code> <br/>
      <code>.am-hide-sm-up</code> <br/>
      <code>.am-hide-sm</code> <br/>
      <code>.am-hide-sm-down</code>
    </td>
  </tr>

  <tr>
    <td>
      <code>.am-show-md-only</code> <br/>
      <code>.am-show-md-up</code> <br/>
      <code>.am-show-md</code> <br/>
      <code>.am-show-md-down</code>
    </td>
    <td>
      <code>.am-hide-md-only</code> <br/>
      <code>.am-hide-md-up</code> <br/>
      <code>.am-hide-md</code> <br/>
      <code>.am-hide-md-down</code>
    </td>
  </tr>

  <tr>
    <td>
      <code>.am-show-lg-only</code> <br/>
      <code>.am-show-lg-up
      </code> <br/> <code>.am-show-lg
    </code> <br/> <code>.am-show-lg-down</code>
    </td>
    <td>
      <code>.am-hide-lg-only</code> <br/>
      <code>.am-hide-lg-up
      </code> <br/> <code>.am-hide-lg
    </code> <br/> <code>.am-hide-lg-down</code>
    </td>
  </tr>
  </tbody>
</table>


`````html
<ul>
  <li class="am-show-sm-only">仅 small 可见</li>
  <li class="am-show-md-up">medium + 可见</li>
  <li class="am-show-md-only">仅 medium 可见</li>
  <li class="am-show-lg-up">large 可见</li>
  <li class="am-show-lg-only">仅 large 可见</li>
</ul>
`````
```html
<ul>
  <li class="am-show-sm-only">仅 small 可见</li>
  <li class="am-show-md-up">medium + 可见</li>
  <li class="am-show-md-only">仅 medium 可见</li>
  <li class="am-show-lg-up">large 可见</li>
  <li class="am-show-lg-only">仅 large 可见</li>
</ul>
```

### 屏幕方向

- 横屏：`.am-show-landscape`, `.am-hide-landscape`
- 竖屏：`.am-show-portrait`, `.am-hide-portrait`

`````html
<ul>
  <li class="am-show-landscape">横屏可见...</li>
  <li class="am-show-portrait">竖屏可见...</li>
</ul>
`````
```html
<ul>
  <li class="am-show-landscape">横屏可见...</li>
  <li class="am-show-portrait">竖屏可见...</li>
</ul>
```
