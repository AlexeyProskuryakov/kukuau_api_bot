Ext.define('Console.view.Notification', {
	extend: 'Ext.window.Window',
	alias: 'widget.notificationWindow',
	title: 'Отправка нотификации сотрудникам',
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
				xtype: 'numberfield',
				name : 'after_min',
				fieldLabel: 'Через (количество минут)',
				width: 250,
				padding:10,
				allowBlank:false,
				itemId:'after_min',
			},
			{
				xtype: 'textfield',
				name : 'text',
				fieldLabel: 'Сообщение',
				width: 550,
				padding:10,
				allowBlank:false,
				itemId:'text',
			}
			]	
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_notification_end'	
		}];
		this.callParent(arguments);
	}
});