const DASHBOARD_JSON = "https://storage.googleapis.com/k8s-artifacts-prod-vuln-dashboard/dashboard.json";
const VULNERABILITY_URL_PREFIX = "https://cve.mitre.org/cgi-bin/cvename.cgi?name=";
const TABLE = "#table";

console.log("Starting up...");

$.getJSON(DASHBOARD_JSON, function (data) {
    $('#pagination-container').pagination({
        dataSource: function (done) {
            var result = [];
            for (var resourceURI in data) {
                result.push(data[resourceURI]);
            }
            done(result);
        },
        pageSize: 20,
        callback: function (data, pagination) {
            $(TABLE).empty();
            headers(data);
            constructTable(data);
        }
    });
})

function constructTable(data) {
    $.each(data, function (index, item) {
        var row = $('<tr/>');
        var link = $('<a href="' + item.ResourceURI + '">' + item.ResourceURI + '</a>');
        row.append($('<td/>').html(link));

        var vul = $('<div>' + item.NumVulnerabilities + '<div/>');
        row.append($('<td/>').html(vul));

        var columnCrit = $('<div>');
        for (var i = 0; i < item.CriticalVulnerabilities.length; i++) {
            $('<a style=\'display :block; width: 100%;\' href="' + VULNERABILITY_URL_PREFIX + item.CriticalVulnerabilities[i] + '">' + item.CriticalVulnerabilities[i] + '</a>').appendTo(columnCrit);
        }
        row.append($('<td/><div/>').html(columnCrit));

        var columnFix = $('<div>');
        for (var i = 0; i < item.FixableVulnerabilities.length; i++) {
            $('<a style=\'display :block; width: 100%;\' href="' + VULNERABILITY_URL_PREFIX + item.FixableVulnerabilities[i] + '">' + item.FixableVulnerabilities[i] + '</a>').appendTo(columnFix);
        }
        row.append($('<td/><div/>').html(columnFix));

        $(TABLE).append(row);
    });
}

function headers(item) {
    var headers = [];
    var header = $('<tr/>');
    for (var resourceURI in item) {
        var row = item[resourceURI];
        for (var k in row) {
            if (k == "ResourceURI" || k == "ImageDigest") {
                continue;
            }
            if ($.inArray(k, headers) == -1) {
                headers.push(k);
                // Creating the header
                header.append($('<th/>').html(k));
            }
        }
    }
    // Appending the header to the table
    $(TABLE).append(header);
    return headers;
}
