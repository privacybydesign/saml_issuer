<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="irma-web-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/server/" />
	<meta name="irma-api-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/api/v2/" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="css/bootstrap.min.css" rel="stylesheet" type="text/css" />
	<script type="text/javascript" src="js/jquery.min.js"></script>
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/bower_components/jwt-decode/build/jwt-decode.js"></script>
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/client/irma.js" async defer></script>
	<script type="text/javascript" src="js/verify.js"></script>

	<script type="text/javascript">
	var jwt = "<?= $jwt ?>";
	</script>

	<title>IRMA Enrollment</title>
</head>

<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Surfnet attributes have been issued!</h2>

				<div id="alert_box"></div>

				<p>
					Congratulations! Your attributes have been successfully loaded onto your IRMA app.
					They should be visible there.
				</p>

				<p>
					You can now use them on any website that accepts these attributes. If you are feeling impatient, you can also immediately test them by pressing the test button.
				</p>

				<button id="verify_surfnet_root_btn" class="btn btn-primary">Test Attributes</button>

				<h3 id="result_header"></h3>
				<h4 id="result_status"></h4>
				<p id="token-content"></p>

			</div>
		</div>
	</div>
</body>
</html>
