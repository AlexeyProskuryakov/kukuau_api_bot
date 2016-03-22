Ext.define('Console.view.Contact', {
	extend: 'Ext.window.Window',
	alias: 'widget.contactWindow',
	title: 'Контакт',
	layout: 'fit',
	autoShow: false,
	width:900,
	height:600,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact window");
		var me = this;
		this.items= [{
			xtype:"form",
			items: [ {
				xtype:"textfield",
				name:"address",
				fieldLabel:"Адресс"
			},{	
				xtype:"textfield",
				name:"description",
				fieldLabel:"Описание"
			},{
				xtype:"grid",
				title:"Способы связи",
				itemId:"profile_contact_links",
				store: 'ContactLinksStore',
				columns:[
				{header: 'Тип',  dataIndex: 'type'},
				{header: 'Значение', dataIndex: 'value', flex:1},
				{header: 'Описание', dataIndex: 'description', flex:1},
				{
					xtype : 'actioncolumn',
					header : 'Delete',
					width : 100,
					align : 'center',
					action:"delete_contact_link",
					items : [
					{
						icon:'img/delete-icon.png',
						tooltip : 'Delete',
						scope : me
					}]
				}
				]
			}
			]}

		// {
		// 	xtype: 'gmappanel',
		// 	id: 'contactAddressMap',
		// 	zoomLevel: 14,
		// 	gmapType: 'map',
		// 	mapConfOpts: ['enableScrollWheelZoom','enableDoubleClickZoom','enableDragging'],
		// 	mapControls: ['GSmallMapControl','GMapTypeControl'],
		// 	setCenter: {
		// 		lat: 39.26940,
		// 		lng: -76.64323
		// 	},
		// 	maplisteners: {
		// 		click: function(mevt){
		// 			Ext.Msg.alert('Lat/Lng of Click', mevt.latLng.lat() + ' / ' + mevt.latLng.lng());
		// 			var input = Ext.get('ac').dom,
		// 			sw = new google.maps.LatLng(39.26940,-76.64323),
		// 			ne = new google.maps.LatLng(39.38904,-76.54848),
		// 			bounds = new google.maps.LatLngBounds(sw,ne);
		// 			var options = {
		// 				location: mevt.latLng,
		// 				radius: '1000',
		// 				types: ['geocode']
		// 			};
		// 		}
		// 	}
		// },
		];
		this.dockedItems=[{
			xtype:'toolbar',
			docked: 'top',
			items: [
			{
				text:'Очистить',
				action: 'clear'
			}
			]
		}];

		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'save'
		},{
			text:'Удалить',
			action: 'delete',
			scope: this
		},
		{
			text:"Добавить связь",
			action:"add_contact_start",
			scope: this,
		}

		];

		this.callParent(arguments);
	}
});