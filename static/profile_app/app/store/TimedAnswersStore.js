Ext.define('Console.store.TimedAnswersStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.TimedAnswer',
    autoLoad: true,
    autoSync: true,
    storeId: 'TimedAnswersStore',
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

