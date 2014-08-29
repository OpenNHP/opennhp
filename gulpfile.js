var path = require('path');
var fs = require('fs-extra');
var _ = require('lodash');
var format = require('util').format;
var exec = require('child_process').exec;

var gulp = require('gulp');
var gutil = require('gulp-util');
var less = require('gulp-less');
var changed = require('gulp-changed');
var watch = require('gulp-watch');
var concat = require('gulp-concat');
var uglify = require('gulp-uglify');
var minifyCSS = require('gulp-minify-css');
var rename = require('gulp-rename');
var bower = require('gulp-bower');
var transport = require('gulp-cmd-transport');
var header = require('gulp-header');
var clean = require('gulp-clean');
var zip = require('gulp-zip');
var replace = require('gulp-replace');


var pkg = require('./package.json');
var transportDir = '.build/ts/';
var buildTmpDir = '.build/tmp/';
var jsPaths = {
    widgets: [
        '*/src/*.js',
        '!{powered_by,switch_mode,toolbar,tech_support,layout*,blank,container}/src/*.js'],
    hbsHelper: ['vendor/amazeui.hbs.helper.js', 'vendor/amazeui.hbs.partials.js']
};

var now = new Date();

var banner = [
    '/*! <%= pkg.name %> - v<%= pkg.version %>',
    '(c) ' + gutil.date(now, 'yyyy') + ' AllMobilize, Inc.',
    '@license <%= pkg.license %>',
    gutil.date(now, 'yyyy-mm-dd HH:mm:ss') + ' */ \r'
].join(' | ');

// write widgets style and tpl
var getWidgetFiles = function() {
    var fsOptions = {encoding: 'utf8'};
    var uiBase = fs.readFileSync('./less/amui.less', fsOptions);
    var WIDGET_DIR = './widget';
    var rejectWidgets = ['.DS_Store', 'blank', 'layout2', 'layout3', 'layout4', 'container', 'powered_by', 'tech_support', 'toolbar', 'switch_mode'];
    var allWidgets = _.reject(fs.readdirSync(WIDGET_DIR), function(widget) {
        return rejectWidgets.indexOf(widget) > -1;
    });

    var partials = '(function(undefined){\n';
    partials += '  var registerAMUIPartials = function(hbs) {\n';

    _.forEach(allWidgets, function(widget) {
        // read widget package.json
        var pkg = fs.readJsonFileSync(path.join(WIDGET_DIR, widget, 'package.json'));
        var srcPath = '../widget/' + widget + '/src/';

        uiBase += '\r\n// ' + widget + '\r\n';

        uiBase += '@import "' + srcPath + pkg.style + '";' + "\r\n";
        _.forEach(pkg.themes, function(item, index) {
            if (!item.hidden && item.name) {
                uiBase += '@import "' + srcPath + widget + '.' + item.name + '.less";' + "\r\n";
            }
        });

        // read tpl
        var tpl = fs.readFileSync(path.join(WIDGET_DIR, widget, 'src', widget + '.hbs'), fsOptions);
        partials += format('    hbs.registerPartial("%s", %s); \n\n', widget, JSON.stringify(tpl));
    });

    fs.writeFileSync('./less/amui.all.less', uiBase);

    partials += '  }; \n\n';
    partials += '  if (typeof module !== \'undefined\' && module.exports) {\n';
    partials += '    module.exports = registerAMUIPartials;\n' +
    '  }\n\n';
    partials += '  this.Handlebars && registerAMUIPartials(Handlebars);\n';
    partials += '}).call(this);\n';

    // write partials
    fs.writeFileSync(path.join('./vendor/amazeui.hbs.partials.js'), partials);
};


// build to dist dir
gulp.task('buildLess', function() {
    gulp.src(['./less/amui.all.less'])
        .pipe(less({
            paths: [path.join(__dirname, 'less'), path.join(__dirname, 'widget/*/src')]
        }))
        .pipe(gulp.dest('./dist/assets/css'))
        // Disable advanced optimizations - selector & property merging, reduction, etc.
        // for Issue #19 https://github.com/allmobilize/amazeui/issues/19
        .pipe(minifyCSS({noAdvanced: true}))
        .pipe(rename({
            suffix: '.min',
            extname: ".css"
        }))
        .pipe(gulp.dest('./dist/assets/css'));
});


gulp.task('bower', function() {
    bower()
        .pipe(gulp.dest('vendor/'))
});


// copy ui js files to build dir

gulp.task('copyWidgetJs', function() {
    gutil.log(gutil.colors.yellow('Start copy UI js files to build dir....'));
    return gulp.src(jsPaths.widgets, {cwd: './widget'})
        .pipe(rename(function(path) {
            path.dirname = ""; // remove widget dir
        }))
        .pipe(gulp.dest(buildTmpDir));
});


// copy widgets js files to build dir

gulp.task('copyUIJs', ['copyWidgetJs'], function() {
    return gulp.src(['*.js', '!./js/zepto.calendar.js'], {
        cwd: './js'
    })
        .pipe(gulp.dest(buildTmpDir));
});

var modules = [];

// gulp cmd transport
gulp.task('transport', ['copyUIJs'], function() {
    return gulp.src(['*.js'], {cwd: buildTmpDir})
        .pipe(rename(function(path) {
            modules.push(path.basename);
        }))
        .pipe(transport({paths: [buildTmpDir]}))
        .pipe(gulp.dest(transportDir));
});

// concat
gulp.task('concat', ['transport'], function() {
    var seajs = path.join(__dirname, 'vendor/seajs/sea.js');
    var seaUse = path.join(__dirname, '/.build/seaUse.js');
    fs.outputFileSync(seaUse, 'seajs.use(' + JSON.stringify(modules) + ');');

    modules = [];

    //[seajs, '*.js', seaUse]
    return gulp.src([seajs, 'core.js', '*!(core)*.js', seaUse], {cwd: transportDir})
        .pipe(concat('amui.js'))
        .pipe(header(banner, {pkg: pkg}))
        .pipe(gulp.dest('./dist/assets/js'))
        .pipe(uglify({
            mangle: {
                except: ['require']
            },
            preserveComments: 'some'
        }))
        .pipe(rename({
            suffix: '.min',
            extname: ".js"
        }))
        .pipe(gulp.dest('./dist/assets/js'))
});


gulp.task('clean', ['concat'], function() {
    gutil.log(gutil.colors.green('Finished build js, cleaning...'));
    gulp.src('./.build', {read: false})
        .pipe(clean({force: true}));
});


gulp.task('hbsHelper', function() {
    gulp.src(jsPaths.hbsHelper)
        .pipe(concat('amui.widget.helper.js'))
        .pipe(gulp.dest('./dist/assets/js'))
        .pipe(uglify({
            mangle: {
                except: ['require']
            },
            preserveComments: 'some'
        }))
        .pipe(rename({
            suffix: '.min',
            extname: ".js"
        }))
        .pipe(gulp.dest('./dist/assets/js'))
});

gulp.task('widgetsFile', getWidgetFiles);

gulp.task('appServer', function() {
    exec('npm start', function (err, stdout, stderr) {
        console.log(stdout);
        console.log(stderr);
    });
});

// Rerun the task when a file changes

gulp.task('watch', function() {
    gulp.watch(['js/*.js', 'widget/*/src/*.js'], ['buildJs']);
    gulp.watch(['less/**/*.less', 'widget/*/src/*.less'], ['buildLess']);
    gulp.watch(['dist/amui*js'], ['copyFiles']);
    gulp.watch(['docs/assets/js/main.js'], ['amazeMain']);
    gulp.watch(['widget/**/*.json', 'widget/**/*.hbs'], ['widgetsFile']);
    gulp.watch(jsPaths.hbsHelper, ['hbsHelper']);
});

gulp.task('zip', function() {
    return gulp.src(['./dist/**', '!dist/demo/**/*', '!dist/test/**/*', '!dist/docs/**/*', '!dist/*.zip',
        'docs/examples/**/*'
    ])
        .pipe(replace(/\{\{assets\}\}/g, 'assets/', {skipBinary: true}))
        .pipe(zip(format('AmazeUI-1.0.0-beta1-%s.zip', gutil.date(now, 'yyyymmdd')), {comment: 'Created on ' + gutil.date(now, 'yyyy-mm-dd HH:mm:ss')}))
        .pipe(gulp.dest('dist'));
});


gulp.task('buildJs', ['copyWidgetJs', 'copyUIJs', 'transport', 'concat', 'clean']);

// gulp.task('init', ['bower', 'buildJs', 'hbsHelper', 'buildLess', 'watch']);

gulp.task('default', ['widgetsFile', 'buildJs', 'buildLess', 'hbsHelper', 'watch']);

gulp.task('preview', ['widgetsFile', 'buildJs', 'buildLess', 'hbsHelper', 'watch', 'appServer']);
