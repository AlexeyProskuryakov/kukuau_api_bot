Ext.define('Console.store.ProfileStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Profile',
    autoLoad: true,
    storeId: 'ProfileStore',
    proxy: {
        type: 'ajax',
        url: '/data/profiles',
        reader: {
            type: 'json',
            root: 'profiles',
            successProperty: 'success'
        }
    }
});