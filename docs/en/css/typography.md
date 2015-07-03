# Typography
---

Many Designers like to use English rather than Chinese. One the reason is that Chinese typography is much harder to handle than English. There are so many different characters with different strokes in Chinese. 

There are few experience and conclusions about Chinese typography in web design for us to consult. Therefore, We did some setting on font for Chinese base on our own experience in Amaze UI.

## Font

Amaze UI mainly use sans-serif fonts, and use monospace font for `<code>`, `<pre>` and some other elements.

### Font CSS

```css
body {
  font-family: "Segoe UI", "Lucida Grande", Helvetica, Arial, "Microsoft YaHei", FreeSans, Arimo, "Droid Sans","wenquanyi micro hei","Hiragino Sans GB", "Hiragino Sans GB W3", Arial, sans-serif;
}
```
- __Segoe UI__ - Windows UI Font( Used since Windows Vista);
- __Helvetica Neue__ Font in iOS7 and OSX 10.10 UI ( Has some problems when using center align, so temporarily use `Lucida Grande` instead）;
- Some Windows users installed `Hiragino Sans GB W3`, which has some rendering problem on windows. So we put `Microsoft YaHei` in front of `Hiragino Sans GB W3`;
- __FreeSans__ - Font in Linux distributions like Ubuntu.

### Font Utility Class

In the implementation of Amaze UI, `font-family` only appear twice in Base style. We also don't recommend you to use too much `font-family` in your project, but you may use utility classes in Amaze UI instead.

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

Technically, Kai is a cursive font, but we use `serif` for English in `.am-kai` because cursive fonts in English is too fancy to distinguish. The definition of 5 fonts can be found in [W3C Docs](http://www.w3.org/TR/CSS2/fonts.html#value-def-generic-family)。

### Webkit Anti-aliasing

What's more, Webkit browsers use anti-aliasing when rendering, so Characters can be slightly thiner in Webkit browsers than other browsers.

__2014.10.10 update:__`-moz-osx-font-smoothing` is added in Firefox for OSX since v25 to achieve the similar effect as Webkit.

```css
body {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

If you don't like it, please reset to the default smooth font.

```css
body {
  -webkit-font-smoothing: subpixel-antialiased;
  -moz-osx-font-smoothing: auto;
}
```

`````html
<h3>Anti-aliasing on <code>-webkit-font-smoothing: antialiased;</code></h3>
<p>
  The quick brown fox jumps over the lazy dog. <br/>
  千万不要因为走的太久，而忘记了我们为什么出发。 <br/>
  千萬不要因為走得太久，而忘記了我們為什麼出發。
</p>
<br/>
<div style="-webkit-font-smoothing: subpixel-antialiased;-moz-osx-font-smoothing: auto">
  <h3>Anti-aliasing off <code>-webkit-font-smoothing: subpixel-antialiased;</code></h3>

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

Reference：

- [-webkit-font-smoothing](http://ued.ctrip.com/blog/wp-content/webkitcss/prop/font-smoothing.html)
- [Better font-rendering on OSX](http://maximilianhoffmann.com/posts/better-font-rendering-on-osx)

### Font Links

__Fonts in operating systems__

- [List of Microsoft Windows fonts](http://en.wikipedia.org/wiki/List_of_Microsoft_Windows_fonts)
- [List of typefaces included with OS X](http://en.wikipedia.org/wiki/List_of_typefaces_included_with_OS_X)
- [Common fonts to all versions of Windows & Mac equivalents](http://www.ampsoft.net/webdesign-l/WindowsMacFonts.html)
- [OS X：Mavericks Fonts](http://support.apple.com/kb/HT5944?viewlocale=zh_CN&locale=zh_CN)
- [OS X：Mountain Lion Fonts](http://support.apple.com/kb/HT5379?viewlocale=zh_CN&locale=zh_CN)
- [iOS 7：Fonts](http://support.apple.com/kb/HT5878?viewlocale=zh_CN&locale=zh_CN)
- [iOS 6：Fonts](http://support.apple.com/kb/HT5484?viewlocale=zh_CN&locale=zh_CN)
- [Windows UI 设计文档](http://dev.windows.com/en-us/design)
- [Supported UI Fonts in Windows Phone][wpfts]
- [Typography | Android Developers](http://developer.android.com/design/style/typography.html)

[wpfts]:http://msdn.microsoft.com/library/windows/apps/hh202920(v=vs.105).aspx#BKMK_SupportedUIFontsinWindowsPhone

__Open source projects about Chinese typography_

- [「漢字標準格式」——印刷品般的漢字網頁排版框架](https://github.com/ethantw/Han)
- [Entry.css - Configurable and readable Chinese text style sheet](https://github.com/zmmbreeze/Entry.css)



## Default Styles for Elements

### `<h1> - <h6>`

`<h1> - <h6>` are bolded and given a margin; `<h1>` is `1.5` times larger than usual; `<h2>` is `1.25` times larger；Others are same as usual.

`````html
<h1>h1 Title 1</h1>
<h2>h2 Title 2</h2>
<h3>h3 Title 3</h3>
<h4>h4 Title 4</h4>
<h5>h5 Title 5</h5>
<h6>h6 Title 6</h6>
`````
```html
<h1>h1 Title 1</h1>
<h2>h2 Title 2</h2>
<h3>h3 Title 3</h3>
<h4>h4 Title 4</h4>
<h5>h5 Title 5</h5>
<h6>h6 Title 6</h6>
```

### `<p>`

`````html
<p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
`````
```html
<p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
```

### `<hr>`

`````html
<hr/>
`````
```html
<hr/>
```

### `<blockquote>`

source is in `<small>` element.

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

### `<pre>`

More details in [Code](/css/code)。

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


### ul/ol

__`<ul>`:__

`````html
<ul>
  <li>Item #1</li>
  <li>Item #2
    <ul>
      <li>Item #1</li>
      <li>Item #2
        <ul>
          <li>Item #1</li>
          <li>Item #2</li>
        </ul>
      </li>
    </ul>
  </li>
  <li>Item #3</li>
  <li>Item #4</li>
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

__`<ol>`:__

`````html
<ol>
  <li>Item #1</li>
  <li>Item #2
    <ol>
      <li>Item #1</li>
      <li>Item #2
        <ol>
          <li>Item #1</li>
          <li>Item #2</li>
        </ol>
      </li>
    </ol>
  </li>
  <li>Item #3</li>
  <li>Item #4</li>
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

### `<dl>`

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

### `<table>`

This is the style after normalize. Find more details in [Table Component](/css/table).

`````html
<table>
  <caption>Title</caption>
  <thead>
    <tr>
      <th>Head #1</th>
      <th>Head #2</th>
      <th>Head #3</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Data #1</td>
      <td>Data #2</td>
      <td>Data #3</td>
    </tr>
    <tr>
      <td>Data #1</td>
      <td>Data #2</td>
      <td>Data #3</td>
    </tr>
  </tbody>
  <tfoot>
    <tr>
      <td>Footer #1</td>
      <td>Footer #2</td>
      <td>Footer #3</td>
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

### `<img>`

More details in [Image Component](/css/image)。

`````html
<img class="am-img-responsive" src="http://7jpqbr.com1.z0.glb.clouddn.com/bing-1.jpg" alt=""/>
`````

```html
<img class="am-img-responsive" src="http://7jpqbr.com1.z0.glb.clouddn.com/bing-1.jpg" alt=""/>

```

### Other Elements

<table class="am-table am-table-bordered am-table-hover am-table-striped">
  <thead>
  <tr>
    <th>Element</th>
    <th>Default Style</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>&lt;a&gt;</code></td>
    <td><a href="#">This is a link</a></td>
  </tr>
  <tr>
    <td><code>&lt;em&gt;</code></td>
    <td><em>I'm in an 'em' element</em></td>
  </tr>
  <tr>
    <td><code>&lt;strong&gt;</code></td>
    <td><strong>Strong element</strong></td>
  </tr>
  <tr>
    <td><code>&lt;code&gt;</code></td>
    <td><code>Inline code</code></td>
  </tr>
  <tr>
    <td><code>&lt;del&gt;</code></td>
    <td><del>Cut me</del></td>
  </tr>
  <tr>
    <td><code>&lt;ins&gt;</code></td>
    <td><ins>I'm inserted</ins></td>
  </tr>
  <tr>
    <td><code>&lt;mark&gt;</code></td>
    <td><mark>I'm marked with mark</mark></td>
  </tr>
  <tr>
    <td><code>&lt;q&gt;</code></td>
    <td><q>I'm in a q element <q>q element in q element</q> q element</q></td>
  </tr>
  <tr>
    <td><code>&lt;abbr&gt;</code></td>
    <td>Responsive web design <abbr title="Responsive web design">RWD</abbr></td>
  </tr>
  <tr>
    <td><code>&lt;dfn&gt;</code></td>
    <td>Define something <dfn title="Enjoy your music, photos and videos, anywhere anytime">DLNA</dfn></td>
  </tr>
  <tr>
    <td><code>&lt;small&gt;</code></td>
    <td><small>I'm smaller</small></td>
  </tr>
  </tbody>
</table>
