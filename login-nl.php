<!DOCTYPE html>
<html lang="nl">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="../css/bootstrap.min.css" rel="stylesheet" type="text/css" />

	<title><?= PROVIDER_NAME ?> attributen</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Laad attributen via <?= PROVIDER_NAME ?></h2>

<?php if (PROVIDER == 'surfnet') { ?>
				<p>
					De attributen van uw hoger onderwijsinstelling kunnen nu in uw
					IRMA app geladen worden via <?= PROVIDER_NAME ?>.
				</p>

				<p>
					Om uw attributen te laden moet u eerst inloggen bij uw onderwijsinstelling.
					Wij ontvangen dan uw attributen. Daarna kunt u ze laden in uw IRMA app.
				</p>


<?php } else if (PROVIDER == 'linkedin') { ?>
				<p>
					Attributen van LinkedIn (naam, email adres, enz.) kunnen eenvoudig in
					de IRMA app geladen worden. Hiervoor dient u in te loggen in LinkedIn
					via de knop hieronder.
				</p>

				<p>
					Na het inloggen kunt u direct uw LinkedIn attributen in de IRMA app
					laden. Wij zullen deze gegevens niet bewaren.
				</p>

<?php } else if (PROVIDER == 'twitter') { ?>
				<p>
					Attributen van Twitter (naam, email adres, enz.) kunnen eenvoudig in
					de IRMA app geladen worden. Hiervoor dient u in te loggen in Twitter
					via de knop hieronder.
				</p>

				<p>
					Twitter geeft ook toestemming tot allerlei informatie (tweets,
					volgers) die wij niet gebruiken. Helaas is het op dit moment
					onmogelijk om dat uit te zetten. Wij downloaden deze informatie niet
					van Twitter dus zullen het zeker niet bewaren.
				</p>

				<p>
					Na het inloggen kunt u direct uw Twitter attributen in de IRMA app
					laden. Wij zullen deze gegevens niet bewaren.
				</p>

<?php } ?>

<?php if (PROVIDER_NAME == 'eduGAIN') { ?>
				<div class="panel panel-info">
					<div class="panel-heading">
						Ondersteuning
					</div>
					<div class="panel-body">
						<p>
							Uitgifte van IRMA-attributen wordt niet door alle onderwijsinstellingen ondersteund die bij eduGAIN
							zijn aangesloten. Het kan dus voorkomen dat uw onderwijsinstelling wel in de lijst voorkomt,
							maar dat u na het bevestigen van uw keuze een foutmelding te zien krijgt.
						</p>
						<p>
							Ontbreekt de koppeling met uw onderwijsinstelling en wilt u
							daar verandering in brengen? Neem dan contact op met de IT-afdeling van uw onderwijsinstelling.
						</p>
					</div>
				</div>
<?php } ?>

				<a href="?action=login" class="btn btn-primary">Log in om attributen te laden</a>
			</div>
		</div>
	</div>
</body>
</html>
