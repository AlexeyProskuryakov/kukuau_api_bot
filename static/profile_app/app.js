Ext.application({
    requires: ['Ext.container.Viewport'],
    name: 'Console',
 
    appFolder: 'profile_app/app',
    controllers: ['Profiles'],
    
    launch: function() {
        Ext.create('Ext.container.Viewport', {
            layout: 'fit',
            items: {
                xtype: 'profilelist',
            }
        });
    }
});
//todo click to pickture for loading it
//todo change html editor and configutre it
//todo list of categories
//todo choose adress at map
//todo change field names