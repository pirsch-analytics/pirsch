import {getScript, ignore} from "./common";

(function () {
    "use strict";

    const script = getScript("#pirschjs");

    if(ignore(script)) {
        return;
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
