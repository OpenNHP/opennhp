# Navbar 工具栏
---

This widget helps create navbars at the bottom of the viewport.

## Usage

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
    "title": "",        // title: text that shown in the navbar

    "link": "",         // link: Link of icon. Use "tel: ***-****-***" for phone number.

    "customIcon": "",   // Only one of customIcon and icon will be applied, so please don't use them at the same time. CustomIcon is used to upload icons. Usage: "customIcon": context.__root + "xxx.png"

    "icon": "",         //icon. For example: The share icon is am-icon-share in amaze UI, so use "icon": "share" here to use share icon.

    "dataApi": ""       //dataApi can be "data-am-navbar-share"( For share module) or "data-am-navbar-qrcode"( For qrcode module). Use 'data-am-navbar-qrcode = YourQRCodeURL' for your own qrcode.
  }
];

return data;
```

## Data Binding

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
      "cols": "", //cols: The number of columns. If the actual columns is more than this value, a list will be added autometically.
  },

  "content": [
    {
      "title": "",        // title: text that shown in the navbar
      "link": "",         // link: Link of icon. Use "tel: ***-****-***" for phone number.
      "customIcon": "",   // Only one of customIcon and icon will be applied, so please don't use them at the same time. CustomIcon is used to upload icons. Usage: "customIcon": context.__root + "xxx.png"
      "icon": "",         //icon. For example: The share icon is am-icon-share in amaze UI, so use "icon": "share" here to use share icon.
      "dataApi": ""       //dataApi can be "data-am-navbar-share"( For share module) or "data-am-navbar-qrcode"( For qrcode module). Use 'data-am-navbar-qrcode = YourQRCodeURL' for your own qrcode.
    }
  ]
}
```
