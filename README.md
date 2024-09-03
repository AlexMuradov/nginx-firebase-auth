### Simple Firebase auth for Nginx Ingress Controller

**Before**: User -> Ingress Controller -> Pod \
**After**: User -> Ingress Controller -> (Auth Service) -> Pod

**Reasons to write**: oauth2-proxy is overcomplicated this should be simple beginner friendly way for centrilized auth.

### How to use:


Go to Firebase -> Project Overview -> Project Settings -> Service Accounts -> Generate New Private Key

Copy content below to **helm/values.yaml** into the **keyJson:**

```json
{
  "type": "service_account",
  "project_id": "**********",
  "private_key_id": "**********",
  "private_key": "**********",
  "client_email": "**********",
  "client_id": "**********",
  "auth_uri": "**********",
  "token_uri": "**********",
  "auth_provider_x509_cert_url": "**********",
  "client_x509_cert_url": "**********",
  "universe_domain": "**********"
}
```

Install charts:

`helm upgrade --install auth <path to helm directory e.g. ./helm/>`

Update your ingress controller rule with annotations below:

```
...
metadata:
  name: application
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "http://auth-simple-auth.default.svc.cluster.local:80/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://<your-login-url>"
...
```

change your `https://<your-service>.<namespace>.svc.cluster.local/auth` depending on namespace, above I gave example for the default namespace.

### Future improvements:

✅ Im-memory store for caching tokens
✅ Redis store for caching tokens (for HA deployments)
✅ Way to store/retrieve UserID and Email after authentificated.
✅ Sidecar implementation for those not using nginx, e.g. AWS ALB Ingress Controller, Azure Front-Door etc.

### License

Do whatever you want with this code at your own risk.