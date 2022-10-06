(function () {
    "use strict";

    // respect Do-Not-Track
    if (navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    // ignore script on localhost
    const script = document.querySelector("#pirschjs");
    const dev = script.hasAttribute("data-dev");

    if (!dev && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.warn("Pirsch ignores hits on localhost. You can enable it by adding the data-dev attribute.");
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

    // get custom attributes
    const attributes = script.getAttributeNames();
    let params = "";

    for (let i = 0; i < attributes.length; i++) {
        if (attributes[i].toLowerCase().startsWith("data-param-")) {
            params += "&" + attributes[i].substring("data-param-".length) + "=" + script.getAttribute(attributes[i]);
        }
    }

    // register hit function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch";
    const clientID = script.getAttribute("data-client-id") || 0;
    const domains = script.getAttribute("data-domain") ? script.getAttribute("data-domain").split(",") || [] : [];
    const disableQueryParams = script.hasAttribute("data-disable-query");
    const disableReferrer = script.hasAttribute("data-disable-referrer");
    const disableResolution = script.hasAttribute("data-disable-resolution");

    function hit() {
        sendHit();

        for (let i = 0; i < domains.length; i++) {
            sendHit(domains[i]);
        }
    }

    function sendHit(hostname) {
        if (!hostname) {
            hostname = location.href;
        } else {
            hostname = location.href.replace(location.hostname, hostname);
        }

        if (disableQueryParams) {
            hostname = (hostname.includes('?') ? hostname.split('?')[0] : hostname);
        }

        const url = endpoint +
            "?nc=" + new Date().getTime() +
            "&client_id=" + clientID +
            "&url=" + encodeURIComponent(hostname.substring(0, 1800)) +
            "&t=" + encodeURIComponent(document.title) +
            "&ref=" + (disableReferrer ? '' : encodeURIComponent(document.referrer)) +
            "&w=" + (disableResolution ? '' : screen.width) +
            "&h=" + (disableResolution ? '' : screen.height) +
            params;
        const req = new XMLHttpRequest();
        req.open("GET", url);
        req.send();
    }

    if (history.pushState) {
        const pushState = history["pushState"];

        history.pushState = function () {
            pushState.apply(this, arguments);
            hit();
        }

        window.addEventListener("popstate", hit);
    }

    if (!document.body) {
        window.addEventListener("DOMContentLoaded", hit);
    } else {
        hit();
    }
})();
