<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="irma-web-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/server/" />
	<meta name="irma-api-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/api/v2/" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="css/bootstrap.min.css" rel="stylesheet" type="text/css" />
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/client/irma.js"></script>
	<script type="text/javascript" src="js/jquery.min.js"></script>
	<script type="text/javascript" src="js/enroll.js"></script>

	<script type="text/javascript">
	var jwt = "<?= $jwt ?>";
	</script>

	<title>Surfnet Enrollment</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Available Surfnet attributes</h2>

				<div id="alert_box"></div>

				<p>The following attributes are available for download:</p>
				<table class="table">
					<tr>
						<th scope="row">Institute</th>
						<td><?= $authenticated_user->institute ?></td>
					</tr>
					<tr>
						<th scope="row">Type</th>
						<td><?= $authenticated_user->type ?></td>
					</tr>
					<tr>
						<th scope="row">ID</th>
						<td><?= $authenticated_user->id ?></td>
					</tr>
					<tr>
						<th scope="row">Full Name</th>
						<td><?= $authenticated_user->fullname ?></td>
					</tr>
					<tr>
						<th scope="row">Given Name</th>
						<td><?= $authenticated_user->firstname ?></td>
					</tr>
					<tr>
						<th scope="row">Family Name</th>
						<td><?= $authenticated_user->familyname ?></td>
					</tr>
					<tr>
						<th scope="row">Email address</th>
						<td><?= $authenticated_user->email ?></td>
					</tr>
				</table>

				<p>Click the Load attributes button to load these attributes into your IRMA app.</p>
				<button id="enroll" class="btn btn-primary">Load attributes into IRMA</button>

				<hr />
				<small>You are logged in as <?= $authenticated_user->fullName ?> (<a href="?action=logout">Logout</a>)</small>
			</div>
		</div>
	</div>
</body>
</html>
