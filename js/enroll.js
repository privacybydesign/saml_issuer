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
		if (msg === 'CANCELLED') {
			onCancel(msg);
		} else {
			onError(msg);
		}
	}

	$("#enroll").on("click", function(e) {
		e.target.disabled = true;
		$('.alert').addClass('hide');
		$.get('?output=irma-session')
			.always(function() {
				e.target.disabled = false;
			})
			.done(function(sessionpackagejson) {
				const sessionpackage = JSON.parse(sessionpackagejson);
				const options = {language: lang, token: sessionpackage.token, server: irma_server_url}
				irma.handleSession(sessionpackage.sessionPtr, options)
					.then(onSuccess, onIrmaError);
			})
			.fail(function(jqXhr) {
				onError(MESSAGES['failed-starting-irma-session']);
			});
	});
});
