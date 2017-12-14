$(function() {
	var success_fun = function(data) {
		$("#result_status").html("Success!");
		$("#result_header").html("Result");
		var fullname = jwt_decode(data).attributes[fullname_attribute];
		$("#token-content").html("<b>Name:</b> " + fullname);
	}

	var cancel_fun = function(data) {
		$("#result_header").html("Result");
		$("#result_status").html("Cancelled!");
	}

	var error_fun = function(data) {
		console.log("Authentication failed!");
		console.log("Error data:", data);
		$("#result_header").html("Result");
		$("#result_status").html("Failed!");
	}

	$("#verify_btn").on("click", function() {
		$("#result_header").text("");
		$("#result_status").text("");
		$("#token-content").text("");
		IRMA.verify(jwt, success_fun, error_fun);
	});

});
