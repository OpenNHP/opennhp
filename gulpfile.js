/* jshint -W097*/
/* jshint node:true */

'use strict';

var path = require('path');
var fs = require('fs-extra');
var _ = require('lodash');
var format = require('util').format;
var exec = require('child_process').exec;

// Temporary solution until gulp 4
// https://github.com/gulpjs/gulp/issues/355
var runSequence = require('run-sequence');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var pkg = require('./package.json');

var config = {
  path: {
    less: [
      './less/amui.less',
      './less/amazeui.widgets.less',
      './less/amazeui.less'
    ],
    widgets: [
      '*/src/*.js',
      '!{powered_by,switch_mode,toolbar,tech_support,layout*,blank,container}' +
      '/src/*.js'],
    hbsHelper: [
      'vendor/amazeui.hbs.helper.js',
      'vendor/amazeui.hbs.partials.js']
  },
  dir: {
    transport: '.build/ts/',
    buildTmp: '.build/tmp/'
  },
  dist: {
    js: './dist/js',
    css: './dist/css'
  },
  js: {
    base: [
      'core.js',
      'util.fastclick.js',
      'util.hammer.js',
      'zepto.outerdemension.js',
      'zepto.extend.data.js',
      'zepto.extend.fx.js',
      'zepto.extend.selector.js'
    ],
    seajs: path.join(__dirname, 'vendor/seajs/sea.js')
  }
};

var dateFormat = 'UTC:yyyy-mm-dd"T"HH:mm:ss Z';

var banner = [
  '/*! <%= pkg.title %> v<%= pkg.version %><%=ver%>',
  'by Amaze UI Team',
  '(c) ' + $.util.date(Date.now(), 'UTC:yyyy') + ' AllMobilize, Inc.',
  'Licensed under <%= pkg.license.type %>',
  $.util.date(Date.now(), dateFormat) + ' */ \n'
].join(' | ');

var seaUse = '';
var seaUseBasic = '';
var seaUseWidgets = '';
var jsWidgets = [];
var plugins;
var allPlugins;
var pluginsUsed;
var pluginsNotUsed;
var jsAll;
var jsAllSorted;
var jsBasic;
var jsBasicSorted;
var jsWidgetsSorted;

// Write widgets style and tpl
var preparingData = function() {
  var fsOptions = {encoding: 'utf8'};
  var uiBase = fs.readFileSync('./less/amui.less', fsOptions);
  var widgetsStyleDeps = [];
  var widgetsStyle = '';
  var widgetsStyleWithDeps = '';
  var WIDGET_DIR = './widget';
  var rejectWidgets = ['.DS_Store', 'blank', 'layout2', 'layout3', 'layout4',
    'container', 'powered_by', 'tech_support', 'toolbar', 'switch_mode'];
  var allWidgets = _.reject(fs.readdirSync(WIDGET_DIR), function(widget) {
    return rejectWidgets.indexOf(widget) > -1;
  });

  var modules = [];
  var modulesBasic = [];
  var modulesWidgets = [];

  allPlugins = fs.readdirSync('./js');
  plugins = fs.readdirSync('./js');

  var partials = '(function(undefined){\n';
  partials += '  \'use strict\';\n\n';
  partials += '  var registerAMUIPartials = function(hbs) {\n';

  allWidgets.forEach(function(widget, i) {
    // read widget package.json
    var pkg = fs.readJsonFileSync(path.
        join(WIDGET_DIR, widget, 'package.json'));
    var srcPath = '../widget/' + widget + '/src/';

    if (i === 0) {
      widgetsStyleDeps = _.union(widgetsStyleDeps, pkg.styleBase);
    }

    widgetsStyleDeps = _.union(widgetsStyleDeps, pkg.styleDependencies);
    jsWidgets.push(pkg.script);

    jsWidgets = _.union(jsWidgets, pkg.jsDependencies);

    widgetsStyle += '\r\n// ' + widget + '\r\n';

    widgetsStyle += '@import "' + srcPath + pkg.style + '";' + '\r\n';
    _.forEach(pkg.themes, function(item, index) {
      if (!item.hidden && item.name) {
        widgetsStyle += '@import "' + srcPath + widget + '.' +
            item.name + '.less";' + '\r\n';
      }
    });

    // read tpl
    var tpl = fs.readFileSync(path.
        join(WIDGET_DIR, widget, 'src', widget + '.hbs'), fsOptions);
    partials += format('    hbs.registerPartial(\'%s\', %s);\n\n',
        widget, JSON.stringify(tpl));
  });

  widgetsStyleDeps.forEach(function(dep) {
    widgetsStyleWithDeps += format('@import "%s";\n', dep);
  });

  fs.writeFileSync('./less/amazeui.less', uiBase + widgetsStyle);

  fs.writeFileSync('./less/amazeui.widgets.less',
      widgetsStyleWithDeps + widgetsStyle);

  /**
   *  Prepare JavaScript Data
   */

    // for amazeui.basic.js
  jsBasic = _.union(config.js.base, allPlugins);

  // for amazeui.js
  jsAll = _.union(jsBasic, jsWidgets);

  jsWidgets = _.union(config.js.base, jsWidgets);

  pluginsNotUsed = _.difference(plugins, jsWidgets);

  pluginsUsed = _.remove(plugins, function(plugin) {
    return pluginsNotUsed.indexOf(plugin) == -1;
  });

  jsWidgets = _.union(config.js.base, pluginsUsed, jsWidgets);

  // seajs.use[''...]
  jsAll.forEach(function(js) {
    var basename = path.basename(js, '.js');
    modules.push(basename);

    if (jsWidgets.indexOf(js) > -1) {
      modulesWidgets.push(basename);
    }

    if (jsBasic.indexOf(js) > -1) {
      modulesBasic.push(basename);
    }
  });

  seaUse = 'seajs.use(' + JSON.stringify(modules) + ');';
  seaUseBasic = 'seajs.use(' + JSON.stringify(modulesBasic) + ');';
  seaUseWidgets = 'seajs.use(' + JSON.stringify(modulesWidgets) + ');';

  // sort for concat
  jsWidgetsSorted = _.union([config.js.seajs], jsWidgets, [seaUseWidgets]);

  jsAllSorted = _.union([config.js.seajs], jsAll);

  jsBasicSorted = _.union([config.js.seajs], jsBasic, [seaUseBasic]);

  partials += '  };\n\n';
  partials += '  if (typeof module !== \'undefined\' && module.exports) {\n';
  partials += '    module.exports = registerAMUIPartials;\n' +
  '  }\n\n';
  partials += '  this.Handlebars && registerAMUIPartials(Handlebars);\n';
  partials += '}).call(this);\n';

  // write partials
  fs.writeFileSync(path.join('./vendor/amazeui.hbs.partials.js'), partials);
};

gulp.task('build:preparing', preparingData);

gulp.task('bower', function() {
  $.bower().
    pipe(gulp.dest('vendor/'));
});

// Build to dist dir.
gulp.task('build:less', function() {
  gulp.src(config.path.less)
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.less({
      paths: [
        path.join(__dirname, 'less'),
        path.join(__dirname, 'widget/*/src')]
    }))
    .pipe($.rename(function(path) {
      if (path.basename === 'amui') {
        path.basename = pkg.name + '.basic';
      }
    }))
    .pipe(gulp.dest(config.dist.css))
    // Disable advanced optimizations - selector & property merging, etc.
    // for Issue #19 https://github.com/allmobilize/amazeui/issues/19
    .pipe($.minifyCss({noAdvanced: true}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.css'
    }))
    .pipe(gulp.dest(config.dist.css));
});

// Copy ui js files to build dir.
gulp.task('build:js:copy:widgets', function() {
  $.util.log($.util.colors.yellow('Start copy UI js files to build dir....'));
  return gulp.src(config.path.widgets, {cwd: './widget'})
      .pipe($.rename(function(path) {
        path.dirname = ''; // remove widget dir
      }))
      .pipe(gulp.dest(config.dir.buildTmp));
});

// Copy core js files to build dir.
gulp.task('build:js:copy:core', function() {
  return gulp.src('*.js', {
    cwd: './js'
  })
    .pipe(gulp.dest(config.dir.buildTmp));
});

// Transport CDM modules
gulp.task('build:js:transport', function() {
  return gulp.src(['*.js'], {cwd: config.dir.buildTmp})
      .pipe($.cmdTransport({paths: [config.dir.buildTmp]}))
      .pipe(gulp.dest(config.dir.transport));
});

// Concat amazeui.js
gulp.task('build:js:concat:all', function() {
  return gulp.src(jsAllSorted, {cwd: config.dir.transport})
      .pipe($.concat(pkg.name + '.js'))
      .pipe($.header(banner, {pkg: pkg, ver: ''}))
      .pipe($.footer('\n<%=use%>', {use: seaUse}))
      .pipe(gulp.dest(config.dist.js))
      .pipe($.uglify({
        mangle: {
          except: ['require']
        }
      }))
      .pipe($.header(banner, {pkg: pkg, ver: ''}))
      .pipe($.rename({
        suffix: '.min',
        extname: '.js'
      }))
      .pipe(gulp.dest(config.dist.js));
});

// Concat amazeui.basic.js
gulp.task('build:js:concat:basic', function() {
  return gulp.src(jsBasicSorted, {cwd: config.dir.transport})
      .pipe($.concat(pkg.name + '.basic.js'))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ basic'}))
      .pipe($.footer('\n<%=use%>', {use: seaUseBasic}))
      .pipe(gulp.dest(config.dist.js))
      .pipe($.uglify({
        mangle: {
          except: ['require']
        }
      }))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ basic'}))
      .pipe($.rename({
        suffix: '.min',
        extname: '.js'
      }))
      .pipe(gulp.dest(config.dist.js));
});

// Concat amazeui.widgets.js
gulp.task('build:js:concat:widgets', function() {
  return gulp.src(jsWidgetsSorted, {cwd: config.dir.transport})
      .pipe($.concat(pkg.name + '.widgets.js'))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ widgets'}))
      .pipe($.footer('\n<%=use%>', {use: seaUseWidgets}))
      .pipe(gulp.dest(config.dist.js))
      .pipe($.uglify({
        mangle: {
          except: ['require']
        }
      }))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ widgets'}))
      .pipe($.rename({
        suffix: '.min',
        extname: '.js'
      }))
      .pipe(gulp.dest(config.dist.js));
});

gulp.task('build:js:clean', function() {
  $.util.log($.util.colors.green('Finished concat js, cleaning...'));
  gulp.src('./.build', {read: false})
      .pipe($.clean({force: true}));
});

gulp.task('build:js:helper', function() {
  gulp.src(config.path.hbsHelper)
      .pipe($.concat(pkg.name + '.widgets.helper.js'))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ helper'}))
      .pipe(gulp.dest(config.dist.js))
      .pipe($.uglify({
        mangle: {
          except: ['require']
        }
      }))
      .pipe($.header(banner, {pkg: pkg, ver: ' ~ helper'}))
      .pipe($.rename({
        suffix: '.min',
        extname: '.js'
      }))
      .pipe(gulp.dest(config.dist.js));
});

gulp.task('build:js', function(cb) {
  runSequence(
    ['build:js:copy:widgets', 'build:js:copy:core'],
    'build:js:transport',
    ['build:js:concat:all', 'build:js:concat:basic', 'build:js:concat:widgets'],
    ['build:js:clean', 'build:js:helper'],
    cb
  );
});

gulp.task('build', function(cb) {
  runSequence('build:preparing', ['build:less', 'build:js'], cb);
});

// Rerun the task when a file changes
gulp.task('watch', function() {
  gulp.watch(['js/*.js', 'widget/*/src/*.js'], ['build:js']);
  gulp.watch(['less/**/*.less', 'widget/*/src/*.less'], ['build:less']);
  gulp.watch(['widget/**/*.json', 'widget/**/*.hbs'], ['build:preparing']);
  gulp.watch(config.path.hbsHelper, ['build:js:helper']);
});

// Task: Make archive
gulp.task('archive', function(cb) {
  runSequence(['build',
      'archive:copy:css', 'archive:copy:js'],
      'archive:zip',
      'archive:clean',
    cb);
});

gulp.task('archive:copy:css', function() {
  return gulp.src('./dist/css/*.css')
      .pipe(gulp.dest('./docs/examples/assets/css'));
});

gulp.task('archive:copy:js', function() {
  return gulp.src([
    './dist/js/*.js',
    './vendor/handlebars/handlebars.min.js',
    './vendor/zepto/zepto.min.js'])
      .pipe(gulp.dest('./docs/examples/assets/js'));
});

gulp.task('archive:zip', function() {
  return gulp.src(['docs/examples/**/*'])
      .pipe($.replace(/\{\{assets\}\}/g, 'assets/', {skipBinary: true}))
      .pipe($.zip(format('AmazeUI-%s-%s.zip',
          pkg.version, $.util.date(Date.now(),'UTC:yyyymmdd')),
          {comment: 'Created on ' + $.util.date(Date.now(), dateFormat)}))
      .pipe(gulp.dest('dist'));
});

gulp.task('archive:clean', function() {
  return gulp.src(['docs/examples/assets/*/amazeui.*',
    './docs/examples/assets/js/handlebars.min.js',
    './docs/examples/assets/js/zepto.min.js'], {read: false})
      .pipe($.clean({force: true}));
});

// Preview server.
gulp.task('appServer', function() {
  exec('npm start', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
  });
});

// gulp.task('init', ['bower', 'build', 'watch']);

gulp.task('default', ['build', 'watch']);

gulp.task('preview', ['build', 'watch', 'appServer']);
