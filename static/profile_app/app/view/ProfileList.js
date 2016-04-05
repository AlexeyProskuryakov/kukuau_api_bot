text_trimmer = function(text, meta){

    var result = text;
    var count_brs = (text.match(/\<br\>/g) || []).length;
    if (count_brs > 2){
        result = text.substring(0, text.indexOf('<br>', text.indexOf('<br>')+1)) + "...";
    }
    if (result.length>155){
        result = result.substring(0,155)+"...";
    }
    return result;
}
Ext.define('Console.view.ProfileList' ,{
    extend: 'Ext.grid.Panel',
    alias: 'widget.profilelist',

    title: 'Список профайлов',
    store: 'ProfileStore',

    initComponent: function() {
        this.columns = [
        {header: 'Имя',  dataIndex: 'name', flex: 1},
        {header: 'Короткое описание',  dataIndex: 'short_description', flex: 1, renderer:text_trimmer},
        {
            header: 'Полное описание', 
            dataIndex: 'text_description', 
            flex: 1,
            renderer:text_trimmer
        },
        {
            xtype: 'booleancolumn', 
            text: 'Включен',
            trueText: 'Да',
            falseText: 'Нет', 
            dataIndex: 'enable'
        },{
            xtype: 'booleancolumn', 
            text: 'Публичен',
            trueText: 'Да',
            falseText: 'Нет', 
            dataIndex: 'public'
        }
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
