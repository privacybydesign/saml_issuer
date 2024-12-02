<!DOCTYPE html>
<html lang="nl">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="../css/bootstrap.min.css" rel="stylesheet" type="text/css" />
    <script type="application/javascript">
        let irma_server_url = "<?= IRMA_SERVER_URL ?>";
        let lang = "<?= LANG ?>";
    </script>
	<script type="text/javascript" src="../js/jquery.min.js"></script>
    <script type="text/javascript" src="../js/yivi.js"></script>
	<script type="text/javascript" src="../js/enroll.js"></script>

	<title><?= PROVIDER_NAME ?> attributen</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Beschikbare <?= PROVIDER_NAME ?> attributen</h2>

				<div id="warning" class="alert alert-warning hide" role="alert">
					<strong>Geannuleerd</strong><br/>
					<small class="message"></small>
				</div>
				<div id="error" class="alert alert-danger hide" role="alert">
					<strong>Fout</strong>: kan credential niet issuen.<br/>
					<small class="message"></small>
				</div>

<?php if (!$irma_attributes || count($irma_attributes) === 0) { ?>
				<div id="no-attributes" class="alert alert-danger" role="alert">
					<strong>Fout</strong>: kan credential niet issuen.<br/>
					<small class="message">Attribuut-issuance is geannuleerd omdat er geen attributen zijn ontvangen.</small>
				</div>
<?php } else { ?>
				<p>De volgende attributen kunnen nu in uw IRMA app geladen worden:</p>
				<table class="table">
<?php   foreach ($irma_attributes as $key => $value) { ?>
					<tr>
						<th scope="row"><?= $ATTRIBUTE_HUMAN_NAMES[$key] ?></th>
<?php     if ($key == 'profileurl') { ?>
						<td><a href="<?= $value ?>"><?= $value ?></a></td>
<?php     } else { ?>
						<td><?= htmlspecialchars($value, ENT_QUOTES|ENT_HTML5) ?></td>
<?php     } ?>
					</tr>
<?php   } ?>
				</table>

				<p>Klik hier om deze attributen in uw IRMA app te laden.</p>
				<button id="enroll" class="btn btn-primary">Laad attributen in IRMA app</button>

				<hr />
				<small>U bent ingelogd als <?= $irma_attributes['firstname'] ?> (<a href="?action=logout">Log uit</a>)</small>
<?php } ?>
			</div>
		</div>
	</div>
</body>
</html>
