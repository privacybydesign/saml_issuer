<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />

	<link href="../css/bootstrap.min.css" rel="stylesheet" type="text/css" />

	<title><?= PROVIDER_NAME ?> attributes</title>
</head>
<body>
	<div class="container">
		<div class="row">
			<div class="col-xs-12 col-md-8 col-lg-6 col-md-offset-2 col-lg-offset-3">
				<h2>Load attributes from <?= PROVIDER_NAME ?></h2>

<?php if (PROVIDER == 'surfnet') { ?>
				<p>
					Attributes from your educational institute can be added to your IRMA
					app via <?= PROVIDER_NAME ?>.
				</p>

				<p>
					In order to load these attributes, you first have to log into your own
					institute. With your permission, the Privacy by Design foundation then
					receives your attributes. Subsequently, they can be loaded into your IRMA app.
				</p>

<?php } else if (PROVIDER == 'linkedin') { ?>
				<p>
					Attributes from LinkedIn can now be loaded into your IRMA app.
					To load these attributes it is necessary that you log in on LinkedIn,
					and give IRMA access to your basic profile data. We use this data only
					during attribute loading.
				</p>

				<p>
					After you have given this permission on LinkedIn the attributes can be
					loaded into your IRMA app.
				</p>

<?php } else if (PROVIDER == 'twitter') { ?>
				<p>
					Attributes from Twitter can now be loaded into your IRMA app.
					To load these attributes it is necessary that you log in on Twitter,
					and give IRMA access to your basic profile data. We use this data only
					during attribute loading.
				</p>

				<p>
					Twitter also gives us access to more information than neccessary, such
					as your tweets and follows, that we do not use. Unfortunately it is
					impossible for us to disable this. However, we do not download or use
					this data in any way.
				</p>

				<p>
					After you have given this permission on Twitter the attributes can be
					loaded into your IRMA app.
				</p>

<?php } ?>

<?php if (PROVIDER_NAME == 'eduGAIN') { ?>
				<div class="panel panel-info">
					<div class="panel-heading">
						Support
					</div>
					<div class="panel-body">
						<p>
							Issuance of IRMA attributes is not supported by all educational institutes affiliated with
							eduGAIN. Therefore, it might happen that your institute is available as an option in the list,
							but when you confirm your choice, an error message is shown.
						</p>
						<p>
							Is your educational institute not connected to this service and
							do you want to change that? Please contact the IT department of your educational institute.
						</p>
					</div>
				</div>
<?php } ?>

				<a href="?action=login" class="btn btn-primary">Log in to load attributes</a>
			</div>
		</div>
	</div>
</body>
</html>
