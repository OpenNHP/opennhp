(function(undefined) {
  'use strict';

  var registerIfCondHelper = function(hbs) {
    hbs.registerHelper('ifCond', function(v1, operator, v2, options) {
      switch (operator) {
        case '==':
          return (v1 == v2) ? options.fn(this) : options.inverse(this);
          break;
        case '===':
          return (v1 === v2) ? options.fn(this) : options.inverse(this);
          break;
        case '<':
          return (v1 < v2) ? options.fn(this) : options.inverse(this);
          break;
        case '<=':
          return (v1 <= v2) ? options.fn(this) : options.inverse(this);
          break;
        case '>':
          return (v1 > v2) ? options.fn(this) : options.inverse(this);
          break;
        case '>=':
          return (v1 >= v2) ? options.fn(this) : options.inverse(this);
          break;
        default:
          return options.inverse(this);
          break;
      }
      return options.inverse(this);
    });
  };

  if (typeof module !== 'undefined' && module.exports) {
    module.exports = registerIfCondHelper;
  }

  this.Handlebars && registerIfCondHelper(this.Handlebars);
}).call(this);
