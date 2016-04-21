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

		this.items = [{
			xtype:"form",
			items:[
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