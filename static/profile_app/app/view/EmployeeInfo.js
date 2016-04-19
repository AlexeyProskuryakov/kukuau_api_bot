Ext.define('Console.view.EmployeeInfo', {
	extend: 'Ext.window.Window',
	alias: 'widget.employeeWindow',
	title: 'Информация о сотруднике',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined,
		profileId:undefined
	},
	initComponent: function() {
		console.log("init employee window");
		store = Ext.create('Ext.data.Store', {
			fields: ['role_name', 'role_id'],
			proxy: {
                type: 'ajax',
                url: '/profile/roles/'+this.getProfileId(),
                reader: {
                    type: 'json',
                    root: 'roles',
                    successProperty: 'success'
                }
            }
		});
		store.load();
		this.items = [{
			xtype:"form",
			items:[
			{
				xtype: 'combobox',
				displayField: 'role_name',
				valueField: 'role_id',
				fieldLabel: 'Роль',
				queryMode: 'local',
				store:store,
				width: 350,
				padding:10,
				allowBlank:false,
				itemId:'role'
			},
			{
				xtype: 'textfield',
				name : 'phone',
				fieldLabel: 'Номер телефона',
				width: 550,
				padding:10,
				allowBlank:false,
				itemId:'phone',
				vtype:'phoneNumber'
			}]	
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_employee_end'	
		}];
		this.callParent(arguments);
	}
});