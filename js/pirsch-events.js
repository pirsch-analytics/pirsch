(function() {
    "use strict";

    // add a dummy function for local development
    window.pirsch = (name, options) => {
        console.log(`Pirsch event: ${name}${options ? " "+JSON.stringify(options) : ""}`);
        return Promise.resolve(null);
    };

    // respect Do-Not-Track
    if(navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    // ignore script on localhost
    const script = document.querySelector("#pirscheventsjs");
    const dev = script.getAttribute("data-dev");

    if(!dev && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.warn("Pirsch ignores events on localhost. You can enable it by adding the data-dev attribute.");
        return;
    }

    // exclude pages
    const exclude = script.getAttribute("data-exclude");
    const paths = exclude ? exclude.split(",") : [];

    for(let i = 0; i < paths.length; i++) {
        if(paths[i].test(location.pathname)) {
            return;
        }
    }

    // register event function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch-event";
    const clientID = script.getAttribute("data-client-id");
    window.pirsch = function(name, options) {
        if(typeof name !== "string" || !name) {
            return Promise.reject("The event name for Pirsch is invalid (must be a non-empty string)! Usage: pirsch('event name', {duration: 42, meta: {key: 'value'}})");
        }

        return new Promise((resolve, reject) => {
            const meta = options && options.meta ? options.meta : {};

            for(let key in meta) {
                if(meta.hasOwnProperty(key)) {
                    meta[key] = String(meta[key]);
                }
            }

            const req = new XMLHttpRequest();
            req.open("POST", endpoint);
            req.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
            req.onload = () => {
                if(req.status >= 200 && req.status < 300) {
                    resolve(req.response);
                } else {
                    reject(req.statusText);
                }
            };
            req.onerror = () => reject(req.statusText);
            req.send(JSON.stringify({
                client_id: clientID,
                url: location.href.substr(0, 1800),
                title: document.title,
                referrer: document.referrer,
                screen_width: screen.width,
                screen_height: screen.height,
                event_name: name,
                event_duration: options && options.duration && typeof options.duration === "number" ? options.duration : 0,
                event_meta: meta
            }));
        });
    }
})();
