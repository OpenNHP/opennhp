<h1><a href="http://amazeui.org/" title="Amaze UI 官网"><img style="float: left" width="240" src="https://raw.githubusercontent.com/allmobilize/amazeui/master/vendor/amazeui/amazeui-b.png" alt="Amaze UI Logo"/></a></h1>
- [Homepage in English](http://translate.google.com/translate?hl=en&sl=zh-CN&u=http://amazeui.org/&prev=search) translated by Google
- [Homepage in English](http://www.microsofttranslator.com/bv.aspx?from=zh-CHS&to=en&a=http%3A%2F%2Famazeui.org%2F) translated by Bing

Amaze UI is an open-sourced responsive front-end framework.

## The Product

### Mobile First

Amaze UI believes in Mobile First. Beginning with mobile phone screens, Amaze UI is extending the adaption to larger screens such as tablet and PC.


### Web Components

Amaze UI contains almost 20 CSS components, 10 JS components and 17 other web components which include 60 different themes. All these components are designed to help you develop more efficiently and create responsive pages with awesome user interface and excellent user experience.


### Localization

Compared with existing front-end frameworks, Amaze UI focuses on optimizing page layout in Chinese by adjusting font to different operating systems automatically. Amaze UI provides better compatibility for currently popular browsers and browsers built in Apps, saving you a lot of time for compatibility debugging.


### Light yet Powerful

Amaze UI puts a lot of efforts on performance. Using CSS 3 for animation makes it more interactive, smooth, efficient and suitable for mobile devices, and allows your web app to load fast.


## Download

Users can download packed templates from the Amaze UI official site.

All documents are saved in the directory of `docs/`. To view the demo more conveniently, we suggest you check the documents by visiting [the official website](http://amazeui.org/).


## Develop

Developers can build extensions on top of Amaze UI.

### Project Structure

```
amazeui
|-- HISTORY.md
|-- LICENSE
|-- README.md
|-- package.json
|-- dist        # Contains all compiled and minified CSS and JavaScript files
|-- docs        # Contains all documentation
|-- fonts       # Icon font, using http://staticfile.org/
|-- gulpfile.js # Gulp config
|-- js          # JavaScript scource
|-- less        # LESS scource
|-- tools       # Related tools
|-- vendor      # Contains external libraries, like Zepto and others that are used by Amaze UI
|   |-- amazeui.partials.js
|   |-- handlebars
|   |-- json.format.js
|   |-- seajs
|   `-- zepto
`-- widget      # Web components
```

### Building

Amaze UI use [gulp.js](http://gulpjs.com/) to build files。

The following shows the steps:

```
npm install -g gulp

git clone https://github.com/allmobilize/amazeui.git

npm install

gulp
```

## Bug feedback & Requests

### Bug feedback

You are welcome to [submit bug report](https://github.com/allmobilize/amazeui/issues) to the Amaze UI team.

To explain your problems clearly, we suggest that you provide a demonstration when you give us feedback.

The following links are pages we have built with online debugging tools, using Amaze UI layouts and scripts. You can fork and send a bug report, linking to example pages.

- [Debug Amaze UI 1.0 in JSBin](http://jsbin.com/qasoxibuje/1/edit?html,output)

### Submit Request

User can submit your requests through Issue system or leave us message on our official website. Any request that match our product concepts will be considered.


## Code Contribution

You are welcome to join our debugging team! You are also very welcome to share the Web components you explored by “Fork” this item and submit request afterwards.


__Development Document__

All the development documents are saved in the directory of `docs/rules`. You can also check those documents on our [official website](http://amazeui.org/).

- [Amaze UI HTML/CSS Specifications](http://amazeui.org/getting-started/html-css)
- [Amaze UI JavaScript Specifications](http://amazeui.org/getting-started/javascript)
- [Amaze UI Web Components Specifications](http://amazeui.org/getting-started/widget)

## Referenced & Used Open-source Projects

* [Zepto.js](https://github.com/madrobby/zepto) ([MIT
License](https://github.com/madrobby/zepto/blob/master/MIT-LICENSE))
* [Sea.js](https://github.com/seajs/seajs) ([MIT License](https://github.com/seajs/seajs/blob/master/LICENSE.md))
* [Handlebars.js](https://github.com/wycats/handlebars.js) ([MIT
License](https://github.com/wycats/handlebars.js/blob/master/LICENSE))
* [normalize.css](https://github.com/necolas/normalize.css) ([MIT
License](https://github.com/necolas/normalize.css/blob/master/LICENSE.md))
* [FontAwesome](https://github.com/FortAwesome/Font-Awesome/) ([CC BY 3.0 License](http://creativecommons.org/licenses/by/3.0/))
* [Bootstrap](https://github.com/twbs/bootstrap) ([MIT License](https://github.com/twbs/bootstrap/blob/master/LICENSE))
* [UIkit](https://github.com/uikit/uikit) ([MIT License](https://github.com/uikit/uikit/blob/master/LICENSE.md))
* [Foundation](https://github.com/zurb/foundation) ([MIT
License](https://github.com/zurb/foundation/blob/master/LICENSE))
* [Framework7](https://github.com/nolimits4web/Framework7) ([MIT
License](https://github.com/nolimits4web/Framework7/blob/master/LICENSE))
* [Alice](https://github.com/aliceui/aliceui.org/) ([MIT
License](https://github.com/aliceui/aliceui.org/blob/master/LICENSE))
* [Arale](https://github.com/aralejs/aralejs.org/) ([MIT
License](https://github.com/aralejs/aralejs.org/blob/master/LICENSE))
* [Pure](https://github.com/yui/pure) ([BSD License](https://github.com/yui/pure/blob/master/LICENSE.md))
* [Semantic UI](https://github.com/Semantic-Org/Semantic-UI) ([MIT
License](https://github.com/Semantic-Org/Semantic-UI/blob/master/LICENSE.md))
* [FastClick](https://github.com/ftlabs/fastclick) ([MIT
License](https://github.com/ftlabs/fastclick/blob/master/LICENSE))
* [screenfull.js](https://github.com/sindresorhus/screenfull.js) ([MIT
License](https://github.com/sindresorhus/screenfull.js/blob/gh-pages/license))
* [FlexSlider](https://github.com/woothemes/FlexSlider) ([GPL 2.0](http://www.gnu.org/licenses/gpl-2.0.html))
* [Hammer.js](https://github.com/hammerjs/hammer.js) ([MIT License](https://github.com/hammerjs/hammer.js/blob/master/LICENSE.md))
* [Flat UI](https://github.com/designmodo/Flat-UI) ([CC BY 3.0 and MIT License](https://github.com/designmodo/Flat-UI#copyright-and-license))
* [store.js](https://github.com/marcuswestin/store.js) ([MIT License](https://github.com/marcuswestin/store.js/blob/master/LICENSE))
* [bootstrap-datepicker.js](http://www.eyecon.ro/bootstrap-datepicker/) ([Apache License 2.0](http://www.eyecon.ro/bootstrap-datepicker/js/bootstrap-datepicker.js))
* [iScroll](http://iscrolljs.com/) ([MIT License](http://iscrolljs.com/#license))

There might be some missing and we will keep updating.

### Developed with Open Source Licensed [WebStorm](http://www.jetbrains.com/webstorm/)

<a href="http://www.jetbrains.com/webstorm/" target="_blank">
<img src="http://ww1.sinaimg.cn/large/005yyi5Jjw1elpp6svs2eg30k004i3ye.gif" width="240" />
</a>
