Ext.define('Console.store.ContactLinksStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.ContactLink',
    autoLoad: true,
    autoSync: true,
    storeId: 'ContactLinksStore',
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

