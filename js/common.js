export function getScript(id) {
    const script = document.querySelector(id);

    if(!script) {
        throw `Pirsch script ${id} tag not found!`;
    }

    return script;
}

export function ignore(script) {
    return dnt() || isLocalhost(script) || !includePage(script) || excludePage(script);
}

export function rewriteHostname(hostname) {
    if (!hostname) {
        hostname = location.href;
    } else {
        hostname = location.href.replace(location.hostname, hostname);
    }

    return hostname;
}

export function rewriteReferrer(hostname) {
    let referrer = document.referrer;

    if (hostname) {
        referrer = referrer.replace(location.hostname, hostname);
    }

    return referrer;
}

function dnt() {
    return navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch");
}

function isLocalhost(script) {
    const dev = script.hasAttribute("data-dev");

    if (!dev && (/^localhost(.*)$|^127(\.[0-9]{1,3}){3}$/is.test(location.hostname) || location.protocol === "file:")) {
        console.info("Pirsch is ignored on localhost. Add the data-dev attribute to enable it.");
        return true;
    }

    return false;
}

function includePage(script) {
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
                return false;
            }
        }
    } catch (e) {
        console.error(e);
    }

    return true;
}

function excludePage(script) {
    try {
        const exclude = script.getAttribute("data-exclude");
        const paths = exclude ? exclude.split(",") : [];

        for (let i = 0; i < paths.length; i++) {
            if (new RegExp(paths[i]).test(location.pathname)) {
                return true;
            }
        }
    } catch (e) {
        console.error(e);
    }

    return false;
}
