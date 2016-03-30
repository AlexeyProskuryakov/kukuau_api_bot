Ext.define('Console.view.UserNameCheck', {
    extend:'Ext.window.Window',
    alias:'widget.UserNameCheckWindow',
    title:'Имя пользователя',
    layout:'fit',
    items: [{
        xtype:"textfield",
        itemId:'user_name',
        name:"user_name",
        fieldLabel:"Введите имя пользователя для которого будет создан сей профайл",
        width: 250,
        padding:100
    }],
    buttons:[{
        text:"Да",
        scope:this,
        action:'newProfile'
    }]
});
