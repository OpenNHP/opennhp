# List
---

This widget helps to build lists of different styles.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `list_news`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
   // Title
  "header": {
    "title": "News feed",
    "link": "###",
    "moreText": "More >",
    "morePosition": "bottom"    // Position of 'more' [bottom|top]
  },

  // Main Contents
  "main": [
    {
      "title": "",            // Title
      "link": "",             // Link
      "date": "",             // Date
      "desc": "",             // Abstract. Support html
      "img": "",              // Thumbnail URL

      // Data Interface
      "thumbAddition": "",    // Additional information shown at the bottom of the thumbnail. Support html. Only avaliable in thumb mode.
      "mainAddition": ""      // Additional information. Support html. Avaliable in all modes.
    }
  ]
};

return data;
```

## Data Interface

```javascript
{
  // id
  "id": "",

  // Customized class name
  "className": "",

  // Theme
  "theme": "",

  "options": {
   // normal、thumb
    "type": "",

    // Required when type is thumb.
    "thumbPosition": "bottom"
    // top - Usually used in full width mode.
    // bottom - Bottom-left | bottom-right
    // left - Right
  },

  // Data
  "content": {
    "header": {
      "title": "News feed", // Title of the list
      "link": "###",
      "className": "",
      "moreText": "More >", // Text of 'more' link
      "morePosition": "bottom" // Position of 'more' [bottom|top]. Won't show 'more' link if this is not set.
    },
    main: [
      {
        "title": "", // Title of the content
        "link": "",  // Link
        "className": "",  // Customized class
        "date": "",  // Date
        "desc": "",  // Abstruct. Support html.
        "img": "",              // Thumbnail URL

        // Data Interface
        "thumbAddition": "",    // Additional information shown at the bottom of the thumbnail. Support html. Only avaliable in thumb mode.
        "mainAddition": ""      // Additional information. Support html. Avaliable in all modes.
      }
    ]
}
```
