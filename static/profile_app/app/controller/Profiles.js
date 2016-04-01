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
    phones_grid = form.getComponent("profile_phones");

    form.loadRecord(profileModel);
    contacts_grid.reconfigure(profileModel.contacts());
    phones_grid.reconfigure(profileModel.phones());
    groups_grid.reconfigure(profileModel.groups());

    var image_src = profileModel.get("image_url");

    if (image_src != ""){
        form.getComponent("profile_image_wrapper").getComponent("profile_image").setSrc(image_src);
    }

    return profile_window;
}
var geocoder = new google.maps.Geocoder();

Ext.define('Console.controller.Profiles', {
    extend: 'Ext.app.Controller',
    views: ['ProfileList', 'UserNameCheck', 'Profile', 'Contact', 'ContactLink', 'Phone', 'GroupChoose'],
    stores: ['ProfileStore', 'ContactsStore', 'ContactLinksStore', 'GroupsStore', 'GroupsGlobalStore', 'ProfileAllowPhoneStore'],
    models: ['Profile'],
    init: function() {
        this.control({
            'viewport > profilelist': {
                itemdblclick: this.editProfile
            },
            'profilelist button[action=new]':{
                click: this.createProfileStart
            },
            'UserNameCheckWindow button[action=newProfile]':{
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


        });
        Ext.widget('profilelist').getStore().load();
    },
    // обновление
    updateProfile: function(button) {
        var win    = button.up('window');
        var form   = win.down('form');
        var values = form.getValues();
        var record = form.getRecord();
        if (record != undefined) {
            if (values.name == ""){
                if (!form.isValid()){
                    return;
                }
            }
            var id = record.get('id'),
            cntcts = [],
            phones = [];
            
            values.id=id,
            Ext.each(record.contacts().data.items, function(item){
                var c_data = item.getData();
                c_data.links = [];
                Ext.each(item.links().data.items, function(l_item){
                    c_data.links.push(l_item.getData());
                });
                cntcts.push(c_data);
            });
            values.contacts = cntcts;
            
            Ext.each(record.phones().data.items, function(p_item){
                var p_data = p_item.getData();
                phones.push(p_data);
            })
            values.phones = phones;
            values.image_url = form.getComponent("profile_image_wrapper").getComponent("profile_image").src;
        } 

        Ext.Ajax.request({
            url: 'profile/update',
            jsonData: values,
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
        id =win.getComponent("user_name").getValue(),
        view = Ext.widget('profilewindow'),
        store = Ext.widget('profilelist').getStore(),
        profile_model = Ext.create("Console.model.Profile", {id:id});
        store.add(profile_model);
        view.down("form").loadRecord(profile_model);
        win.destroy();
        view.show();


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
        map_cmp = c_form.getComponent("contact_map"),
        center = {lat:54.858088, "lng": 83.110492};

        if (!(record instanceof Ext.EventObjectImpl)){   
            c_form.loadRecord(record);
            var cl_grid = c_form.getComponent("profile_contact_links");
            cl_grid.reconfigure(record.links());
            if ((record.get("lat") != 0.0) || (record.get("lon") != 0.0)) {
                center = {lat:record.get("lat"), lng:record.get("lon")};
            } 
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
            if (c_model == undefined){
                c_model = Ext.create("Console.model.Contact", c_form.getValues());
                c_form.loadRecord(c_model);
                c_form.getComponent("profile_contact_links").reconfigure(c_model.links());
            } 
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
        c_model = parent_form.getRecord();

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
            var phone_model = Ext.create("Console.model.ProfileAllowPhone", {id:guid(), value:phone_cmp.getValue()})
            profile_model.phones().add(phone_model);
            win.destroy();
        }        
    },

    deletePhone:function(grid, row, index){
        grid.getStore().removeAt(index);
    },

    addGroupStart:function(button){
        var win    = button.up('window'),
        c_view = Ext.widget("groupWindow", {"parent":win});
        c_view.show();
    },

    addGroupEnd:function(button){
        var win = button.up('window'),
        profile_model = win.getParent().down("form").getRecord(),
        phone_cmp = win.down('form').getComponent("phone_value");
        if (phone_cmp.validate()){
            var phone_model = Ext.create("Console.model.ProfileAllowPhone", {id:guid(), value:phone_cmp.getValue()})
            profile_model.phones().add(phone_model);
            win.destroy();
        }        
    },

    deleteGroup:function(grid, row, index){
        grid.getStore().removeAt(index);
    }, 

    chooseGroup:function(){
        Ext.widget('groupGlobalGrid').getStore().load();
    }

});
