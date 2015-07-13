# Header
---

This widget helps build headers that can be used as a navigate bar, especially in mobile pages.

## Fix top

Add the `.am-header-fixed` class to the default style.

__Default style：__

```css
.am-header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1010;
}
```

```html
<header data-am-widget="header" class="am-header am-header-default am-header-fixed">
  <div class="am-header-left am-header-nav">
    <a href="#left-link" class="">
      <i class="am-header-icon am-icon-home"></i>
    </a>
  </div>
  <h1 class="am-header-title">
    <a href="#title-link">Amaze UI</a>
  </h1>
  <div class="am-header-right am-header-nav">
    <a href="#right-link" class="">
      <i class="am-header-icon am-icon-bars"></i>
    </a>
  </div>
</header>
```

If there is an header fixed to the top of a page, our JS will add the `.am-with-fixed-header` class to the `<body>`, which has a top padding of 49px. This padding can be adjusted according to the need in the project.

```css
.am-with-fixed-header {
  padding-top: @am-header-height;
}
```


## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `header`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {

  "left": [
      {
        "link": "",         // url : http://xxx.xxx.xxx
        "title": "",        // Text of the link.
        "icon": "",         // Name of icons in Amaze UI font icon: http://www.amazeui.org/css/icon
        "customIcon": ""    // URL of customized icon. If this is set a value, the icon above will not be shown.
      }
  ],

  "title": "",			//Title. Support HTML.

  "right": [ // Same as the left
      {
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": ""
      }
  ]
};

return data;
```

## Data Interface

```javascript
{
  "id": "",

  "className": "",

  "theme": "",

  "options": {
    "fixed": false
  },

  "content": {
    "left": [{
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": "",
        "className": ""
    }],

    "title": "",

    "right": [{
        "link": "",
        "title": "",
        "icon": "",
        "customIcon": "",
        "className": ""
    }]
  }
}
```
