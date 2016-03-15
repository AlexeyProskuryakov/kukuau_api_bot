Ext.define('Console.view.Contact', {
	extend: 'Ext.form.Panel',
	alias: 'widget.contactPanel',
	header:false,
	title: 'Контакт',
	layout: 'fit',
	autoShow: false,
	width:400,
	height:400,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact window");

		this.items= [
		{
			xtype:"textfield",
			name:"address",
			fieldLabel:"Адресс:"
		},{
			xtype: 'gmappanel',
			id: 'contactAddressMap',
			zoomLevel: 14,
			gmapType: 'map',
			mapConfOpts: ['enableScrollWheelZoom','enableDoubleClickZoom','enableDragging'],
			mapControls: ['GSmallMapControl','GMapTypeControl'],
			setCenter: {
				lat: 39.26940,
				lng: -76.64323
			},
			maplisteners: {
				click: function(mevt){
					Ext.Msg.alert('Lat/Lng of Click', mevt.latLng.lat() + ' / ' + mevt.latLng.lng());
					var input = Ext.get('ac').dom,
					sw = new google.maps.LatLng(39.26940,-76.64323),
					ne = new google.maps.LatLng(39.38904,-76.54848),
					bounds = new google.maps.LatLngBounds(sw,ne);
					var options = {
						location: mevt.latLng,
						radius: '1000',
						types: ['geocode']
					};
				}
			}
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
		}];
		
		this.callParent(arguments);
	}
});