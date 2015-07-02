# Web Widgets
---

[Web Components](http://www.w3.org/TR/components-intro/) is pretty impressive, but is not fully supported by browsers. In this way, Amaze UI Web Widgets are implemented according to Web Components, but using technics more commonly supported by browsers. Implementing some parts that repeatedly used in web development in to widgets, Amaze UI Web Widgets help developers to build websites more effecient.

## Widget Structure

<div>
  <img src="/i/docs/widget.jpg" alt="Widget Structure" style="max-width: 400px" class="am-center"/>
</div>

As shown above, Amaze UI Web Widgets is discribed in a `package.json` file. Each widget consists of template(hbs), styles(LESS) and interaction(JS). Among these three parts, styles and interaction is developed based on [CSS](/css) 、[JS plugins](/javascript). Template is developed using [Handlebars](http://handlebarsjs.com/) as template engine.

Widgets provided by Amaze UI can be find in [Web Widgets](/widgets) page.

## Share Widgets

If you want to share your own Web Widgets, you can [fork our project](https://github.com/allmobilize/amazeui/fork), and develop according to [Development Docs](/getting-started/widget). Don't forget to send our pull request after complete your widget.

Once your widgets pass our test, you will be able to find them on the official website of Amaze UI.

We sincerely welcome everyone to join us, and provide more widgets to help web developers.