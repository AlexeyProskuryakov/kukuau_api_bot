var view = undefined;

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


Ext.define('Console.controller.Profiles', {
    extend: 'Ext.app.Controller',
    
    views: ['ProfileList', 'Profile', 'Contact'],
    stores: ['ProfileStore'],
    models: ['Profile'],
    init: function() {
        console.log("controller init", this);
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
            'contactwindow button[action=add_contact_end]':{
                click:this.addContact
            }
        });
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
                params: values,
                success: function(response){
                    var data=Ext.decode(response.responseText);
                    if(data.success){
                        var store = Ext.widget('profilelist').getStore();
                        store.load();
                        Ext.Msg.alert('Обновление',data.message);
                    }
                    else{
                        Ext.Msg.alert('Обновление','Не удалось обновить книгу в библиотеке');
                    }
                }
            })
        } else{
            var id = record.get('id');
            values.id=id;
            Ext.Ajax.request({ //todo this is update and create....
                url: 'app/data/update.php',
                params: values,
                success: function(response){
                    var data=Ext.decode(response.responseText);
                    if(data.success){
                        var store = Ext.widget('profilelist').getStore();
                        store.load();
                        Ext.Msg.alert('Обновление',data.message);
                    }
                    else{
                        Ext.Msg.alert('Обновление','Не удалось обновить книгу в библиотеке');
                    }
                }
            });
        }
    },
    // создание
    createProfile: function(button) {
        view = Ext.widget('profilewindow');
    },
    // удаление
    deleteProfile: function(button) {
        var win    = button.up('window'),
        form   = win.down('form'),
        id = form.getRecord().get('id');
        Ext.Ajax.request({
            url: 'app/data/delete.php',
            params: {id:id},
            success: function(response){
                var data=Ext.decode(response.responseText);
                if(data.success){
                    Ext.Msg.alert('Удаление',data.message);
                    var store = Ext.widget('profilelist').getStore();
                    var record = store.getById(id);
                    store.remove(record);
                    form.getForm.reset();
                }
                else{
                    Ext.Msg.alert('Удаление','Не удалось удалить книгу из библиотеки');
                }
            }
        });
    },
    clearForm: function(grid, record) {
        view.down('form').getForm().reset();
    },
    editProfile: function(grid, record) {
        view = Ext.widget('profilewindow');
        view.down('form').loadRecord(record);
    },

    showContactForm: function(button){
        var win    = button.up('window');
        var form   = win.down('form');
        var c_view = Ext.widget("contactwindow", {"parent":win});
    },

    addContact:function(button){
        var win    = button.up('window');
        var form   = win.down('form');
        var parent = win.getParent();
        var record = parent.down("form").getRecord();
        console.log(record);

        
    }

});
