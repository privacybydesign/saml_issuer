#!/bin/bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)

SK=`realpath "$dir/../../../saml-sk"`
PK=`realpath "$dir/../../../saml-pk"`

if [ ! -e "$SK.pem" ]; then
    # Generate a private key in PEM format
    echo generating $SK.pem
    openssl genrsa -out ${SK}.pem 2048

    # Calculate corresponding public key, saved in PEM format
    echo generating $PK.pem
    openssl rsa -in ${SK}.pem -pubout -outform PEM -out ${PK}.pem
fi
