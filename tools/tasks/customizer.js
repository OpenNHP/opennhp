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
var cstPath = path.join(__dirname, '../../dist/customized');
var DEFAULTS = {
  dist: cstPath,
  tmp: path.join(cstPath, '.tmp'),
  js: path.join(cstPath, '.tmp/js/amazeui.custom.js'),
  less: path.join(cstPath, '.tmp/less/amazeui.custom.less'),
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
  ]
};

var configFile = (path.join(__dirname, './config.json'));

if (!fs.existsSync(configFile)) {
  throw new Error('config.json is not exists.');
}

var config = require(configFile);
var less = [];
var js = [];

config.style.forEach(function(file) {
  less.push(format('@import "%s";', file));
});

config.js.forEach(function(file) {
  js.push(format('require("./%s");', file));
});

file.write(DEFAULTS.less, less.join('\n'));
file.write(DEFAULTS.js, js.join('\n'));

gulp.task('customizer', function(cb) {
  runSequence(['customizer:less', 'customizer:js'],
    'customizer:clean',
    cb);
});

gulp.task('customizer:less', function() {
  gulp.src(DEFAULTS.less)
    .pipe($.less({
      paths: [
        path.join(__dirname, '../../less'),
        path.join(__dirname, '../../widget/header/src')]
    }))
    .pipe($.autoprefixer({
      browsers: config.AUTOPREFIXER_BROWSERS ||
      DEFAULTS.AUTOPREFIXER_BROWSERS}))
    .pipe(gulp.dest(cstPath))
    .pipe($.size({showFiles: true, title: 'source'}))
    .pipe($.minifyCss({noAdvanced: true}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.css'
    }))
    .pipe(gulp.dest(cstPath))
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
    .pipe(gulp.dest(cstPath))
    .pipe($.uglify())
    .pipe($.rename({
      suffix: '.min',
      extname: '.js'
    }))
    .pipe(gulp.dest(cstPath))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('customizer:clean', function(cb) {
  del(DEFAULTS.tmp, cb);
});
