<html>


<body>
<h1>Apply for an IRMA card</h1>


<?php
    require_once ($_SERVER['DOCUMENT_ROOT'] . '/../support/php/prelude.inc');

    $mysqli = new mysqli("mysql-irma.science.ru.nl", "irma_user", $admin_password, "irma");
if ($mysqli->connect_errno) {
    echo "Failed to connect to MySQL: (" . $mysqli->connect_errno . ") " . $mysqli->connect_error;
}

    $result = $mysqli->query("TRUNCATE TABLE PilotParticipants");
    print $result;
    print "<br>";
    // $result = $mysqli->query ("INSERT INTO PilotParticipants (eduPersonPrincipalName, givenName, surname) VALUES ('u949100@ru.nl', 'Ronny', 'Wichers Schreur')");
    print $result;
    print "<br>";
    $result = $mysqli->query ("SELECT * FROM  PilotParticipants");
    var_dump ($result->fetch_fields ());
    print "Done!";
?>
</body>

</html>
        
