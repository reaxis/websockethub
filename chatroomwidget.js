// Copyright © 2014 Hraban Luyat <hraban@0brg.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

"use strict";

(function () {

    // build the chat widget
    function chatroomwidget_main($) {
        function myloc() {
            return $('[rel=canonical]').prop('href')
                || (window.location.origin + window.location.pathname);
        }
        function getWebsocketUrl() {
            if (typeof chatroomwidgetServerUrl !== "undefined") {
                return chatroomwidgetServerUrl;
            }
            // Include protocol to prevent leaking between HTTP and HTTPS.
            // Exclude querystring & anchor to normalize for sessions &c
            // some naive normalization, too, why not (I know why not)
            var norm = myloc().replace(/\/+/, '/');
            return "ws://websockethub.com/" + encodeURIComponent(norm);
        }
        function isScrolledToBottom(el) {
            return el.scrollHeight === (el.offsetHeight + el.scrollTop);
        }
        function scrollToBottom(el) {
            el.scrollTop = el.scrollHeight;
        }
        var oldTitle = document.title;
        var name;
        var id = "chatroomwidget_88ed71a";
        var ws;
        var $input = $("<form style='flex-shrink: 0'>")
            .append($("<input id="+id+"input style='width: 100%'>"))
            .submit(function (e) {
                e.preventDefault();
                var node = document.getElementById(id+"input");
                ws.send(name + ": " + node.value);
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
                    width: "300px",
                    height: "500px",
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
            .append($("<div id="+id+"banner><p>What nickname do you want?<p></div>")
                .css({
                    "background": "white",
                    "text-align": "center",
                    "top": 0,
                    "height": "100%",
                    "position": "absolute",
                    "width": "100%",
                    "z-index": 1001,
                }).append($('<form><input name=name>').submit(function (e) {
                    e.preventDefault();
                    name = this.name.value;
                    $('#'+id+'banner').remove();
                    $('#'+id).find('input')[0].focus();
                })))
            .append($("<div id="+id+"messages>").css({
                    'white-space': 'pre-wrap',
                    'font-family': 'sans-serif',
                    'word-wrap': 'break-word',
                }))
            .append($input)
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
            }).on("click", "input", function () {
                return false;
            }).click(function(e) {
                e.preventDefault();
                if ($(this).hasClass("mini")) {
                    toMaxi(this);
                } else {
                    toMini(this);
                }
            }).appendTo('body');
        toMini($node[0]);
        var ws = new WebSocket(getWebsocketUrl());
        ws.onopen = function onopen() {
        };
        ws.onmessage = function onmessage(evt) {
            var el = document.getElementById(id+'messages');
            var atbottom = isScrolledToBottom(el);
            var i = evt.data.indexOf(':');
            var name = evt.data.slice(0, i + 1);
            var msg = evt.data.slice(i + 1);
            $('<div>')
                .text(msg)
                .prepend($('<span class=name style="font-weight: bold;">').text(name))
                .appendTo(el);
            if (atbottom || $('#'+id).hasClass("mini")) {
                scrollToBottom(el);
            }
            if (document.hidden) {
                document.title = "new messages - " + oldTitle;
            }
        };
        document.addEventListener("visibilitychange", function () {
            if (!document.hidden) {
                document.title = oldTitle;
            }
        });
    }

    // Load chatroomwidget if jQuery exists, exponentially back off if it
    // doesn't and retry until the wait time is longer than one second (total
    // wait time is cumulative). All numbers were chosen because they felt
    // right. The reason for doing this is allowing this script to be included
    // before the jquery library and still work. If jquery just doesn't load,
    // I'll load it myself.
    function chatroomwidget_loader(retry_ms) {
        if (typeof jQuery === 'undefined') {
            if (retry_ms === undefined) {
                throw "failed to load jquery";
            }
            // try again a couple times
            if (retry_ms > 1000) {
                var el = document.createElement("script");
                el.src = "http://static.websockethub.com/jquery-2.1.0.min.js";
                el.onload = chatroomwidget_loader;
                document.body.appendChild(el);
            } else {
                setTimeout(function () {
                    chatroomwidget_loader(retry_ms * 1.4);
                }, retry_ms);
            }
        } else {
            chatroomwidget_main(jQuery);
        }
    }

    if (typeof define === "function" && define.amd) {
        define("chatroomwidget", ["jquery"], chatroomwidget_main);
    } else {
        chatroomwidget_loader(20);
    }

})();
