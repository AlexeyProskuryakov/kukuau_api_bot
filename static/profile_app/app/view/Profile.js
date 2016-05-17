Ext.define('Console.view.Profile', {
    extend: 'Ext.window.Window',
    alias: 'widget.profilewindow',

    title: 'Профайл',
    layout: 'fit',
    autoDestroy: true,
    maximizable : true,
    autoShow: false,
    width: 800,
    height: 600,

    // overflowY:'scroll',
    // autoScroll: true,
    // layout: 'fit',

    initComponent: function() {
        console.log("init profile");
        var me = this
        this.items = [{
            xtype: 'form',
            layout:'accordion',
            // autoScroll:true,
            items: [
            new Ext.form.FormPanel({
                frame: true,
                autoDestroy: true,
                title: "Иконка",
                collapsible: true,
                collapsed: true,
                itemId: "profile_image_wrapper",
                layout: 'column',
                fileUpload: true,
                items: [{
                    xtype: 'image',
                    src: "/img/no_icon.gif",
                    width: 100,
                    height: 100,
                    name: 'image_url',
                    padding: 10,
                    id: "profile_image",
                    itemId: "profile_image"
                }, {
                    xtype: 'filefield',
                    name: 'img_file',
                    fieldLabel: 'Иконка',
                    width: 380,
                    padding: 10,
                    buttonText: 'Выбрать иконку'
                }, ],
                buttons: [{
                    text: 'Загрузить иконку',
                    handler: function() {
                        var form = this.up('form').getForm(),
                        p_model = me.down('form').getRecord(),
                        panel = this;
                        var profile_id;
                        if (p_model == undefined) {
                            profile_id = guid();
                        } else {
                            profile_id = p_model.get("id");
                        }


                        if (form.isValid()) {
                            image_cmp = this.up("form").getComponent("profile_image");
                            form.submit({
                                headers: {
                                    'Content-Type': 'multipart/form-data'
                                },
                                url: '/profile/upload_img/' + profile_id,
                                waitMsg: 'Жульк жульк...',
                                success: function(form, action) {
                                    if (action.result.success == false){
                                        Ext.Msg.alert("Ошибка",action.result.error);
                                        return;
                                    }
                                    image_cmp.setSrc(action.result.url);
                                },
                                failure:function(form, action){
                                    Ext.Msg.alert("Ошибка", action.result.error);
                                }
                            });
                        }
                    }
                }]
            }), 
            {
                xtype: "grid",
                title: "Доступно телефонам",
                itemId: "profile_phones",
                store: 'ProfileAllowPhoneStore',
                collapsible: true,
                collapsed: true,
                defaults:{
                    xtype:"panel",
                },
                columns: [{
                    header: "Номер телефона",
                    dataIndex: 'value',
                    flex: 1
                }, {
                    xtype: 'actioncolumn',
                    header: 'Delete',
                    width: 100,
                    align: 'center',
                    action: "delete_phone",
                    items: [{
                        icon: 'img/delete-icon.png',
                        tooltip: 'Delete',
                        scope: me
                    }]
                }],
                buttons: [{
                    text: "Добавить телефон",
                    action: "add_phone_start",
                    scope: this,
                }]
            },
            {
                xtype: "grid",
                title: "Принадлежит группам",
                itemId: "profile_groups",
                store: 'GroupsStore',
                collapsible: true,
                collapsed: true,
                columns: [{
                    header: "Название группы",
                    dataIndex: 'name',
                    flex: 1

                }, {
                    header: "Описание",
                    dataIndex: 'description',
                    flex: 1
                }, {
                    xtype: 'actioncolumn',
                    header: 'Delete',
                    width: 100,
                    align: 'center',
                    action: "delete_group",
                    items: [{
                        icon: 'img/delete-icon.png',
                        tooltip: 'Delete',
                        scope: me
                    }]
                }],
                buttons: [{
                    text: "Добавить группу",
                    action: "add_group_start",
                    scope: this,
                }]
            },
            {
                xtype: "grid",
                title: "Контакты",
                itemId: "profile_contacts",
                store: 'ContactsStore',
                name: 'contacts',
                markDirty: false,
                collapsible: true,
                collapsed: true,
                columns: [{
                    header: 'Адрес',
                    dataIndex: 'address',
                    flex: 1
                }, {
                    header: 'Описание',
                    dataIndex: 'description',
                    markDirty: false,
                    flex: 1, 
                    renderer:function(item, meta){
                        if (item == "") {
                            var result = "";
                            meta.record.links().each(function(r,i){
                                var sep = ", ";
                                if (i == 0){
                                    sep = "";
                                }
                                if (r.get('value') != ''){
                                    result += sep + r.get('value');
                                }
                            });
                            if (result.length > 50){
                                result = result.substring(0,49) + "...";
                            }
                            return result;
                        }
                        return item;
                    }
                }, {
                    xtype: 'actioncolumn',
                    header: 'Delete',
                    width: 100,
                    align: 'center',
                    action: "delete_contact",
                    items: [{
                        icon: 'img/delete-icon.png',
                        tooltip: 'Delete',
                        scope: me
                    }]
                }],
                buttons: [{
                    text: "Добавить контакт",
                    action: "add_contact_start",
                    scope: this,
                }]
            },
            {
                xtype: "grid",
                title: "Фичи",
                itemId: "profile_features",
                store: 'FeaturesStore',
                collapsible: true,
                collapsed: true,
                columns: [{
                    header: "Название фичи",
                    dataIndex: 'name',
                    flex: 1

                }, {
                    header: "Var",
                    dataIndex: 'var',
                    flex: 1
                }, {
                    xtype: 'actioncolumn',
                    header: 'Delete',
                    width: 100,
                    align: 'center',
                    action: "delete_feature",
                    items: [{
                        icon: 'img/delete-icon.png',
                        tooltip: 'Delete',
                        scope: me
                    }]
                }],
                buttons: [{
                    text: "Добавить фичу",
                    action: "add_feature_start",
                    scope: this,
                }]
            },
            {
                xtype: "grid",
                title: "Сотрудники",
                itemId: "profile_employees",
                store: 'EmployeesStore',
                collapsible: true,
                collapsed: true,
                columns: [
                {
                    header: "Имя",
                    dataIndex: 'name',
                    flex: 1

                }, {
                    header: "Телефон",
                    dataIndex: 'phone',
                    flex: 1
                },
                {
                    xtype: 'actioncolumn',
                    header: 'Delete',
                    width: 100,
                    align: 'center',
                    action: "delete_employee",
                    items: [{
                        icon: 'img/delete-icon.png',
                        tooltip: 'Delete',
                        scope: me
                    }]
                }],
                buttons: [{
                    text: "Добавить сотрудника",
                    action: "add_employee_start",
                    scope: this,
                }]
            },
            {
                xtype:'form',
                title: "Настройка бота",
                collapsible:true,
                collapsed:true,
                itemId:'profile_bot_config_wrapper',
                items: [
                {
                    xtype: "grid",
                    title: "Автоответы",
                    collapsible: true,
                    collapsed: true,
                    itemId: "answers",
                    store: 'TimedAnswersStore',
                    columns: [{
                        header: "Через (минуты)",
                        dataIndex: 'after_min',
                        width:200,
                    }, {
                        header: "Ответ",
                        dataIndex: 'text',
                        flex: 1
                    }, {
                        xtype: 'actioncolumn',
                        header: 'Delete',
                        width: 100,
                        align: 'center',
                        action: "delete_answer",
                        items: [{
                            icon: 'img/delete-icon.png',
                            tooltip: 'Delete',
                            scope: me
                        }]
                    }],
                    buttons: [{
                        text: "Добавить автоответ",
                        action: "add_answer_start",
                        scope: this,
                    }]
                }, {
                    xtype: "grid",
                    title: "Нотификации",
                    itemId: "notifications",
                    store: 'TimedAnswersStore',
                    collapsible: true,
                    collapsed: true,
                    defaults:{
                        xtype:"panel",
                        autoscroll  : true
                    },
                    columns: [{
                        header: "Через (минуты)",
                        dataIndex: 'after_min',
                        width:200,
                    },
                    {
                        header: "Ответ",
                        dataIndex: 'text',
                        flex: 1
                    },
                    {
                        xtype: 'actioncolumn',
                        header: 'Delete',
                        width: 100,
                        align: 'center',
                        action: "delete_notification",
                        items: [{
                            icon: 'img/delete-icon.png',
                            tooltip: 'Delete',
                            scope: me
                        }]
                    }],
                    buttons: [{
                        text: "Добавить нотификацию",
                        action: "add_notification_start",
                        scope: this,
                    }]
                },
                {
                    xtype: 'textareafield',
                    fieldLabel: 'Информация',
                    itemId:'information',
                    dataIndex: 'information',
                    width: 580,
                    padding: 10,
                } 
                ]
            }, 
            new Ext.form.FormPanel({
                frame: true,
                autoDestroy: true,
                title: "Основная информация",
                collapsible: true,
                collapsed: false,
                itemId: "profile_main_information",
                // layout: 'fit',
                defaults: {
                    xtype: 'form',
                },
                // layout: 'fit',
                autoScroll:true,
                items: [
                {
                    xtype: 'checkbox',
                    inputValue: true,
                    name: 'enable',
                    fieldLabel: "Включен",
                    padding: 10
                },
                {
                    xtype: 'checkbox',
                    name: 'public',
                    inputValue: true,
                    fieldLabel: "Публичен",
                    padding: 10
                },
                {
                    xtype: 'textfield',
                    name: 'name',
                    fieldLabel: 'Имя',
                    width: 400,
                    padding: 10,
                    allowBlank: false
                },
                {
                    xtype: 'htmleditor',
                    name: 'short_description',
                    enableColors: false,
                    enableFontSize: false,
                    enableLists: false,
                    enableSourceEdit: false,
                    enableAlignments: false,
                    enableFont: false,
                    height: 100,
                    width:600,
                    grow: true,
                    fieldLabel: 'Слоган',
                    padding: 10, 
                    allowBlank:false,
                },
                {
                    xtype: 'htmleditor',
                    name: 'text_description',
                    enableFont: false,
                    enableColors: false,
                    enableFontSize: false,
                    enableLists: false,
                    enableSourceEdit: false,
                    enableAlignments: false,
                    height: 100,
                    width:600,
                    grow: true,
                    fieldLabel: 'Описание',
                    padding: 10,
                    allowBlank:false
                }
                ]
            }),
            
            ]
        }
        ];

        this.buttons = [{
            text: 'Сохранить',
            scope: this,
            action: 'save'
        }, {
            text: 'Удалить',
            action: 'delete',
            scope: this
        }


        ];

        this.callParent(arguments);}
    });