# AVG-Grid
---

Average Grid (used to be Block Grid) can be used to arrange contents. Use `ul` / `ol` to create columns in same width.

Responsive Breakpoints:

<table class="am-table am-table-bordered am-table-striped">
  <thead>
  <tr>
    <th style="width: 160px">Class</th>
    <th>Interval</th>
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

Different from Grid, **number here represents the number of divisions**. For Instance, `.am-avg-sm-2` will set the width of `<li>` to be `50%`.

Considering generality (for menus, pictures etc), `padding` is not set for `<li>`. You may set it as your wish.

Attention: AVG Grid __can be only apply to `<ul>` / `<ol>` stucture__.

~~The following examples use styles below:~~

These codes have been integrated into [thumbnail](/css/thumbnail?_ver=2.x).

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

**All rights of pictures used in following samples belongs to [Microsoft Bing](http://www.bing.com).**

## Usage

When there is only `.am-avg-sm-*`, it will be apply to all screens.

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

## Responsive Design

You may add more responsive classes as you wish. Scale the window to check the responsive effect.

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

## More Example

Someone asked us to provide more examples, so here it is:

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

