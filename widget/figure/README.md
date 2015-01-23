# Figure 单张图片
---

Figure 组件，用于放置单张图片。多图请使用 [Gallery](/widgets/gallery)。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `figure`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
  "img": "",          // 图片的路径
  "imgAlt": "",       // 图片描述，可访性和 SEO 用途，原网站图片的alt属性，如果alt为空或者不填写这项，则调用 figcaption
  "figcaption": ""    // 图片标题，显示在图标下方，这里的imgalt和figcaption选择一个填写就可以了
};

return data;
```

## 数据接口

如果觉得图片体积太大耗费流量，可以 `img` 里传缩略图， `rel` 里传大图地址。

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {
    "figcaptionPosition": "bottom", // 图标标题位置 top - 图片上方， bottom - 图片下方
    "zoomble": false // 是否启用图片缩放功能 ['auto'|true|false]
                      // 'auto' - 根据图片宽度自动判断，图大宽度大于窗口宽度时开启，否则不开启
                      // false 不开启；其他转型后非 false 的值，开启
                     // 此选项会作为 pureview 选项值 data-am-figure="{pureview: {{zoomable}} }"
  },
  "content": [
    {
      "img": "", // 图片（缩略图）路径
      "rel": "", // 大图路径
      "imgAlt": "", // 图片alt描述，如果为空则读取 figcaption
      "figcaption": "" // 图片标题
    }
  ]
}
```
