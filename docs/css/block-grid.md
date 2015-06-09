# AVG-Grid
---

Average Grid，均分网格（原 Block Grid），使用 `ul` / `ol` 创建等分列，用于内容的排列。

响应式断点为：

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th style="width: 160px">Class</th>
    <th>区间</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td><code>.am-avg-sm-*</code></td>
    <td><code>0 - 640px</code></td>
  </tr>
  <tr>
    <td><code>.am-avg-md-*</code></td>
    <td><code>641px - 1024px</code></td>
  </tr>
  <tr>
    <td><code>.am-avg-lg-*</code></td>
    <td><code>1025px + </code></td>
  </tr>
  </tbody>
</table>

与布局网格不同的是，这里的**数字表示几等分**，而不是占 12 等分中的几列，比如 `.am-avg-sm-2` 会将子元素 `<li>` 的宽度设置为 `50%`。

考虑到通用性（菜单、图片）等，`<li>` 没有设置 `padding`，使用时需根据需求自行设置。

另外需要注意的 AVG Grid __只能用于 `<ul>` / `<ol>` 结构__。

~~下面的演示中，添加了以下自定义样式（Less）：~~

这部分代码已经整合到[缩略图](/css/thumbnail?_ver=2.x)，无需再添加。

```css
.doc-block-grid {
  margin-left: -5px;
  margin-right: -5px;
  > li {
    padding: 0 5px 10px 5px;
    img {
      border: 1px solid #CCC;
      padding: 2px;
      background: #FFF;
    }
  }
}
```

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 基本使用

只添加 `.am-avg-sm-*`，应用于所有屏幕尺寸。

`````html
<ul class="am-avg-sm-4 am-thumbnails">
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
</ul>
`````

```html
<ul class="am-avg-sm-4 am-thumbnails">
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
</ul>
```

## 响应式

按需增加更多响应式 class，缩放窗口可以查看响应效果。

`````html
<ul class="am-avg-sm-2 am-avg-md-3 am-avg-lg-4 am-thumbnails">
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
</ul>
`````

```html
<ul class="am-avg-sm-2 am-avg-md-3 am-avg-lg-4 am-thumbnails">
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-2.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-3.jpg" /></li>
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-4.jpg" /></li>
</ul>
```

## 九宫格

看到同学提这样的需求，就简单写一个示例。

<style>
  .boxes {
    width: 300px;
  }

  .boxes .box {
    height: 100px;
    color: #eee;
    line-height: 100px;
    text-align: center;
    font-weight: bold;
    transition: transform .25s ease;
  }

  .boxes .box:hover {
    font-size: 250%;
    transform: rotate(360deg);
    -webkit-animation: heart .45s ease-in-out .15s infinite;
    animation: heart .45s ease-in-out .15s infinite;
  }

  .box-1 {
    background-color: red;
  }

  .box-2 {
    background-color: orange;
  }

  .box-3 {
    background-color: #0000ff;
  }

  .box-4 {
    background-color: #008000;
  }

  .box-5 {
    background-color: red;
  }

  .box-6 {
    background-color: orange;
  }

  .box-7 {
    background-color: #0000ff;
  }

  .box-8 {
    background-color: #008000;
  }

  .box-9 {
    background-color: red;
  }

  @-webkit-keyframes heart {
    0% {
      font-size: 150%;
    }

    100% {
      font-size: 300%;
    }
  }

  @keyframes heart {
    0% {
      font-size: 150%;
    }

    100% {
      font-size: 300%;
    }
  }
</style>

`````html
<ul class="am-avg-sm-3 boxes">
  <li class="box box-1">1</li>
  <li class="box box-2">2</li>
  <li class="box box-3">3</li>
  <li class="box box-4">4</li>
  <li class="box box-5">5</li>
  <li class="box box-6">6</li>
  <li class="box box-7">7</li>
  <li class="box box-8">8</li>
  <li class="box box-9">9</li>
</ul>
`````

```html
<ul class="am-avg-sm-3 boxes">
  <li class="box box-1">1</li>
  <li class="box box-2">2</li>
  <li class="box box-3">3</li>
  <li class="box box-4">4</li>
  <li class="box box-5">5</li>
  <li class="box box-6">6</li>
  <li class="box box-7">7</li>
  <li class="box box-8">8</li>
  <li class="box box-9">9</li>
</ul>
```
```css
.boxes {
  width: 300px;
}

.boxes .box {
  height: 100px;
  color: #eee;
  line-height: 100px;
  text-align: center;
  font-weight: bold;
  transition: all .2s ease;
}

.boxes .box:hover {
  font-size: 250%;
  transform: rotate(360deg);
}

.box-1 {
  background-color: red;
}

.box-2 {
  background-color: orange;
}

.box-3 {
  background-color: #0000ff;
}

.box-4 {
  background-color: #008000;
}

.box-5 {
  background-color: red;
}

.box-6 {
  background-color: orange;
}

.box-7 {
  background-color: #0000ff;
}

.box-8 {
  background-color: #008000;
}

.box-9 {
  background-color: red;
}
```

