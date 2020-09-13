function Pirsch(options) {
    if(!options) {
        options = {};
    }

    window.addEventListener("DOMContentLoaded", function() {
        var endpoint = options.endpoint || "/pirsch";
        var tenantID = options.tenant_id || 0;
        var nocache = new Date().getTime();
        var location = window.location;
        var referrer = document.referrer;
        var width = window.screen.width;
        var height = window.screen.height;
        var url = endpoint+"?nocache="+nocache+"&tenant_id="+tenantID+"&location="+location+"&referrer="+referrer+"&width="+width+"&height="+height;
        var req = new XMLHttpRequest();
        req.open("GET", url);
        req.send();
    });
}
