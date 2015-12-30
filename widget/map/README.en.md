# Map
---

**This widget calls Baidu map(speedy version) API, and only works on touch devices.**

If the address positioning is not precise enough, you may use longitude/latitude positioning instead. The address will be dismissed if longitude and latitude is provided.

__How to get longitude/latitude?__Open [Baidu map position picker](http://api.map.baidu.com/lbsapi/getpoint/index.html), and click on the position. Then click on the copy icon on the upper right corner to get your longitude and latitude.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `map`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- This widget don't need data-binding.

## API

```javascript
{
  // ID
  "id": "",

  // Customized class
  "className": "",

  // Theme
  "theme": "",

  // Options
  "options": {
    "name": "", // Name of this label
    "address": "", // Address. Used to position the label.
    "longitude": "", // Longtitude. Used to improve the precision.
    "latitude": "" // Latitude. Used to improve the precision.
    "zoomControl": Boolean, // Whether allow zooming.
    "scaleControl": Boolean, // Whether show scale on map.
    "setZoom": Number, // Zoom level.
    "icon": "" // Icon of label.
  }
}
```
