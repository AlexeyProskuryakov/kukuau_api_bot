Ext.define("Console.model.ContactLink",{
	extend:"Ext.data.Model",
	idProperty:'id',
	fields:[
	'id',
	'type',
	'value',
	'description',
	'order_number'
	], 
	associations: [{
		type: 'belongsTo',
		model: 'Console.model.Contact',
		name: 'contact'
	}]
});