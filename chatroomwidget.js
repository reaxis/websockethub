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

    function $(id, parent) {
        return (parent || document).querySelector(id);
    }

    function $all(selector, func, parent) {
        Array.prototype.forEach.call((parent || document).querySelectorAll(selector), func);
    }

    function getWebsocketUrl() {
        if (typeof chatroomwidgetServerUrl !== "undefined") {
            return chatroomwidgetServerUrl;
        }

        // Include protocol to prevent leaking between HTTP and HTTPS.
        // Exclude querystring & anchor to normalize for sessions &c
        var myloc = window.location.origin + window.location.pathname;
        return "ws://websockethub.com/" + encodeURIComponent(myloc);
    }

    // build the chat widget
    function chatroomwidget_main() {
        function isScrolledToBottom(el) {
            return el.scrollHeight === (el.offsetHeight + el.scrollTop);
        }

        function scrollToBottom(el) {
            el.scrollTop = el.scrollHeight;
        }

        var name,
            id = "chatroomwidget_88ed71a",
            ws;

        var $node = document.createElement("div");
        $node.id = id;
        $node.style.position = "absolute";
        $node.style.right = "50px";
        $node.style.top = "50px";
        $node.style.border = "solid thin black";
        $node.style.backgroundColor = "white";
        $node.style.zIndex = 1000;
        $node.style.display = "flex";
        $node.style.flexDirection = "column";
        $node.style.justifyContent = "flex-end";
        $node.addEventListener("click", function(event) {
            event.preventDefault();

            if ($node.className === "mini") {
                toMaxi();
            } else {
                toMini();
            }
        });

        var banner = document.createElement("div");
        banner.id = id + "banner";
        banner.style.background = "white";
        banner.style.textAlign = "center";
        banner.style.top = 0;
        banner.style.height = "100%";
        banner.style.position = "absolute";
        banner.style.width = "100%";
        banner.style.zIndex = 1001;

        var bannerText = document.createElement("p");
        bannerText.appendChild(document.createTextNode("What nickname do you want?"));
        banner.appendChild(bannerText);

        var form2 = document.createElement("form");
        var input2 = document.createElement("input");

        form2.appendChild(input2);

        form2.addEventListener("submit", function(event) {
            event.preventDefault();

            name = input2.value;
            $("#" + id + "banner").parentNode.removeChild($("#" + id + "banner"));
            input2.focus();
        });

        banner.appendChild(form2);

        $node.appendChild(banner);

        var messages = document.createElement("div");
        messages.id = id + "messages";
        messages.style.whiteSpace = "pre-wrap";
        messages.style.fontFamily = "sans-serif";

        $node.appendChild(messages);

        // chat input
        var form = document.createElement("form");
        form.style.flexShrink = 0;

        var input = document.createElement("input");
        input.id = id + "input";
        input.style.width = "100%";

        form.addEventListener("submit", function(event) {
            event.preventDefault();
            ws.send("<" + name + "> " + input.value);
            input.value = "";
        });

        form.appendChild(input);

        $node.appendChild(form);

        // prevent clicking on inputs from minimizing/maximizing
        $all("input", function(el) {
            el.addEventListener("click", function(event) {
                event.stopPropagation();
            });
        }, $node);

        document.body.appendChild($node);

        function toMini() {
            $node.className = "mini";
            $node.style.width = "100px";
            $node.style.height = "100px";
            $node.style.fontSize = "6pt";
            $all("#" + id + " > :not(#" + id + "messages)", function(el) {
                el.style.display = "none";
            });
            messages.style.overflow = "hidden";
        }

        function toMaxi() {
            $node.className = "";
            $node.style.width = "300px";
            $node.style.height = "500px";
            $node.style.fontSize = "12pt";

            $all("#" + id + " > :not(#" + id + "messages)", function(el) {
                el.style.display = "block";
            });
            messages.style.overflowY = "auto";
            messages.style.overflowX = "hidden";

            scrollToBottom(messages);
        }

        toMini($node);

        var ws = new WebSocket(getWebsocketUrl());
        ws.onopen = function() {
        };
        ws.onmessage = function(event) {
            var el = $("#" + id + "messages");

            var message = document.createElement("div");
            message.appendChild(document.createTextNode(event.data));
            el.appendChild(message);

            if (isScrolledToBottom(el) || $("#" + id).className === "mini") {
                scrollToBottom(el);
            }
        };
    }

    // Load chatroomwidget if jQuery exists, exponentially back off if it
    // doesn't and retry until the wait time is longer than one second (total
    // wait time is cumulative). All numbers were chosen because they felt
    // right. The reason for doing this is allowing this script to be included
    // before the jquery library and still work.
    function chatroomwidget_loader() {
        chatroomwidget_main();
    }

    if (typeof define === "function" && define.amd) {
        define("chatroomwidget", [], chatroomwidget_main);
    } else {
        chatroomwidget_loader();
    }

})();
