Ext.define('Console.view.ProfileList' ,{
    extend: 'Ext.grid.Panel',
    alias: 'widget.profilelist',

    title: 'Список профайлов',
    store: 'ProfileStore',

    initComponent: function() {
        console.log("profile list init");
        this.columns = [
        {header: 'Имя',  dataIndex: 'name', flex: 1},
        {header: 'Короткое описание',  dataIndex: 'short_description', flex: 1},
        {header: 'Полное описание', dataIndex: 'text_description', flex: 1},
        {header: 'Адресс', dataIndex: 'address', flex: 1},
        ];

        this.buttons = [{
            text: 'Добавить новый профайл',
            scope: this,
            action: 'new'
        }];
        this.callParent(arguments);
    }
});
