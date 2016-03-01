Ext.Loader.setConfig({
    enabled: true,
    paths: { 'Classes' : 'ext_classes' }
});

Ext.require('Classes.Profile');

Ext.define('Person.Panel', {
        alias: 'widget.profile_panel',
        extend: 'Ext.panel.Panel',
        title: 'Персональная панель',
        html : '<a href="/">Домой</a>'
});

Ext.onReady(function(){
    Ext.create('Ext.container.Viewport', {
        layout: 'fit',
        items: [{
            xtype:"profile_panel"
        }]
    });
});

