'use strict';

/* jshint node: true */

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
var replace = require('gulp-replace');

gulp.task('sass:copy', function() {
  return gulp.src('./less/*.less')
    .pipe($.rename(function(path) {
      path.extname = '.scss';
    }))
    .pipe(gulp.dest('./dist/scss'));
});

gulp.task('sass:convert', function() {
  return gulp.src('./dist/scss/*.scss')
    // change less/ dir to scss/ on imports
    .pipe(replace(/\/less\//g, '/scss/'))
    // gradient filter
    .pipe(replace(/filter:\s*e\(\%\("progid:DXImageTransform\.Microsoft\.gradient\(startColorstr='\%d',\s*endColorstr='\%d',\s*GradientType=(\d)\)",\s*argb\((@.+)*\),\s*argb\((@.*)\)\)\)/g, function(all, $1, $2, $3) {
      return format("filter: progid:DXImageTransform.Microsoft.gradient(startColorstr='\#{ie-hex-str(%s)}', endColorstr='\#{ie-hex-str(%s)}', GradientType=%d)", $2, $3, $1);
    }))
    // reset-filter()
    .pipe(replace(/filter:\s*e\(\%\(.+\)\);/g, 'filter: progid:DXImageTransform.Microsoft.gradient(enabled = false);'))
    .pipe(replace(/\.less/g, '.scss'))                                 // change .less extensions to .scss on imports
    .pipe(replace(/@/g, '$'))                                          // convert variables
    .pipe(replace(/ e\(/g, ' unquote('))                               // convert escape function
    // covert mixin
    .pipe(replace(/\.([\w-]*)\s*\((.*)\)\s*\{/g, function(match, $1, $2) {
      return '@mixin ' + $1 + '(' + $2.replace(/;/g, ',') + ') {';
    }))
    // .size() in .square
    .pipe(replace('.size($size; $size);', '@include size($size, $size);'))
    .pipe(replace(/\.clearfix(\(\))?;/g, '@include clearfix;'))
    .pipe(replace(/\.tab-focus(\(\))?;/g, '@include tab-focus;'))
    // media query variables
    .pipe(replace(/[\$|@]media\s*(\$\w+-\w+)\s*\{/g, '@media #{$1} {'))
    // comment empty mixins
    .pipe(replace(/@mixin ([\w\-]*)\s*\((.*)\)\s*\{\s*\}/g, '// @mixin $1($2){}'))
    // hook calls
    .pipe(replace(/\.(hook[a-zA-Z\-\d]+);/g, '// @include $1();'))
    .pipe(replace(/\$(import|media|font-face|page|-ms-viewport|keyframes|-webkit-keyframes)/g, '@$1')) // replace valid '@' statements
    .pipe(replace(/(\$[\w\-]*)\s*:(.*);\n/g, '$1: $2 !default;\n'))    // make variables optional
    .pipe(replace(/\$\{/g, '#{$'))                                      // string literals: from: /~"(.*)"/g, to: '#{"$1"}'
    .pipe(replace(/~("[^"]+")/g, 'unquote($1)'))                       // string literals: for real
    .pipe(gulp.dest('./dist/scss'));
});

gulp.task('sass', function(cb) {
  runSequence('sass:copy', 'sass:convert', cb);
});
