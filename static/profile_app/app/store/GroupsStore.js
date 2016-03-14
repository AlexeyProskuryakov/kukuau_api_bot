Ext.define('Console.store.GroupsStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Group',
    autoLoad: true,
    autoSync: true,
    storeId: 'GroupsStore',
    proxy: {
        type: 'memory',
        reader: {
            type: 'json',
            root: 'data',
        },
        writer: {
            type: 'json',
            root: 'data',
        },
    }
});

