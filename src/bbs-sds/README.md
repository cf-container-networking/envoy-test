# Envoy SDS API implementation

This uses the Diego BBS to provide an Envoy SDS API server.

```
cf push sds
cf create-security-group access-bbs <(echo '[ { "protocol": "tcp", "destination": "10.244.0.0/24", "ports": "8889", "description": "Allow sds to reach bbs" } ]')
cf bind-security-group access-bbs orgWithSDS spaceWithSDS
cf restart sds
```

Then
```
curl http://sds.bosh-lite.com/v1/registration/$(cf app --guid someApp)
```
should return the hosts for that app.

Caveats:
- Client certificate and key to access the BBS are currently hardcoded into the `main.go`!
- BBS IP address is hardcoded into the `manifest.yml`
- BBS port is hardcoded in `main.go`
