var path = require('path');
var format = require('util').format;
var hbs = require('hbs');
require('../../vendor/amazeui.hbs.helper')(hbs);
var fs = require('fs-extra');
var _ = require('lodash');
var rootDir = path.join(__dirname, '../../');
var widgetDir = path.join(rootDir, 'widget');
var lessDir = path.join(rootDir, 'less');
var distDir = path.join(rootDir, 'dist');

var config = {
    rootDir: rootDir,
    widgetDir: widgetDir
};

var components = {};

var rejectWidgets = ['.DS_Store', 'blank', 'layout2', 'layout3', 'layout4', 'container'];
var allWidgets = _.reject(fs.readdirSync(widgetDir), function(widget) {
    return rejectWidgets.indexOf(widget) > -1;
});

_.forEach(allWidgets, function(widget) {
    if (widget !== '.DS_Store') {

        // read widget package.json
        var pkg = fs.readJsonSync(path.join(widgetDir, widget, 'package.json'));

        var srcDir = path.join(widgetDir, widget, 'src');

        // hbs partials
        hbs.registerPartials(srcDir);

        if (!pkg.hidden && pkg.type !== 'layout') {
            var widgetName = (pkg.localName['en']) ? (pkg.localName['en']) : widget;
            var demos = [];
            var tpl = fs.readFileSync(path.join(srcDir, widget + '.hbs'), 'utf-8');

            _.forEach(pkg.themes, function(theme, index) {
                if (theme.demos.length) {
                    _.forEach(theme.demos, function(data, i) {
                        demos.push({
                            title: format('%s（%s）', theme.name, data.desc || theme.desc),
                            url: format('%s/%s/%d', widget, theme.name, i),
                            data: {
                                theme: theme.name,
                                options: data.data.options || {},
                                content: data.data.content || (pkg.demoContent && pkg.demoContent)
                            }
                        });
                    });
                }
            });

            components[widget] = {
                name: widgetName,
                localName: pkg.localName,
                tpl: tpl,
                demos: demos
            };
        }
    }
});


exports.components = components;
exports.config = config;