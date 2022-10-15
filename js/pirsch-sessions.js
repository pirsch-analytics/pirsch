(function () {
    "use strict";

    // respect Do-Not-Track
    if (navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    // ignore script on localhost
    const script = document.querySelector("#pirschsessionsjs");
    const dev = script.hasAttribute("data-dev");

    if (!dev && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.warn("Pirsch ignores sessions on localhost. You can enable it by adding the data-dev attribute.");
        return;
    }

    // include pages
    try {
        const include = script.getAttribute("data-include");
        const paths = include ? include.split(",") : [];

        if (paths.length) {
            let found = false;

            for (let i = 0; i < paths.length; i++) {
                if (new RegExp(paths[i]).test(location.pathname)) {
                    found = true;
                    break;
                }
            }

            if (!found) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
    }

    // exclude pages
    try {
        const exclude = script.getAttribute("data-exclude");
        const paths = exclude ? exclude.split(",") : [];

        for (let i = 0; i < paths.length; i++) {
            if (new RegExp(paths[i]).test(location.pathname)) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
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
