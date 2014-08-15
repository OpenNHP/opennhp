# Header
---
用于自动转码项目页头部分

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

  "left": "[{
    "link": "",         //url : http://xxx.xxx.xxx。
    "icon": "",         //class : 使用 Amaze UI 字体图标 http://www.amazeui.org/css/icon，优先使用字体图标。
    "img": ""           //img : 图片
  }]",                  //left、right 可以写多个

  "title": "",			//可写 html 标签（可选）

  "right": "[{
      "link": "",
      "icon": "",
      "img": ""
    }]"
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
  },

  "content": {
    "left": [{
        "link": "",
        "icon": "",
        "img": ""
    }],

    "title": "",

    "right": [{
        "link": "",
        "icon": "",
        "img": ""
    }]
  }
}
```
