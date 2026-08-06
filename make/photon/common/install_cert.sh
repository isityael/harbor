#!/bin/sh

set -e

ORIGINAL_LOCATION=$(dirname "$0")
TARGET_BUNDLE="${HARBOR_CA_BUNDLE:-${SSL_CERT_FILE:-}}"

if [ -z "$TARGET_BUNDLE" ]; then
    for candidate in \
        /etc/pki/tls/certs/ca-bundle.crt \
        /etc/ssl/certs/ca-certificates.crt \
        /etc/ssl/cert.pem; do
        if [ -f "$candidate" ]; then
            TARGET_BUNDLE="$candidate"
            break
        fi
    done
fi

if [ -z "$TARGET_BUNDLE" ]; then
    echo "No CA bundle path found, skip appending CA certificates"
    exit 0
fi

if [ ! -w "$TARGET_BUNDLE" ]; then
    echo "CA bundle $TARGET_BUNDLE is not writable, skip appending CA certificates"
    exit 0
fi

if [ ! -f "$ORIGINAL_LOCATION/ca-bundle.crt.original" ]; then
    cp "$TARGET_BUNDLE" "$ORIGINAL_LOCATION/ca-bundle.crt.original"
fi

cp "$ORIGINAL_LOCATION/ca-bundle.crt.original" "$TARGET_BUNDLE"

# Install /etc/harbor/ssl/{component}/ca.crt to trust CA.
echo "Appending internal tls trust CA to $TARGET_BUNDLE ..."
if [ -d /etc/harbor/ssl ]; then
    find /etc/harbor/ssl -maxdepth 2 -name ca.crt -type f | while IFS= read -r caFile; do
        cat "$caFile" >> "$TARGET_BUNDLE"
        echo "Internal tls trust CA $caFile appended ..."
    done
fi
echo "Internal tls trust CA appending is Done."

if [ -d /harbor_cust_cert ] && [ -n "$(find /harbor_cust_cert -mindepth 1 -maxdepth 1 -print -quit)" ]; then
    echo "Appending trust CA to $TARGET_BUNDLE ..."
    for z in /harbor_cust_cert/*; do
        case ${z} in
            *.crt | *.ca | *.ca-bundle | *.pem)
                if [ -d "$z" ]; then
                    echo "$z is directory, skip it ..."
                else
                    cat "$z" >> "$TARGET_BUNDLE"
                    echo " $z Appended ..."
                fi
                ;;
            *) echo "$z is Not ca file ..." ;;
        esac
    done
    echo "CA appending is Done."
fi
