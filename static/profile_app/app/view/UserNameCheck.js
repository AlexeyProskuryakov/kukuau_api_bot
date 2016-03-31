var userCheckVType = {
        userCheck: function(val, field){
            var userCheckRegex = /^[a-z0-9_]+$/;
            return userCheckRegex.test(val);
        },
        userCheckText: 'Имя пользователя должно быть латинским маленькими буквами, цифрами, если нужны пробелы используйте нижнее подчеркивание. Например: diana_fata_morgana',
        userCheckMask: /[a-z0-9_]/
    };
Ext.apply(Ext.form.field.VTypes, userCheckVType);


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
        padding:100,
        vtype:"userCheck"
    }],
    buttons:[{
        text:"Да",
        scope:this,
        action:'newProfile'
    }]
});
