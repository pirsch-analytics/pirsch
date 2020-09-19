(function() {
    "use strict";

    var script = document.querySelector("#pirschjs");
    var endpoint = script.getAttribute("data-endpoint") || "/pirsch";
    var tenantID = script.getAttribute("data-tenant-id") || 0;
    var trackLocalhost = script.hasAttribute("data-track-localhost");

    if(!trackLocalhost && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(window.location.hostname) || window.location.protocol === "file:")) {
        console.warn("Pirsch ignores hits on localhost. You can enable it by adding the data-track-localhost attribute.");
        return;
    }

    var attributes = script.getAttributeNames();
    var params = "";

    for(let i = 0; i < attributes.length; i++) {
        if(attributes[i].toLowerCase().startsWith("data-param-")) {
            params += "&"+attributes[i].substr("data-param-".length)+"="+script.getAttribute(attributes[i]);
        }
    }

    function hit() {
        var nocache = new Date().getTime();
        var location = window.location;
        var referrer = document.referrer;
        var width = window.screen.width;
        var height = window.screen.height;
        var url = endpoint+
            "?nocache="+ nocache+
            "&tenantid="+tenantID+
            "&location="+location+
            "&referrer="+referrer+
            "&width="+width+
            "&height="+height+
            params;

        var req = new XMLHttpRequest();
        req.open("GET", url);
        req.send();
    }

    if(window.history.pushState) {
        var pushState = window.history["pushState"];

        window.history.pushState = function() {
            pushState.apply(this, arguments);
            hit();
        }

        window.addEventListener("popstate", hit)
    }

    if(!document.body) {
        window.addEventListener("DOMContentLoaded", hit);
    } else {
        hit();
    }
})();
