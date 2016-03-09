var fileUploadPhoto = {
            xtype: 'fileuploadfield',
            id: 'photo',
            buttonOnly: false,
            buttonText: "123",
            fieldLabel: 'Image',
            name: 'photo' ,
            buttonCfg: {
                text: '',
                iconCls: 'upload-icon'
            }
        };

        
var fileUploadPhotoPreview = new Ext.Component({
    autoEl: { 
        tag: 'img', src: 'animal.gif', id: 'photoPreview'
    }
});


var mydata = new Ext.FormPanel({
        labelWidth:100,
        url: 'xyz.php',
        fileUpload: true,
        monitorValid:true,
        xtype:"form",
        frame: true,
        border: false,
        autoScroll: true, 
        items: {
            xtype: 'tabpanel',
            activeTab: 0,
 
            defaults: { bodyStyle: 'padding:10px' },
            items: [{ 
                id: 'tabImages',
                title: 'My Imgae',
                layout: 'form',
                autoScroll: true, 
                items: [
                fileUploadPhoto,fileUploadPhotoPreview
                ]}
        ] //TAB END
    
                    
        }, //Items End
        buttons: [{
            text: 'Save',
            id: "buttonOK",
            type: "submit",
            formBind:true,
            handler: function() {
                mydata.getForm().submit({
                    waitMsg:'Saving...',
                    reset:false,               
                    params: {
                        moreinfo: 'moreparams'
                    },
                    success: function(form, action){                                                                           
                            oldPreview = fileUploadPhotoPreview.autoEl.src;
                            newPhoto = "butterfly.jpg";
                            fileUploadPhotoPreview.el.dom.src=newPreview;                                
                            mydata.findById('photo').reset();                                
                            alert ("ok");
                        }
                        
                     
                    },
                    failure: function( form, action){
                        alert ("error");
                    }
                })
            }
        }]
        
    });