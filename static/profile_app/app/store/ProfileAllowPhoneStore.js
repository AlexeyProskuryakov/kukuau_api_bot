Ext.define('Console.store.ProfileAllowPhoneStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.ProfileAllowPhone',
    autoLoad: true,
    autoSync: true,
    storeId: 'ProfileAllowPhoneStore',
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
    },
});

