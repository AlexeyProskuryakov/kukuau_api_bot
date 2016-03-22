var view = undefined;

function guid() {
  function s4() {
    return Math.floor((1 + Math.random()) * 0x10000)
    .toString(16)
    .substring(1);
}
return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
s4() + '-' + s4() + s4() + s4();
}

function createContactComponent(type, value, descrption){
    var contactInfoComponent = {
        xtype:"bar",
        items:[
        {
            xtype:'label',
            text:type,
        },{
            xtype:'label',
            text:value, 
        },{
            xtype:'label',
            text:descrption
        }
        ]
    }
    return contactInfoComponent;
}

function createProfileForm(profileModel){
    var profile_window = Ext.widget('profilewindow');
    var form = profile_window.down('form');

    form.loadRecord(profileModel);

    var contacts_grid = form.getComponent('profile_contacts');
    contacts_grid.reconfigure(profileModel.contacts());
    
    var image = Ext.getCmp("profile_image");
    image.setSrc(profileModel.get("image_url"));

    return profile_window;
}

Ext.define('Console.controller.Profiles', {
    extend: 'Ext.app.Controller',
    views: ['ProfileList', 'Profile', 'Contact'],
    stores: ['ProfileStore', 'ContactsStore', 'ContactLinksStore', 'GroupsStore'],
    models: ['Profile'],
    init: function() {
        this.control({
            'viewport > profilelist': {
                itemdblclick: this.editProfile
            },
            'profilelist button[action=new]':{
                click: this.createProfile
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
            'ContactLinkWindow button[action=add_contact_end]':{
                click:this.addContact
            }
        });
        Ext.widget('profilelist').getStore().load();
    },
    // обновление
    updateProfile: function(button) {
        var win    = button.up('window');
        var form   = win.down('form');
        var values = form.getValues();
        var record = form.getRecord();
        if (record == undefined) {
            Ext.Ajax.request({
                url: 'profile/create',
                jsonData: values,
                success: function(response){
                    var data=Ext.decode(response.responseText);
                    if(data.success){
                        var store = Ext.widget('profilelist').getStore();
                        store.load();
                        view.hide();
                    }
                    else{
                        Ext.Msg.alert('Обновление','Что-то пошло не так...');
                    }
                }
            })
        } else{
            var id = record.get('id');
            values.id=id;
            cntcts = [];
            Ext.each(record.contacts().data.items, function(item){
                cntcts.push(item.getData());
            });
            values.contacts = cntcts;
            Ext.Ajax.request({
                url: 'profile/update',
                jsonData: values,
                success: function(response){
                    var data=Ext.decode(response.responseText);
                    if(data.success){
                        var store = Ext.widget('profilelist').getStore();
                        store.load();
                        view.hide();
                    }
                    else{
                        Ext.Msg.alert('Обновление','Что-то пошло не так...');
                    }
                }
            });
        }
    },
    // создание
    createProfile: function(button) {
        view = Ext.widget('profilewindow');
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
                            if(data.success){
                                var store = Ext.widget('profilelist').getStore();
                                var record = store.getById(id);
                                store.remove(record);
                                view.hide()
                            }
                            else{
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

    showContactForm: function(button, record){
        var win    = button.up('window');
        console.log("scf: ",win);
        var c_view = Ext.widget("contactWindow", {"parent":win});
        if (!(record instanceof Ext.EventObjectImpl)){
            var c_form = c_view.down("form");
            c_form.loadRecord(record);
            console.log(c_form);
            var cl_grid = c_form.getComponent("profile_contact_links");
            cl_grid.reconfigure(record.links());

        }
        c_view.show();
    },

    addContact:function(button){
        var win    = button.up('window');
        var form   = win.down('form');
        var contact_model = form.getRecord();
        var profile_win = win.getParent();
        var profile_form = profile_win.down("form");
        var profile_model = profile_form.getRecord();
        if (profile_model == undefined){
            var values = profile_form.getValues();
            values.id = guid();
            var store = Ext.widget('profilelist').getStore();
            profile_model = Ext.create("Console.model.Profile", values);
            store.add(profile_model);
            var contacts_grid = profile_form.getComponent('profile_contacts');
            contacts_grid.reconfigure(profile_model.contacts());
            profile_form.loadRecord(profile_model);
        }
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

    deleteContact: function(button){
        var row_id = button.rowValues.recordId,
        store = button.up('form').getRecord().contacts(),
        record = store.getById(row_id);
        store.remove(record);

    }

});
