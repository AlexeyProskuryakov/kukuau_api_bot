Ext.define("Console.model.Contact",{
	extend:"Ext.data.Model",
	idProperty:'id',
	fields:[
	{name:'id', type:'int'},
	'address',
	'description',
	{name:'lat', mapping:'geo.lat', type:'float'},
	{name:'lon', mapping:'geo.lon', type:'float'},
	'order_number'
	],
	associations: [{
		type: 'hasMany',
		model: 'Console.model.ContactLink',
		name: 'links'
	}]
});