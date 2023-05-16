## kustomize/helm

```{image} img/kustomize.svg
:width: 200px
```
```{image} img/helm.svg
:width: 200px
```

### Deploy

YAML Files:
```{revealjs-fragments}
* Deployment
* Service
* Ingress
* ConfigMap
* Secret
```

### Deploy

```{revealjs-code-block} console
---
data-line-numbers: 1-4|5-8
---
deploy/
├── deployment.yaml
├── ingress.yaml
└── service.yaml
$ kubectl apply -f deploy
deployment.apps/app created
ingress.networking.k8s.io/app created
service/app created
```

### Kustomize

```{revealjs-code-block} yaml
---
data-line-numbers: 1|2-5|6|7|8-9
---
# kustomization.yaml
resources:
- deployment.yaml
- service.yaml
- ingress.yaml
namePrefix: dev-
namespace: development
commonLabels:
  environment: development
```

### Kustomize

```{revealjs-code-block} console
$ kustomize build
$ kubectl kustomize
$ kubectl apply -k
```

### Kustomize

```{revealjs-code-block} console
---
data-line-numbers: 1-5|6-9
---
deploy/
├── deployment.yaml
├── ingress.yaml
├── kustomization.yaml
└── service.yaml
$ kubectl apply -k deploy
service/dev-app created
deployment.apps/dev-app created
ingress.networking.k8s.io/dev-app created
```

### Environment

```{revealjs-fragments}
* Dev
* Test(ift/lt)
* Stage
* Prod
```

### Kustomize Overlays
```{revealjs-code-block} console
---
data-line-numbers: 2-6|7-8|9-10|11-12
---
deploy/
├── base
│   ├── deployment.yaml
│   ├── ingress.yaml
│   ├── kustomization.yaml
│   └── service.yaml
├── dev
│   └── kustomization.yaml
├── prod
│   └── kustomization.yaml
└── stage
    └── kustomization.yaml
```

### Kustomize Overlays

```{revealjs-code-block} yaml
---
data-line-numbers: 1-7|8-14
---
# deploy/dev/kustomization.yaml
resources:
- ../base
namePrefix: dev-
namespace: development
commonLabels:
  environment: development
# deploy/prod/kustomization.yaml
resources:
- ../base
namePrefix: prod-
namespace: production
commonLabels:
  environment: production
```

### Kustomize Overlays

```{revealjs-code-block} console
---
data-line-numbers: 1-4|5-8
---
$ kubectl apply -k deploy/dev
service/dev-app created
deployment.apps/dev-app created
ingress.networking.k8s.io/dev-app created
$ kubectl apply -k deploy/prod
service/prod-app created
deployment.apps/prod-app created
ingress.networking.k8s.io/prod-app created
```

### Kustomize built-ins

```{revealjs-fragments}
* Transformers
* Generators
```

### Kustomize Transformers

```{revealjs-fragments}
* namePrefix
* nameSuffix
* commonAnnotations
* images
* commonLabels
* namespace
* replicas
* patches
```

### Kustomize Generators

```{revealjs-fragments}
* configMapGenerator
* secretGenerator
* helmCharts
```

### Helm

```{revealjs-fragments}
* Chart
* Release
* Repository
```

### Helm Chart

```{revealjs-code-block} console
---
data-line-numbers: 1-2|4-17|5|6|7-16|10-14|8|9|15-16|17
---
$ helm create chart
Creating chart
$ tree chart/
chart/
├── Chart.yaml
├── charts
├── templates
│   ├── NOTES.txt
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── hpa.yaml
│   ├── ingress.yaml
│   ├── service.yaml
│   ├── serviceaccount.yaml
│   └── tests
│       └── test-connection.yaml
└── values.yaml
```

### Helm Chart

```{revealjs-code-block} yaml
---
data-line-numbers: 1|2|3|4|5|6|7|8
---
# chart/Chart.yaml
apiVersion: v2
name: chart
description: A Helm chart for Kubernetes
type: application
version: 0.1.0
appVersion: "1.16.0"
dependencies: []
```

### Helm Templates

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|3-6|8-13|15
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "chart.fullname" . }}
  labels:
    {{- include "chart.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "chart.selectorLabels" . | nindent 4 }}
```

### Helm Templates

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|4|5-17|7|8-10|11-13|14-17
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  drink: {{ .Values.favorite.drink | default "tea" | quote }}
  {{ if eq .Values.favorite.drink "coffee" }}
  mug: "true"
  {{ end }}
  {{- with .Values.favorite }}
  food: {{ .food | upper | quote }}
  {{- end }}
  toppings: |-
  {{- range .Values.pizzaToppings }}
  - {{ . | title | quote }}
  {{- end }}
```

### Helm Template Objects

```{revealjs-fragments}
* Release
* Values
* Chart
* Files
* Capabilities
* Template
```

### Helm Template Functions

```{revealjs-fragments}
* Logic and Flow Control
* String
* Regular Expressions
* Encoding
* Math
* Cryptographic and Security

and many others(github.com/Masterminds/sprig)
```

### Helm Repo

```{revealjs-code-block} console
---
data-line-numbers: 1-5|6-7|8-10|11-14|15-19|20-24
---
$ helm search hub wordpress
URL                                                 CHART VERSION APP VERSION DESCRIPTION
https://hub.helm.sh/charts/bitnami/wordpress        7.6.7         5.2.4       Web publishing platform...
https://hub.helm.sh/charts/presslabs/wordpress-...  v0.6.3        v0.6.3      Presslabs WordPress
https://hub.helm.sh/charts/presslabs/wordpress-...  v0.7.1        v0.7.1      A Helm chart for deploy...
$ helm repo add brigade https://brigadecore.github.io/charts
"brigade" has been added to your repositories
$ helm repo list
NAME            URL
brigade         https://brigadecore.github.io/charts
$ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "opensearch" chart repository
Update Complete. ⎈Happy Helming!⎈
$ helm search repo brigade
NAME                          CHART VERSION APP VERSION DESCRIPTION
brigade/brigade               1.3.2         v1.2.1      Brigade provides event-driven scripting of...
brigade/brigade-github-app    0.4.1         v0.2.1      The Brigade GitHub App, an advanced gateway...
brigade/brigade-github-oauth  0.2.0         v0.20.0     The legacy OAuth GitHub Gateway for Brigade
$ helm show all brigade/brigade
# chart
# values
# crds
# readme
```

### Helm Release

```{revealjs-code-block} console
---
data-line-numbers: 1-9|10-12|13-17|18-27|28-31|32-33|34-38|39-40
---
$ helm install grafana/grafana --generate-name \
  -f values.yaml --set key=value
NAME: grafana-1684182240
LAST DEPLOYED: Mon May 15 23:24:01 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
NOTES:
...
$ helm list
NAME               NAMESPACE REVISION UPDATED             STATUS   CHART          APP VERSION
grafana-1684182240 default   1        2023-05-15 23:24:01 deployed grafana-6.56.4 9.5.2
$ helm get all grafana-1684182240
# hooks
# manifest
# notes
# values
$ helm upgrade grafana-1684182240 grafana/grafana \
  --set key=value2
Release "grafana-1684182240" has been upgraded. Happy Helming!
NAME: grafana-1684182240
LAST DEPLOYED: Mon May 15 23:29:51 2023
NAMESPACE: default
STATUS: deployed
REVISION: 2
NOTES:
...
$ helm history grafana-1684182240
REVISION UPDATED                  STATUS     CHART          APP VERSION DESCRIPTION
1        Mon May 15 23:24:01 2023 superseded grafana-6.56.4 9.5.2       Install complete
2        Mon May 15 23:29:51 2023 deployed   grafana-6.56.4 9.5.2       Upgrade complete
$ helm rollback grafana-1684182240 1
Rollback was a success! Happy Helming!
$ helm history grafana-1684182240
REVISION UPDATED                  STATUS     CHART          APP VERSION DESCRIPTION
1        Mon May 15 23:24:01 2023 superseded grafana-6.56.4 9.5.2       Install complete
2        Mon May 15 23:29:51 2023 superseded grafana-6.56.4 9.5.2       Upgrade complete
3        Mon May 15 23:31:50 2023 deployed   grafana-6.56.4 9.5.2       Rollback to 1
$ helm uninstall grafana-1684182240
release "grafana-1684182240" uninstalled
```
