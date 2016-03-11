Ext.define("Console.model.Contact",{
	extend:"Ext.data.Model",
	fields:[
	'id',
	'type',
	'value',
	'description'
	],
	 // proxy: {
  //                  type: 'ajax',
  //                  api: {
  //                           read: '/profile/contact/read',
  //                           create: '/profile/contact/create',
  //                           update: '/profile/contact/update',
  //                           destroy: '/profile/contact/destroy'
  //                      }
  //           },
	belongsTo: {model:'Console.model.Profile', name:'profile'}
});