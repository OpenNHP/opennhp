# Figure
---

The Figure widget is used to place single image. Please use [Gallery](/widgets/gallery) for multiple images.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of Figure widget is `figure`. See [Accordion](/widgets/accordion) for more details

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "img": "",          // Path of image.
  "imgAlt": "",       // Alternative text. If this is empty then use figcaption.
  "figcaption": ""    // The caption of image. Shown at the bottom of image. 
};

return data;
```

## Data Interface

If the image is too big, use the thumbnail in `img`, and use the original image in `rel`.

```javascript
{
  "id": "",
  "className": "",
  "theme": "",
  "options": {
    "figcaptionPosition": "bottom", // Caption position: top | bottom
    "zoomble": false // Whether allow zooming: ['auto'|true|false]
                      // 'auto' - Judge by the width of image. Allow zooming if the width of image is larger than viewport.
                     // This option will be used as the data-am-figure="{pureview: {{zoomable}} }" option in pureview.
  },
  "content": [
    {
      "img": "", // Path of image(thumbnail)
      "rel": "", // Path of original image.
      "imgAlt": "", // Alternative text. If this is empty then use figcaption.
      "figcaption": "" // Caption.
    }
  ]
}
```
