<!DOCTYPE html>
<html lang="nl">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link href="../css/style.css" rel="stylesheet" type="text/css" />
	<script type="text/javascript" src="../js/jquery.min.js"></script>
	<script type="text/javascript" src="../js/yivi.js"></script>
	<script type="text/javascript" src="../js/enroll.js"></script>
	<title><?= PROVIDER_NAME ?> attributen</title>
</head>
<body data-irma-server-url="<?= h(IRMA_SERVER_URL) ?>" data-lang="<?= h(LANG) ?>">
	<div id="container">
		<header>
			<div class="header-spacer"></div>
			<h1>Beschikbare <?= PROVIDER_NAME ?> attributen</h1>
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
				<p>De volgende attributen kunnen nu in uw Yivi app geladen worden:</p>
				<table class="table">
<?php   foreach ($irma_attributes as $key => $value) { ?>
					<tr>
						<th scope="row"><?= $ATTRIBUTE_HUMAN_NAMES[$key] ?></th>
<?php     if ($key == 'profileurl' && ($profile_url = safe_http_url($value)) !== '') { ?>
						<td><a href="<?= h($profile_url) ?>"><?= h($profile_url) ?></a></td>
<?php     } else { ?>
						<td><?= htmlspecialchars($value, ENT_QUOTES|ENT_HTML5) ?></td>
<?php     } ?>
					</tr>
<?php   } ?>
				</table>

				<p>Klik hier om deze attributen in uw Yivi app te laden.</p>

				<hr />
				<p class="small-text">U bent ingelogd als <?= h($irma_attributes['firstname'] ?? '') ?> (<a href="?action=logout">Log uit</a>)</p>
<?php } ?>
			</div>
		</main>
<?php if ($irma_attributes && count($irma_attributes) > 0) { ?>
		<footer>
			<div class="actions">
				<a href="?action=logout" class="secondary">Log uit</a>
				<button id="enroll" class="primary">Laad attributen in Yivi app</button>
			</div>
		</footer>
<?php } ?>
	</div>
</body>
</html>
