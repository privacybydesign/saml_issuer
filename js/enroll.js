$(function() {
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
		IRMA.issue(jwt, success_fun, showWarning, showError);
	});
});
