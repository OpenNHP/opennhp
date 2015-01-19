/* jshint -W097*/
/* jshint node:true */

'use strict';

var path = require('path');
var fs = require('fs-extra');
var _ = require('lodash');
var format = require('util').format;
var browserify = require('browserify');
var transform = require('vinyl-transform');
var del = require('del');
var bistre = require('bistre');

// Temporary solution until gulp 4
// https://github.com/gulpjs/gulp/issues/355
var runSequence = require('run-sequence');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var pkg = require('./package.json');

var config = {
  path: {
    less: [
      // './less/amui.less',
      // './less/amazeui.widgets.less',
      './less/amazeui.less',
      './less/themes/flat/amazeui.flat.less'
    ],
    fonts: './fonts/*',
    widgets: [
      '*/src/*.js',
      '!{powered_by,switch_mode,toolbar,tech_support,layout*,blank,container}' +
      '/src/*.js'],
    hbsHelper: [
      'vendor/amazeui.hbs.helper.js',
      'vendor/amazeui.hbs.partials.js'],
    buildTmp: '.build/temp/'
  },
  dist: {
    js: './dist/js',
    css: './dist/css',
    fonts: './dist/fonts'
  },
  js: {
    base: [
      'core.js',
      'util.fastclick.js',
      'util.hammer.js'
    ]
  },

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

var dateFormat = 'UTC:yyyy-mm-dd"T"HH:mm:ss Z';

var banner = [
  '/*! <%= pkg.title %> v<%= pkg.version %><%=ver%>',
  'by Amaze UI Team',
  '(c) ' + $.util.date(Date.now(), 'UTC:yyyy') + ' AllMobilize, Inc.',
  'Licensed under <%= pkg.license.type %>',
  $.util.date(Date.now(), dateFormat) + ' */ \n'
].join(' | ');

var initAll = '\'use strict\';\n\n' + 'var $ = require(\'jquery\');\n\n';
var initBasic = '';
var initWidgets = '';
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

    if (basename !== 'amazeui' || basename !== 'amazeui.legacy') {
      initAll += 'require(\'./' + basename + '\');\n';
    }

    if (jsWidgets.indexOf(js) > -1) {
      modulesWidgets.push(basename);
    }

    if (jsBasic.indexOf(js) > -1) {
      modulesBasic.push(basename);
    }
  });

  initAll += '\nmodule.exports = $.AMUI;\n';

  // initAll = 'require(' + JSON.stringify(modules) + ');';
  initBasic = 'require(' + JSON.stringify(modulesBasic) + ');';
  initWidgets = 'require(' + JSON.stringify(modulesWidgets) + ');';

  // sort for concat
  jsWidgetsSorted = _.union(jsWidgets, [initWidgets]);

  jsAllSorted = _.union(jsAll);

  jsBasicSorted = _.union(jsBasic, [initBasic]);

  partials += '  };\n\n';
  partials += '  if (typeof module !== \'undefined\' && module.exports) {\n';
  partials += '    module.exports = registerAMUIPartials;\n' +
  '  }\n\n';
  partials += '  this.Handlebars && registerAMUIPartials(this.Handlebars);\n';
  partials += '}).call(this);\n';

  // write partials
  fs.writeFileSync(path.join('./vendor/amazeui.hbs.partials.js'), partials);
  fs.writeFileSync(path.join('./js/amazeui.js'), initAll);
};

gulp.task('build:preparing', preparingData);

gulp.task('build:clean', function(cb) {
  del([
    config.dist.css,
    config.dist.js
  ], cb);
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
    .pipe($.autoprefixer({browsers: config.AUTOPREFIXER_BROWSERS}))
    .pipe($.replace('//dn-staticfile.qbox.me/font-awesome/4.2.0/', '../'))
    .pipe(gulp.dest(config.dist.css))
    .pipe($.size({showFiles: true, title: 'source'}))
    // Disable advanced optimizations - selector & property merging, etc.
    // for Issue #19 https://github.com/allmobilize/amazeui/issues/19
    .pipe($.minifyCss({noAdvanced: true}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.css'
    }))
    .pipe(gulp.dest(config.dist.css))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('build:fonts', function() {
  gulp.src(config.path.fonts)
    .pipe(gulp.dest(config.dist.fonts));
});

// Copy ui js files to build dir.
gulp.task('build:js:copy:widgets', function() {
  $.util.log($.util.colors.yellow('Start copy widgets js to build dir....'));
  return gulp.src(config.path.widgets, {cwd: './widget'})
    .pipe($.rename(function(path) {
      path.dirname = ''; // remove widget dir
    }))
    .pipe(gulp.dest(config.path.buildTmp));
});

// Copy core js files to build dir.
gulp.task('build:js:copy:core', function() {
  return gulp.src('*.js', {
    cwd: './js'
  })
    .pipe(gulp.dest(config.path.buildTmp));
});

gulp.task('build:js:browserify', function() {
  var bundler = transform(function(filename) {
    var b = browserify({
      entries: filename,
      basedir: path.join(__dirname, config.path.buildTmp)
    });
    return b.bundle();
  });

  return gulp.src(pkg.name + '.js',
    {cwd: path.join(__dirname, config.path.buildTmp)})
    .pipe(bundler)
    .pipe(gulp.dest(config.dist.js))
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.jsbeautifier({config: '.jsbeautifyrc'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify())
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.js'
    }))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('build:js:fuckie', function() {
  var bundler = transform(function(filename) {
    var b = browserify({
      entries: filename,
      basedir: path.join(__dirname, config.path.buildTmp)
    });
    return b.bundle();
  });

  return gulp.src('amazeui.legacy.js',
    {cwd: path.join(__dirname, config.path.buildTmp)})
    .pipe(bundler)
    .pipe(gulp.dest(config.dist.js))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ Old IE Fucker'}))
    .pipe($.jsbeautifier({config: '.jsbeautifyrc'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify())
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.js'
    }))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

// Concat AMD
gulp.task('build:js:amd', function() {
  return gulp.src('amazeui.js', {cwd: config.path.buildTmp})
    .pipe($.amdBundler({
      baseDir: config.path.buildTmp
    }))
    .pipe($.concat(pkg.name + '.amd.js'))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ AMD'}))
    .pipe($.jsbeautifier({config: '.jsbeautifyrc'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify())
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ AMD'}))
    .pipe($.rename({
      suffix: '.min',
      extname: '.js'
    }))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('build:js:clean', function(cb) {
  $.util.log($.util.colors.green('Finished concat js, cleaning...'));
  del('./.build', cb);
});

gulp.task('build:js:helper', function() {
  gulp.src(config.path.hbsHelper)
    .pipe($.concat(pkg.name + '.widgets.helper.js'))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ helper'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify())
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
    ['build:js:browserify'],
    ['build:js:fuckie'],
    // ['build:js:amd'],
    ['build:js:clean', 'build:js:helper'],
    cb
  );
});

gulp.task('build', function(cb) {
  runSequence(
    'build:preparing',
    'build:clean',
    ['build:less', 'build:fonts', 'build:js'],
    cb);
});

// Rerun the task when a file changes
gulp.task('watch', function() {
  gulp.watch(['widget/**/*.json', 'widget/**/*.hbs'], ['build:preparing']);
  gulp.watch(['js/*.js', 'widget/*/src/*.js'], ['build:js']);
  gulp.watch(['less/**/*.less', 'widget/*/src/*.less'], ['build:less']);
  gulp.watch(config.path.hbsHelper, ['build:js:helper']);
});

// Task: Make archive
gulp.task('archive', function(cb) {
  runSequence([
      'archive:copy:css',
      'archive:copy:fonts',
      'archive:copy:js',
      'archive:copy:polyfill'
    ],
    'archive:zip',
    'archive:clean', cb);
});

gulp.task('archive:copy:css', function() {
  return gulp.src('./dist/css/*.css')
    .pipe($.replace('//dn-staticfile.qbox.me/font-awesome/4.2.0/', '../'))
    .pipe(gulp.dest('./docs/examples/assets/css'));
});

gulp.task('archive:copy:fonts', function() {
  return gulp.src('./fonts/*')
    .pipe(gulp.dest('./docs/examples/assets/fonts'));
});

gulp.task('archive:copy:js', function() {
  return gulp.src([
    './dist/js/*.js',
    './node_modules/handlebars/dist/handlebars.min.js',
    './node_modules/jquery/dist/cdn/jquery-2.1.1.min.js'])
    .pipe($.rename(function(file) {
      if (file.basename.indexOf('jquery-') > -1) {
        file.basename = 'jquery.min';
      }
    }))
    .pipe(gulp.dest('./docs/examples/assets/js'));
});

gulp.task('archive:copy:polyfill', function() {
  return gulp.src([
    './vendor/polyfill/*.js'])
    .pipe(gulp.dest('./docs/examples/assets/js/polyfill'));
});

gulp.task('archive:zip', function() {
  return gulp.src(['docs/examples/**/*'])
    .pipe($.replace(/\{\{assets\}\}/g, 'assets/', {skipBinary: true}))
    .pipe($.zip(format('AmazeUI-%s-%s.zip',
        pkg.version, $.util.date(Date.now(), 'UTC:yyyymmdd')),
      {comment: 'Created on ' + $.util.date(Date.now(), dateFormat)}))
    .pipe(gulp.dest('dist'));
});

gulp.task('archive:clean', function(cb) {
  del(['docs/examples/assets/*/amazeui.*',
    'docs/examples/assets/js/*',
    'docs/examples/assets/fonts',
    '!docs/examples/assets/js/app.js'
  ], cb);
});

// Preview server.
gulp.task('appServer', function(callback) {
  $.nodemon({
    script: 'tools/app/app.js',
    env: {
      NODE_ENV: 'development'
    },
    stdout: false
  }).on('readable', function() {
    this.stdout
      .pipe(bistre({time: true}))
      .pipe(process.stdout);
    this.stderr
      .pipe(bistre({time: true}))
      .pipe(process.stderr);
  });
  callback();
});

// gulp.task('init', ['bower', 'build', 'watch']);

gulp.task('default', ['build', 'watch']);

gulp.task('preview', ['build', 'watch', 'appServer']);

// tasks
require('./tools/tasks/');

gulp.task('customize', ['customizer']);
