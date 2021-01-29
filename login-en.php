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
					Your personal data at your educational institution (such as
					name, email or registration number) can be loaded into your
					IRMA app via a cooperation between <?= PROVIDER_NAME ?> and the
					Privacy by Design Foundation.
				</p>

				<p>
					In order to load your attributes you first login to your own
					institution. The foundation then receives your data via
					<?= PROVIDER_NAME ?> and loads them into your IRMA app. The
					foundation immediately removes your data. In your IRMA app the
					data remain valid for three months.
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
						Does this work for me?
					</div>
					<div class="panel-body">
						<p>
							Loading your data works if, firstly, your institution is part of
							eduGAIN, and secondly, your institution allows data transfer to
							IRMA (the Privacy by Design Foundation).
						</p>
						<p>
							The button below shows if your institution has joined eduGAIN.
							If subsequent loading into IRMA fails, your institution has
							not agreed to the required data transfer. You can then contact
							the IT department of your own institution.
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
