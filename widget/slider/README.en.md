# Slider
---

See more details about this slider in [Slider JS plugin](/javascript/slider?_ver=2.x)。

**All rights of pictures used in following samples belong to [Microsoft Bing](http://www.bing.com).**

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `slider`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var content = [
  {
    "img": "",      // Source URL of image. Required.
    "link": "",     // Link of this image. Optional.
    "thumb": "",    // URL of thumbnail. Optional
    "desc": ""      // Additional information of this image. Support HTML.
  }
];

return content;
```

## 数据接口

```javascript
{
  // id
  "id": "",

  // Customized Class
  "className": "",

  // Theme
  "theme": "",

  // Configuration. Use JSON to config the slider.
  "sliderConfig": "{}",

  //Content (* means required)
  "content": [
    {
      "img": "",      // Source URL of image. Required.
      "link": "",     // Link of this image. Optional.
      "className": "",
      "thumb": "",    // URL of thumbnail. Optional
      "desc": ""      // Additional information of this image. Support DOM object.
    }
  ]
}
```

## Slider Configuration

Attentions: Callback function can't be passed through JSON string. Please use slider through JS if callback is needed:

- Add the `am-slider-manual` class to disable the default initialization;
- Initialize slider menually in JS.

```javascript
$(function() {
  $('.am-slider-manual').flexslider({
  // options
  });
});
```

## FAQ

### Why does my slider crash in some browsers?
The quatation marks in example are encoded **unnecessarily**, which may cause some error. The solution is replacing the `&quot;` to `"`.**The codes copied using the `Copy` button has been using `"` since May, 15th, 2015**。

```html
<!-- Problem Codes -->
<div data-am-widget="slider" class="am-slider am-slider-d3" data-am-slider='{&quot;controlNav&quot;:&quot;thumbnails&quot;,&quot;directionNav&quot;:false}'>

<!-- Correct Codes -->
<div data-am-widget="slider" class="am-slider am-slider-d3" data-am-slider='{"controlNav":"thumbnails","directionNav":false}'>
```

### How to set width/height?

The widget doesn't provide width/height interface. But the width is set to `100%`, and the height is decided by the contents. Please use css to change the width/height of container to adjust the width/height of this widget.
