Ext.define("Console.model.ContactLink",{
	extend:"Ext.data.Model",
	idProperty:'id',
	fields:[
	'id',
	'type',
	'value',
	'description',
	{name:'order_number', type:'int'}
	], 
	associations: [{
		type: 'belongsTo',
		model: 'Console.model.Contact',
		name: 'contact'
	}]
});