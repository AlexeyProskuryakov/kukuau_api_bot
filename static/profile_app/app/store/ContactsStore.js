Ext.define('Console.store.ContactsStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Contact',
    autoLoad: true,
    autoSync: true,
    storeId: 'ContactsStore',
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
    sorters: [{
         property: 'order_number',
         direction: 'ASC'
     }]
});

