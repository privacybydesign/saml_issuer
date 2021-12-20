$(function() {
	function onSuccess(data) {
		var fullname = data.disclosed[0][0].rawvalue;
		$("#token-content")
			.removeClass('hide')
			.find('.name').text(fullname);
	}

	function onCancel(msg) {
		console.warn('SESSION CANCELLED');
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
		if (msg === "Aborted") {
			onCancel(msg);
		} else {
			onError(msg);
		}
	}

	$("#verify_btn").on("click", function() {
		$('.alert').addClass('hide');
		$("#token-content").addClass('hide');
		irma.newPopup({
			language: lang,
			session: {
				url: irma_server_url,
				start: {
					url: () => '?action=done&output=irma-session',
				},
			},
		})
			.start()
			.then(onSuccess, onIrmaError);
	});

});
