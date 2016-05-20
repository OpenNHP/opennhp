'use strict';

var path = require('path');
var fs = require('fs');
var _ = require('lodash');
var format = require('util').format;
var webpack = require('webpack-stream');
var del = require('del');
var runSequence = require('run-sequence');
var gulp = require('gulp');
var $ = require('gulp-load-plugins')();
var file = require('../lib/file');
var cstmzPath = path.join(__dirname, '../../dist/customized');
var cstmzTmp = path.join(__dirname, '../../.cstmz-tmp');
var DEFAULTS = {
  dist: cstmzPath,
  tmp: cstmzTmp,
  js: path.join(cstmzTmp, 'js/amazeui.custom.js'),
  less: path.join(cstmzTmp, 'less/amazeui.custom.less'),
  AUTOPREFIXER_BROWSERS: [
    'ie >= 8',
    'ie_mob >= 10',
    'ff >= 30',
    'chrome >= 34',
    'safari >= 7',
    'opera >= 23',
    'ios >= 7',
    'android >= 2.3',
    'bb >= 10'
  ],
  widgetBase: [
    'variables.less',
    'mixins.less',
    'base.less',
    'grid.less',
    'block-grid.less',
    'icon.less',
    'utility.less'
  ]
};

var configFile = (path.join(__dirname, './config.json'));

if (!fs.existsSync(configFile)) {
  return gulp.task('customizer', function() {
    throw new Error('config.json is not exists.');
  });
}

var config = require(configFile);
var less = [
  '@import "variables.less";',
  '@import "mixins.less";',
  '@import "base.less";'
];
var js = [
  'require("../../js/core");'
];

gulp.task('customizer:preparing', function(callback) {
  config.style.forEach(function(file) {
    less.push(format('@import "%s";', file));
  });

  config.js.forEach(function(file) {
    js.push(format('require("../../js/%s");', file));
  });

  // widgets
  if (config.widgets) {
    if (config.widgets.length) {
      DEFAULTS.widgetBase.forEach(function(base) {
        less.push(format('@import "%s";', base));
      });
    }

    config.widgets.forEach(function(widget) {
      js.push(format(
        'require("../../widget/%s/src/%s");',
        widget.name,
        widget.name)
      );
      less.push(format('@import "../../widget/%s/src/%s.less";',
        widget.name, widget.name));
      var pkg = require(path.join('../../widget', widget.name, 'package.json'));

      pkg.styleDependencies.forEach(function(dep) {
        less.push(format('@import "%s";', dep));
      });

      if (widget.theme) {
        widget.theme.forEach(function(theme) {
          less.push(format('@import "../../widget/%s/src/%s";', widget.name,
            theme));
        });
      }
    });
  }

  file.write(DEFAULTS.less, _.uniq(less).join('\n'));
  file.write(DEFAULTS.js, _.uniq(js).join('\n'));

  callback();
});

gulp.task('customizer:less', function() {
  return gulp.src(DEFAULTS.less)
    .pipe($.less({
      paths: [
        path.join(__dirname, '../../less')
      ]
    }))
    .pipe($.autoprefixer({
      browsers: config.AUTOPREFIXER_BROWSERS ||
      DEFAULTS.AUTOPREFIXER_BROWSERS
    }))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.size({showFiles: true, title: 'source'}))
    .pipe($.cleanCss({
      advanced: false,
      compatibility: 'ie8'
    }))
    .pipe($.rename({
      suffix: '.min',
      extname: '.css'
    }))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('customizer:js', function() {
  return gulp.src(DEFAULTS.js)
    .pipe(webpack({
      output: {
        filename: 'amazeui.custom.js',
        library: 'AMUI',
        libraryTarget: 'umd'
      },
      externals: [
        {
          jquery: {
            root: 'jQuery',
            commonjs2: 'jquery',
            commonjs: 'jquery',
            amd: 'jquery'
          }
        }
      ]
    }))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.uglify({
      output: {
        ascii_only: true
      }
    }))
    .pipe($.rename({suffix: '.min'}))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('customizer:clean', function() {
  return del(DEFAULTS.tmp);
});

gulp.task('customizer', function(cb) {
  runSequence(
    'customizer:preparing',
    ['customizer:less', 'customizer:js'],
    'customizer:clean',
    cb);
});
