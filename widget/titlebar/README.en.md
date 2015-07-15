# Titlebar
---

This widget can be used as header or title.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `titlebar`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "title": "",		// Title
  "link": "",       // Title link. Optional
  "nav": [          // Additional links.
    {
      "link": "",
      "title": "",
      "className": ""
    }
  ]
};

return data;
```

## Data Binding

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": null,

  "content": {
    "title": "",
    "link": "",
    "className": "",
    "nav": [
      {
        "link": "",
        "title": "",
        "className": ""
      }
    ]
  }
}
```
