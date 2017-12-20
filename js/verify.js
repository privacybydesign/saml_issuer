$(function() {
	function onSuccess(data) {
		var fullname = jwt_decode(data).attributes[fullname_attribute];
		$("#token-content")
			.removeClass('hide')
			.find('.name').text(fullname);
	}

	function onCancel(msg) {
		console.warn('CANCEL:', msg);
		$('.alert.alert-warning')
			.removeClass('hide')
			.find('.message').text(msg);
	}

	function onError(msg) {
		console.error('ERROR:', msg);
		$('.alert.alert-danger')
			.removeClass('hide')
			.find('.message').text(msg);
	}

	$("#verify_btn").on("click", function() {
		$('.alert').addClass('hide');
		$("#token-content").addClass('hide');
		IRMA.verify(jwt, onSuccess, onCancel, onError);
	});

});
