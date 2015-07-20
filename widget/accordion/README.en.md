# Accordion
---

This widget help building a beautiful accordion panel.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

First, import Handlebars library and `amui.widget.helper.js`(See `widget.html` file in [Amaze UI HTML Template](/getting-started)).

Then use either following ways to bind data.

__1. Insert the script of widget into the template:__

```html
<script type="text/x-handlebars-template" id="my-tpl">
  {{>accordion accordionData}}
</script>
```

Then query the template and pass the data to it.

```javascript
$(function() {
  var $tpl = $('#my-tpl'),
      tpl = $tpl.text(),
      template = Handlebars.compile(tpl),
      data = {
        accordionData: {
          "theme": "basic",
          "content": [
            {
              "title": "Title 1",
              "content": "Content 1",
              "active": true
            },
            {
              "title": "Title 2",
              "content": "Content 2"
            },
            {
              "title": "Title 3",
              "content": "Content 3"
            }
          ]
        }
      },
      html = template(data);

  $tpl.before(html);
});
```

The rendered HTML looks like this:

```html
<section data-am-widget="accordion" class="am-accordion am-accordion-basic doc-accordion-class"
         id="doc-accordion-example" data-accordion-settings="{  }">
  <dl class="am-accordion-item am-active">
    <dt class="am-accordion-title">Title 1</dt>
    <dd class="am-accordion-content">Content 1</dd>
  </dl>
  <dl class="am-accordion-item">
    <dt class="am-accordion-title">Title 2</dt>
    <dd class="am-accordion-content">Content 2</dd>
  </dl>
  <dl class="am-accordion-item">
    <dt class="am-accordion-title">Title 3</dt>
    <dd class="am-accordion-content">Content 3</dd>
  </dl>
</section>
```

If there are other widget or templates in the same page, we recommend using this way. Maintenance will be easier in this way.

__Second, Directly render the widget with Handlebars：__

```javascript
var template = Handlebars.compile('{{>accordion}}'),
    data = {
      accordionData: {
        "id": "doc-accordion-example",
        "className": "doc-accordion-class",
        "theme": "basic",
        "content": [
          {
            "title": "Title 1",
            "content": "Content 1",
            "active": true
          },
          {
            "title": "Title 2",
            "content": "Content 2"
          },
          {
            "title": "Title 3",
            "content": "Content 3"
          }
        ]
      }
    },
    html = template(data.accordionData);

$('body').append(html);
```

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = [
  {
    "title": "",    // Title. Support using html
    "content": ""   // Content. Support using html
  }
];

return data;
```

## Data Interface

```javascript
{
  // ID
  "id": "",

  // Customized class
  "className": "",

  // Theme
  "theme": "",

  "options": {
    "multiple": false // Allow opening multiple panel at the same time. Default value is FALSE.
  },

  // Content(* means required)
  "content": [
    {
      "title": "", // Title. Support using html
      "content": "", // Content. Support using html
      "active": false // Whether activate current panel.
      // New in Amaze UI 2.3
      "disabled": null // Whether diable current panel.
    }
  ]
}
```

## Attention:

- **Don't use `padding`/`margin`/`border` for `.am-accordion-bd`.**
