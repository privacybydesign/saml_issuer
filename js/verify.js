$(function() {
	var success_fun = function(data) {
		$("#result_status").html("Succes!");
		$("#result_header").html("Resultaat");
		var fullname = jwt_decode(data).attributes[fullname_attribute];
		$("#token-content").html("<b>Naam:</b> " + fullname);
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

	$("#verify_btn").on("click", function() {
		$("#result_header").text("");
		$("#result_status").text("");
		$("#token-content").text("");
		IRMA.verify(jwt, success_fun, error_fun);
	});

});
