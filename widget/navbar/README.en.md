# Navbar 工具栏
---

底部工具栏组件。

## 使用方法

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `navbar`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title": "",        // title 值：要显示的文本

    "link": "",         // link 值 ：如果是电话则写"tel:0101245678"

    "customIcon": "",   // customIcon与下边的icon选用其中之一，customIcon用于上传自定义的小图标，写法："customIcon": context.__root + "xxx.png"

    "icon": "",         //icon 值,例如：分享的图标在AMUI中是am-icon-share ，则此时的icon写法为： "icon": "share"

    "dataApi": ""       //dataApi 值可以填写"data-am-navbar-share"（用于分享模块） 或者"data-am-navbar-qrcode"(用于二维码模块)。若要使用自己上传的二维码 'data-am-navbar-qrcode = 二维码地址' 即可
  }
];

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
      "cols": "", //cols 值 ：列数，若数据列数较多，会自动添加一个列表
  },

  "content": [
    {
      "title": "", // title 值：要显示的文本
      "link": "", // link 值 ：链接地址，电话则写"tel:0101245678"
      "className": "",
      "customIcon": "", // customIcon与icon选用其中之一，customIcon用于上传的小图标/ 写法： "customIcon": "xxx.png"
      "icon": "",//icon 值：使用icon font 例如：分享的图标 am-icon-share ，icon写法： "icon": "share" ,更多的查看amui icon 目录下
      "dataApi": ""//dataApi 值为 data-am-navbar-share（用于分享模块） 或者 data-am-navbar-qrcode(用于二维码模块) 若使用自己上传的二维码 'data-am-navbar-qrcode = 二维码地址' 即可
      }
  ]
}
```
