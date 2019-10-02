$(function() {
	function onSuccess(data) {
		var fullname = data.disclosed[0][0].rawvalue;
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

	function onIrmaError(msg) {
		if (msg === "CANCELLED") {
			onCancel();
		} else {
			onError(msg);
		}
	}

	$("#verify_btn").on("click", function() {
		$('.alert').addClass('hide');
		$("#token-content").addClass('hide');
		$("#verify_btn").attr("disabled", true);
		const options = {language: lang, token: verification_session.token, server: irma_server_url};
		irma.handleSession(verification_session.sessionPtr, options)
			.then(onSuccess, onIrmaError);
	});

});
