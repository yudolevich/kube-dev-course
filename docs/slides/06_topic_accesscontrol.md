## Access Control

```{image} img/acctl.svg
:width: 200px
```

### Access Control

```{image} img/acctl-overview.svg
:width: 700px
```

### Authentication

```{revealjs-fragments}
![](img/user.svg)

![](img/sa.svg)
```

### Authentication strategies

````{revealjs-fragments}
* X509 Client Certs
  ```bash
    openssl req -subj "/CN=user/O=app1/O=app2" -new -key key.pem
  ```
* Bearer Token
  ```bash
    Authorization: Bearer $TOKEN
  ```
* Authentication proxy
  ```bash
    X-Remote-User: user
    X-Remote-Group: app1
  ```
````

### Authentication OIDC

```{image} img/oidc-example.svg
:width: 700px
```

### Authentication resources

### ServiceAccount

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4-5|6|7-8|9-10
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: default
automountServiceAccountToken: false
imagePullSecrets:
- name: myregistrykey
secrets:
- default-token-r4vrb
```

### TokenRequest

```{revealjs-code-block} yaml
---
data-line-numbers: 1|2-3|5-6|7-10|12-13
---
$ k create token default -o yaml
apiVersion: authentication.k8s.io/v1
kind: TokenRequest
metadata:
  name: default
  namespace: default
spec:
  audiences:
  - https://kubernetes.default.svc.cluster.local
  expirationSeconds: 3600
status:
  expirationTimestamp: ""
  token: eyJhbGciOiJSUzI1NiIsImtpZC...
```

### TokenReview

```{revealjs-code-block} bash
---
data-line-numbers: 1-6|7-12|13|14-15|16|17-23
---
k create -o yaml -f - << EOF
kind: TokenReview
apiVersion: authentication.k8s.io/v1
spec:
  token: eyJhbGc...
EOF
apiVersion: authentication.k8s.io/v1
kind: TokenReview
metadata:
  creationTimestamp: null
spec:
  token: eyJhbGc...
status:
  audiences:
  - https://kubernetes.default.svc.cluster.local
  authenticated: true
  user:
    groups:
    - system:serviceaccounts
    - system:serviceaccounts:default
    - system:authenticated
    uid: 54228ff0-d504-4416-a870-e3de7d53dc7c
    username: system:serviceaccount:default:default
```

### CertificateSigningRequest

```{revealjs-code-block} bash
---
data-line-numbers: 1-5|7|8|9|10-11|13|14|16|17-24
---
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: myuser
spec:
  request: LS0t...
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 86400  # one day
  usages:
  - client auth
EOF
k certificate approve myuser
k get csr myuser -o jsonpath='{.status}' | jq
{
  "certificate": "LS0tL...",
  "conditions": [
    {
      "lastTransitionTime": "",
      "lastUpdateTime": "",
      "message": "This CSR was approved by kubectl certificate approve.",
      "reason": "KubectlApprove",
      "status": "True",
      "type": "Approved"
```

### Authorization

```{revealjs-fragments}
* Node
* ABAC
* RBAC
* Webhook
```

### RBAC Authorization

```{revealjs-fragments}
* Role
* RoleBinding
* ClusterRole
* ClusterRoleBinding
```

### Authorization resources

### Role

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4-5|6-9
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: pod-reader
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
```

### ClusterRole

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4-5|6-9
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  # "namespace" omitted since ClusterRoles are not namespaced
  name: secret-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]
```

### RoleBinding

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4-5|6-10|11-15|20-22|24-26|27-30
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-pods
  namespace: default
subjects:
# You can specify more than one "subject"
- kind: User
  name: jane # "name" is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  # "roleRef" specifies the binding to a Role / ClusterRole
  kind: Role #this must be Role or ClusterRole
  name: pod-reader # name of the Role or ClusterRole
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-secrets
  # This only grants permissions within the "development" namespace.
  namespace: development
subjects:
- kind: User
  name: dave # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```

### ClusterRoleBinding

```{revealjs-code-block} yaml
---
data-line-numbers: 1-3|4-5|6-9|10-13
---
apiVersion: rbac.authorization.k8s.io/v1
# allows anyone in the "manager" group to read secrets in any namespace.
kind: ClusterRoleBinding
metadata:
  name: read-secrets-global
subjects:
- kind: Group
  name: manager # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```

### SubjectAccessReview

```{revealjs-fragments}
* SubjectAccessReview
* LocalSubjectAccessReview
* SelfSubjectAccessReview
```

### SubjectAccessReview

```{revealjs-code-block} bash
---
data-line-numbers: 1|3-4|6-10|11-12|13|15-17
---
k create -f - -o jsonpath='{.status}' <<EOF | jq
---
apiVersion: authorization.k8s.io/v1
kind: SubjectAccessReview
spec:
  resourceAttributes:
    namespace: default
    verb: get
    resource: pods
    version: v1
  groups:
    - system:masters
  user: kubernetes-admin
EOF
{
  "allowed": true
}
```

### SelfSubjectAccessReview

```{revealjs-code-block} console
---
data-line-numbers: 1-2
---
$ k auth can-i get pods
yes
```

### SelfSubjectRulesReview

```{revealjs-code-block} bash
---
data-line-numbers: 1|2-3|4-5|7-19
---
k create -f - -o jsonpath='{.status.resourceRules}'<<EOF| jq
kind: SelfSubjectRulesReview
apiVersion: authorization.k8s.io/v1
spec:
  namespace: default
EOF
[
  {
    "apiGroups": [
      "*"
    ],
    "resources": [
      "*"
    ],
    "verbs": [
      "*"
    ]
  },
...
```

### SelfSubjectRulesReview

```{revealjs-code-block} console
---
data-line-numbers: 1|2-17
---
$ k auth can-i --list
Resources  Non-Resource URLs   Resource Names   Verbs
*.*        []                  []               [*]
           [*]                 []               [*]
           [/api/*]            []               [get]
           [/api]              []               [get]
           [/apis/*]           []               [get]
           [/apis]             []               [get]
           [/healthz]          []               [get]
           [/healthz]          []               [get]
           [/livez]            []               [get]
           [/livez]            []               [get]
           [/openapi/*]        []               [get]
           [/openapi]          []               [get]
           [/readyz]           []               [get]
           [/readyz]           []               [get]
...
```

### Admission Control

```{image} img/adm-ctl.png
:width: 700px
```

### LimitRange

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|5-6|7-8|9-10|11-12|13-14|15
---
apiVersion: v1
kind: LimitRange
metadata:
  name: cpu-resource-constraint
spec:
  limits:
  - default: # this section defines default limits
      cpu: 500m
    defaultRequest: # this section defines default requests
      cpu: 500m
    max: # max and min define the limit range
      cpu: "1"
    min:
      cpu: 100m
    type: Container
```

### ResourceQuota

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|6-15|16-26
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-resources
spec:
  hard:
    requests.cpu: "1"
    requests.memory: 1Gi
    limits.cpu: "2"
    limits.memory: 2Gi
    configmaps: "10"
    persistentvolumeclaims: "4"
    pods: "4"
    secrets: "10"
    services: "10"
status:
  used:
    configmaps: "1"
    limits.cpu: "0"
    limits.memory: "0"
    persistentvolumeclaims: "1"
    pods: "1"
    requests.cpu: "0"
    requests.memory: "0"
    secrets: "1"
    services: "1"
```

### Pod Security Standards

```{revealjs-fragments}
* Privileged
* Baseline
* Restricted
```

### Pod Security Admission
```{revealjs-fragments}
* enforce
* audit
* warn
```

```{revealjs-code-block} yaml
apiVersion: v1
kind: Namespace
metadata:
  name: my-restricted-namespace
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/warn-version: latest
```

### Admission Webhooks

```{revealjs-fragments}
* ValidatingWebhookConfiguration
* MutatingWebhookConfiguration
```

### Admission Webhooks

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|5|6|7-12|13-17|18-21
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: "pod-policy.example.com"
webhooks:
- name: "pod-policy.example.com"
  rules:
  - apiGroups:   [""]
    apiVersions: ["v1"]
    operations:  ["CREATE"]
    resources:   ["pods"]
    scope:       "Namespaced"
  clientConfig:
    service:
      namespace: "example-namespace"
      name: "example-service"
    caBundle: <CA_BUNDLE>
  admissionReviewVersions: ["v1"]
  failurePolicy: Fail
  sideEffects: None
  timeoutSeconds: 5
```
