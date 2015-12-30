# Typography 文字排版
---

很多设计师喜欢用英文，因为中文排版真的不太容易搞。中文字符多、不同字符的笔画差异也很大，使用英文时高大上的设计稿，替换成中文以后，可能会大打折扣。

网页设计中，针对中文排版的研究很少，没有太多现成的结论供参考。Amaze UI 只能根据一些经验，在字体设置、字号上做一些更适合中文的设置。

## 字体

Amaze UI 主要使用非衬线字体（sans-serif），在 `<code>`、`<pre>` 等元素上则设置了等宽字体（monospace）。

### 字体定义

```css
body {
  font-family: "Segoe UI", "Lucida Grande", Helvetica, Arial, "Microsoft YaHei", FreeSans, Arimo, "Droid Sans","wenquanyi micro hei","Hiragino Sans GB", "Hiragino Sans GB W3", Arial, sans-serif;
}
```
- __Segoe UI__ - Windows UI 字体（从 Windows Vista 开始使用）；
- __Helvetica Neue__ 是 iOS7 及 OSX 10.10 UI 字体（在部分文字垂直居中的场景有一些小问题，暂时先使用 `Lucida Grande`）；
- 有些 Windows 用户安装了丽黑字体，丽黑字体在 Windows 上渲染出来很模糊，所以把微软雅黑放在苹果丽黑之前；
- __FreeSans__ - 包括 Ubuntu 之类的 Linux 分发版包含的字体。

### 字体辅助 Class

在 Amaze UI 的实践中，`font-family` 设置只在 Base 样式中出现了一两次。在具体项目中，我们也不建议到处设置 `font-family`，虽然 Amaze UI 提供了设置字体的辅助 class。

- `.am-sans-serif` __非衬线字体__，Amaze UI 主要使用的。
- `.am-serif` __衬线字体__，中文显示宋体，Amaze UI 中未使用。
- `.am-kai` __数字英文显示衬线字体，中文显示楷体__。和 `.am-serif` 的区别仅在于中文字体，Amaze UI 中把 `<blockquote>` 的字体设置成了 `.am-kai`。
- `.am-monospace` __等宽字体__，Amaze UI 源代码中使用。

下面为几种字体系列的演示：

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

严格来说，楷体属于手写体系列（cursive），不过英文的手写字过于花哨，所以在 `.am-kai` 中英文使用 `serif`。关于五种字体的定义，可以查看 [W3C 文档](http://www.w3.org/TR/CSS2/fonts.html#value-def-generic-family)。

### Webkit 反锯齿

另外，在 Webkit 浏览器下还设置了反锯齿平滑渲染，渲染出来要纤细一些，其他内核的浏览器上看着稍粗一些。

__2014.10.10 update:__ OSX 平台的 Firefox 从 v25 增加了 `-moz-osx-font-smoothing`，实现类似 Webkit 的字体渲染效果。

```css
body {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

如果你不喜欢，可以重置成浏览器的默认平滑字体。

```css
body {
  -webkit-font-smoothing: subpixel-antialiased;
  -moz-osx-font-smoothing: auto;
}
```

`````html
<h3>开启反锯齿 <code>-webkit-font-smoothing: antialiased;</code></h3>
<p>
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走的太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>
<br/>
<div style="-webkit-font-smoothing: subpixel-antialiased;-moz-osx-font-smoothing: auto">
  <h3>未开启反锯齿 <code>-webkit-font-smoothing: subpixel-antialiased;</code></h3>

  <p>
    The quick brown fox jumps over the lazy dog. <br/>
    千万不要因为走的太久，而忘记了我们为什么出发。 <br/>
    千萬不要因為走得太久，而忘記了我們為什麼出發。
  </p>
</div>
`````

```html
<h3>-webkit-font-smoothing: antialiased;</h3>
<p>
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走的太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>
<br/>
<div style="-webkit-font-smoothing: subpixel-antialiased; -moz-osx-font-smoothing: auto">
  <h3>-webkit-font-smoothing: subpixel-antialiased;</h3>
  <p>
    The quick brown fox jumps over the lazy dog. <br/>
    千万不要因为走的太久，而忘记了我们为什么出发。 <br/>
    千萬不要因為走得太久，而忘記了我們為什麼出發。
  </p>
</div>
```

参考链接：

- [-webkit-font-smoothing](http://ued.ctrip.com/blog/wp-content/webkitcss/prop/font-smoothing.html)
- [Better font-rendering on OSX](http://maximilianhoffmann.com/posts/better-font-rendering-on-osx)

### 字体相关链接

__主流系统中附带的字体__

- [List of Microsoft Windows fonts](http://en.wikipedia.org/wiki/List_of_Microsoft_Windows_fonts)
- [List of typefaces included with OS X](http://en.wikipedia.org/wiki/List_of_typefaces_included_with_OS_X)
- [Common fonts to all versions of Windows & Mac equivalents](http://www.ampsoft.net/webdesign-l/WindowsMacFonts.html)
- [OS X：Mavericks 附带的字体](http://support.apple.com/kb/HT5944?viewlocale=zh_CN&locale=zh_CN)
- [OS X：Mountain Lion 附带的字体](http://support.apple.com/kb/HT5379?viewlocale=zh_CN&locale=zh_CN)
- [iOS 7：字体列表](http://support.apple.com/kb/HT5878?viewlocale=zh_CN&locale=zh_CN)
- [iOS 6：字体列表](http://support.apple.com/kb/HT5484?viewlocale=zh_CN&locale=zh_CN)
- [Windows UI 设计文档](http://dev.windows.com/en-us/design)
- [Supported UI Fonts in Windows Phone][wpfts]
- [Typography | Android Developers](http://developer.android.com/design/style/typography.html)

[wpfts]:http://msdn.microsoft.com/library/windows/apps/hh202920(v=vs.105).aspx#BKMK_SupportedUIFontsinWindowsPhone

__中文排版相关开源项目__

- [「漢字標準格式」——印刷品般的漢字網頁排版框架](https://github.com/ethantw/Han)
- [Entry.css - 可配置的、更适合阅读的中文文章样式库](https://github.com/zmmbreeze/Entry.css)



## 元素基本样式

### 标题 `<h1> - <h6>`

`<h1> - <h6>` 保持粗体，设置了边距；`<h1>` 为正常字号的 `1.5` 倍；`<h2>` 为正常字号的 `1.25` 倍；其他保持正常字号。

`````html
<h1>h1 标题1</h1>
<h2>h2 标题2</h2>
<h3>h3 标题3</h3>
<h4>h4 标题4</h4>
<h5>h5 标题5</h5>
<h6>h6 标题6</h6>
`````
```html
<h1>h1 标题1</h1>
<h2>h2 标题2</h2>
<h3>h3 标题3</h3>
<h4>h4 标题4</h4>
<h5>h5 标题5</h5>
<h6>h6 标题6</h6>
```

### 段落 `<p>`

`````html
<p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
`````
```html
<p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
```

### 分隔线 `<hr>`

`````html
<hr/>
`````
```html
<hr/>
```

### 引用 `<blockquote>`

来源放到 `<small>` 元素里面。

`````html
<blockquote>
  <p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
  <small>马尔克斯 ——《百年孤独》</small>
</blockquote>
`````

```html
<blockquote>
  <p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
  <small>马尔克斯 ——《百年孤独》</small>
</blockquote>
```

### 代码块 `<pre>`

更多细节查看 [Code](/css/code)。

`````html
<pre style="background-color:#f8f8f8;color:#555">window.addEventListener('load', function() {
    FastClick.attach(document.body);
}, false);
</pre>
`````

```html
<pre>window.addEventListener('load', function() {
    FastClick.attach(document.body);
}, false);
</pre>
```


### 列表 ul/ol

__无序列表（`<ul>`）：__

`````html
<ul>
  <li>条目 #1</li>
  <li>条目 #2
    <ul>
      <li>条目 #1</li>
      <li>条目 #2
        <ul>
          <li>条目 #1</li>
          <li>条目 #2</li>
        </ul>
      </li>
    </ul>
  </li>
  <li>条目 #3</li>
  <li>条目 #4</li>
</ul>
`````
```html
<ul>
  <li>...</li>
  <li>...
    <ul>
      <li>
        <ul>
          <li>...</li>
        </ul>
      </li>
    </ul>
  </li>
</ul>
```

__有序列表（`<ol>`）：__

`````html
<ol>
  <li>条目 #1</li>
  <li>条目 #2
    <ol>
      <li>条目 #1</li>
      <li>条目 #2
        <ol>
          <li>条目 #1</li>
          <li>条目 #2</li>
        </ol>
      </li>
    </ol>
  </li>
  <li>条目 #3</li>
  <li>条目 #4</li>
</ol>
`````
```html
<ol>
  <li>...</li>
  <li>...
    <ol>
      <li>
        <ol>
          <li>...</li>
        </ol>
      </li>
    </ol>
  </li>
</ol>
```

### 定义列表

`````html
<dl>
  <dt>响应式网页设计</dt>
  <dd>自适应网页设计（英语：Responsive web design，通常缩写为RWD）是一种网页设计的技术做法，该设计可使网站在多种浏览设备（从桌面电脑显示器到移动电话或其他移动产品设备）上阅读和导航，同时减少缩放、平移和滚动。</dd>
  <dt>响应式网页设计（RWD）的元素</dt>
  <dd>采用 RWD 设计的网站 使用 CSS3 Media queries，即一种对 @media 规则的扩展，以及流式的基于比例的网格和自适应大小的图像以适应不同大小的设备。</dd>
</dl>
`````
```html
<dl>
  <dt>响应式网页设计</dt>
  <dd>自适应网页设计（英语：Responsive web design，通常缩写为RWD）是一种网页设计的技术做法，该设计可使网站在多种浏览设备（从桌面电脑显示器到移动电话或其他移动产品设备）上阅读和导航，同时减少缩放、平移和滚动。</dd>
  <dt>响应式网页设计（RWD）的元素</dt>
  <dd>采用 RWD 设计的网站 使用 CSS3 Media queries，即一种对 @media 规则的扩展，以及流式的基于比例的网格和自适应大小的图像以适应不同大小的设备。</dd>
</dl>
```

### 表格 `<table>`

这里只是 normalize 以后的样式，更多样式查看 [Table 组件](/css/table)。

`````html
<table>
  <caption>表格标题</caption>
  <thead>
    <tr>
      <th>表头 #1</th>
      <th>表头 #2</th>
      <th>表头 #3</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>数据 #1</td>
      <td>数据 #2</td>
      <td>数据 #3</td>
    </tr>
    <tr>
      <td>数据 #1</td>
      <td>数据 #2</td>
      <td>数据 #3</td>
    </tr>
  </tbody>
  <tfoot>
    <tr>
      <td>表页脚 #1</td>
      <td>表页脚 #2</td>
      <td>表页脚 #3</td>
    </tr>
  </tfoot>
</table>
`````
```html
<table>
  <caption>...</caption>
  <thead>
    <tr>
      <th>...</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>...</td>
    </tr>
  </tbody>
  <tfoot>
    <tr>
      <td>...</td>
    </tr>
  </tfoot>
</table>
```

### 图片 `<img>`

更多样式查看 [Image 组件](/css/image)。

`````html
<img class="am-img-responsive" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
`````

```html
<img class="am-img-responsive" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>

```

### 其他元素

<table class="am-table am-table-bordered am-table-hover am-table-striped">
  <thead>
  <tr>
    <th>元素</th>
    <th>基本样式</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>&lt;a&gt;</code></td>
    <td><a href="#">这是一个超链接</a></td>
  </tr>
  <tr>
    <td><code>&lt;em&gt;</code></td>
    <td><em>我在 em 元素里</em></td>
  </tr>
  <tr>
    <td><code>&lt;strong&gt;</code></td>
    <td><strong>strong 元素，壮吧</strong></td>
  </tr>
  <tr>
    <td><code>&lt;code&gt;</code></td>
    <td><code>行内代码</code></td>
  </tr>
  <tr>
    <td><code>&lt;del&gt;</code></td>
    <td><del>在我的胸口划一刀</del></td>
  </tr>
  <tr>
    <td><code>&lt;ins&gt;</code></td>
    <td><ins>强势插入的元素</ins></td>
  </tr>
  <tr>
    <td><code>&lt;mark&gt;</code></td>
    <td><mark>我被贴上 mark 标签了</mark></td>
  </tr>
  <tr>
    <td><code>&lt;q&gt;</code></td>
    <td><q>我在 q 元素里面 <q>q 元素里面的 q元素</q> q 元素</q></td>
  </tr>
  <tr>
    <td><code>&lt;abbr&gt;</code></td>
    <td>响应式设计缩写 <abbr title="Responsive web design">RWD</abbr></td>
  </tr>
  <tr>
    <td><code>&lt;dfn&gt;</code></td>
    <td>定义一个东西 <dfn title="Enjoy your music, photos and videos, anywhere anytime">DLNA</dfn></td>
  </tr>
  <tr>
    <td><code>&lt;small&gt;</code></td>
    <td><small>我比别人小一些</small></td>
  </tr>
  </tbody>
</table>
