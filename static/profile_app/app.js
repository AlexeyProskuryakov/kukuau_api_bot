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