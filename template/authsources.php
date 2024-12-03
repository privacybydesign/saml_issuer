<?php

$common = array();

$config = array(
  // This is a authentication source which handles admin authentication.
  'admin' => array(
    // The default is to use core:AdminPassword, but it can be replaced with
    // any authentication source.
    'core:AdminPassword',
  ),

  // Default authentication source for Surfconext federation
  // as per
  //  <http://www.surfnet.nl/Documents/handleiding_201002_simpleSAMLphp.pdf>.
  // new documentation:
  //  <https://wiki.surfnet.nl/display/surfconextdev/Get+Conexted>
  // used metadata (also check config-metarefresh.php)
  // https://wiki.surfnet.nl/display/surfconextdev/Metadata+of+your+Service

  'surfnet' => array(
    'saml:SP',
    'EntityAttributes' => array(
      'http://macedir.org/entity-category' => array(
        'http://refeds.org/category/research-and-scholarship',
      ),
    ),
    'metadata.sign.enable' => FALSE,
    'entityID' => 'https://{{ gateway_primary_domain }}/simplesamlphp/module.php/saml/sp/metadata.php/surfnet',
    'privatekey' => 'saml.pem',
    'certificate' => 'saml.crt',
    'discoURL' => 'https://service.seamlessaccess.org/ds/',
    'name' => array(
      'en' => 'IRMA attributes',
      'nl' => 'IRMA attributen',
    ),
    'description' => array(
      'en' => 'Issue IRMA attributes based on your federated identity',
      'nl' => 'Verkrijg IRMA attributen gebaseerd op uw federatieve identiteit',
    ),
    'OrganizationName' => array(
      'en' => 'Privacy by Design Foundation',
      'nl' => 'Stichting Privacy by Design',
    ),
    'OrganizationURL' => array(
      'en' => 'https://{{ gateway_primary_domain }}/en',
      'nl' => 'https://{{ gateway_primary_domain }}',
    ),
    // We would like to get the mail and displayName attributes
    // See https://wiki.surfnet.nl/display/surfconextdev/Attributes+in+SURFconext
    'attributes' => array(
      'urn:oid:0.9.2342.19200300.100.1.3', // 'urn:mace:dir:attribute-def:mail'
      'urn:oid:1.3.6.1.4.1.25178.1.2.9', // 'urn:mace:terena.org:attribute-def:schacHomeOrganization'
      'urn:oid:1.3.6.1.4.1.5923.1.1.1.1', // 'urn:mace:dir:attribute-def:eduPersonAffiliation'
      'urn:oid:0.9.2342.19200300.100.1.1', // 'urn:mace:dir:attribute-def:uid'
      'urn:oid:2.16.840.1.113730.3.1.241', // displayname
      'urn:oid:1.3.6.1.4.1.5923.1.1.1.6', // ePPN
      'urn:oid:2.5.4.42', // 'urn:mace:dir:attribute-def:givenName'
      'urn:oid:2.5.4.4', // 'urn:mace:dir:attribute-def:sn'
    ),
    // But only the mail attributes are strictly required
    'attributes.required' => array(
      'urn:oid:0.9.2342.19200300.100.1.3',
    ),
    // set the corresponding name format for attributes
    'attributes.NameFormat' => 'urn:oasis:names:tc:SAML:2.0:attrname-format:uri',
    // Only expose HTTP-POST binding
    'acs.Bindings' => array (
      'urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST'
    ),
    'UIInfo' => array(
      'Logo' => array(
        array(
          'url'    => 'https://{{ gateway_primary_domain }}/images/irma.png',
          'height' => 300,
          'width'  => 300,
          'lang'   => 'en',
        ),
      ),
      'DisplayName' => array(
        'en' => 'IRMA attributes',
        'nl' => 'IRMA attributen',
      ),
      'Description' => array(
        'en' => 'Issue IRMA attributes based on your federated identity',
        'nl' => 'Verkrijg IRMA attributen gebaseerd op uw federatieve identiteit',
      ),
      'InformationURL' => array(
        'en' => 'https://{{ gateway_primary_domain }}/issuance-surfconext/',
        'nl' => 'https://{{ gateway_primary_domain }}/uitgifte-surfconext/',
      ),
      'PrivacyStatementURL' => array(
        'en' => 'https://{{ gateway_primary_domain }}/privacy-policy-en/',
        'nl' => 'https://{{ gateway_primary_domain }}/privacy-policy/',
      ),
    ),
  ),

  'linkedin' => array(
    'authoauth2:OpenIDConnect',
    'issuer' => 'https://www.linkedin.com/oauth',
    'clientId' => getenv('LINKEDIN_CLIENT_ID'),
    'clientSecret' => getenv('LINKEDIN_CLIENT_SECRET'),
    'scopes' => array('openid', 'email', 'profile'),
  ),

  'twitter' => [
    'authoauth2:OAuth2',
    'urlAuthorize' => 'https://www.twitter.com/i/oauth2/authorize',
    'urlAccessToken' => 'https://api.twitter.com/2/oauth2/token',
    'urlResourceOwnerDetails' => 'https://www.twitter.com/userinfo',
    'clientId' => getenv('TWITTER_CLIENT_ID'),
    'clientSecret' => getenv('TWITTER_CLIENT_SECRET'),
    'scopes' => ['users.read'],
  ]
);
