

Ext.define('Console.view.Profile', {
    extend: 'Ext.window.Window',
    alias: 'widget.profilewindow',

    title: 'Профайл',
    layout: 'fit',
    autoShow: true,
    width:800,
    height:800,

    initComponent: function() {
        console.log("init profile");

        this.items = [{
            xtype: 'form',
            items: [ 
            {
                xtype: 'fileuploadfield',
                buttonOnly: false,
                buttonText: "Загрузить",
                fieldLabel: 'Картинка профиля',
                name: 'image' ,
                
                padding:10
            },{
                xtype: 'textfield',
                name : 'name',
                fieldLabel: 'Имя',
                width:400,
                padding:10
            },{
                xtype: 'htmleditor',
                name: 'short_description',
                grow: true,
                fieldLabel: 'Короткое описание',
                
                padding:10
            },{
                xtype: 'htmleditor',
                name : 'text_description',
                grow: true,
                fieldLabel: 'Длинное описание',
                
                padding:10
            },{
                xtype:'textfield',
                name:'address',
                fieldLabel:"Адресс",
                width:800,
                padding:10
            }, 

            {
                xtype:"toolbar",
                name:"contacts",
                items:[

                ]
            },
            {
                xtype:"button",
                scale:"small",
                padding:10,
                margin: '0 0 10 0',
                text:"Добавить контакт",
                action:"add_contact_start"
            }]
        }];

        this.dockedItems=[{
            xtype:'toolbar',
            docked: 'top',
            items: [
            {
                text:'Очистить',
                iconCls:'save-icon',
                action: 'clear'
            },{
                text:'Удалить',
                iconCls:'delete-icon',
                action: 'delete'
            }]
        }];

        this.buttons = [{
            text: 'Сохранить',
            scope: this,
            action: 'save'
        }];

        this.callParent(arguments);
    }
});