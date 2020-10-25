const DASHBOARD_JSON = "https://storage.googleapis.com/k8s-artifacts-prod-vuln-dashboard/dashboard.json";
const VULNERABILITY_URL_PREFIX = "https://cve.mitre.org/cgi-bin/cvename.cgi?name=";
const TABLE = "#table";
var JSON;
var filtering = false;

console.log("Starting up");

$.getJSON(DASHBOARD_JSON, function (data) {
    console.log(data);
    JSON = data;
    constructTable();
})

function constructTable() {
    $(TABLE).empty();
    // Getting the all column names 
    var headers = Headers();

    // Traversing the JSON data 
    for (var resourceURI in JSON) {
        var ignoreRow = false;
        var row = $('<tr/>');
        for (var index = 0; index < headers.length; index++) {
            var val = JSON[resourceURI][headers[index]];
            
            // If there is any key, which is matching 
            // with the column name 
            if (val == null) val = "";
            if (filtering && headers[index] == "NumVulnerabilities"
            && val != "" && parseInt(val) <= 0) {
                ignoreRow = true;
                break;
            }
            if (headers[index] == "ResourceURI") {
                var link = $('<a href="' + val + '">' + val + '</a>')
                row.append($('<td/>').html(link));
            } else if (headers[index] == "CriticalVulnerabilities"
            || headers[index] == "FixableVulnerabilities") {
                var column = $('<div>');
                for (var i = 0; i < val.length; i++) {
                    $('<a style=\'display :block; width: 100%;\' href="' + VULNERABILITY_URL_PREFIX + val[i] + '">' + val[i] + '</a>').appendTo($div);
                }
                row.append($('<td/>').html(column));
            } else row.append($('<td/>').html(val));
        }

        if (ignoreRow) continue;
        // Adding each row to the table 
        $(TABLE).append(row);
    }
}

function Headers() {
    var headers = [];
    var header = $('<tr/>');
    for (var resourceURI in JSON) {
        var row = JSON[resourceURI];
        for (var k in row) {
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

function ToggleFixableVulnerabiliyFilter() {
    filtering = !filtering;
    constructTable();
}
