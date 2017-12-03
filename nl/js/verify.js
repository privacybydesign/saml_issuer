$(function() {
	var success_fun = function(data) {
		$("#result_status").html("Succes!");
		$("#result_header").html("Resultaat");
		var surfnet_id = jwt_decode(data).attributes["pbdf.pbdf.surfnet.fullname"];
		$("#token-content").html("<b>Naam:</b> " + surfnet_id);
	}

	var cancel_fun = function(data) {
		$("#result_header").html("Resultaat");
		$("#result_status").html("Geannuleerd!");
	}

	var error_fun = function(data) {
		console.log("Authentication failed!");
		console.log("Error data:", data);
		$("#result_header").html("Resultaat");
		$("#result_status").html("Mislukt!");
	}

	$("#verify_surfnet_root_btn").on("click", function() {
		$("#result_header").text("");
		$("#result_status").text("");
		$("#token-content").text("");
		IRMA.verify(jwt, success_fun, error_fun);
	});

});
