<!DOCTYPE html>
<html lang="nl">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link href="../css/style.css" rel="stylesheet" type="text/css" />
	<title><?= PROVIDER_NAME ?> attributen geladen</title>
</head>
<body>
	<div id="container">
		<header>
			<div class="header-spacer"></div>
			<h1><?= PROVIDER_NAME ?> attributen zijn geladen</h1>
			<?php
			$path = preg_replace('#^/(nl|en)/#', '/', $_SERVER['REQUEST_URI']);
			$path = preg_replace('/\?.*$/', '', $path);
			$qs = $_SERVER['QUERY_STRING'] ? '?' . $_SERVER['QUERY_STRING'] : '';
			?>
			<div class="lang-switch">
				<a href="/nl<?= $path . $qs ?>" class="active">NL</a>
				<a href="/en<?= $path . $qs ?>">EN</a>
			</div>
		</header>
		<main>
			<div class="content">
				<div class="imageContainer">
					<img src="../images/done.png" alt="Succes" />
				</div>

				<p>
					Gefeliciteerd! Uw attributen zijn nu in uw Yivi app geladen. Daar zijn ze nu zichtbaar.
				</p>

				<p>
					U kunt deze attributen gebruiken op iedere website die ze accepteert.
				</p>
			</div>
		</main>
		<footer>
			<div class="actions">
				<div></div>
				<a href="/uitgifte/" class="primary">Terug naar attribuut-uitgifte</a>
			</div>
		</footer>
	</div>
</body>
</html>
