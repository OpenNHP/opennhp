# Pagination
---

___This widget need to be improved in future version!__

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `pagination`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "prevTitle": "Prev",  //(Optional) The texts shown in prev button
  "prevLink": "#",        //(Optional) Href of prev link

  "nextTitle": "Next",  //(Optional) The texts shown in next button
  "nextLink": "#",        //(Optional) Href of next link

  "firstTitle": "First", //(Optional) The texts shown in first page button
  "firstLink": "#",       //(Optional) Href of first link

  "lastTitle": "Last",  //(Optional) The texts shown in last page button
  "lastLink": "#",        //(Optional) Href of last link

  "total": "",            // (Optional) Show the total page number. Only available in "select" theme. Only show current page number if assign an empty string to total.

  "page": [
    {
      "title": "1",
      "link": "#"     //Link to first page.
    },
    {
      "title": "2",
      "link": "#"
    }
  ]
};

return data;
```

## 数据结构

```javascript
{
  "id": "",

  // Customized Class
  "className": "",

  // Theme
  "theme": "default",

  "options": {
    "select": ""
  },

  "content": {
    // Previous Page
    "prevTitle": "上一页",
    "prevLink": "#",

    // First Page
    "firstTitle": "第一页",
    "firstLink": "#",

    // Optional
    "nextTitle": "下一页",
    "nextLink": "#",

    // Optional
    "lastTitle": "最末页",
    "lastLink": "#",

    "total": "", // Total page number

    "page": [
      {
        "title": "1",
        "link": "#",
        "className": ""
      }
    ]
  }
}
```
