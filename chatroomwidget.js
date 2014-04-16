"use strict";

var debug;

// appended odd string to avoid shadowing existing vars
function chatroomwidget_88ed71a($) {
    function isScrolledToBottom(el) {
        return el.scrollHeight === (el.offsetHeight + el.scrollTop);
    }
    function scrollToBottom(el) {
        el.scrollTop = el.scrollHeight;
    }

    var id = "chatroomwidget_88ed71a";
    if ($ === undefined) {
        throw "chatroomwidget.js requires jQuery";
    }
    var ws;
    var $input = $("<form style='flex-shrink: 0'>")
        .append($("<input id="+id+"input style='width: 100%'>").click(function (e) {
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
                "font-size": '6pt',
            })
            .find('> :not(#'+id+'messages)')
                .hide();
        $('#'+id+'messages').css('overflow', 'hidden');
    }
    function toMaxi(node) {
        $(node)
            .removeClass("mini")
            .css({
                width: "250px",
                height: "200px",
                "font-size": '12pt',
            })
            .find('> :not(#'+id+'messages)')
                .show();
        scrollToBottom($('#'+id+'messages').css({
            "overflow-y": "auto",
            "overflow-x": "hidden",
        })[0]);
    }
    var $node = $('<div id=' + id + '>')
        .append($("<div id="+id+"messages>").css({
                'white-space': 'pre-wrap',
            })
        ).append($input)
        .css({
            position: "absolute",
            right: "50px",
            top: "50px",
            border: "solid thin black",
            "background-color": "white",
            "z-index": 1000,
            display: "flex",
            "flex-direction": "column",
            "justify-content": "flex-end",
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
        var el = document.getElementById(id+'messages');
        var atbottom = isScrolledToBottom(el);
        $('<div>').text(evt.data).appendTo(el);
        if (atbottom || $('#'+id).hasClass("mini")) {
            scrollToBottom(el);
        }
    };
    debug = {ws: ws};
}

if (typeof define === "function" && define.amd) {
    define(["jquery"], chatroomwidget_88ed71a);
} else {
    chatroomwidget_88ed71a(jQuery);
}
