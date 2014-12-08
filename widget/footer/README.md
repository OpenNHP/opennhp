# Footer 页脚
---

用于页面底部显示版本切换、版权等信息。

## 使用方法

### 直接使用

- 拷贝演示中的代码，粘贴到 Amaze UI HTML 模板（[点此下载](/getting-started)） `<body>` 区域；
- 将示例代码中的内容替换为自己的内容。

### 使用 Handlebars

本组件 Handlebars partial 名称为 `footer`，使用细节参照[折叠面板组件](/widgets/accordion)。

### 云适配 WebIDE

- 将组件拖入编辑界面；
- 点击右侧面板里的【数据采集】按钮，按以下格式采集数据。

```javascript
var data = {
  "lang": context.__lang,     // 默认，无需改动。若改动赋值为“en”则为英文，否则为中文！
  "owner": "",                // 网站名字 可以填写公司名称或者其他内容。
  "companyInfo": [
    {                       // 网站信息
      "detail": ""        // 必要时可加 a 标签跳转到某个页面，网站的详细信息，在页面中的footer部分就可以看到这里的文字
    },
    {
      "detail": ""        // 必要时可加 a 标签跳转到某个页面，网站的详细信息，在页面中的footer部分就可以看到这里的文字
    }
  ]
};
return data;
```

__数据说明：__

- `switchName`：可写：手机版等，默认 `云适配版`；
- `owner`：网站名字；
- `slogan`：弹出层广告语；
- `companyInfo`：填写公司相关信息，此为一对象数组，每一条信息通过对detail赋值来实现！

### Add to Homescreen

- 目前仅支持 iOS，需要在 header 里设置 icon。

参考链接：

- [Safari Web Content Guide: Configuring Web Applications](https://developer.apple.com/library/ios/documentation/AppleApplications/Reference/SafariWebContent/ConfiguringWebApplications/ConfiguringWebApplications.html)
- [iOS Human Interface Guidelines: Icon and Image Sizes](https://developer.apple.com/library/ios/documentation/UserExperience/Conceptual/MobileHIG/IconMatrix.html#//apple_ref/doc/uid/TP40006556-CH27-SW1)
- [Add to Homescreen - Google Chrome Mobile -- Google Developers](https://developers.google.com/chrome/mobile/docs/installtohomescreen)
- [Everything you always wanted to know about touch icons](http://mathiasbynens.be/notes/touch-icons)

添加到桌面图标功能默认开启，可以通过设置以下代码禁用：

```javascript
window.AMUI_NO_ADD2HS = true;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
     "modal": "",
     "techSupprtCo": "",
     "techSupprtNet": "",
     "textPosition": ""
  },

  "content": {
    "lang": "",
    "switchName": "",
    "owner": "",
    "slogan": "",
    "companyInfo": [
      {
        "detail": ""
      }
    ]
  }
}
```
