import {getScript, ignore} from "./common";

(function () {
    "use strict";

    const script = getScript("#pirschsessionsjs");

    if(ignore(script)) {
        return;
    }

    // register session function
    const endpoint = script.getAttribute("data-endpoint") || "/session";
    const clientID = script.getAttribute("data-client-id") || 0;
    const domains = script.getAttribute("data-domain") ? script.getAttribute("data-domain").split(",") || [] : [];

    function extendSession() {
        sendExtendSession();

        for (let i = 0; i < domains.length; i++) {
            sendExtendSession(domains[i]);
        }
    }

    function sendExtendSession(hostname) {
        if (!hostname) {
            hostname = location.href;
        } else {
            hostname = location.href.replace(location.hostname, hostname);
        }

        const url = endpoint +
            "?nc=" + new Date().getTime() +
            "&client_id=" + clientID +
            "&url=" + encodeURIComponent(hostname.substring(0, 1800));
        const req = new XMLHttpRequest();
        req.open("POST", url);
        req.send();
    }

    const interval = Number.parseInt(script.getAttribute("data-interval-ms"), 10) || 60_000;

    const intervalHandler = setInterval(() => {
        extendSession();
    }, interval);

    window.pirschClearSession = () => {
        clearInterval(intervalHandler);
    }
})();
