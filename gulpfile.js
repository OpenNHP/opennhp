/* jshint -W097*/
/* jshint node:true */

'use strict';

var path = require('path');
var fs = require('fs-extra');
var _ = require('lodash');
var format = require('util').format;
var browserify = require('browserify');
var watchify = require('watchify');
var collapser = require('bundle-collapser/plugin');
var derequire = require('derequire/plugin');
var del = require('del');
var bistre = require('bistre');
var source = require('vinyl-source-stream');
var buffer = require('vinyl-buffer');

// Temporary solution until gulp 4
// https://github.com/gulpjs/gulp/issues/355
var runSequence = require('run-sequence');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var pkg = require('./package.json');

var config = {
  path: {
    less: [
      './less/amazeui.less',
      './less/themes/flat/amazeui.flat.less'
    ],
    fonts: './fonts/*',
    widgets: [
      '*/src/*.js',
      '!{layout*,blank,container}' +
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
  ],
  uglify: {
    output: {
      ascii_only: true
    }
  }
};

var dateFormat = 'isoDateTime';

var banner = [
  '/*! <%= pkg.title %> v<%= pkg.version %><%=ver%>',
  'by Amaze UI Team',
  '(c) ' + $.util.date(Date.now(), 'UTC:yyyy') + ' AllMobilize, Inc.',
  'Licensed under <%= pkg.license %>',
  $.util.date(Date.now(), dateFormat) + ' */ \n'
].join(' | ');

var jsEntry;
var plugins;
var WIDGET_DIR = './widget/';

// Write widgets style and tpl
var preparingData = function() {
  jsEntry = ''; // empty string
  var fsOptions = {encoding: 'utf8'};

  // less
  var uiBase = fs.readFileSync('./less/amui.less', fsOptions);
  var widgetsStyle = '';

  var rejectWidgets = ['.DS_Store', 'blank', 'layout2', 'layout3', 'layout4',
    'container'];
  var allWidgets = _.reject(fs.readdirSync(WIDGET_DIR), function(widget) {
    return rejectWidgets.indexOf(widget) > -1;
  });

  plugins = _.union(config.js.base, fs.readdirSync('./js'));

  jsEntry += '\'use strict\';\n\n' + 'var $ = require(\'jquery\');\n';

  plugins.forEach(function(plugin, i) {
    var basename = path.basename(plugin, '.js');

    if (basename !== 'amazeui' && basename !== 'amazeui.legacy') {
      jsEntry += (basename === 'core' ? 'var UI = ' : '') +
        'require("./' + basename + '");\n';
    }
  });

  // widgets partial
  var partials = '(function(undefined){\n';
  partials += '  \'use strict\';\n\n';
  partials += '  var registerAMUIPartials = function(hbs) {\n';

  // get widgets dependencies
  allWidgets.forEach(function(widget, i) {
    // read widget package.json
    var pkg = fs.readJsonFileSync(path.
      join(WIDGET_DIR, widget, 'package.json'));
    // ./widget/header/src/header
    var srcPath = WIDGET_DIR + widget + '/src/' + widget;

    widgetsStyle += '\r\n// ' + widget + '\r\n';
    widgetsStyle += '@import ".' + srcPath + '.less";' + '\r\n';
    pkg.themes.forEach(function(item, index) {
      if (!item.hidden && item.name) {
        widgetsStyle += '@import ".' + srcPath + '.' + item.name +
          '.less";' + '\r\n';
      }
    });

    // add to entry
    jsEntry += 'require(".' + srcPath + '.js");\n';

    // read tpl
    var hbs = fs.readFileSync(path.join(srcPath + '.hbs'), fsOptions);
    partials += format('    hbs.registerPartial(\'%s\', %s);\n\n',
      widget, JSON.stringify(hbs));
  });

  // end jsEntry
  jsEntry += '\nmodule.exports = $.AMUI = UI;\n';
  fs.writeFileSync(path.join('./js/amazeui.js'), jsEntry);

  partials += '  };\n\n';
  partials += '  if (typeof module !== \'undefined\' && module.exports) {\n';
  partials += '    module.exports = registerAMUIPartials;\n' +
  '  }\n\n';
  partials += '  this.Handlebars && registerAMUIPartials(this.Handlebars);\n';
  partials += '}).call(this);\n';

  // write partials
  fs.writeFileSync(path.join('./vendor/amazeui.hbs.partials.js'), partials);

  // write less
  fs.writeFileSync('./less/amazeui.less', uiBase + widgetsStyle);
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
    .pipe($.replace('//dn-staticfile.qbox.me/font-awesome/4.3.0/', '../'))
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

var bundleInit = function() {
  var b = browserify(_.assign({}, watchify.args, {
    entries: './js/amazeui.js',
    basedir: __dirname,
    standalone: 'AMUI',
    paths: ['./js']
  }));

  if (process.env.NODE_ENV !== 'travisci') {
    b = watchify(b);
    b.on('update', function() {
      bundle(b);
    });
  }

  b.plugin(derequire);
  b.plugin(collapser);
  b.on('log', $.util.log);
  bundle(b);
};

var bundle = function(b) {
  return b.bundle()
    .on('error', $.util.log.bind($.util, 'Browserify Error'))
    .pipe(source('amazeui.js'))
    .pipe(buffer())
    .pipe($.replace('{{VERSION}}', pkg.version))
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify(config.uglify))
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.rename({suffix: '.min'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
};

gulp.task('build:js:browserify', bundleInit);

gulp.task('build:js:fuckie', function() {
  return gulp.src('vendor/polyfill/*.js')
    .pipe($.concat('amazeui.ie8polyfill.js'))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ IE8 Fucker'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify(config.uglify))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ IE8 Fucker'}))
    .pipe($.rename({suffix: '.min'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('build:js:helper', function() {
  gulp.src(config.path.hbsHelper)
    .pipe($.concat(pkg.name + '.widgets.helper.js'))
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ Handlebars helper'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify())
    .pipe($.header(banner, {pkg: pkg, ver: ' ~ Handlebars helper'}))
    .pipe($.rename({suffix: '.min'}))
    .pipe(gulp.dest(config.dist.js));
});

gulp.task('build:js', function(cb) {
  runSequence(
    ['build:js:browserify', 'build:js:fuckie'],
    ['build:js:helper'],
    cb);
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
  // gulp.watch(['js/*.js', 'widget/*/src/*.js'], ['build:js']);
  gulp.watch(['less/**/*.less', 'widget/*/src/*.less'], ['build:less']);
  gulp.watch(config.path.hbsHelper, ['build:js:helper']);
});

// Task: Make archive
gulp.task('archive', function(cb) {
  runSequence([
      'archive:copy:css',
      'archive:copy:fonts',
      'archive:copy:js'
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
    './node_modules/jquery/dist/jquery.min.js'])
    .pipe($.replace('\n//# sourceMappingURL=jquery.min.map', ''))
    .pipe(gulp.dest('./docs/examples/assets/js'));
});

gulp.task('archive:zip', function() {
  return gulp.src(['docs/examples/**/*'])
    .pipe($.replace(/\{\{assets\}\}/g, 'assets/', {skipBinary: true}))
    .pipe($.zip(format('AmazeUI-%s.zip', pkg.version)))
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
