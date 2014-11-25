'use strict';

var fs = require('fs');
var path = require('path');

var file = {};

file.mkdir = function(dirpath, mode) {
  // get from grunt.file
  if (fs.existsSync(dirpath)) {
    return;
  }

  if (!mode) {
    mode = parseInt('0777', 8) & (~process.umask());
  }

  dirpath.split(path.sep).reduce(function(parts, part) {
    parts += part + '/';
    var subpath = path.resolve(parts);
    if (!fs.existsSync(subpath)) {
      fs.mkdirSync(subpath, mode);
    }
    return parts;
  }, '');
};

file.read = function(filepath) {
  var content = fs.readFileSync(filepath, {
    encoding: 'utf8'
  });
  if (content.charAt(0) === 0xFEFF) {
    content = content.slice(1);
  }
  return content;
};

file.write = function(filepath, content) {
  file.mkdir(path.dirname(filepath));
  return fs.writeFileSync(filepath, content);
};

module.exports = file;
