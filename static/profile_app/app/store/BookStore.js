Ext.define('BookApp.store.BookStore', {
    extend: 'Ext.data.Store',
    model: 'BookApp.model.Book',
    autoLoad: true,
    storeId: 'BookStore',
    proxy: {
        type: 'ajax',
        url: '/data/books',
        reader: {
            type: 'json',
            root: 'books',
            successProperty: 'success'
        }
    }
});