Ext.define('Console.view.Profile', {
    extend: 'Ext.window.Window',
    alias: 'widget.profilewindow',

    title: 'Профайл',
    layout: 'fit',
    autoShow: false,
    autoScroll:false,
    width:520,
    height:1024,

    initComponent: function() {
        console.log("init profile");
        var me = this
        this.items = [{
            xtype: 'form',
            items: [ 
            new Ext.form.FormPanel({
                frame: true,
                layout: 'column',
                defaults: {
                    xtype: 'form',
                },
                fileUpload:true,
                items:[
                {
                    xtype: 'image',
                    src:"/img/no_icon.gif",
                    width:100,
                    height:100,
                    name:'image_url',
                    padding:10,
                    id:"profile_image",
                    itemId:"profile_image"
                },{
                   xtype: 'filefield',
                   name: 'img_file',
                   fieldLabel: 'Фотокарточка',
                   allowBlank: false,
                   width:380,
                   padding:10,
                   buttonText: 'Выбрать фотокарточку'
               },
               ],
               buttons: [{
                text: 'Загрузить фотокарточку',
                handler: function() {
                    var form = this.up('form').getForm(),
                        p_model = me.down('form').getRecord(),
                        panel = this;
                    var profile_id;
                    if (p_model == undefined){
                        profile_id = guid();
                    } else {
                        profile_id = p_model.get("id");
                    }
                    

                    if(form.isValid()){
                        form.submit({
                            headers: { 'Content-Type': 'multipart/form-data' },
                            url: '/profile/upload_img/'+profile_id,
                            waitMsg: 'Загрузка фотокарточки...',
                            success: function(form, action) {
                                console.log(form, action);
                                Ext.getCmp("profile_image").setSrc(action.result.url);
                            },
                        });
                    }
                }
            }]
        }),{
                xtype:'checkbox',
                inputValue:true,
                name:'enable',
                fieldLabel:"Включен",
                padding:10
            },
            {
                xtype:'checkbox',
                name:'public',
                inputValue:true,
                fieldLabel:"Публичен",
                padding:10
            },
            {
                xtype: 'textfield',
                name : 'name',
                fieldLabel: 'Имя',
                width:400,
                padding:10
            },{
                xtype: 'htmleditor',
                name: 'short_description',
                enableColors:false,
                enableFontSize:false,
                enableLists:false,
                enableSourceEdit:false,
                enableAlignments:false,
                enableFont:false,
                height:170,
                grow: true,
                fieldLabel: 'Слоган',
                padding:10
            },{
                xtype: 'htmleditor',
                name : 'text_description',
                height:100,
                enableFont:false,
                enableColors:false,
                enableFontSize:false,
                enableLists:false,
                enableSourceEdit:false,
                enableAlignments:false,
                height:270,
                grow: true,
                fieldLabel: 'Описание',
                padding:10
            },
            {
                xtype:"grid",
                title:"Контакты",
                itemId:"profile_contacts",
                store: 'ContactsStore',
                name:'contacts',
                columns:[
                {header: 'Адрес',  dataIndex: 'address', flex:1},
                {header: 'Описание', dataIndex: 'description', flex:1},
                {
                    xtype : 'actioncolumn',
                    header : 'Delete',
                    width : 100,
                    align : 'center',
                    action:"delete_contact",
                    items : [
                    {
                        icon:'img/delete-icon.png',
                        tooltip : 'Delete',
                        scope : me
                    }]
                }
                ],
            },
            ]
        }];

        this.buttons = [{
            text: 'Сохранить',
            scope: this,
            action: 'save'
        },{
            text:'Удалить',
            action: 'delete',
            scope: this
        },
        {
            text:"Добавить контакт",
            action:"add_contact_start",
            scope: this,
        }

        ];

        this.callParent(arguments);
    }
});