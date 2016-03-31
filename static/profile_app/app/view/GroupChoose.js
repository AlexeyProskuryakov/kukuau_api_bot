Ext.define('Console.view.Group', {
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
		var store = Ext.create('Ext.data.Store', {
			fields: ['name', 'description'],
			proxy: {
				type: 'ajax',
				url: '/profile/all_groups',
				reader: {
					type: 'json',
					root: 'groups',
					successProperty: 'success'
				}
			}
		});
		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: 'combo',
				name : 'type',
				store:store,
				queryMode: 'local',
				displayField: 'name',
				valueField: 'name',
				fieldLabel: 'Имя группы',
				typeAhead: true,
				typeAheadDelay: 100,
				hideTrigger: true,
				width: 350,
				padding:10
			},
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