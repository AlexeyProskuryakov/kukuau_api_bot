Ext.define('Console.store.EmployeesStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Employee',
    autoLoad: true,
    autoSync: true,
    storeId: 'EmployeesStore',
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

