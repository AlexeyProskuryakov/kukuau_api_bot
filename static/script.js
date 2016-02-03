var messages_updated = Math.round( Date.now() / 1000 );
var contacts_updated = Math.round( Date.now() / 1000 );


var message_for = $("#with").prop("value");

function paste_message(message){
    var text_message = "<div class='media msg'><div class='media-body'><small class='pull-right time'><i class='fa fa-clock-o'></i>{{Time}}</small><h5 class='media-heading'>{{From}}</h5><small class='col-lg-10'>{{Body}}</small></div></div>";
    var result = Mustache.render(text_message,message);
    $("#chat-wrapper").prepend(result);
}

function update_messages(){
    data = {m_for: message_for, after:messages_updated}
    $.ajax({type:"POST",
        url:            "/new_messages",
        contentType:    'application/json',
        data:           JSON.stringify(data),
        dataType:       'json',
        success:        function(x){
            x.messages.forEach(function(message){
                paste_message(message);
            });
            messages_updated = x.next_;

        }
    });
    return true;
}


function set_contact_new_message(contact_id, count){
    var c_w = $("#s-"+contact_id);
    console.log("c_w: ",c_w);
    c_w.text("("+count+")");
}

function add_new_contact(contact){
    var c_text = "<div class='contat' id='{{ID}}'><a class='btn btn-primary btn-med' href='/chat?with={{ID}}'> {{#IsTeam}} Команда {{/IsTeam}}{{Name}} ({{NewMessagesCount}})</a></div>";
    var result = Mustache.render(c_text, contact);
    console.log(result);
    $("#contacts").append(result);
}

function update_contacts(){
    var exists = $(".contact");
    var ex_values = new Array();
    for (var k in exists){
        if (exists[k]["id"] != undefined){
            ex_values.push(exists[k]["id"]);
        }
    }
    console.log("exist:",ex_values)
    data = {after: contacts_updated, exist:ex_values}
    $.ajax({type:"POST",
        url:            "/new_contacts",
        contentType:    'application/json',
        data:           JSON.stringify(data),
        dataType:       'json',
        success:        function(x){
            x.old.forEach(function(c){
                console.log("old:",c);
                set_contact_new_message(c.ID, c.NewMessagesCount);
            });
            x['new'].forEach(function(c){
                console.log("new",c);
                add_new_contact(c);
            });
            // contacts_updated = x.next_;
        }
    });
    return true;
}

$("#chat-form-message").keydown(function(e){
    if (e.ctrlKey && e.keyCode == 13) {
        $("#chat-form").submit();
    }
});
setInterval(function(){
    // console.log("will ask about new messages");
    update_messages();
    update_contacts();
    return true;
}, 5000);




