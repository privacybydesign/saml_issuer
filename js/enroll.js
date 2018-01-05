var MESSAGES = {
	'en': {
		'failed-getting-jwt': 'Failed to load attributes: cannot get the JWT from the server. Is there a problem with the connection?',
	},
	'nl': {
		'failed-getting-jwt': 'Kan attributen niet laden: JWT kan niet opgehaald worden van de server. Is er een probleem met de verbinding?',
	},
}[$(document.documentElement).prop('lang')];

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

	$("#enroll").on("click", function(e) {
		e.target.disabled = true;
		$('.alert').addClass('hide');
		$.get('?output=jwt')
			.always(function() {
				e.target.disabled = false;
			})
			.done(function(jwt) {
				IRMA.issue(jwt, onSuccess, onCancel, onError);
			})
			.fail(function(jqXhr) {
				onError(MESSAGES['failed-getting-jwt']);
			});
	});
});
