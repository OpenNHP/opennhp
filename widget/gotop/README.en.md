# GoTop
---

This widget create an button to help users to go back to the top. When using `fixed` style, the visibility of button will be determined by the position of scroll bar. (Visible only when it is over 50px away from top).

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `gotop`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.


```javascript
var data = {
  "title":      "Go top"	  // Title (not shown in some styles)
  "icon":       "arrow-up"	// Name of icon. Use Icon Font.
  "customIcon": ""          // URL of customized icon.
};

return data;
```


## Data Interface

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "content": {
    "title":      "Go top"    // Title (not shown in some styles)
    "icon":       "arrow-up"  // Name of icon. Use Icon Font.
    "customIcon": ""          // URL of customized icon.
  }
}
```
