'use strict';
var util = require('util');
var path = require('path');
var yeoman = require('yeoman-generator');
var chalk = require('chalk');

var thisTime = new Date();


var AmuiwidgetGenerator = yeoman.generators.Base.extend({
  init: function() {
    this.pkg = yeoman.file.readJSON(path.join(__dirname, '../package.json'));

    this.on('end', function() {
      if (!this.options['skip-install']) {
        //this.npmInstall();
      }
    });
  },

  askFor: function() {
    var done = this.async();

    // have Yeoman greet the user
    console.log(this.yeoman);

    // replace it with a short and sweet description of your generator
    console.log(chalk.magenta('You\'re using the fantastic Amuiwidget generator.'));

    var prompts = [
      {
        name: 'widgetName',
        message: 'What widget name do you want?'
      },
      {
        name: 'version',
        message: 'Widget version?',
        default: '1.0.0'
      },
      {
        name: 'authorName',
        message: 'Author name?'
      },
      {
        name: 'authorEmail',
        message: 'Author email?'
      },
      {
        name: "description",
        message: "Widget description?",
        default: '模块描述'
      }
    ];

    this.prompt(prompts, function(props) {
      this.widgetName = props.widgetName;
      this.version = props.version;
      this.authorName = props.authorName;
      this.authorEmail = props.authorEmail;
      this.description = props.description;

      this.createDate = thisTime.getFullYear() + '.' + (thisTime.getMonth() + 1) + '.' + thisTime.getDate();

      done();
    }.bind(this));
  },

  app: function() {
    var widgetName = this.widgetName;
    var srcPath = widgetName + '/src';

    this.mkdir(widgetName);
    this.mkdir(srcPath);

    this.directory('examples', widgetName + '/examples');

    //
    this.copy('_package.json', widgetName + '/package.json');
    this.copy('_README.md', widgetName + '/README.md');
    this.copy('_HISTORY.md', widgetName + '/HISTORY.md');

    // src
    this.copy('src/icon.png', srcPath + '/' + widgetName + '.png');
    this.copy('src/script.js', srcPath + '/' + widgetName + '.js');
    this.copy('src/style.less', srcPath + '/' + widgetName + '.less');
    this.copy('src/theme.default.less', srcPath + '/' + widgetName + '.default.less');
    this.copy('src/tpl.hbs', srcPath + '/' + widgetName + '.hbs');


    console.log('Directories initialization done!');
  },

  projectfiles: function() {
    // this.copy('editorconfig', '.editorconfig');
    // this.copy('jshintrc', '.jshintrc');
  }
});

module.exports = AmuiwidgetGenerator;
