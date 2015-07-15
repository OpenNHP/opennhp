# Paragraph
---

This is a paragraph widget that can help organize texts, images, links and etc.
文本段落组件，可用于放置文本、图片、链接等。

**All rights of pictures used in following samples belong to [Microsoft Bing](http://www.bing.com).**

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `paragraph`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "content": ""   // Content of paragraph
};

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
    "imgLightbox": false, // Whether enable light box.
    "tableScrollable": false // Whether allow vertical scroll.
  },

  "content": {
    "content": ""
  }
}
```
