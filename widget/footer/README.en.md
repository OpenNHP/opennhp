# Footer
---

This widget provide a footer for displaying informations such as versions and Copy rights.

## Usage

### Copy and Paste

- Copy the codes in examples, and paste it to the `<body>` of the Amaze UI HTML template([Download](/getting-started));
- Replace the contents in examples with your own contents.

### Using Handlebars

The Handlebars partial of this widget is `footer`. See [Accordion](/widgets/accordion) for more details.

### Allmobilize WebIDE

- Drag the widget to the edit interface;
- Click the `Data binding` button on the right panel, and bind data using following format.

```javascript
var data = {
  "lang": context.__lang,     // This value can be determined autometically.
  "owner": "",                // Name of website, company or person.
  "companyInfo": [
    {                       // Website information
      "detail": ""        // The detail information of this website. Support using HTML, such as an "a" to create a link to some other page. This will be shown in footer.
    },
    {
      "detail": ""        // Same as the detail above.
    }
  ]
};
return data;
```

__Explanation：__

- `switchName`：Can be "Mobile Version", "Desktop Version". Default value is `"Allmobilize Version"`;
- `owner`：Name of website.
- `slogan`：Your own popup slogan.
- `companyInfo`：Detail information. This is an array, and the each piece of information should be assigned to an element.

### Add to Homescreen

- Only supported by iOS currently. Require setting icon in header.

Reference:

- [Safari Web Content Guide: Configuring Web Applications](https://developer.apple.com/library/ios/documentation/AppleApplications/Reference/SafariWebContent/ConfiguringWebApplications/ConfiguringWebApplications.html)
- [iOS Human Interface Guidelines: Icon and Image Sizes](https://developer.apple.com/library/ios/documentation/UserExperience/Conceptual/MobileHIG/IconMatrix.html#//apple_ref/doc/uid/TP40006556-CH27-SW1)
- [Add to Homescreen - Google Chrome Mobile -- Google Developers](https://developers.google.com/chrome/mobile/docs/installtohomescreen)
- [Everything you always wanted to know about touch icons](http://mathiasbynens.be/notes/touch-icons)

The add to homescreen is defaultly enabled, and can be disabled by using:

```javascript
window.AMUI_NO_ADD2HS = true;
```

## Data Interface

```javascript
{
  "id": "",

  "className": "",

  "theme": "default",

  "options": {
     "modal": "",
     "techSupprtCo": "",
     "techSupprtNet": "",
     "textPosition": ""
  },

  "content": {
    "lang": "",
    "switchName": "",
    "owner": "",
    "slogan": "",
    "companyInfo": [
      {
        "detail": ""
      }
    ]
  }
}
```
