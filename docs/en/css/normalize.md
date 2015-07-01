# Normalize
---

When styles need to be unified between browsers, [Reset](http://meyerweb.com/eric/tools/css/reset/ ) used to be very popular. 

```css
* {
    margin: 0;
    padding: 0;
    border:0;
}
```

As time goes by, Reset is no longer a cut edge technic. Instead, [normalize.css](https://github.com/necolas/normalize.css) becomes more popular. Normalize.css can keep identifiability when unifying the styles, while reset can't.

We also use normalize.css in Amaze UI, but with some customize:

- Add `-webkit-font-smoothing: antialiased;` to `html`.
- `<hgroup>` has been removed from W3C standard, so we don't recommend you to use it.
- Max width is set to `100%` for `<img>`.
- Margin is set to `0` for `<figure>`.
- Add `vertical-align: top; resize: vertical;` to `<textarea>`.
- Remove `<dfn>`.
- Remove `<h1>`