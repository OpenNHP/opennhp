# Tabs
---

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `tabs`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title":   "",   // Title of this tab
    "content": "",   // Content of this tab
    "active":  false // Whether activate this tab. true | false. Only allow one tab to be active.
  }
];

return data;
```

## 数据接口

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": {
    "noSwipe": false; // Whether enable touch events.
  }

  "content": [
    {
      "title": "",
      "content": "",
      "active": ""
    }
  ]
}
```

## FAQ

### How to disable touch events?

Add the `data-am-tabs-noswipe="1"` class.

```html
<div data-am-widget="tabs" class="am-tabs am-tabs-default" data-am-tabs-noswipe="1">
```
