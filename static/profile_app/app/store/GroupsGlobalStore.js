Ext.define('Console.store.GroupsGlobalStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Group',
    storeId: 'GroupsGlobalStore',
    proxy: {
        type: 'ajax',
        url: '/profile/all_groups',
        reader: {
            type: 'json',
            root: 'groups',
            successProperty: 'success'
        }
    }
});

