<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link href="../css/style.css" rel="stylesheet" type="text/css" />
	<title><?= PROVIDER_NAME ?> attributes loaded</title>
</head>
<body>
	<div id="container">
		<header>
			<div class="header-spacer"></div>
			<h1><?= PROVIDER_NAME ?> attributes have been loaded</h1>
			<?php
			$path = preg_replace('#^/(nl|en)/#', '/', $_SERVER['REQUEST_URI']);
			$path = preg_replace('/\?.*$/', '', $path);
			$qs = $_SERVER['QUERY_STRING'] ? '?' . $_SERVER['QUERY_STRING'] : '';
			?>
			<div class="lang-switch">
				<a href="/nl<?= $path . $qs ?>">NL</a>
				<a href="/en<?= $path . $qs ?>" class="active">EN</a>
			</div>
		</header>
		<main>
			<div class="content">
				<div class="imageContainer">
					<img src="../images/done.png" alt="Success" />
				</div>

				<p>
					Congratulations! Your attributes have been added to your Yivi
					app. They should now be visible there.
				</p>

				<p>
					You can now use these attributes on any website that accepts them.
				</p>
			</div>
		</main>
		<footer>
			<div class="actions">
				<div></div>
				<a href="?action=logout" class="primary">Log out</a>
			</div>
		</footer>
	</div>
</body>
</html>
