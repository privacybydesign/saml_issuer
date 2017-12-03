<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="irma-web-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/server/" />
	<meta name="irma-api-server" value="https://privacybydesign.foundation/tomcat/irma_api_server/api/v2/" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="css/bootstrap.min.css" rel="stylesheet" type="text/css" />
	<script type="text/javascript" src="js/jquery.min.js"></script>
	<script type="text/javascript" src="https://privacybydesign.foundation/tomcat/irma_api_server/client/irma.js"></script>
	<script type="text/javascript" src="js/enroll.js"></script>

	<script type="text/javascript">
	var jwt = "<?= $jwt ?>";
	</script>

	<title>Surfnet attributen</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Beschikbare SURFconext attributen</h2>

				<div id="alert_box"></div>

				<p>De volgende attributen kunnen nu in
				uw IRMA app geladen worden.</p>
				<table class="table">
					<tr>
						<th scope="row">Instituut</th>
						<td><?= $authenticated_user->institute ?></td>
					</tr>
					<tr>
						<th scope="row">Type (student/medewerker)</th>
						<td><?= $authenticated_user->type ?></td>
					</tr>
					<tr>
						<th scope="row">Unieke identificatie</th>
						<td><?= $authenticated_user->id ?></td>
					</tr>
					<tr>
						<th scope="row">Volledige naam</th>
						<td><?= $authenticated_user->fullname ?></td>
					</tr>
					<tr>
						<th scope="row">Voornaam</th>
						<td><?= $authenticated_user->firstname ?></td>
					</tr>
					<tr>
						<th scope="row">Achternaam</th>
						<td><?= $authenticated_user->familyname ?></td>
					</tr>
					<tr>
						<th scope="row">E-mail adres</th>
						<td><?= $authenticated_user->email ?></td>
					</tr>
				</table>

				<p>Klik hier om deze attributen in uw
				IRMA app te laden.</p>
				<button id="enroll" class="btn
				btn-primary">Laad attributen in IRMA
				app</button>

				<hr />
				<small>U bent ingelogd als
				<?= $authenticated_user->fullname ?>
				(<a href="?action=logout">Log
				uit</a>)</small>
			</div>
		</div>
	</div>
</body>
</html>
