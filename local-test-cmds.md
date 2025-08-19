What to try

Check what the server is really speaking

openssl s_client -connect localhost:8443 -showcerts


If you see a cert chain → it’s TLS.

If it hangs or shows garbage → it’s not TLS.

Verify with your trusted bundle

openssl s_client -connect localhost:8443 -CAfile spiffe-root.crt


Try with SVID client cert

curl -vk \
  --cert /tmp/spire-assets/svid.0.pem \
  --key /tmp/spire-assets/svid.0.key \
  https://localhost:8443