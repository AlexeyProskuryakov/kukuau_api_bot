Ext.define('Console.view.NewGroupAdd', {
	extend: 'Ext.window.Window',
	alias: 'widget.newGroupWindow',
	title: 'Новая группа',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init new group window");
		this.items= [{
			xtype:"form",
			items:[
       		{
				xtype: 'textfield',
				name : 'name',
				fieldLabel: 'Имя',
				width: 550,
				padding:10
			},{
				xtype: 'textfield',
				name : 'description',
				fieldLabel: 'Описание',
				width: 550,
				padding:10
			}
			]
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_global_group_end'
		}];

		this.callParent(arguments);
	}
	

});