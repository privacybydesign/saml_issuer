$(function() {
	// Setup verification request
	var sprequest = {
		"validity": 60,
		"request": {
			"content": [
				{
					"label": "Surfnet",
					"attributes": ["pbdf.pbdf.surfnet.fullname"]
				},
			]
		}
	};

	var success_fun = function(data) {
		console.log("Authentication successful!");
		console.log("Authentication token:", data);
		$("#result_status").html("Success!");
		$("#result_header").html("Result");
		var json_token  = jwt_decode(data);
		var attributes = json_token.attributes;
		var surfnet_id = attributes["pbdf.pbdf.surfnet.fullname"];
		json_string  = JSON.stringify(surfnet_id, null, 2);
		$("#token-content").html("<b>Name:</b> " + surfnet_id);
	}

	var cancel_fun = function(data) {
		console.log("Authentication cancelled!");
		$("#result_header").html("Result");
		$("#result_status").html("Cancelled!");
	}

	var error_fun = function(data) {
		console.log("Authentication failed!");
		console.log("Error data:", data);
		$("#result_header").html("Result");
		$("#result_status").html("Failure!");
	}

	$("#verify_surfnet_root_btn").on("click", function() {
		console.log("Button clicked");
		$("#result_header").text("");
		$("#result_status").text("");
		$("#token-content").text("");
		var jwt = IRMA.createUnsignedVerificationJWT(sprequest);
		IRMA.verify(jwt, success_fun, error_fun);
	});

});
