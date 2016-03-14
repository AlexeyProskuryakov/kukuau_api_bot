document.getElementById( 'chat-end' ).scrollIntoView();

var messages_updated = Math.round( Date.now() / 1000 );
var contacts_updated = Math.round( Date.now() / 1000 );


var message_for = $("#with").prop("value");

function paste_message(message){
    var text_message = "<div class='media msg'><div class='media-body'><small class='pull-right time'><i class='fa fa-clock-o'></i>{{Time}}</small><h5 class='media-heading'>{{From}}</h5><small class='col-lg-10'>{{Body}}</small></div></div>";
    var result = Mustache.render(text_message,message);
    $("#chat-wrapper").append(result);
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
    c_w.text("("+count+")");
}

function add_new_contact(contact){
    console.log("add new contact",contact);
    var c_text = "<div class='contact' id='{{ID}}'><a class='bg-success a-contact' href='/chat?with={{ID}}'> {{#IsTeam}} Команда {{/IsTeam}}{{Name}} <span class='small' id='s-{{ID}}'>({{NewMessagesCount}})<span></a></div>";
    var result = Mustache.render(c_text, contact);
    if (contact.IsTeam == true) {
        $("#team-contacts").prepend(result);
    } else if (contact.IsPassersby == true){
        $("#man-contacts").prepend(result);
    } else {
        $("#contacts").prepend(result);
    }

}

function delete_chat(between){
    $.ajax({
        type:"POST",
        url:"/delete_chat/"+between,
        dataType:"json",
        success: function(x){
            $("#removed").text(x.removed);
            $("#removed").show(500);

        }
    });
}

function update_contacts(){
    var exists = $(".contact");
    var ex_values = new Array();
    for (var k in exists){
        if (exists[k]["id"] != undefined){
            ex_values.push(exists[k]["id"]);
        }
    }
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

    update_messages();
    update_contacts();
    return true;
}, 5000);

$("#chat-form-message").focus();


function delete_all(){
     $.ajax({
        type:"POST",
        url:            "/delete_all",
        success:        function(x){
            if (x.ok == true) {
                console.log(x);
                text = "<div><p class='bg-success'>Удалено шагов: {{steps_removed}}</p><p class='bg-success'>Обновленно пользователей: {{peoples_updated}}</p><p class='bg-success'>Удалено комманд: {{teams_removed}}</p><p class='bg-success'>Удалено сообщений от комманд: {{messages_removed}}</p></div>";
                el = Mustache.render(text, x);
                $("#delete-result").prepend(el);
            }
        }
    });
}

function send_messages_from_klichat(){
    var to_winner = $("#to-winner").val(),
        to_not_winner = $("#to-not-winner").val(),
        winners = [],
        winners_chbx = $(".winner:checked").each(function(x, obj){
            winners.push(obj.id);
        });
    console.log("message for winner: ", to_winner, "to not winner: ", to_not_winner, "winners: ", winners);
    $.ajax({
        type:           "POST",
        url:            "/send_messages_at_quest_end",
        data:           JSON.stringify({to_winner:to_winner, to_not_winner:to_not_winner, winners:winners}),
        dataType:       'json',
        success:        function(x){
                    if (x.ok == true) {
                        console.log(x);
                        text = "<div><p class='bg-success'>Сообщения поставлены в очередь на отправление.</p></div>"
                        el = Mustache.render(text, x);
                        $("#send-result").prepend(el);
                    }
        }
    });
}



