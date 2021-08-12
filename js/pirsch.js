(function() {
    "use strict";

    if(navigator.doNotTrack == "1") {
        return;
    }

    var script = document.querySelector("#pirschjs");
    var endpoint = script.getAttribute("data-endpoint") || "/pirsch";
    var clientID = script.getAttribute("data-client-id") || 0;
    var trackLocalhost = script.hasAttribute("data-track-localhost");

    if(!trackLocalhost && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.warn("Pirsch ignores hits on localhost. You can enable it by adding the data-track-localhost attribute.");
        return;
    }

    var attributes = script.getAttributeNames();
    var params = "";

    for(var i = 0; i < attributes.length; i++) {
        if(attributes[i].toLowerCase().startsWith("data-param-")) {
            params += "&"+attributes[i].substr("data-param-".length)+"="+script.getAttribute(attributes[i]);
        }
    }

    function hit() {
        var url = endpoint+
            "?nc="+ new Date().getTime()+
            "&client_id="+clientID+
            "&url="+location.href.substr(0, 1800)+
            "&t="+document.title+
            "&ref="+document.referrer+
            "&w="+screen.width+
            "&h="+screen.height+
            params;
        var req = new XMLHttpRequest();
        req.open("GET", url);
        req.send();
    }

    if(history.pushState) {
        var pushState = history["pushState"];

        history.pushState = function() {
            pushState.apply(this, arguments);
            hit();
        }

        window.addEventListener("popstate", hit);
    }

    if(!document.body) {
        window.addEventListener("DOMContentLoaded", hit);
    } else {
        hit();
    }
})();
