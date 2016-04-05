Ext.define('Console.view.GroupChoose', {
	extend: 'Ext.window.Window',
	alias: 'widget.groupWindow',
	title: 'Группы в которые будет входить профайл',
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
				xtype: "grid",
				markDirty:false,
				title: "Выберите группу или добавьте новую",
				alias: "widget.groupGlobalGrid",
				itemId: "choose_group_grid",
				store: 'GroupsGlobalStore',
				columns: [{
					header: "Имя группы",
					dataIndex: 'name',
					flex: 1
				}, {
					header: "Описание",
					dataIndex: 'description',
					flex: 1
				}, { 
					xtype : 'checkcolumn', 
					text : 'Выбрать',
					dataIndex:'_active'
				}],
			},
			]
		}];
		this.buttons = [{
			text: 'OK',
			scope: this,
			action: 'add_group_end'
		},
		{
			text: "Добавить новую группу",
			action: "add_global_group_start",
			scope: this,
		}
		];

		this.callParent(arguments);
	}
	

});
