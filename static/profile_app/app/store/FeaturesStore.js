Ext.define('Console.store.FeaturesStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Feature',
    autoLoad: true,
    autoSync: true,
    storeId: 'FeaturesStore',
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

