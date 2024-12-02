var MESSAGES = {
	'en': {
		'failed-starting-irma-session': 'Failed to load attributes: cannot start IRMA session at IRMA server. Is there a problem with the connection?',
	},
	'nl': {
		'failed-starting-irma-session': 'Kan attributen niet laden: IRMA session kan niet gestart worden bij de IRMA server. Is er een probleem met de verbinding?',
	},
}[$(document.documentElement).prop('lang')];

$(function() {
	function onSuccess() {
		window.location = "?action=done";
	}

	function onCancel(msg) {
		console.warn('SESSION CANCELLED');
		$('#warning')
			.removeClass('hide')
			.find('.message').text(msg);
	}

	function onError(msg) {
		console.error('ERROR:', msg);
		$('#error')
			.removeClass('hide')
			.find('.message').text(msg);
	}

	function onIrmaError(msg) {
		if (msg === 'Aborted') {
			onCancel(msg);
		} else {
			onError(msg);
		}
	}

	$("#enroll").on("click", function() {
		$('.alert').addClass('hide');
		yivi.newPopup({
			language: lang,
			session: {
				url: irma_server_url,
				start: {
					url: () => '?output=irma-session',
				},
			},
		})
			.start()
			.then(onSuccess, onIrmaError);
	});
});
