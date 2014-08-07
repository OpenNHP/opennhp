define(function(require, exports, module) {
    var $ = window.Zepto;

    require('./zepto.extend.fx');
    require('./zepto.extend.selector');
    require('./zepto.extend.data');

    /**
     * @via https://github.com/Semantic-Org/Semantic-UI/blob/master/src/modules/accordion.js
     * @license https://github.com/Semantic-Org/Semantic-UI/blob/master/LICENSE.md
     */

    $.fn.accordion = function(parameters) {
        var $allModules = $(this),

            query = arguments[0],
            methodInvoked = (typeof query == 'string'),
            queryArguments = [].slice.call(arguments, 1),
            returnedValue;

        $allModules.each(function() {
            var settings = ( $.isPlainObject(parameters) )
                    ? $.extend(true, {}, $.fn.accordion.settings, parameters)
                    : $.extend({}, $.fn.accordion.settings),

                className = settings.className,
                namespace = settings.namespace,
                selector = settings.selector,

                eventNamespace = '.' + namespace,
                moduleNamespace = 'module-' + namespace,

                $module = $(this),
                $title = $module.find(selector.title),
                $content = $module.find(selector.content),
                $item = $module.find(selector.item),

                element = this,
                instance = $module.data(moduleNamespace),
                module;

            module = {

                initialize: function() {
                    // initializing
                    $title.on('click' + eventNamespace, module.event.click);
                    module.instantiate();
                },

                instantiate: function() {
                    instance = module;
                    $module.data(moduleNamespace, module);
                },

                destroy: function() {
                    $module.removeData(moduleNamespace);
                    $title.off(eventNamespace);
                },

                event: {
                    click: function() {
                        var $activeTitle = $(this),
                            index = $item.index($activeTitle.parent(selector.item));

                        module.toggle(index);
                    }
                },

                toggle: function(index) {
                    var $activeItem = $item.eq(index),
                        contentIsOpen = $activeItem.hasClass(className.active);

                    if (contentIsOpen) {
                        if (settings.collapsible) {
                            module.close(index);
                        }
                    }
                    else {
                        module.open(index);
                    }
                },

                open: function(index) {
                    var $activeItem = $item.eq(index),
                        $activeContent = $activeItem.next(selector.content);

                    if (!settings.multiple) {
                        $item.not($activeItem).removeClass(className.active);
                    }

                    $activeItem.addClass(className.active);

                    $activeContent.animate(settings.duration, settings.easing, function() {
                        $.proxy(settings.onOpen, $activeContent)();
                        $.proxy(settings.onChange, $activeContent)();
                    });

                },

                close: function(index) {
                    var $activeItem = $item.eq(index),
                        $activeContent = $activeItem.find(selector.content);

                    $activeItem.removeClass(className.active);

                    $activeContent.animate(settings.duration, settings.easing, function() {
                        $.proxy(settings.onClose, $activeContent)();
                        $.proxy(settings.onChange, $activeContent)();
                    });
                },

                setting: function(name, value) {
                    if ($.isPlainObject(name)) {
                        $.extend(true, settings, name);
                    }
                    else if (value !== undefined) {
                        settings[name] = value;
                    }
                    else {
                        return settings[name];
                    }
                },
                internal: function(name, value) {
                    if (value !== undefined) {
                        if ($.isPlainObject(name)) {
                            $.extend(true, module, name);
                        }
                        else {
                            module[name] = value;
                        }
                    }
                    else {
                        return module[name];
                    }
                },

                invoke: function(query, passedArguments, context) {
                    var
                        object = instance,
                        maxDepth,
                        found,
                        response
                        ;
                    passedArguments = passedArguments || queryArguments;
                    context = element || context;
                    if (typeof query == 'string' && object !== undefined) {
                        query = query.split(/[\. ]/);
                        maxDepth = query.length - 1;
                        $.each(query, function(depth, value) {
                            var camelCaseValue = (depth != maxDepth)
                                    ? value + query[depth + 1].charAt(0).toUpperCase() + query[depth + 1].slice(1)
                                    : query
                                ;
                            if ($.isPlainObject(object[camelCaseValue]) && (depth != maxDepth)) {
                                object = object[camelCaseValue];
                            }
                            else if (object[camelCaseValue] !== undefined) {
                                found = object[camelCaseValue];
                                return false;
                            }
                            else if ($.isPlainObject(object[value]) && (depth != maxDepth)) {
                                object = object[value];
                            }
                            else if (object[value] !== undefined) {
                                found = object[value];
                                return false;
                            }
                            else {
                                return false;
                            }
                        });
                    }
                    if ($.isFunction(found)) {
                        response = found.apply(context, passedArguments);
                    }
                    else if (found !== undefined) {
                        response = found;
                    }
                    if ($.isArray(returnedValue)) {
                        returnedValue.push(response);
                    }
                    else if (returnedValue !== undefined) {
                        returnedValue = [returnedValue, response];
                    }
                    else if (response !== undefined) {
                        returnedValue = response;
                    }
                    return found;
                }
            };
            if (methodInvoked) {
                if (instance === undefined) {
                    module.initialize();
                }
                module.invoke(query);
            }
            else {
                if (instance !== undefined) {
                    module.destroy();
                }
                module.initialize();
            }
        });
        return (returnedValue !== undefined) ? returnedValue : this;
    };

    $.fn.accordion.settings = {
        name: 'Accordion',
        namespace: 'accordion',

        multiple: false,
        collapsible: true,

        duration: 500,
        easing: 'ease-in-out',

        onOpen: function() {
        },
        onClose: function() {
        },
        onChange: function() {
        },

        className: {
            active: 'am-active'
        },

        selector: {
            item: '.am-accordion-item',
            title: '.am-accordion-title',
            content: '.am-accordion-content'
        }
    };
});