Ext.define('Console.view.ProfileList' ,{
    extend: 'Ext.grid.Panel',
    alias: 'widget.profilelist',

    title: 'Список профайлов',
    store: 'ProfileStore',

    initComponent: function() {
        this.columns = [
        {header: 'Имя',  dataIndex: 'name', flex: 1},
        {header: 'Короткое описание',  dataIndex: 'short_description', flex: 1},
        {header: 'Полное описание', dataIndex: 'text_description', flex: 1},
        {header: 'Адресс', dataIndex: 'address', flex: 1},
        {xtype: 'booleancolumn', 
        text: 'Включен',
        trueText: 'Да',
        falseText: 'Нет', 
        dataIndex: 'enable'},
        {xtype: 'booleancolumn', 
        text: 'Публичен',
        trueText: 'Да',
        falseText: 'Нет', 
        dataIndex: 'public'}
        ];

        this.buttons = [{
            text: 'Добавить новый профайл',
            scope: this,
            action: 'new'
        }];
        this.dockedItems=[{
            xtype:'toolbar',
            docked: 'top',
            items: [
            {
                text:'В старую консоль',
                handler: function() {
                    window.location = '/';
                }
            }
            ]
        }];
        this.callParent(arguments);
    }
});
