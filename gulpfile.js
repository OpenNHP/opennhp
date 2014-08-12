var path = require('path');
var fs = require('fs-extra');

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


var pkg = require('./package.json');
var transportDir = '.build/ts/';
var buildTmpDir = '.build/tmp/';
var jsPaths = {
    widgets: [
        '*/src/*.js',
        '!{powered_by,switch_mode,toolbar,tech_support,layout*,blank,container}/src/*.js']
};

var now = new Date();

var banner = [
    '/*! <%= pkg.name %> - v<%= pkg.version %>',
        '(c) ' + gutil.date(now, 'yyyy') + ' AllMobilize, Inc.',
    '@license <%= pkg.license %>',
        gutil.date(now, 'yyyy-mm-dd HH:mm:ss') + ' */ \r'
].join(' | ');


// build to dist dir

gulp.task('buildLess', function() {
    gulp.src(['./less/amui.all.less'])
        //.pipe(watch())
        .pipe(less({
            paths: [path.join(__dirname, 'less'), path.join(__dirname, 'widget/*/src')]
        }))
        .pipe(gulp.dest('./dist/assets/css'))
        .pipe(minifyCSS())
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
    var seajs = path.join(__dirname,'vendor/seajs/sea.js');
    var seaUse = path.join(__dirname,'/.build/seaUse.js');
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
    gulp.src(['docs/assets/helper/handlebars.js', 'vendor/amazeui.partials.js'])
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


// Rerun the task when a file changes

gulp.task('watch', function() {
    gulp.watch(['js/*.js', 'widget/*/src/*.js'], ['buildJs']);
    gulp.watch(['less/**/*.less', 'widget/*/src/*.less'], ['buildLess']);
    gulp.watch(['dist/amui*js'], ['copyFiles']);
    gulp.watch(['docs/assets/js/main.js'], ['amazeMain']);
});



gulp.task('zip', function () {
    return gulp.src(['./docs/boilerplate/**', './dist/**', '!dist/demo/**/*', '!dist/test/**/*', '!dist/docs/**/*', '!dist/*.zip',
    'docs/examples/blog.html',
    'docs/examples/landing.html',
    'docs/examples/login.html',
    'docs/examples/sidebar.html'
    ])
        .pipe(zip('AmazeUI-1.0.0-beta1.zip', {comment: 'Created on ' + gutil.date(now, 'yyyy-mm-dd HH:mm:ss')}))
        .pipe(gulp.dest('dist'));
});


gulp.task('buildJs', ['copyWidgetJs', 'copyUIJs', 'transport', 'concat', 'clean']);

gulp.task('init', ['bower', 'buildJs', 'hbsHelper', 'buildLess', 'watch']);

gulp.task('default', ['buildJs', 'buildLess', 'hbsHelper', 'watch']);
