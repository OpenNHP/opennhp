# Gallery
---

This widget helps build galleries to organize images.

**All rights of pictures used in following samples belong to [Microsoft Bing](http://www.bing.com).**

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `gallery`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "img": "",      // Thumbnail URL
    "link": "",     // Image URL
    "title": "",    // Title
    "desc": ""      // Additional information. Support DOM object.
  }
];
return data;
```

## Specify a Thumbnail

Considering the limit of data traffic, this gallery widget allow using thumbnails in list and show the original image when use click on it.

### Using `data-rel`

Set the `data-rel` attribute to be the URL of original image.

`````html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview: 1}">
  <li>
    <div class="am-gallery-item">
      <img src="https://farm4.staticflickr.com/3835/15329524682_2642280b33_z.jpg"
           alt="远方 有一个地方 那里种有我们的梦想" data-rel="https://farm4.staticflickr.com/3835/15329524682_554d4c0886_k.jpg"/>
      <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
      <div class="am-gallery-desc">2375-09-26</div>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <img src="https://farm3.staticflickr.com/2941/15346557971_dec5c9ac36_z.jpg"
           alt="某天 也许会相遇 相遇在这个好地方" data-rel="https://farm3.staticflickr.com/2941/15346557971_d8f3d52978_k.jpg"/>
      <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
      <div class="am-gallery-desc">2375-09-26</div>
    </div>
  </li>
</ul>
`````
```html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview: 1}">
  <li>
    <div class="am-gallery-item">
      <img src="https://farm4.staticflickr.com/3835/15329524682_2642280b33_z.jpg"
           alt="远方 有一个地方 那里种有我们的梦想" data-rel="https://farm4.staticflickr.com/3835/15329524682_554d4c0886_k.jpg"/>
      <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
      <div class="am-gallery-desc">2375-09-26</div>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <img src="https://farm3.staticflickr.com/2941/15346557971_dec5c9ac36_z.jpg"
           alt="某天 也许会相遇 相遇在这个好地方" data-rel="https://farm3.staticflickr.com/2941/15346557971_d8f3d52978_k.jpg"/>
      <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
      <div class="am-gallery-desc">2375-09-26</div>
    </div>
  </li>
</ul>
```

### Using link

Use `data-am-gallery="{pureview:{target: 'a'}}"` for `target`.

This way is better considering accessibility. Users can still open the link to see the original image even if they have disabled the Javascript.

`````html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview:{target: 'a'}}">
  <li>
    <div class="am-gallery-item">
      <a href="https://farm4.staticflickr.com/3835/15329524682_554d4c0886_k.jpg" title="远方 有一个地方 那里种有我们的梦想"><img src="https://farm4.staticflickr.com/3835/15329524682_2642280b33_z.jpg"
           alt="远方 有一个地方 那里种有我们的梦想"/>
      <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
      <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="https://farm3.staticflickr.com/2941/15346557971_d8f3d52978_k.jpg" title="某天 也许会相遇 相遇在这个好地方"><img src="https://farm3.staticflickr.com/2941/15346557971_dec5c9ac36_z.jpg"
           alt="某天 也许会相遇 相遇在这个好地方"/>
      <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
      <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
</ul>
`````
```html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview:{target: 'a'}}">
  <li>
    <div class="am-gallery-item">
      <a href="https://farm4.staticflickr.com/3835/15329524682_554d4c0886_k.jpg" title="远方 有一个地方 那里种有我们的梦想"><img src="https://farm4.staticflickr.com/3835/15329524682_2642280b33_z.jpg" alt="远方 有一个地方 那里种有我们的梦想"/>
        <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
        <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="https://farm3.staticflickr.com/2941/15346557971_d8f3d52978_k.jpg" title="某天 也许会相遇 相遇在这个好地方"><img src="https://farm3.staticflickr.com/2941/15346557971_dec5c9ac36_z.jpg" alt="某天 也许会相遇 相遇在这个好地方"/>
        <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
        <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
</ul>
```

## Wechat image viewer

Wechat image viewer will be used when opening an image in wechat. It can be disabled using following attributes.

```html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview:{weChatImagePreview: false}}">
```

`````html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview:{weChatImagePreview: false}}">
  <li>
    <div class="am-gallery-item">
      <a href="https://farm4.staticflickr.com/3835/15329524682_554d4c0886_k.jpg" title="远方 有一个地方 那里种有我们的梦想"><img src="https://farm4.staticflickr.com/3835/15329524682_2642280b33_z.jpg"
                                                                                                                 alt="远方 有一个地方 那里种有我们的梦想"/>
        <h3 class="am-gallery-title">远方 有一个地方 那里种有我们的梦想</h3>
        <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
  <li>
    <div class="am-gallery-item">
      <a href="https://farm3.staticflickr.com/2941/15346557971_d8f3d52978_k.jpg" title="某天 也许会相遇 相遇在这个好地方"><img src="https://farm3.staticflickr.com/2941/15346557971_dec5c9ac36_z.jpg"
                                                                                                                alt="某天 也许会相遇 相遇在这个好地方"/>
        <h3 class="am-gallery-title">某天 也许会相遇 相遇在这个好地方</h3>
        <div class="am-gallery-desc">2375-09-26</div></a>
    </div>
  </li>
</ul>
`````
```html
<ul data-am-widget="gallery" class="am-gallery am-avg-sm-2 am-gallery-imgbordered" data-am-gallery="{pureview:{weChatImagePreview: false}}">
  <li>
    ...
  </li>
</ul>
```

## 数据接口

```javascript
{
  // id
  "id": "",

  // 自定义 class
  "className": "",

  // 主题
  "theme": "",

  "options": {
    "cols": 1,  // 列数
    "gallery": false // 是否开启点击图片全屏显示大图功能
  },

  //内容（*为必备项）
  "content": [
    {
      "img": "", // *
      "link": "", // 链接
      "title": "", // *图片标题
      "className": "", // 仅在设置了 link 后有效
      "desc": "" // 附加信息，支持DOM，为高级定制提供DOM接口
    }
  ]
}
```
