# pam\_oidc

**pam_oidc** authenticates users with an OpenID Connect (OIDC) token.

Linux-PAM extensions are used, so currently the module only supports Linux. Contributions to support other operating systems are welcomed.

## Example Usage

In `/etc/pam.d/service`:

```
auth required pam_oidc.so <options>
```

Example for Google:

```
auth required pam_oidc.so issuer=https://accounts.google.com aud=12345-v12345.apps.googleusercontent.com
```

### Options

#### issuer

Required.

The issuer URL. The OpenID configuration should be available at _issuer_/.well-known/openid-configuration

#### aud

Required.

The audience value to expect. Tokens signed by the issuer but for a different audience will be rejected. This prevents tokens issued for a different purpose from being used for authentication.

#### user\_template

Default: `{{.Subject}}`

A Go [text/template](http://pkg.go.dev/text/template) that, when rendered with the JWT/OIDC claims, provides the expected username.

For example, `{{.Subject}}` would mean that users are expected to authenticate with the JWT `sub` claim as their username.

The `trimPrefix` and `trimSuffix` functions are available. For example `{{.Subject | trimSuffix "@example.com"}}` would mean a user whose token subject is `jdoe@example.com` would authenticate as `jdoe`.

#### groups\_claim\_key

Default: `groups`

The name of the key within the token claims that specifies which groups a user is a member of.

If the token uses a key other than `groups` (e.g., `{"roles":["a", "b", "c"]}`), specifies `groups_claim_key=roles`.

#### authorized\_groups

Default: (no value)

If specified, a comma-separated list of groups required for authentication to pass. A user must be a member of _at least_ one of the groups in the list, if specified.

#### require\_acr

Default: (no value)

If specified, the required value of the `acr` claim in the token for authentication to pass.

#### require\_acrs

Default: (no value)

If specified, a comma-separated list of acrs one of which must match the `acr` claim in the token for authentication to pass.

#### http\_proxy

Default: (no value)

If specified, an HTTP proxy used to connect to the issuer to discover OpenID Connect parameters.

## Local Testing

A Vagrant VM is available for local testing:

```
vagrant up
```

By default, PAM is setup with Percona Server to accept OpenID Connect tokens from the Google Cloud SDK using email address as the username:

```
gcloud auth login
gcloud auth print-identity-token
```

Within the VM, create a database user to authenticate using PAM:

```
vagrant ssh

# within the Vagrant VM
sudo mysql -u root

# within the MySQL monitor
CREATE USER 'jdoe@gmail.com'@'%' IDENTIFIED WITH auth_pam;
```

With the token from `gcloud auth print-identity-token`, attempt to login:

```
TOKEN="..." # paste from `gcloud auth print-identity-token`

# The token must be specified using --password=... because it is too long for
# MySQL to accept interactively
mysql --user="jdoe@gmail.com" --password="$TOKEN"
```

To debug failures, check the auth logs:

```
sudo tail -f /var/log/auth.log
```
