Ext.define('Console.view.GroupChoose', {
	extend: 'Ext.window.Window',
	alias: 'widget.groupWindow',
	title: 'Имя группы в которую входит профайл',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init group window");
		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: 'textfield',
				name : 'value',
				fieldLabel: 'Описание',
				itemId: 'description',
				width: 550,
				padding:10,
			}
			]
		}];
		this.buttons = [{
			text: 'OK',
			scope: this,
			action: 'add_group_end'
		}];

		this.callParent(arguments);
	}
	

});