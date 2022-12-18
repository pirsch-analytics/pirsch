import {getScript, ignore, rewriteHostname, rewriteReferrer} from "./common";

(function () {
    "use strict";

    // add a dummy function for local development
    window.pirsch = (name, options) => {
        console.log(`Pirsch event: ${name}${options ? " " + JSON.stringify(options) : ""}`);
        return Promise.resolve(null);
    };

    const script = getScript("#pirscheventsjs");

    if(ignore(script)) {
        return;
    }

    // register event function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch-event";
    const clientID = script.getAttribute("data-client-id");
    const domains = script.getAttribute("data-domain") ? script.getAttribute("data-domain").split(",") || [] : [];
    const disableQueryParams = script.hasAttribute("data-disable-query");
    const disableReferrer = script.hasAttribute("data-disable-referrer");
    const disableResolution = script.hasAttribute("data-disable-resolution");
    const rewrite = script.getAttribute("data-dev");

    window.pirsch = function (name, options) {
        if (typeof name !== "string" || !name) {
            return Promise.reject("The event name for Pirsch is invalid (must be a non-empty string)! Usage: pirsch('event name', {duration: 42, meta: {key: 'value'}})");
        }

        return new Promise((resolve, reject) => {
            const meta = options && options.meta ? options.meta : {};

            for (let key in meta) {
                if (meta.hasOwnProperty(key)) {
                    meta[key] = String(meta[key]);
                }
            }

            sendEvent(rewrite, name, options, meta, resolve, reject);

            for (let i = 0; i < domains.length; i++) {
                sendEvent(domains[i], name, options, meta, resolve, reject);
            }
        });
    }

    function sendEvent(hostname, name, options, meta, resolve, reject) {
        const referrer = rewriteReferrer(hostname);
        hostname = rewriteHostname(hostname);

        if (disableQueryParams) {
            hostname = (hostname.includes('?') ? hostname.split('?')[0] : hostname);
        }

        if (navigator.sendBeacon(endpoint, JSON.stringify({
            client_id: clientID,
            url: hostname.substring(0, 1800),
            title: document.title,
            referrer: (disableReferrer ? '' : encodeURIComponent(referrer)),
            screen_width: (disableResolution ? 0 : screen.width),
            screen_height: (disableResolution ? 0 : screen.height),
            event_name: name,
            event_duration: options && options.duration && typeof options.duration === "number" ? options.duration : 0,
            event_meta: meta
        }))) {
            resolve();
        } else {
            reject("error queuing event request");
        }
    }
})();
