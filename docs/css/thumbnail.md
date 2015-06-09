# Thumbnail
---

Thumbnail 组件主要用于显示图片列表及图文混排列表。

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 基本样式

在 `<img>` 添加 `.am-thumbnail` 类；也可以在 `<img>` 外面添加一个容器，如 `<div>`、`<figure>`、`<a>` 等，再将 class 添加到容器上。

### 结合网格使用

`````html
<div class="am-g">
  <div class="am-u-sm-4">
    <img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-2.jpg" alt=""/>
    </a>
  </div>
  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-3.jpg" alt=""/>
    </figure>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-sm-4">
    <img class="am-thumbnail" src="..." alt=""/>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="..." alt=""/>
    </a>
  </div>

  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="..." alt=""/>
    </figure>
  </div>
</div>
```

### 结合 AVG Grid 使用

添加 AVG Grid class 同时添加 `.am-thumbnails`。

`````html
<ul class="am-avg-sm-3 am-thumbnails">
  <li><img class="am-thumbnail" src="http://s.amazeui.org/media/i/demos/bing-1.jpg"
           alt=""/></li>

  <li><a href="#" class="am-thumbnail">
    <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
  </a></li>

  <li>
    <figure class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
    </figure>
  </li>
</ul>
`````

```html
<ul class="am-avg-sm-3 am-thumbnails">
  <li>
    <img class="am-thumbnail" src="..." alt=""/>
  </li>

  <li>
    <a href="#" class="am-thumbnail">
      <img src="..." alt=""/>
    </a>
  </li>

  <li>
    <figure class="am-thumbnail">
      <img src="..." alt=""/>
    </figure>
  </li>
</ul>
```

## 标题

`````html
<div class="am-g">
  <div class="am-u-sm-4">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <h3 class="am-thumbnail-caption">图片标题 #1</h3>
    </div>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">图片标题 #2</figcaption>
    </a>
  </div>
  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">图片标题 #3</figcaption>
    </figure>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-sm-4">
    <div class="am-thumbnail">
      <img src="..." alt=""/>
      <h3 class="am-thumbnail-caption">图片标题 #1</h3>
    </div>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="..." alt=""/>
      <figcaption class="am-thumbnail-caption">图片标题 #2</figcaption>
    </a>
  </div>
  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="..." alt=""/>
      <figcaption class="am-thumbnail-caption">图片标题 #3</figcaption>
    </figure>
  </div>
</div>
```


## 图文混排

在`am-thumbnail`内加入`.am-caption`可以添加任何类型的HTML内容标题、段落、或按钮。


`````html
<div class="am-g">
  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>百年孤独选</h3>
        <p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
        <p>
          <button class="am-btn am-btn-primary">孤独</button>
          <button class="am-btn am-btn-default">百年</button>
        </p>
      </div>
    </div>
  </div>

  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>百年孤独选</h3>
        <p>无论走到哪里，都应该记住，过去都是假的，回忆是一条没有尽头的路，一切以往的春天都不复存在，就连那最坚韧而又狂乱的爱情归根结底也不过是一种转瞬即逝的现实。</p>
        <p>
          <button class="am-btn am-btn-primary">孤独</button>
          <button class="am-btn am-btn-default">百年</button>
        </p>
      </div>
    </div>
  </div>

</div>
`````
```html
<div class="am-g">
  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="..." alt=""/>
      <div class="am-thumbnail-caption">
        <h3>图片标题</h3>
        <p>...</p>
        <p>
          <button class="am-btn am-btn-primary">按钮</button>
          <button class="am-btn am-btn-default">按钮</button>
        </p>
      </div>
    </div>
  </div>

  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="..." alt=""/>
      <div class="am-thumbnail-caption">
        <h3>图片标题</h3>
        <p>...</p>
        <p>
          <button class="am-btn am-btn-primary">按钮</button>
          <button class="am-btn am-btn-default">按钮</button>
        </p>
      </div>
    </div>
  </div>
</div>
```
