'use strict';

var $ = require('jquery');
var UI = require('./core');
UI.support.geolocation = window.navigator && window.navigator.geolocation;

var geo = UI.support.geolocation;

var Geolocation = function(options) {
  this.options = options || {};
};

Geolocation.MESSAGES = {
  unsupportedBrowser: 'Browser does not support location services',
  permissionDenied: 'You have rejected access to your location',
  positionUnavailable: 'Unable to determine your location',
  timeout: 'Service timeout has been reached'
};

Geolocation.ERROR_CODE = {
  0: 'unsupportedBrowser',
  1: 'permissionDenied',
  2: 'positionUnavailable',
  3: 'timeout'
};

Geolocation.prototype.get = function(options) {
  var _this = this;
  options = $.extend({}, this.options, options);
  var deferred = new $.Deferred();

  if (geo) {
    this.watchID = geo.getCurrentPosition(function(position) {
      deferred.resolve.call(_this, position);
    }, function(error) {
      deferred.reject(Geolocation.MESSAGES[Geolocation.ERROR_CODE[error.code]]);
    }, options);
  } else {
    deferred.reject(Geolocation.MESSAGES.unsupportedBrowser);
  }

  return deferred.promise();
};

Geolocation.prototype.watch = function(options) {
  if (!geo) {
    return;
  }

  options = $.extend({}, this.options, options);

  if (!$.isFunction(options.done)) {
    return;
  }

  this.clearWatch();

  var fail = $.isFunction(options.fail) ? options.fail : null;

  this.watchID = geo.watchPosition(options.done, fail, options);

  return this.watchID;
};

Geolocation.prototype.clearWatch = function() {
  if (!geo || !this.watchID) {
    return;
  }
  geo.clearWatch(this.watchID);
  this.watchID = null;
};

$.AMUI.Geolocation = Geolocation;

module.exports = Geolocation;
