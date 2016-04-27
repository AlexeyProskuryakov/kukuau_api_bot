function call_message_func(action, context_str, message_id){
	var context = JSON.parse(context_str);
	data = {action:action, context:context, message_id:message_id};
	$.ajax({type:"POST",
		url:            url_prefix+"/message_function",
		contentType:    'application/json',
		data:           JSON.stringify(data),
		dataType:       'json',
		success:        function(x){
			if (x.ok==true){
				if (action != 'start'){
					$("#"+message_id).remove();
				}
				$("#state-"+message_id).text(x.result);
			}
		}
	});
}

function pasteNewOrder(message){
	var text_message =  "<div class='msg' id={{MessageID}}>"+
	"<h4 class='media-heading'><a href='"+url_prefix+"?with={{From}}'>{{FromName}}</a> "+
	"<small class='time'>{{time}}</small>"+
	"</h4>"+
	"<div class='msg-with-data'>"+
	"<h4>{{Body}}</h4>"+
	"<table class='table table-condensed table-bordered table-hover table-little-text'>"+
	"{{#AdditionalData}}"+
	"{{#Value}}"+
	"<tr><td>{{Name}}</td><td>{{Value}}</td></tr>"+
	"{{/Value}}"+
	"{{/AdditionalData}}"+
	"</table>"+
	"{{#AdditionalFuncs}}"+
	" <button class='btn btn-default btn-sm' onclick='call_message_func(\"{{Action}}\", \"{{&Context}}\", \"{{MessageID}}\")'> {{Name}} </button> "+
	"{{/AdditionalFuncs}}"+
	"<div class='status'><h5> Статус: <big id='state-{{MessageID}}'>{{RelatedOrderState}}</big></h5></div>"+
	"</div>"+
	"</div>";
	for (var i=0; i < message.AdditionalFuncs.length; i++){
		message.AdditionalFuncs[i].Context = JSON.stringify(message.AdditionalFuncs[i].Context).replace(/\"/gi,"\\x22");
	}
	var result = Mustache.render(text_message, message);
	$("#orders-wrapper").append(result);
}

function update_orders(){
	var except = [];
	$(".msg").each(function(x,item){
    	except.push(item.getAttribute('id'));
	});
	var data = {except:except}
	$.ajax({
		type:"POST",
		url: url_prefix+"/order_page_supply",
		dataType:       'json',
		data:           JSON.stringify(data),
		success:        function(x){
			if (x.ok==true){
				x.order_messages.forEach(function(message){
					if ($("#"+message.MessageID).length != 0) {
						return;
					}			
					pasteNewOrder(message);	
				});
			}
		}
	});
}

setInterval(function(){
	update_orders();
	return true;
}, 5000);