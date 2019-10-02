<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="../css/bootstrap.min.css" rel="stylesheet" type="text/css" />
    <script type="application/javascript">
        let irma_server_url = "<?= IRMA_SERVER_URL ?>";
        let lang = "<?= LANG ?>";
    </script>
	<script type="text/javascript" src="../js/jquery.min.js"></script>
	<script type="text/javascript" src="../js/irma.js"></script>
	<script type="text/javascript" src="../js/enroll.js"></script>

	<title><?= PROVIDER_NAME ?> attributes</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Available <?= PROVIDER_NAME ?> attributes</h2>

				<div id="warning" class="alert alert-warning hide" role="alert">
					<strong>Cancelled</strong><br/>
					<small class="message"></small>
				</div>
				<div id="error" class="alert alert-danger hide" role="alert">
					<strong>Error</strong>: cannot issue credential.<br/>
					<small class="message"></small>
				</div>

<?php if (!$irma_attributes || count($irma_attributes) === 0) { ?>
				<div id="no-attributes" class="alert alert-danger" role="alert">
					<strong>Error</strong>: cannot issue credential.<br/>
					<small class="message">Issuance was aborted because no attributes were received.</small>
				</div>
<?php } else { ?>
				<p>The attributes below can be added to your IRMA app:</p>
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

				<p>Click here to add this attributes to your IRMA app.</p>
				<button id="enroll" class="btn btn-primary">Load attributes in IRMA app</button>

				<hr />
				<small>You are logged in as <?= $irma_attributes['fullname'] ?> (<a href="?action=logout">Log out</a>)</small>
<?php } ?>
			</div>
		</div>
	</div>
</body>
</html>
