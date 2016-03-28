Ext.define('Console.view.Contact', {
	extend: 'Ext.window.Window',
	alias: 'widget.contactWindow',
	title: 'Контакт',
	layout: 'fit',
	autoShow: false,
	width:900,
	height:900,
	config:{
		parent:undefined,

	},
	initComponent: function() {
		console.log("init contact window");
		var me = this;
		var geocoder = new google.maps.Geocoder();

		this.items= [{
			xtype:"form",
			items: [ {
				xtype:"textfield",
				itemId:'address',
				name:"address",
				fieldLabel:"Адрес",
				width: 750,
				padding:10
			},{	
				xtype:"textfield",
				name:"description",
				fieldLabel:"Примечание",
				width: 750,
				padding:10
			},{	
				xtype:"textfield",
				name:"lat",
				itemId:"lat",
				fieldLabel:"Долгота",
				width: 250,
				padding:10
			},{	
				xtype:"textfield",
				itemId:'lon',
				name:"lon",
				fieldLabel:"Широта",
				width: 250,
				padding:10
			},

			{
				xtype: 'gmappanel',
				itemId: 'contact_map',
				zoomLevel: 14,
				width:880,
				height:400,
				gmapType: 'map',
				mapConfOpts: ['enableScrollWheelZoom','enableDoubleClickZoom','enableDragging'],
				mapControls: [],
				maplisteners: {
					click: function(mevt){
						console.log(mevt);
						var lat = mevt.latLng.lat(),
						lon = mevt.latLng.lng(),
						geocdr = new google.maps.Geocoder;

						me.down('form').getComponent("lat").setValue(lat);
						me.down('form').getComponent("lon").setValue(lon);

						geocdr.geocode({location:mevt.latLng}, function(results, status){
							console.log(results, status);
							if (status === google.maps.GeocoderStatus.OK) {	
								me.down('form').getComponent('address').setValue(results[0].formatted_address);
							} else {
								console.log("failed locate because status is ", status)
							}
						});
						this.addMarker(mevt.latLng, mevt.latLng, true);
						
						console.log(me, lat, lon);
					}
				}
				
			},
			{
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
				],

			}
			]}
			];
			this.buttons = [{
				text: 'Сохранить',
				scope: this,
				action: 'save_contact'
			},{
				text:"Добавить связь",
				action:"add_contact_link",
				scope: this,
			}];
			this.callParent(arguments);
		}
	});