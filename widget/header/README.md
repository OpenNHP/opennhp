# Header
---

页头组件，为移动页面顶部的导航条设计。

## 顶部固定

在默认样式的基础上添加 `.am-header-fixed`。

__默认样式：__

```css
.am-header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1010;
}
```

```html
<header data-am-widget="header" class="am-header am-header-default am-header-fixed">
  <div class="am-header-left am-header-nav">
    <a href="#left-link" class="">
      <i class="am-header-icon am-icon-home"></i>
    </a>
  </div>
  <h1 class="am-header-title">
    <a href="#title-link">Amaze UI</a>
  </h1>
  <div class="am-header-right am-header-nav">
    <a href="#right-link" class="">
      <i class="am-header-icon am-icon-bars"></i>
    </a>
  </div>
</header>
```

如果页面中有固定顶部的 Header，JS 会在 `<body>` 上添加 `.am-with-fixed-header`，这个 class 下面默认设置了 `padding-top: 49px`，可以根据具体情况做调整。

```css
.am-with-fixed-header {
  padding-top: @am-header-height;
}
```


## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `header`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {

  "left": [
      {
        "link": "",         // url : http://xxx.xxx.xxx
        "title": "",        // 链接标题
        "icon": "",         // 字体图标名称: 使用 Amaze UI 字体图标 http://www.amazeui.org/css/icon
        "customIcon": ""    // 自定义图标 URL，设置此项后当前链接不再显示 icon
      }
  ],

  "title": "",			//可写 html 标签

  "right": [ // 右侧字段含义同左侧
      {
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": ""
      }
  ]
};

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": {
    "fixed": false
  },

  "content": {
    "left": [{
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": "",
        "className": ""
    }],

    "title": "",

    "right": [{
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": "",
        "className": ""
    }]
  }
}
```
