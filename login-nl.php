<!DOCTYPE html>
<html lang="nl">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link href="../css/style.css" rel="stylesheet" type="text/css" />
	<title><?= PROVIDER_NAME ?> attributen</title>
</head>
<body>
	<div id="container">
		<header>
			<div class="header-spacer"></div>
			<h1>Laad attributen via <?= PROVIDER_NAME ?></h1>
			<?php
			$path = preg_replace('#^/(nl|en)/#', '/', $_SERVER['REQUEST_URI']);
			$path = preg_replace('/\?.*$/', '', $path);
			$qs = $_SERVER['QUERY_STRING'] ? '?' . $_SERVER['QUERY_STRING'] : '';
			?>
			<div class="lang-switch">
				<a href="/nl<?= h($path . $qs) ?>" class="active">NL</a>
				<a href="/en<?= h($path . $qs) ?>">EN</a>
			</div>
		</header>
		<main>
			<div class="content">
<?php if (PROVIDER == 'surfnet') { ?>
				<p>
					Uw persoonsgegevens bij uw hoger onderwijsinstelling (zoals
					naam, e-mail, of registratienummer) kunnen in uw
					Yivi-app geladen worden, via een samenwerking tussen <?= PROVIDER_NAME ?>
					en de stichting Privacy by Design.
				</p>

				<p>
					Om uw attributen te laden logt u eerst in bij uw eigen onderwijsinstelling.
					De stichting ontvangt dan uw attributen via <?= PROVIDER_NAME ?>
					en laadt ze in uw Yivi-app. Vervolgens verwijdert de stichting uw gegevens.
					Uw gegevens in de Yivi-app blijven drie maanden geldig.
				</p>


<?php } else if (PROVIDER == 'linkedin') { ?>
				<p>
					Attributen van LinkedIn (naam, email adres, enz.) kunnen eenvoudig in
					de Yivi app geladen worden. Hiervoor dient u in te loggen in LinkedIn
					via de knop hieronder.
				</p>

				<p>
					Na het inloggen kunt u direct uw LinkedIn attributen in de Yivi app
					laden. Wij zullen deze gegevens niet bewaren.
				</p>

<?php } else if (PROVIDER == 'twitter') { ?>
				<p>
					Attributen van Twitter (naam, email adres, enz.) kunnen eenvoudig in
					de Yivi app geladen worden. Hiervoor dient u in te loggen in Twitter
					via de knop hieronder.
				</p>

				<p>
					Twitter geeft ook toestemming tot allerlei informatie (tweets,
					volgers) die wij niet gebruiken. Helaas is het op dit moment
					onmogelijk om dat uit te zetten. Wij downloaden deze informatie niet
					van Twitter dus zullen het zeker niet bewaren.
				</p>

				<p>
					Na het inloggen kunt u direct uw Twitter attributen in de Yivi app
					laden. Wij zullen deze gegevens niet bewaren.
				</p>

<?php } ?>

<?php if (PROVIDER_NAME == 'eduGAIN') { ?>
				<div class="panel">
					<div class="panel-heading">
						Werkt dit ook voor mij?
					</div>
					<div class="panel-body">
						<p>
							Om uw gegevens te kunnen laden moet allereerst uw onderwijsinstelling aangesloten zijn bij
							eduGAIN. Daarnaast moet uw instelling instemmen met het doorgeven van persoonsgegevens aan
							Yivi (de Stichting Privacy by Design).
						</p>
						<p>
							Via de onderstaande knop kunt u zien of uw instelling bij eduGAIN aangesloten is. Als het
							laden in Yivi daarna niet lukt, heeft uw instelling daarvoor geen toestemming gegeven.
							U kunt daarover contact opnemen met de IT-afdeling van uw eigen instelling.
						</p>
					</div>
				</div>
<?php } ?>
			</div>
		</main>
		<footer>
			<div class="actions">
				<div></div>
				<a href="?action=login" class="primary">Log in om attributen te laden</a>
			</div>
		</footer>
	</div>
</body>
</html>
