(function () {
    "use strict";

    // add a dummy function for local development
    window.pirsch = (name, options) => {
        console.log(`Pirsch event: ${name}${options ? " " + JSON.stringify(options) : ""}`);
        return Promise.resolve(null);
    };

    // respect Do-Not-Track
    if (navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    // ignore script on localhost
    const script = document.querySelector("#pirscheventsjs");
    const dev = script.hasAttribute("data-dev");

    if (!dev && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.warn("Pirsch ignores events on localhost. You can enable it by adding the data-dev attribute.");
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

    // register event function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch-event";
    const clientID = script.getAttribute("data-client-id");
    const domains = script.getAttribute("data-domain") ? script.getAttribute("data-domain").split(",") || [] : [];
    const disableQueryParams = script.hasAttribute("data-disable-query");
    const disableReferrer = script.hasAttribute("data-disable-referrer");
    const disableResolution = script.hasAttribute("data-disable-resolution");

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

            sendEvent(null, name, options, meta, resolve, reject);

            for (let i = 0; i < domains.length; i++) {
                sendEvent(domains[i], name, options, meta, resolve, reject);
            }
        });
    }

    function sendEvent(hostname, name, options, meta, resolve, reject) {
        if (!hostname) {
            hostname = location.href;
        } else {
            hostname = location.href.replace(location.hostname, hostname);
        }

        if (disableQueryParams) {
            hostname = (hostname.includes('?') ? hostname.split('?')[0] : hostname);
        }

        if (navigator.sendBeacon(endpoint, JSON.stringify({
            client_id: clientID,
            url: hostname.substring(0, 1800),
            title: document.title,
            referrer: (disableReferrer ? '' : encodeURIComponent(document.referrer)),
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
