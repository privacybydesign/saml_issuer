$(function() {
	var success_fun = function(data) {
		$("#result_status").html("Success!");
		$("#result_header").html("Result");
		var surfnet_id = jwt_decode(data).attributes["pbdf.pbdf.surfnet.fullname"];
		$("#token-content").html("<b>Full name:</b> " + surfnet_id);
	}

	var cancel_fun = function(data) {
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
		$("#result_header").text("");
		$("#result_status").text("");
		$("#token-content").text("");
		IRMA.verify(jwt, success_fun, error_fun);
	});

});
