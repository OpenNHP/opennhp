# Menu
---

This wiget helps to create a menu.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `menu`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title": "",            // Title of the top level menu.
    "link": "",             // Link of the top level menu.
    "className": "",
    "subMenu": [
      {                   // SubMenu
        "title": "",    // Title of this sumbmenu
        "link": "",     // Link of this sumbmenu
        "target": ""
      }
    ],
    "subCols": 3            // Number of columns in subMenu.
  }
];

return data;
```

## Data binding

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": {
    "cols": 5, // Number of columns in top menu. [1-12]

    "offCanvasFlip": false, // Only avilable in offcanvas theme. The offCanvas will be on the left if this is set to false, and on the right if this is true.
  },

  // Data
  "content": [
    {
      "title": "", // Title of the top level menu.
      "link": "", // Link of the top level menu.
      "className": "", // The customized class of "li" in the menu
      "subCols": "", // Number of columns in subMenu.
      "channelLink": "Enter channel &raquo;", // Show text in submenu and link to this channel.
      "subMenu": [{
        "title": "", // Title of the submenu.
        "link": "", // Link of the submenu
        "className": "", // Class of the menu
      }]
    }
  ]
}
```
