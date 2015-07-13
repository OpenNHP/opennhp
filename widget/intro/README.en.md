# Intro
---

Images and texts can be put in this widget to form a introduction of website or company.


## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `intro`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "title": "",  // Title
  "more": { // More link
    "title": "",
    "link": ""
  },
  "left": "", //Content on the left
  "right": "" //Content on the right
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
		"leftCols": "",
		"RightCols": "",
		"position":""
	},

	"content": {
		"title": "",
		"more": {
      "link": "",
      "title": "",
      "className": ""
    },
		"left": "",
		"right": ""
	}
}
```
