# Gallery 图片画廊
---

图片画廊组件，用于展示图片为主体的内容。

**演示图标版权归[微软 Bing](http://www.bing.com) 所有。**

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `gallery`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = [
  {
    "img": "",      // 图片地址
    "link": "",     // 填写点击图片弹出放大时的图片路径
    "title": "",    // 图片标题
    "desc": ""      // 附加信息，支持DOM，为高级定制提供DOM接口
  }
];
return data;
```

## 指定缩略图

基于节省流量考虑，可以指定缩略图，用户点击放大的时候再显示大图。

### 使用 `data-rel`

将大图放在 `<img>` 的 `data-rel` 属性上。

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

### 使用链接

设置 `target`: `data-am-gallery="{pureview:{target: 'a'}}"`。

从可访性的角度而言，这种方式更好一些：用户再禁用 JavaScript 以后仍然可以打开链接查看大图。

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

## 不使用微信图片查看器

如果在微信中打开时不想调用微信的图片查看器，可以通过以下选项关闭：

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
