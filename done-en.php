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

	<title><?= PROVIDER_NAME ?> attributes loaded</title>
</head>

<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2><?= PROVIDER_NAME ?> attributes have been loaded</h2>

				<div class="alert alert-warning hide" role="alert">
					<strong>Cancelled</strong><br/>
					<small class="message"></small>
				</div>
				<div class="alert alert-danger hide" role="alert">
					<strong>Error</strong>: cannot verify credential.<br/>
					<small class="message"></small>
				</div>

				<p>
					Congratulations! Your attributes have been added to your IRMA
					app. They should now be visible there.
				</p>

				<p>
					You can now use these attributes on any website that accpets them. <br/>
					By clicking the button below you can test this, by disclosing your name.
				</p>

				<p>
					<button id="verify_btn" class="btn btn-primary">Disclose name attribute</button>
					<a href="/issuance/">Back to attribute issuance</a>
				</p>

				<div id="token-content" class="hide">
					<h3>Result</h3>
					<p><strong>Name</strong>: <span class="name"></span></p>
				</div>
			</div>
		</div>
	</div>
</body>
</html>
