Ext.define('Console.store.FeaturesGlobalStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Feature',
    storeId: 'FeaturesGlobalStore',
    proxy: {
        type: 'ajax',
        url: '/profile/all_features',
        reader: {
            type: 'json',
            root: 'features',
            successProperty: 'success'
        }
    }
});

