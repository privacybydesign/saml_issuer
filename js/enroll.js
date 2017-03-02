$(function() {
	var iprequest = {
		timeout: 60,
		request: {
			"credentials": [
				{
					"credential": "irma-demo.Surfnet.root",
					"attributes": {
						"userID": userID,
						"securityHash": "00000000",
					}
				},
				{
					"credential": "irma-demo.MijnOverheid.fullName",
					"attributes": {
						"firstnames": firstnames,
						"firstname": firstname,
						"familyname": familyname,
						"prefix": prefix,
					}
				}
			],
		}
	};

	var showWarning = function(msg) {
		$("#alert_box").html('<div class="alert alert-warning" role="alert">'
			 + '<strong>Warning:</strong> ' + msg + '</div>');
	};

	var showError = function (msg) {
		$("#alert_box").html('<div class="alert alert-danger" role="alert">'
			 + '<strong>Error:</strong> ' + msg + '</div>');
	};

	var success_fun = function(data) {
		window.location = "done.html";
	};

	$("#enroll").on("click", function() {
		var jwt = IRMA.createUnsignedIssuanceJWT(iprequest);
		IRMA.issue(jwt, success_fun, showWarning, showError);
	});
});
