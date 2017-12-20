<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="irma-web-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/server/" />
	<meta name="irma-api-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/api/v2/" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="../css/bootstrap.min.css" rel="stylesheet" type="text/css" />
	<script type="text/javascript" src="../js/jquery.min.js"></script>
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/bower_components/jwt-decode/build/jwt-decode.js"></script>
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/client/irma.js" async defer></script>
	<script type="text/javascript" src="../js/verify.js"></script>

	<script type="text/javascript">
	var jwt = "<?= $jwt ?>";
	var fullname_attribute = "<?= $fullname_attribute ?>";
	</script>

	<title><?= PROVIDER_NAME ?> attributen geladen</title>
</head>

<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2><?= PROVIDER_NAME ?> attributen zijn geladen</h2>

				<div class="alert alert-warning hide" role="alert">
					<strong>Geannuleerd</strong><br/>
					<small class="message"></small>
				</div>
				<div class="alert alert-danger hide" role="alert">
					<strong>Fout</strong>: kan credential niet tonen.<br/>
					<small class="message"></small>
				</div>

				<p>
					Gefeliciteerd! Uw attributen zijn nu in uw IRMA app geladen. Daar zijn ze nu zichtbaar.
				</p>

				<p>
					U kunt deze attributen gebruiken op iedere website die ze accepteert.<br/>
					Door op de knop hieronder te klikken kunt u dit testen; u toont dan uw naam-attribuut.
				</p>

				<p>
					<button id="verify_btn" class="btn btn-primary">Toon naam-attribuut</button>
					<a href="/uitgifte/">Keer terug naar de attribuut-uitgifte pagina</a>
				</p>

				<div id="token-content" class="hide">
					<h3>Resultaat</h3>
					<p><strong>Naam</strong>: <span class="name"></span></p>
				</div>
			</div>
		</div>
	</div>
</body>
</html>
