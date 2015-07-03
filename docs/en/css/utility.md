# Utility
---

Defines some useful classes. Then difference between utility and LESS mixin is that mixin is called in style sheets, but utility is called in HTML. For example, if you want to remove the float style of an element, just add the `.am-cf` to this element.

## Layout

### Container

#### Basic Container

~~`.am-container`，盒模型为 `border-box`，水平居中对齐，清除浮动。~~

`.am-container` is in [Grid](/css/grid?_ver=2.x).

#### Horizontal Scrollable

`.am-scrollable-horizontal` shows horizontal scrollbar when the contents overflow.

`````html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    <thead>
    <tr>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
    </tr>
    <tr>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
    </tr>
    <tr>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
      <td>data</td>
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

#### Vertical Scrollable

`.am-scrollable-vertical` shows vertical scrollbar when the contents overflow. Default height is `240px`.

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

### Float

- `.am-cf` - Clear Float

- `.am-nbfc` - Use `overflow: hidden;` to create a new <abbr title="Block formatting context">BFC</abbr> to clear float. ([Reference](https://developer.mozilla.org/en-US/docs/Web/Guide/CSS/Block_formatting_context))

```css
.am-cf {
  .clearfix();
}
```

- `.am-fl` - Right float
- `.am-fr` - Left float
- `.am-center` - Center

```css
.am-center {
  display: block;
  margin-left: auto;
  margin-right: auto;
}
```

Example：

`````html
<div class="am-cf" style="padding: 10px; border: 1px dashed #ddd">
  <button class="am-btn am-btn-success am-fl">Left Float</button>
  <button class="am-btn am-btn-success am-fr">Right Float</button>
</div>
`````
```html
<div class="am-cf">
  <button class="am-btn am-btn-success am-fl">Left Float</button>
  <button class="am-btn am-btn-success am-fr">Right Float</button>
</div>
```

### Vertical Alignment

Vertical alignment is achieved by setting the height of "ghost" element (use pseudo element) to <code>100%</code>, and setting the `vertical-align` attribute of the element need to be aligned.（[Reference](https://developer.mozilla.org/zh-CN/docs/Web/CSS/vertical-align)）。

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Class</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-vertical-align</code></td>
    <td>Add this class to parent container. Height of parent container need to be specified.</td>
  </tr>
  <tr>
    <td><code>.am-vertical-align-middle</code></td>
    <td>Center Aligned Element</td>
  </tr>
  <tr>
    <td><code>.am-vertical-align-bottom</code></td>
    <td>Bottom Aligned Element</td>
  </tr>
  </tbody>
</table>

#### Vertical Center Align

`````html
<div class="am-vertical-align" style="height: 150px; border: 1px dashed #ddd;">
  <div class="am-vertical-align-middle">
    XX in the air
  </div>
</div>
`````

```html
<div class="am-vertical-align" style="height: 150px;">
  <div class="am-vertical-align-middle">
    XX in the air
  </div>
</div>
```

#### Vertical Bottom Align

`````html
<div class="am-vertical-align" style="height: 150px; border: 1px dashed #ddd;">
  <div class="am-vertical-align-bottom">
    DOWN to the bottom
  </div>
</div>
`````

```html
<div class="am-vertical-align" style="height: 150px;">
  <div class="am-vertical-align-bottom">
    DOWN to the bottom
  </div>
</div>
```

#### Reference

- [CSS 实现水平、垂直居中](http://coding.smashingmagazine.com/2013/08/09/absolute-horizontal-vertical-centering-css/)
- [Centering in the Unknown
](http://css-tricks.com/centering-in-the-unknown/)
- [Cube Layout.css](http://thx.alibaba-inc.com/cube/doc/layout/)

### Display

#### Attributes

- `.am-block` - Set `display` to `block`
- `.am-inline` - Set `display` to `inline`
- `.am-inline-block` - Set `display` to `inline-block`

#### Hide

Use `.am-hide` class。

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

### Padding and Margin

#### Size

- `xs` - 5px
- `sm` - 10px
- default - 16px
- `lg` - 24px
- `xl` - 32px

#### Class List

Classes without size have default size (16px). `{size}` could be one of `0, xs, sm, lg, xl`.

- **v2.4:** Add Padding and Margin with `0` value.

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
      Horizontal Margin <br/>
      <code>.am-margin-horizontal</code> <br/>
      <code>.am-margin-horizontal-{size}</code>
    </td>
    <td>
      Horizontal Padding <br/>
      <code>.am-padding-horizontal</code> <br/>
      <code>.am-padding-horizontal-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      Vertical Margin <br/>
      <code>.am-margin-vertical</code> <br/>
      <code>.am-margin-vertical-{size}</code>
    </td>
    <td>
      Vertical Padding <br/>
      <code>.am-padding-vertical</code> <br/>
      <code>.am-padding-vertical-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      Left Margin <br/>
      <code>.am-margin-left</code> <br/>
      <code>.am-margin-left-{size}</code>
    </td>
    <td>
      Left Padding <br/>
      <code>.am-padding-left</code> <br/>
      <code>.am-padding-left-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      Right Margin <br/>
      <code>.am-margin-right</code> <br/>
      <code>.am-margin-right-{size}</code>
    </td>
    <td>
      Right Padding <br/>
      <code>.am-padding-right</code> <br/>
      <code>.am-padding-right-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      Top Margin <br/>
      <code>.am-margin-top</code> <br/>
      <code>.am-margin-top-{size}</code>
    </td>
    <td>
      Top Padding <br/>
      <code>.am-padding-top</code> <br/>
      <code>.am-padding-top-{size}</code>
    </td>
  </tr>
  <tr>
    <td>
      Bottom Margin <br/>
      <code>.am-margin-bottom</code> <br/>
      <code>.am-margin-bottom-{size}</code>
    </td>
    <td>
      Bottom Padding <br/>
      <code>.am-padding-bottom</code> <br/>
      <code>.am-padding-bottom-{size}</code>
    </td>
  </tr>
</table>


## Text Tools

### Font

- `.am-sans-serif` __Sans serif font__ is the main font of Amaze UI.
- `.am-serif` __Serif font__ is not used in Amaze UI.
- `.am-kai` __Use serif for English text and numbers, and use Kai for Chinese characters__. The only difference between `.am-kai` and `.am-serif` is on the Chinese characters. Amaze UI use `.am-kai` in `<blockquote>`.
- `.am-monospace` __monospace font__ is used in Amaze UI source code. 

This example shows the difference among these fonts.

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

### Text Color

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

### Light Color Link

Default color of link is the primary color. Add the `.am-link-muted` class  to change the color to gray.

`````html
<a href="##" class="am-link-muted">Super Link</a>

<h3 class="am-link-muted"><a href="##">Super Link</a></h3>

<ul class="am-link-muted">
  <li><a href="##">Super Link</a></li>
  <li><a href="##">Super Link</a></li>
</ul>
`````
```html
<a href="" class="am-link-muted">...</a>

<h3 class="am-link-muted"><a href="">...</a></h3>

<ul class="am-link-muted">
  <li><a href="">...</a></li>
</ul>
```

### Text Size

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


Frequently-used font size:

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


### Alignment

Alignment classes can be responsive.

<table class="am-table am-table-bd am-table-striped">
  <thead>
    <tr><th>Left Align</th>
    <th>Right Align</th>
    <th>Center Align</th>
    <th>Justify</th></tr>
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

### Vertical Alignment

- `.am-text-top` - Top Align
- `.am-text-middle` - Center Align
- `.am-text-bottom` - Bottom Align

`````html
<div class="am-g">
  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-top">Top Aligned</span>
  </div>

  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-middle">Center Aligned</span>
  </div>

  <div class="am-u-md-4">
    <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView/1/w/120/h/120" alt="" width="48" height="48"/>
    <span class="am-text-bottom">Bottom Aligned</span>
  </div>
</div>
`````

### Wrap and Truncate

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th style="min-width: 130px">Class</th>
    <th>Description</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-text-truncate</code></td>
    <td>Wrap is disabled, and cut off the overflow contents (end with <code>...</code>).</td>
  </tr>
  <tr>
    <td><code>.am-text-break</code></td>
    <td>Auto wrap when the content overflow. Mostly used in English Typography</td>
  </tr>
  <tr>
    <td><code>.am-text-nowrap</code></td>
    <td>Wrap is diabled, but don't cut off the overflow contents</td>
  </tr>
  </tbody>
</table>

#### Single Line Truncation

`.am-text-truncate`. The `display` attribute of element should be `block` or `inline-block`.

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

Reference:

- https://developer.mozilla.org/en-US/docs/Web/CSS/white-space
- https://developer.mozilla.org/en-US/docs/Web/CSS/text-overflow
- [IE8 & 9 white-space nowrap 失效](http://www.99css.com/archives/811)

#### Multiple Lines Truncation

When websites are only designed for PC, we can control the number of lines by controlling the length of text in back-end.

But this don't works in responsive web design, so we need to control the length of text in front-end.

Webkit browsers can truncate multiple lines by using `-webkit-line-clamp`. However, this attribute don't works in other browsers. One solution is to set the height of container to `line-height * number of lines` and cut off the overflow part.

```css
.line-clamp {
  overflow : hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2; // Use your number of lines here
  -webkit-box-orient: vertical;
}
```

If you are considering about browsers with other core, you may use Mixin in Amaze UI.

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

Of course, some JS plugins can achieve this effort on all browsers, but we don't recommend using JS here.

__Reference:__

- [-webkit-line-clamp](http://dropshado.ws/post/1015351370/webkit-line-clamp)
- [Line Clampin’ - Truncating Multiple Line Text](http://css-tricks.com/line-clampin/)
- [CSS Ellipsis: How to Manage Multi-Line Ellipsis in Pure CSS](http://www.mobify.com/blog/multiline-ellipsis-in-pure-css/)
- [Clamp.js](https://github.com/josephschmitt/Clamp.js)
- [TextTailor.JS](https://github.com/jpntex/TextTailor.js)


### Image Replacement

CSS Image Replacement is a technic with long history, and also is developing with the development of front-end technology.

Image Replacement give consideration to display effect, accessibility and SEO, so we suggest developers to use image replacement for their logo and titles.

Use `.am-text-ir` class to achieve image replacement with background image.

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

### Text Wrap Utilities

Use `float` to achieve effects similar to HTML `align` attribute. Float of parent container should be cleared. Comparing to `.am-fl` and `.am-fr`, `margin` is added to float element.

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


## Responsive Utilities

### Viewport Size

`.am-[show|hide]-[sm|md|lg][-up|-down|-only]`. Adjust window size to check the effect.

Class explanation：

- `show` and `hide` are just like their name;
- `sm`、`md`、`lg` is the abbreviation of the interval of screen size. More details in Grid;
- `down` means smaller than interval, and `up` means larger than interval. `only` means only in this interval.

Now we can translate following classes:

- `.am-show-sm-only`: Show only in `sm` interval
- `.am-show-sm-up`: Show when larger than `sm` interval
- `.am-show-sm`: Show in `sm` interval, and inherit to `md` and `lg` interval.
- `.am-show-md-down`: Show when smaller than `md` interval

<table class="am-table am-table-bd am-table-striped">
  <thead>
  <tr>
    <th>Show utility class</th>
    <th>Hide utility class</th>
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

### Screen Orientation

- Landscape：`.am-show-landscape`, `.am-hide-landscape`
- Portrait：`.am-show-portrait`, `.am-hide-portrait`

`````html
<ul>
  <li class="am-show-landscape">Show in landscape...</li>
  <li class="am-show-portrait">Show in portrait...</li>
</ul>
`````
```html
<ul>
  <li class="am-show-landscape">Show in landscape...</li>
  <li class="am-show-portrait">Show in portrait...</li>
</ul>
```
