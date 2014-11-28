/* jshint node: true */
'use strict';

var path = require('path');
var fs = require('fs');
var _ = require('lodash');
var format = require('util').format;
var browserify = require('browserify');
var transform = require('vinyl-transform');
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
  'require("./core");'
];

config.style.forEach(function(file) {
  less.push(format('@import "%s";', file));
});

config.js.forEach(function(file) {
  js.push(format('require("./%s");', file));
});

// widgets
if (config.widgets) {
  if (config.widgets.length) {
    DEFAULTS.widgetBase.forEach(function(base) {
      less.push(format('@import "%s";', base));
    });
  }

  config.widgets.forEach(function(widget) {
    js.push(format('require("./%s");', widget.name));
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

gulp.task('customizer', function(cb) {
  runSequence(['customizer:less', 'customizer:js'],
    'customizer:clean',
    cb);
});

gulp.task('customizer:less', function() {
  gulp.src(DEFAULTS.less)
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
    .pipe($.minifyCss({noAdvanced: true}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.css'
    }))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('customizer:js', function(cb) {
  runSequence(
    [
      'customizer:js:copy:core',
      'customizer:js:copy:widgets'
    ],
    'customizer:js:browserify',
    cb);
});

// Copy ui js files to build dir.
gulp.task('customizer:js:copy:widgets', function() {
  $.util.log($.util.colors.yellow('Start copy widgets js to build dir....'));
  return gulp.src([
    '*/src/*.js',
    '!{layout*,blank,container}' +
    '/src/*.js'], {cwd: './widget'})
    .pipe($.rename(function(path) {
      path.dirname = ''; // remove widget dir
    }))
    .pipe(gulp.dest(DEFAULTS.tmp + '/js'));
});

// Copy core js files to build dir.
gulp.task('customizer:js:copy:core', function() {
  return gulp.src('*.js', {
    cwd: './js'
  })
    .pipe(gulp.dest(DEFAULTS.tmp + '/js'));
});

gulp.task('customizer:js:browserify', function() {
  var bundler = transform(function(filename) {
    var b = browserify({
      entries: filename,
      basedir: path.join(__dirname, DEFAULTS.tmp, 'js')
    });
    return b.bundle();
  });

  return gulp.src(DEFAULTS.js,
    {cwd: path.join(__dirname, DEFAULTS.tmp, 'js')})
    .pipe(bundler)
    .pipe(gulp.dest(cstmzPath))
    .pipe($.uglify())
    .pipe($.rename({
      suffix: '.min',
      extname: '.js'
    }))
    .pipe(gulp.dest(cstmzPath))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('customizer:clean', function(cb) {
  del(DEFAULTS.tmp, cb);
});
