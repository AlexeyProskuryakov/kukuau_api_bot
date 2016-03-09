Ext.define('BookApp.view.Book', {
    extend: 'Ext.window.Window',
    alias: 'widget.bookwindow',
 
    title: 'Книга',
    layout: 'fit',
    autoShow: true,
 
    initComponent: function() {
        console.log("init book");
        this.items = [{
                xtype: 'form',
                items: [{
                        xtype: 'textfield',
                        name : 'name',
                        fieldLabel: 'Название'
                    },{
                        xtype: 'textfield',
                        name : 'author',
                        fieldLabel: 'Автор'
                    },{
                        xtype: 'numberfield',
                        name : 'year',
                        fieldLabel: 'Год',
                        minValue: 1000,
                }]
            }];

        this.dockedItems=[{
            xtype:'toolbar',
            docked: 'top',
            items: [{
                text:'Создать',
                iconCls:'new-icon',
                action: 'new'
            },{
                text:'Сохранить',
                iconCls:'save-icon',
                action: 'save'
            },{
                text:'Удалить',
                iconCls:'delete-icon',
                action: 'delete'
            }]
        }];

        this.buttons = [{
                text: 'Очистить',
                scope: this,
                action: 'clear'
        }];
 
        this.callParent(arguments);
    }
});