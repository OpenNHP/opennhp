# Thumbnail
---

Thumbnail is usually used in list of images and list of combination of images and text.

**All rights of pictures used in following samples belong to [Microsoft Bing](http://www.bing.com).**

## Default Style

Add the `.am-thumbnail` class to `<img>` to create a thumbnail with defatult style; or add a container outside the `<img>`, such as `<div>`, `<figure>`, `<a>` etc, and add this class to the container.

### Using with Grid

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

### Using with AVG Grid

Add AVG Grid class and `.am-thumbnails` to the `<ul>` elements.

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

## Title
`````html
<div class="am-g">
  <div class="am-u-sm-4">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <h3 class="am-thumbnail-caption">Caption #1</h3>
    </div>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">Caption #2</figcaption>
    </a>
  </div>
  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">Caption #3</figcaption>
    </figure>
  </div>
</div>
`````

```html
<div class="am-g">
  <div class="am-u-sm-4">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <h3 class="am-thumbnail-caption">Caption #1</h3>
    </div>
  </div>

  <div class="am-u-sm-4">
    <a href="#" class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">Caption #2</figcaption>
    </a>
  </div>
  <div class="am-u-sm-4">
    <figure class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bing-1.jpg" alt=""/>
      <figcaption class="am-thumbnail-caption">Caption #3</figcaption>
    </figure>
  </div>
</div>
```


## With Text

Add a div with `.am-thumbnail-caption` into `am-thumbnail`, and then you can use any kind of HTML content inside the caption.


`````html
<div class="am-g">
  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>One Hundred Years of Solitude</h3>
        <p>wherever they might be they always remember that the past was a lie, that memory has no return, that every spring gone by could never be recovered, and that the wildest and most tenacious love was an ephemeral truth in the end.</p>
        <p>
          <button class="am-btn am-btn-primary">Solitude</button>
          <button class="am-btn am-btn-default">One Hundred Years</button>
        </p>
      </div>
    </div>
  </div>

  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>One Hundred Years of Solitude</h3>
        <p>wherever they might be they always remember that the past was a lie, that memory has no return, that every spring gone by could never be recovered, and that the wildest and most tenacious love was an ephemeral truth in the end.</p>
        <p>
          <button class="am-btn am-btn-primary">Solitude</button>
          <button class="am-btn am-btn-default">One Hundred Years</button>
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
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>One Hundred Years of Solitude</h3>
        <p>...</p>
        <p>
          <button class="am-btn am-btn-primary">Solitude</button>
          <button class="am-btn am-btn-default">One Hundred Years</button>
        </p>
      </div>
    </div>
  </div>

  <div class="am-u-sm-6">
    <div class="am-thumbnail">
      <img src="http://s.amazeui.org/media/i/demos/bw-2014-06-19.jpg?imageView2/0/w/600" alt=""/>
      <div class="am-thumbnail-caption">
        <h3>One Hundred Years of Solitude</h3>
        <p>...</p>
        <p>
          <button class="am-btn am-btn-primary">Solitude</button>
          <button class="am-btn am-btn-default">One Hundred Years</button>
        </p>
      </div>
    </div>
  </div>

</div>
```
