'use strict';

var path = require('path');
var fs = require('fs');
var format = require('util').format;
var _ = require('lodash');
var webpack = require('webpack-stream');
var del = require('del');
var bistre = require('bistre');

// Temporary solution until gulp 4
// https://github.com/gulpjs/gulp/issues/355
var runSequence = require('run-sequence');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var pkg = require('./package.json');
var excluded = require('./tools/excluded');
var components = require('./components.json');

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
    compress: {
      warnings: false
    },
    output: {
      /* eslint-disable camelcase */
      ascii_only: true
      /* eslint-enable camelcase */
    }
  }
};
var NODE_ENV = process.env.NODE_ENV;
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

  var excludedWidgets = ['.DS_Store', 'blank', 'layout2', 'layout3', 'layout4',
    'container'].concat(excluded.widgets);
  var allWidgets = fs.readdirSync(WIDGET_DIR).filter(function(widget) {
    return excludedWidgets.indexOf(widget) === -1;
  });

  // 剔除排除配置中不打包的样式
  var excludedStyleDep = [];
  var includedStyleDep = [];
  var getStyleDep = function(type, plugin) {
    var basename = path.basename(plugin, '.js');

    if (components.js[basename]) {
      components.js[basename].depStyle.forEach(function(dep) {
        if (dep.indexOf('ui.') > -1) {
          type === 'excluded' ?
            excludedStyleDep.push(dep) : includedStyleDep.push(dep);
        }
      });
    }
  };

  plugins = _.union(config.js.base, fs.readdirSync('./js'));

  jsEntry += '\'use strict\';\n\n' + 'var $ = require(\'jquery\');\n';

  plugins.forEach(function(plugin) {
    var basename = path.basename(plugin, '.js');

    if (basename !== 'amazeui' && basename !== 'amazeui.legacy' &&
      (excluded.plugins.indexOf(basename) === -1)) {
      jsEntry += (basename === 'core' ? 'var UI = ' : '') +
        'require("./' + basename + '");\n';

      getStyleDep('included', plugin);
    }
  });
  excluded.plugins.forEach(function(plugin) {
    getStyleDep('excluded', plugin);
  });

  // widgets partial
  var partials = '(function(undefined){\n';
  partials += '  \'use strict\';\n\n';
  partials += '  var registerAMUIPartials = function(hbs) {\n';

  // get widgets dependencies
  allWidgets.forEach(function(widget) {
    // read widget package.json
    var pkg = require(path.join(__dirname, WIDGET_DIR, widget, 'package.json'));
    // ./widget/header/src/header
    var srcPath = WIDGET_DIR + widget + '/src/' + widget;

    widgetsStyle += '\r\n// ' + widget + '\r\n';
    widgetsStyle += '@import ".' + srcPath + '.less";' + '\r\n';
    pkg.themes.forEach(function(item) {
      if (!item.hidden && item.name && item.name !== 'one') {
        widgetsStyle += '@import ".' + srcPath + '.' + item.name +
          '.less";' + '\r\n';
      }
    });

    // 将 widget 依赖的样式推入数组
    pkg.styleDependencies.forEach(function(file) {
      if (file.indexOf('ui.') > -1) {
        includedStyleDep.push(file);
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

  // replace excluded style
  includedStyleDep = _.uniq(includedStyleDep);
  excludedStyleDep = _.uniq(excludedStyleDep);

  var intersectionDep = _.intersection(includedStyleDep, excludedStyleDep);
  excludedStyleDep = _.xor(excludedStyleDep, intersectionDep);

  // console.log(includedStyleDep);
  // console.log(excludedStyleDep);

  excludedStyleDep.forEach(function(dep) {
    var regExp = new RegExp('(@import "' + dep + '";)');
    uiBase = uiBase.replace(regExp, '// $1');
  });

  // write less
  fs.writeFileSync('./less/amazeui.less', uiBase + widgetsStyle);
};

gulp.task('build:preparing', preparingData);

gulp.task('build:clean', function() {
  return del([
    config.dist.css,
    config.dist.js
  ]);
});

// Build to dist dir.
gulp.task('build:less', function() {
  gulp.src(config.path.less)
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.plumber({errorHandler: function(err) {
      // 处理编译less错误提示  防止错误之后gulp任务直接中断
      // $.notify.onError({
      //           title:    "编译错误",
      //           message:  "错误信息: <%= error.message %>",
      //           sound:    "Bottle"
      //       })(err);
      console.log(err);
      this.emit('end');
    }}))
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
    .pipe($.replace('//dn-amui.qbox.me/font-awesome/4.5.0/', '../'))
    .pipe(gulp.dest(config.dist.css))
    .pipe($.size({showFiles: true, title: 'source'}))
    // Disable advanced optimizations - selector & property merging, etc.
    // for Issue #19 https://github.com/allmobilize/amazeui/issues/19
    .pipe($.cleanCss({
      advanced: false,
      // @see https://github.com/jakubpawlowicz/clean-css#how-to-set-a-compatibility-mode
      compatibility: 'ie8'
    }))
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

gulp.task('build:js:pack', function() {
  return gulp.src('js/amazeui.js')
    .pipe(webpack({
      watch: !(NODE_ENV === 'travisci' || NODE_ENV === 'production'),
      output: {
        filename: 'amazeui.js',
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
    .pipe($.replace('{{VERSION}}', pkg.version))
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.uglify(config.uglify))
    .pipe($.header(banner, {pkg: pkg, ver: ''}))
    .pipe($.rename({suffix: '.min'}))
    .pipe(gulp.dest(config.dist.js))
    .pipe($.size({showFiles: true, title: 'minified'}))
    .pipe($.size({showFiles: true, gzip: true, title: 'gzipped'}));
});

gulp.task('build:js', function(cb) {
  runSequence(
    ['build:js:pack', 'build:js:fuckie'],
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
    'archive:clean',
    cb
  );
});

gulp.task('archive:copy:css', function() {
  return gulp.src('./dist/css/*.css')
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
