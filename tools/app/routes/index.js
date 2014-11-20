'use strict';

var express = require('express');
var format = require('util').format;
var hbs = require('hbs');
var lib = require('../lib');
var router = express.Router();
var componentsData = lib.components;

/* GET home page. */
router.get('/', function(req, res) {
  res.render('index', {
    title: 'Amaze UI Components Demo',
    components: componentsData,
    componentsData: JSON.stringify(componentsData),
    header: {
      content: {
        title: 'Amaze UI'
      }
    }
  });
});

router.get('/:component', function(req, res) {
  var component = req.params.component;
  res.redirect('/#' + component);
});

router.get('/:component/:theme', function(req, res) {
  var component = req.params.component;
  res.redirect('/#' + component);
});

router.get('/:component/:theme/:id', function(req, res) {
  var params = req.params;
  var component = params.component;
  var theme = params.theme;
  var id = params.id;
  var demoIndex = -1;
  var reqCpt = componentsData[component];

  if (!reqCpt) {
    return res.status(404).render('error', {
      message: '组件不存在！',
      error: {}
    });
  }

  reqCpt.demos.forEach(function(item, index) {
    if (item.url == format('%s/%s/%s', component, theme, id)) {
      demoIndex = index;
    }
  });

  if (demoIndex < 0) {
    return res.send('数据不存在！');
  }

  var demoData = reqCpt.demos[demoIndex];
  var demoCompiler = hbs.handlebars.compile(reqCpt.tpl);

  res.render('index', {
    demoDetail: 1,
    title: format('%s - %s | Amaze UI Components Demo',
        reqCpt.localName.en, demoData.title),
    components: componentsData,
    componentsData: JSON.stringify(componentsData),
    header: {
      content: {
        left: [{
          link: '/',
          icon: 'home'
        }],
        title: demoData.title
      }
    },
    content: demoCompiler(demoData.data)
  });
});

module.exports = router;
