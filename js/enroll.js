$(function() {
	function onSuccess(data) {
		window.location = "?action=done";
	}

	function onCancel(msg) {
		console.warn('CANCEL:', msg);
		$('.alert.alert-warning')
			.removeClass('hide')
			.find('.message').text(msg);
	};

	function onError(msg) {
		console.error('ERROR:', msg);
		$('.alert.alert-danger')
			.removeClass('hide')
			.find('.message').text(msg);
	};

	$("#enroll").on("click", function() {
		$('.alert').addClass('hide');
		IRMA.issue(jwt, onSuccess, onCancel, onError);
	});
});
