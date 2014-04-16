"use strict";

var debug;

// appended odd string to avoid shadowing existing vars
function chatroomwidget_88ed71a($) {
    var id = "chatroomwidget_88ed71a";
    if ($ === undefined) {
        throw "chatroomwidget.js requires jQuery";
    }
    var ws;
    var $input = $("<form>")
        .append($("<input id="+id+"input>").click(function (e) {
                // don't hide
                return false;
            })
        ).submit(function (e) {
            e.preventDefault();
            var node = document.getElementById(id+"input");
            ws.send(node.value);
            node.value = "";
        });
    function toMini(node) {
        $(node)
            .addClass("mini")
            .css({
                width: "100px",
                height: "100px",
            })
            .find('> :not(#'+id+'messages)')
                .hide();
    }
    function toMaxi(node) {
        $(node)
            .removeClass("mini")
            .css({
                width: "250px",
                height: "500px",
            })
            .find('> :not(#'+id+'messages)')
                .show();
    }
    var $node = $('<div id=' + id + '>')
        .append("<div id="+id+"messages style='white-space: pre-wrap'>")
        .append($input)
        .css({
            position: "absolute",
            right: "50px",
            top: "50px",
            border: "solid thin black",
            "background-color": "white",
            "z-index": 1000,
        }).click(function(e) {
            e.preventDefault();
            if ($(this).hasClass("mini")) {
                toMaxi(this);
            } else {
                toMini(this);
            }
        }).appendTo('body');
    toMini($node[0]);
    var ws = new WebSocket("ws://echo.websocket.org/");
    ws.onopen = function onopen() {
    };
    ws.onmessage = function onmessage(evt) {
        $('<div>').text(evt.data).appendTo('#'+id+'messages');
    };
    debug = {ws: ws};
}

if (typeof define === "function" && define.amd) {
    define(["jquery"], chatroomwidget_88ed71a);
} else {
    chatroomwidget_88ed71a(jQuery);
}
