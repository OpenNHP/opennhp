'use strict';

var express = require('express');
var path = require('path');
var logger = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var hbs = require('hbs');
var errorHandler = require('errorhandler');

// var amuiHelper = require('amui-hbs-helper')(hbs);

var rootDir = path.join(__dirname, '../../');
var appDir = path.join(rootDir, 'tools', 'app');

var routes = require('./routes/index');

var app = express();

// view engine setup
app.set('views', path.join(appDir, 'views'));
app.set('view engine', 'hbs');

app.use(logger('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({extended: true}));
app.use(cookieParser());
app.use(require('less-middleware')(path.join(appDir, 'public'),
  {
    debug: true,
    parser: {
      paths: ['./']
    }
  }
));

app.use(express.static(path.join(appDir, 'public')));

// ./dist
app.use(express.static(path.join(rootDir, 'dist')));
app.use(express.static(path.join(rootDir, 'widget')));
app.use(express.static(path.join(rootDir, 'node_modules')));

app.use('/', routes);

// catch 404 and forward to error handler
/*app.use(function(req, res, next) {
 var err = new Error('Not Found');
 err.status = 404;
 next(err);
 });*/

// error handlers

// development error handler
// will print stacktrace
if (app.get('env') === 'development') {
  app.use(errorHandler());
  app.use(function(err, req, res, next) {
    res.status(err.status || 500);
    res.render('error', {
      message: err.message,
      error: err
    });
  });
}

// production error handler
// no stacktraces leaked to user
app.use(function(err, req, res, next) {
  res.status(err.status || 500);
  res.render('error', {
    message: err.message,
    error: {}
  });
});

app.set('port', process.env.PORT || 3008);

var server = app.listen(app.get('port'), function() {
  console.log('Amaze UI preview server listening on port ' + server.address().port);
});
