var view = undefined;

function guid(is_str) {
    function s4() {
        return Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
    }
    function hash(str) {
        var hash = 0, i, chr, len;
        if (str.length === 0) return hash;
        for (i = 0, len = str.length; i < len; i++) {
            chr   = str.charCodeAt(i);
            hash  = ((hash << 5) - hash) + chr;
            hash |= 0; // Convert to 32bit integer
        }
        return hash;
    }
    var result = s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4();
    if (is_str == undefined || is_str == false){
        return hash(result);
    }
    return result;
}

function createProfileForm(profileModel){
    var profile_window = Ext.widget('profilewindow'),
    form = profile_window.down('form'),
    contacts_grid = form.getComponent('profile_contacts'), 
    groups_grid = form.getComponent("profile_groups"),
    phones_grid = form.getComponent("profile_phones"),
    features_grid = form.getComponent("profile_features"),
    employees_grid = form.getComponent("profile_employees");

    form.loadRecord(profileModel);
    contacts_grid.reconfigure(profileModel.contacts());
    phones_grid.reconfigure(profileModel.phones());
    groups_grid.reconfigure(profileModel.groups());
    features_grid.reconfigure(profileModel.features());
    employees_grid.reconfigure(profileModel.employees());

    var image_src = profileModel.get("image_url");
    
    if (image_src != ""){
        form.getComponent("profile_image_wrapper").getComponent("profile_image").setSrc(image_src);
    }

    profile_window.setTitle(profile_window.title+" ("+profileModel.get("id")+")");

    var  answers_grid = form.getComponent("profile_bot_config_wrapper").getComponent("answers"),
    notifications_grid = form.getComponent("profile_bot_config_wrapper").getComponent("notifications"),
    information_field = form.getComponent("profile_bot_config_wrapper").getComponent("information"),
    botConfigRaw = profileModel.get('botconfig'),
    botConfig = Ext.create('Console.model.BotConfig', botConfigRaw);

    if (botConfigRaw.answers != undefined && botConfigRaw.answers != null){
        botConfigRaw.answers.forEach(function(x){
            x._id = guid();
            botConfig.answers().add(Ext.create('Console.model.TimedAnswer', x))
        })
        answers_grid.reconfigure(botConfig.answers());
    }

    if (botConfigRaw.notifications != undefined && botConfigRaw.notifications != null){
        botConfigRaw.notifications.forEach(function(x){
            x._id = guid();
            botConfig.notifications().add(Ext.create('Console.model.TimedAnswer', x))
        })
        notifications_grid.reconfigure(botConfig.notifications());
    }

    information_field.setValue(botConfig.get("information"));

    profileModel.setBotConfig(botConfig);
    return profile_window;
}
var geocoder = new google.maps.Geocoder();

Ext.define('Console.controller.Profiles', {
    extend: 'Ext.app.Controller',
    views: ['ProfileList', 'UserNameCheck', 'Profile', 'Contact', 'ContactLink', 'Phone', 'GroupChoose', 'NewGroupAdd', 'FeatureChoose', 'EmployeeInfo', 'Notification', 'AutoAnswer'],
    stores: ['ProfileStore', 'ContactsStore', 'ContactLinksStore', 'GroupsStore', 'GroupsGlobalStore', 'ProfileAllowPhoneStore', 'FeaturesStore', 'FeaturesGlobalStore', 'EmployeesStore','TimedAnswersStore'],
    models: ['Profile','BotConfig'],
    config:{
        group_global_storage: undefined,
        feature_global_storage: undefined
    },
    init: function() {
        this.control({
            'viewport > profilelist': {
                itemdblclick: this.editProfile
            },
            'profilelist button[action=new]':{
                click: this.createProfileStart
            },
            'UserNameCheckWindow button[action=new_profile]':{
                click: this.createProfileEnd
            },
            'profilewindow button[action=save]': {
                click: this.updateProfile
            },
            'profilewindow button[action=delete]': {
                click: this.deleteProfile
            },
            'profilewindow button[action=clear]': {
                click: this.clearForm
            },

            'profilewindow button[action=add_contact_start]':{
                click: this.showContactForm
            },
            
            'profilewindow grid[itemId=profile_contacts]':{
                itemdblclick: this.showContactForm
            },

            'profilewindow actioncolumn[action=delete_contact]':{
                click: this.deleteContact
            },

            //contact links
            'contactLinkWindow button[action=add_contact_end]':{
                click:this.addContact
            },
            'contactLinkWindow button[action=save_contact_link]':{
                click:this.saveContactLink
            },
            'contactWindow grid[itemId=profile_contact_links]':{
                itemdblclick:this.showContactLinkForm
            },
            'contactWindow button[action=save_contact]':{
                click:this.saveContact
            },
            'contactWindow button[action=add_contact_link]':{
                click:this.showContactLinkForm
            },
            'contactWindow actioncolumn[action=delete_contact_link]':{
                click:this.deleteContactLink
            },

            //phones
            'profilewindow button[action=add_phone_start]':{
                click:this.addPhoneStart
            },

            'phoneWindow button[action=add_phone_end]':{
                click:this.addPhoneEnd
            },

            'profilewindow actioncolumn[action=delete_phone]':{
                click: this.deletePhone
            },

            //groups
            'profilewindow button[action=add_group_start]':{
                click:this.addGroupStart
            },

            'groupWindow button[action=add_group_end]':{
                click:this.addGroupEnd
            },

            'profilewindow actioncolumn[action=delete_group]':{
                click: this.deleteGroup
            },        
            'groupWindow button[action=add_global_group_start]':{
                click: this.addGlobalGroupStart
            },

            'groupWindow grid[itemId=choose_group_grid]':{
                itemdblclick: this.changeGlobalGroup
            },

            'newGroupWindow button[action=add_global_group_end]':{
                click: this.addGlobalGroupEnd
            },

            //features
            'profilewindow button[action=add_feature_start]':{
                click: this.addFeatureStart
            },

            'featureWindow button[action=add_feature_end]':{
                click: this.addFeatureEnd
            }, 
            'profilewindow actioncolumn[action=delete_feature]':{
                click: this.deleteFeature
            },

            //employees
            'profilewindow button[action=add_employee_start]':{
                click: this.addEmployeeStart
            },
            
            'employeeWindow button[action=add_employee_end]':{
                click: this.addEmployeeEnd
            },
            
            'profilewindow actioncolumn[action=delete_employee]':{
                click: this.deleteEmployee
            },

            'profilewindow grid[itemId=profile_employees]':{
               itemdblclick: this.changeEmployee
           },       
            //bot configuration
            'profilewindow grid[itemId=answers]':{
                itemdblclick: this.changeAnswer
            },
            
            'profilewindow grid[itemId=notifications]':{
                itemdblclick: this.changeNotification
            },
            'profilewindow button[action=add_answer_start]':{
                click: this.createAnswer
            },
            'answerWindow button[action=add_answer_end]':{
                click: this.createAnswerEnd
            },

            'profilewindow button[action=add_notification_start]':{
                click: this.createNotification
            },
            'notificationWindow button[action=add_notification_end]':{
                click: this.createNotificationEnd
            },
            'profilewindow actioncolumn[action=delete_answer]':{
                click: this.deleteAnswer
            },
            'profilewindow actioncolumn[action=delete_notification]':{
                click: this.deleteNotification
            },
        });

Ext.widget('profilelist').getStore().load();
this.group_global_storage = Ext.create("Console.store.GroupsGlobalStore").load();
this.feature_global_storage = Ext.create("Console.store.FeaturesGlobalStore").load();
},
    // обновление
    updateProfile: function(button) {
        var win    = button.up('window');
        var form   = win.down('form');
        var profile_main_values = form.getComponent("profile_main_information").getValues();
        var record = form.getRecord();
        if (record != undefined) {
            if (profile_main_values.name == ""){
                if (!form.isValid()){
                    return;
                }
            }
            var id = record.get('id'),
            cntcts = [],
            phones = [],
            groups = [],
            features = [],
            employees = [];

            
            profile_main_values.id=id,
            Ext.each(record.contacts().data.items, function(item){
                var c_data = item.getData();
                c_data.links = [];
                Ext.each(item.links().data.items, function(l_item){
                    c_data.links.push(l_item.getData());
                });
                cntcts.push(c_data);
            });
            profile_main_values.contacts = cntcts;
            
            Ext.each(record.phones().data.items, function(p_item){
                var p_data = p_item.getData();
                phones.push(p_data);
            });
            profile_main_values.phones = phones;

            Ext.each(record.groups().data.items, function(g_item){
                groups.push(g_item.getData());
            });
            profile_main_values.groups = groups;

            Ext.each(record.features().data.items, function(f_item){
                features.push(f_item.getData());
            });
            profile_main_values.features = features;

            Ext.each(record.employees().data.items, function(e_item){
                employees.push(e_item.getData());
            });
            profile_main_values.employees = employees;

            profile_main_values.image_url = form.getComponent("profile_image_wrapper").getComponent("profile_image").src;

            //bot configuration
            var botConfig = record.getBotConfig();

            profile_main_values.botconfig = {answers:[], notifications:[],information:form.getComponent("profile_bot_config_wrapper").getComponent("information").getValue()};
            Ext.each(botConfig.answers().data.items, function(a_item){
                profile_main_values.botconfig.answers.push(a_item.getData());
            });
            Ext.each(botConfig.notifications().data.items, function(n_item){
                profile_main_values.botconfig.notifications.push(n_item.getData());
            });
            console.log("profile main values", profile_main_values);
        } 

        Ext.Ajax.request({
            url: 'profile/update',
            jsonData: profile_main_values,
            success: function(response){
                var data=Ext.decode(response.responseText);
                if(data.success){
                    var store = Ext.widget('profilelist').getStore();
                    store.load();
                    win.destroy();
                }
                else{
                    Ext.Msg.alert('Обновление','Что-то пошло не так...');
                }
            }
        });
        
    },
    createProfileStart:function(button){
        var view = Ext.widget("UserNameCheckWindow");
        view.show();
    },
    // создание
    createProfileEnd: function(button) {
        var win = button.up("window"),
        cmp = win.getComponent("user_name"),
        id = cmp.getValue();
        if (id != "" && cmp.validate()){
            var view = Ext.widget('profilewindow'),
            store = Ext.widget('profilelist').getStore(),
            profile_model = Ext.create("Console.model.Profile", {id:id});
            profile_model.setBotConfig(Ext.create('Console.model.BotConfig'));

            store.add(profile_model);
            view.down("form").loadRecord(profile_model);
            win.destroy();
            view.setTitle(view.title+" ("+id+")");
            view.show();
        } else {
            cmp.markInvalid("Введите имя правильно, оно необходимо.");
        }

    },
    // удаление
    deleteProfile: function(button) {
        var win    = button.up('window'),
        form   = win.down('form'),
        id = form.getRecord().get('id');
        var q_w = Ext.create('Ext.window.Window', {
            title: 'Уверены?',
            width: 300,
            height: 200,
            items:[{
                xtype: 'button',
                text: 'Да!',
                scale   : 'large',
                style:'margin-left:110px; margin-top:60px;',
                handler:function(){
                    Ext.Ajax.request({
                        url: 'profile/delete',
                        jsonData: {id:id},
                        success: function(response){
                            var data=Ext.decode(response.responseText);
                            if (data.success) {
                                var store = Ext.widget('profilelist').getStore();
                                var record = store.getById(id);
                                store.remove(record);
                                view.hide()
                            } else {
                                Ext.Msg.alert('Удаление','Что-то пошло не так...');
                            }
                        }
                    });            
                    q_w.hide();
                }
            }],
        });
        q_w.show();
        
    },

    clearForm: function(grid, record) {
        view.down('form').getForm().reset();
    },

    editProfile: function(grid, record) {
        view = createProfileForm(record);
        view.show();
    },
    addContact:function(button){
        var win    = button.up('window');
        var form   = win.down('form');
        var contact_model = form.getRecord();
        var profile_win = win.getParent();
        var profile_form = profile_win.down("form");
        var profile_model = profile_form.getRecord();
        var store = profile_model.contacts();
        var values = form.getValues();
        if (contact_model == undefined) {
            values.id = guid();
            store.add(values);
        } else{
            values.id = contact_model.getId();
            var rec = store.getById(contact_model.getId());
            rec.set(values);
        }
        win.hide();
    },

    deleteContact: function(grid, row, index){
        grid.getStore().removeAt(index);
    },

    showContactForm: function(button, record){
        var win    = button.up('window'),
        c_view = Ext.widget("contactWindow", {"parent":win}),
        c_form = c_view.down("form"),
        map_cmp = c_form.down("form").getComponent("contact_map"),
        center = {lat:54.858088, "lng": 83.110492};

        if (!(record instanceof Ext.EventObjectImpl)){   
            c_form.loadRecord(record);
            var cl_grid = c_form.getComponent("profile_contact_links");
            cl_grid.reconfigure(record.links());
            if ((record.get("lat") != 0.0) || (record.get("lon") != 0.0)) {
                center = {lat:record.get("lat"), lng:record.get("lon")};
            } 
        } else {
            c_model = Ext.create("Console.model.Contact", {id:guid()});
            c_form.loadRecord(c_model);
            c_form.getComponent("profile_contact_links").reconfigure(c_model.links());
        } 
        var p_model = win.down("form").getRecord();  
        var marker = new google.maps.Marker({
            position: center,
            map: map_cmp.getMap()
        });    
        map_cmp.addMarkers([center]);    
        map_cmp.setCenter = center;
        c_view.show();
    },

    showContactLinkForm: function(button, record){
        var win = button.up('window'),
        cl_view = Ext.widget("contactLinkWindow", {"parent":win}),
        cl_form = cl_view.down("form"),
        c_form = win.down("form");

        if (!(record instanceof Ext.EventObjectImpl)){
            cl_form.loadRecord(record);
        } else {
            var onf = cl_form.getForm().findField("order_number"),
            c_model = win.down("form").getRecord();
            var cl_store = c_model.links();
            onf.setValue(cl_store.count()+1);    
        }
        cl_view.show();

    },

    saveContactLink:function(button){
        var win    = button.up('window'),
        form   = win.down('form'),
        cl_model = form.getRecord(),
        values = form.getValues(),
        parent_form = win.getParent().down("form"),
        c_model = parent_form.getRecord(),
        cmp_type = form.getComponent('cl_type'),
        cmp_value = form.getComponent('cl_value');
        if (!cmp_type.validate()){
            cmp_type.markInvalid("Указание типа обязательно!");
            return;
        }
        if (!cmp_value.validate()){
            cmp_value.markInvalid("Указание значения обязательно!");
            return;
        }
        l_store = c_model.links();
        if (cl_model != undefined){
            var cl_id = cl_model.getId(),
            stored_cl_rec = l_store.getById(cl_id);

            values.id = cl_id;
            stored_cl_rec.set(values);            
        } else {
            values.id = guid();
            l_store.add(values);
        }
        
        win.hide();
    },
    deleteContactLink:function(grid, row, index){
        grid.getStore().removeAt(index);
    },

    saveContact:function(button){
        console.log("saving contact");

        var win = button.up("window"),
        c_store = win.getParent().down('form').getRecord().contacts(),
        form = win.down("form"),
        c_values = form.getValues(),
        c_model = form.getRecord();
        
        if (c_model == undefined){
            c_values.id = guid();
            c_model = Ext.create("Console.model.Contact", c_values);
            c_store.add(c_model);
        } else {
            var c_id = c_model.getId(),
            c_rec = c_store.getById(c_id);

            c_values.id = c_id;
            c_model.set(c_values);
            if (c_rec != null){
                c_rec.set(c_values);    
            } else {
                c_store.add(c_model);
            }
        } 
        win.getParent().down('form').getComponent("profile_contacts").reconfigure(c_store);
        win.hide();
    },

    addPhoneStart:function(button){
        var win    = button.up('window'),
        c_view = Ext.widget("phoneWindow", {"parent":win});
        c_view.show();
    },

    addPhoneEnd:function(button){
        var win = button.up('window'),
        profile_model = win.getParent().down("form").getRecord(),
        phone_cmp = win.down('form').getComponent("phone_value");

        if (phone_cmp.validate()){
            var p_value = phone_cmp.getValue(),
            p_store = profile_model.phones(),
            contains = false;
            p_store.each(function(record, id){
                if (record.get('value') == p_value){
                    contains = true;
                }
            });
            if (contains){
                phone_cmp.markInvalid("Такой телефон уже существует!");
                return;
            }
            var phone_model = Ext.create("Console.model.ProfileAllowPhone", {id:guid(), value:phone_cmp.getValue()});

            p_store.add(phone_model);
            win.getParent().down("form").getComponent('profile_phones').reconfigure(p_store);

            win.destroy();
        }        
    },

    deletePhone:function(grid, row, index){
        grid.getStore().removeAt(index);
    },

    addGroupStart:function(button){
        var win    = button.up('window'),
        c_view = Ext.widget("groupWindow", {"parent":win}),
        profile_model = win.down("form").getRecord(),
        new_cgg_store = Ext.create("Console.store.GroupsGlobalStore"),
        cggCmp = c_view.down("form").getComponent("choose_group_grid");

        var profile_groups = [];
        Ext.each(profile_model.groups().data.items, function(item){
            profile_groups.push(item.getData()['name']);
        });

        console.log("profile groups is: ", profile_groups);
        this.group_global_storage.each(function(record, item){
            if (profile_groups.indexOf(record.get('name')) < 0){
                record.set("_active", false);
                record.dirty = false;
                new_cgg_store.add(record);
            }
        });
        cggCmp.reconfigure(new_cgg_store);

        c_view.show();
    },

    addGroupEnd:function(button){
        var win = button.up('window'),
        profile_form = win.getParent().down("form"),
        profile_model = profile_form.getRecord(),
        ggs = win.down('form').getComponent("choose_group_grid").getStore(),
        gps = profile_model.groups();

        Ext.each(ggs.data.items, function(item){
            var data = item.getData();
            if (data['_active'] == true){
                data['_active'] = false;
                gps.add(Ext.create('Console.model.Group', data));
                console.log("add group at profile", data);
            }
        })
        profile_form.getComponent("profile_groups").reconfigure(gps);
        win.destroy();
    },

    addGlobalGroupStart: function(button){
        var win = button.up('window'),
        ng_view = Ext.widget("newGroupWindow", {parent:win});
        ng_view.show();
    },

    addGlobalGroupEnd: function(button){
        var win = button.up('window'),
        ggs = win.getParent().down('form').getComponent('choose_group_grid').getStore(),
        ggv = win.down('form').getValues();
        if (ggv['name'] != ''){
            ggv['_active'] = true;
            var ggm = Ext.create('Console.model.Group', ggv);
            ggs.add(ggv);
            win.destroy();    
        } else {
            win.down("form").getComponent("group_name").markInvalid("Имя обязательно!");
        }
        
    },
    deleteGroup:function(grid, row, index){
        grid.getStore().removeAt(index);
    }, 

    addFeatureStart: function(button){
        console.log("adding new feature...");
        var win = button.up("window"),
        c_view = Ext.widget("featureWindow", {"parent":win}),
        profile_model = win.down("form").getRecord(),
        new_cfg_store = Ext.create("Console.store.FeaturesGlobalStore"),
        cfgCmp = c_view.down("form").getComponent("choose_feature_grid");

        var profile_features = [];
        Ext.each(profile_model.features().data.items, function(item){
            profile_features.push(item.getData()['name']);
        });

        console.log("profile features is: ", profile_features);
        this.feature_global_storage.each(function(record, item){
            if (profile_features.indexOf(record.get('name')) < 0){
                record.set("_active", false);
                record.dirty = false;
                new_cfg_store.add(record);
            }
        });
        cfgCmp.reconfigure(new_cfg_store);

        c_view.show();
    },

    addFeatureEnd:function(button){
        console.log("ending add feature");
        var win = button.up('window'),
        profile_form = win.getParent().down("form"),
        profile_model = profile_form.getRecord(),
        fgs = win.down('form').getComponent("choose_feature_grid").getStore(),
        fps = profile_model.features();

        Ext.each(fgs.data.items, function(item){
            var data = item.getData();
            if (data['_active'] == true){
                data['_active'] = false;
                fps.add(Ext.create('Console.model.Feature', data));
                console.log("add feature at profile", data);
            }
        })
        profile_form.getComponent("profile_features").reconfigure(fps);
        win.destroy();
    },

    deleteFeature:function(grid, row, index){
        console.log("delete feature", row, index, grid);
        grid.getStore().removeAt(index);
    }, 

    addEmployeeStart:function(button){
        console.log("adding new employee...");
        var win = button.up("window"),
        pModel = win.down("form").getRecord();
        var e_view = Ext.widget("employeeWindow", {"parent":win, profileId:pModel.get("id")});
        e_view.show();
    },

    changeEmployee: function(grid, row, index){
        console.log('change employee');
        var win = grid.up("window"),
        pModel = win.down("form").getRecord();

        var e_view = Ext.widget("employeeWindow", {"parent":win, profileId:pModel.get("id")}),
        e_form = e_view.down('form'),
        phoneCmp = e_form.getComponent('phone');

        phoneCmp.setValue(row.data.phone);
        e_view.show();  
    },

    addEmployeeEnd:function(button){
        console.log("adding new employee...");
        var e_win = button.up("window"),
        pForm = e_win.getParent().down("form"),
        pModel = pForm.getRecord(),
        e_form = e_win.down('form'),
        phoneComp = e_form.getComponent('phone');
        if (phoneComp.validate()){
            var phone = phoneComp.getValue();
            console.log('phone: ',phone);

            Ext.Ajax.request({
                url:"profile/employee/"+phone,
                success:function(x){
                    var data=Ext.decode(x.responseText);
                    if (data.success==true && data.employee != null ){
                        var employeeData = data.employee;
                        employeeData['phone'] = phone;
                        eModel = Ext.create("Console.model.Employee", employeeData);

                        pModel.employees().add(eModel);
                        var eGrid = pForm.getComponent("profile_employees");
                        eGrid.reconfigure(pModel.employees());
                        e_win.destroy();
                    } else {
                        Ext.Msg.alert('Ошибка!', 'Сотрудник с таким номером телефона не зарегистрирован в KliChat.');         
                    }
                }
            });
        }

    },
    deleteEmployee:function(grid, row, index){
        console.log("delete feature", row, index, grid);
        grid.getStore().removeAt(index);
    }, 

    createAnswer:function(button){
        console.log("start create answer");
        var win = button.up("window"),
        pModel = win.down("form").getRecord();
        var e_view = Ext.widget("answerWindow", {"parent":win, profileId:pModel.get("id")});
        e_view.show();
    },

    createAnswerEnd:function(button){
        console.log("end create answer");
        var win = button.up("window"),
        form = win.down('form'),
        pWin = win.getParent(),
        pModel = pWin.down("form").getRecord(),
        after_min = form.getComponent('after_min'),
        text = form.getComponent('text'),
        answers_grid = pWin.down('form').getComponent("profile_bot_config_wrapper").getComponent("answers"),
        botConfig = pModel.getBotConfig();

        if (text.validate() && after_min.validate()){

            var record = form.getRecord();
            if (record == undefined){
                botConfig.answers().add(Ext.create('Console.model.TimedAnswer', {after_min:after_min.getValue(), text:text.getValue()}))
                pModel.setBotConfig(botConfig);
            } else {
                var stored = botConfig.answers().getById(record.get("_id"));
                stored.set(record.getData());
            }
            answers_grid.reconfigure(botConfig.answers());
            win.destroy();
        }

    },
    createNotification:function(button){
        console.log("start create notification");
        var win = button.up("window"),
        pModel = win.down("form").getRecord();
        var e_view = Ext.widget("notificationWindow", {"parent":win, profileId:pModel.get("id")});
        e_view.show();
    },

    createNotificationEnd:function(button){
        console.log("end create notification");
        var win = button.up("window"),
        pWin = win.getParent(),
        pModel = pWin.down("form").getRecord(),
        after_min = win.down('form').getComponent('after_min'),
        text = win.down('form').getComponent('text'),
        notifications_grid = pWin.down('form').getComponent("profile_bot_config_wrapper").getComponent("notifications"),
        botConfig = pModel.getBotConfig();

        if (text.validate() && after_min.validate()){
            botConfig.notifications().add(Ext.create('Console.model.TimedAnswer', {after_min:after_min.getValue(), text:text.getValue()}))
            notifications_grid.reconfigure(botConfig.notifications());
            pModel.setBotConfig(botConfig);
            win.destroy();
        }

    },

    deleteAnswer:function(grid, row, index){
        console.log("delete feature", row, index, grid);
        grid.getStore().removeAt(index);
    }, 

    deleteNotification:function(grid, row, index){
        console.log("delete feature", row, index, grid);
        grid.getStore().removeAt(index);
    }, 

    changeAnswer:function(grid, row, index){
        console.log("start change answer");
        var win = grid.up("window"),
        pModel = win.down("form").getRecord(),
        e_view = Ext.widget("answerWindow", {"parent":win, profileId:pModel.get("id")}),
        e_form = e_view.down('form');
        
        e_form.loadRecord(row);
        e_view.show();
    }
});
