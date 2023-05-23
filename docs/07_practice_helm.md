# Kustomize/Helm

## Установка
Утилита `kustomize` существует в двух вариантах - как отдельная утилита, которую 
[можно скачать со страницы с релизами][kustomize-releases], и как встроенная в утилиту `kubectl`. Для
практических занятий воспользуемся встроенным функционалом.


Утилиту `helm` необходимо [скачать со страницы с релизами][helm-releases] и разместить в каталоге из
переменной `PATH`, вариант для Linux:
```console
curl -Ls https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz | tar xvz linux-amd64/helm
chmod +x linux-amd64/helm
sudo mv linux-amd64/helm /usr/local/bin/helm
```

## Kustomize
Опробуем работу `kustomize` в варианте, когда у нас есть базовая конфигурация и дополнительные модификации
для различных сред. В качестве основного приложения для деплоя возьмем образ nginx и в зависимости от
среды развертывания модифицируем хостнейм для Ingress и сообщение выдаваемое при запросе.

Создадим директории и неймспейсы для деплоя:
```console
$ mkdir base dev prod
$ kubectl create ns development
$ kubectl create ns production
```

### Base

Подготовим базовую конфигурацию из ресурсов Deployment, Service и Ingress, а также файл `kustomization` в
котором перечислены данные ресурсы:
```bash
cat <<EOF>> base/deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
EOF

cat <<EOF>> base/service.yaml
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: ClusterIP
EOF

cat <<EOF>> base/ingress.yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx
spec:
  rules:
  - host: localhost
    http:
      paths:
      - backend:
          service:
            name: nginx
            port:
              number: 80
        path: /
        pathType: Prefix
EOF

cat <<EOF>> base/kustomization.yaml
resources:
- deployment.yaml
- service.yaml
- ingress.yaml
EOF
```

### Development

Создадим `kustomize` конфигурацию для develop среды:
```bash
cat <<EOF>> dev/kustomization.yaml
resources:
- ../base
namespace: development
commonLabels:
  environment: development
configMapGenerator:
- name: index
  literals:
  - |
    index.html=develop server

  options:
    disableNameSuffixHash: true
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
        - name: nginx
          volumeMounts:
          - name: index
            mountPath: /usr/share/nginx/html
        volumes:
        - name: index
          configMap:
            name: index
patchesJson6902:
- target:
    group: networking.k8s.io
    version: v1
    kind: Ingress
    name: nginx
  patch: |-
    - op: replace
      path: /spec/rules/0/host
      value: dev.127.0.0.1.nip.io
    - op: add
      path: /spec/ingressClassName
      value: nginx
EOF
```

Здесь мы модифицируем лейблы и неймспейс, генерируем ConfigMap из одной строки, а также делаем патчи:
* для Deployment, добавляя `volume` с ConfigMap
* для Ingress, указывая поля `host` и `ingressClassName`

### Production

Аналогично создадим `kustomize` конфигурацию для production среды:
```bash
cat <<EOF>> prod/kustomization.yaml
resources:
- ../base
namespace: production
commonLabels:
  environment: production
configMapGenerator:
- name: index
  literals:
  - |
    index.html=PRODUCTION SERVER!!!

  options:
    disableNameSuffixHash: true
patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      spec:
        containers:
        - name: nginx
          volumeMounts:
          - name: index
            mountPath: /usr/share/nginx/html
        volumes:
        - name: index
          configMap:
            name: index
patchesJson6902:
- target:
    group: networking.k8s.io
    version: v1
    kind: Ingress
    name: nginx
  patch: |-
    - op: replace
      path: /spec/rules/0/host
      value: prod.127.0.0.1.nip.io
    - op: add
      path: /spec/ingressClassName
      value: nginx
EOF
```
Здесь мы меняем те же параметры, что и для develop среды.

### Deploy

Получается следующая структура каталогов:
```console
.
├── base
│   ├── deployment.yaml
│   ├── ingress.yaml
│   ├── kustomization.yaml
│   └── service.yaml
├── dev
│   └── kustomization.yaml
└── prod
    └── kustomization.yaml
```

Используя `kubectl apply -k` применим обе конфигурации:
```console
$ kubectl apply -k dev/
configmap/index created
service/nginx created
deployment.apps/nginx created
ingress.networking.k8s.io/nginx created
$ kubectl apply -k prod/
configmap/index created
service/nginx created
deployment.apps/nginx created
ingress.networking.k8s.io/nginx created
```

После деплоя можно наблюдать созданные ресурсы в каждом неймспейсе:
```console
$ kubectl get deploy,po,svc,ingress -n development
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx   1/1     1            1           24m
NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-6b9b58c86f-5888p   1/1     Running   0          16m
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/nginx   ClusterIP   10.96.242.213   <none>        80/TCP    24m
NAME                              CLASS   HOSTS                     ADDRESS     PORTS   AGE
ingress.networking.k8s.io/nginx   nginx   dev.127.0.0.1.nip.io   localhost   80      24m

$ kubectl get deploy,po,svc,ingress -n production
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx   1/1     1            1           8m20s
NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-79ffd89469-mzhtf   1/1     Running   0          8m20s
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/nginx   ClusterIP   10.96.189.189   <none>        80/TCP    8m20s
NAME                              CLASS   HOSTS                      ADDRESS     PORTS   AGE
ingress.networking.k8s.io/nginx   nginx   prod.127.0.0.1.nip.io   localhost   80      8m20s
```

Также видно, что в зависимости от среды приложение ведет себя по разному:
```console
$ curl dev.127.0.0.1.nip.io
develop server
$ curl prod.127.0.0.1.nip.io
PRODUCTION SERVER!!!
```

После можно удалить созданные неймспейсы вместе с ресурсами:
```console
$ kubectl delete ns development production
```

## Helm

Повторим такой же сценарий, но уже с использованием утилиты `helm`.

### Chart

Создадим новый чарт с именем *nginx* и удалим генерируемые шаблоны по-умолчанию, так как они нам не
понадобятся:
```console
$ helm create nginx
Creating nginx
$ rm -r nginx/templates/*
```

### Templates

Создадим шаблоны ресурсов Deployment, Service, Ingress и ConfigMap взяв за основу базовые ресурсы
из примера с `kustomize` и параметризовав те части, которые будут зависеть от среды развертывания:
```bash
cat <<EOF>> nginx/templates/deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: {{ .Values.image }}
        name: nginx
        volumeMounts: {{- toYaml .Values.volumeMounts | nindent 8 }}
      volumes: {{- toYaml .Values.volumes | nindent 6 }}
EOF
cat <<EOF>> nginx/templates/service.yaml
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: ClusterIP
EOF
cat <<EOF>> nginx/templates/ingress.yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx
spec:
  ingressClassName: {{ .Values.ingress.className }}
  rules:
  - host: {{ .Values.ingress.host }}
    http:
      paths:
      - backend:
          service:
            name: nginx
            port:
              number: 80
        path: /
        pathType: Prefix
EOF
cat <<EOF>> nginx/templates/configmap.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: index
data:
  index.html: |
    {{ .Values.indexData }}
EOF
```
Для шаблонизации используется [шаблонизатор языка go][go-template], подробнее о
[работе с шаблонами в helm можно почитать тут][helm-template-guide].

### Values

Также создадим файл `values.yaml`, который будет содержать параметры вставляемые в шаблон по-умолчанию и
отдельно файлы с параметрами для каждой среды:
```bash
cat <<EOF>> nginx/values.yaml
image: nginx

volumeMounts:
- name: index
  mountPath: /usr/share/nginx/html
volumes:
- name: index
  configMap:
    name: index

indexData: "index string"

ingress:
  host: localhost
  className: nginx
EOF
cat <<EOF>> nginx/values-dev.yaml
indexData: "develop server"

ingress:
  host: dev.127.0.0.1.nip.io
EOF
cat <<EOF>> nginx/values-prod.yaml
indexData: "PRODUCTION SERVER!!!"

ingress:
  host: prod.127.0.0.1.nip.io
EOF
```

### Deploy

В итоге у нас получится следующая структура каталогов:
```console
nginx/
├── Chart.yaml
├── charts
├── templates
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── ingress.yaml
│   └── service.yaml
├── values-dev.yaml
├── values-prod.yaml
└── values.yaml
```

Установим наш чарт с разными параметрами в отдельные неймспейсы:
```console
$ helm install nginx ./nginx/ -f nginx/values-dev.yaml --namespace development --create-namespace
NAME: nginx
LAST DEPLOYED: Tue May 23 00:01:23 2023
NAMESPACE: development
STATUS: deployed
REVISION: 1
TEST SUITE: None
$ helm install nginx ./nginx/ -f nginx/values-prod.yaml --namespace production --create-namespace
NAME: nginx
LAST DEPLOYED: Tue May 23 00:03:24 2023
NAMESPACE: production
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

После установки также можно наблюдать созданные ресурсы с разными параметрами:
```console
$ kubectl get deploy,po,svc,ingress -n development
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx   1/1     1            1           37s
NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-796d76f6f9-dn6vf   1/1     Running   0          37s
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/nginx   ClusterIP   10.96.147.211   <none>        80/TCP    38s
NAME                              CLASS   HOSTS                  ADDRESS     PORTS   AGE
ingress.networking.k8s.io/nginx   nginx   dev.127.0.0.1.nip.io   localhost   80      37s
$ kubectl get deploy,po,svc,ingress -n production
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx   1/1     1            1           35s
NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-796d76f6f9-vbmd9   1/1     Running   0          35s
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/nginx   ClusterIP   10.96.229.104   <none>        80/TCP    35s
NAME                              CLASS   HOSTS                   ADDRESS     PORTS   AGE
ingress.networking.k8s.io/nginx   nginx   prod.127.0.0.1.nip.io   localhost   80      36s

$ curl dev.127.0.0.1.nip.io
develop server
$ curl prod.127.0.0.1.nip.io
PRODUCTION SERVER!!!
```



[kustomize-releases]:https://github.com/kubernetes-sigs/kustomize/releases
[helm-releases]:https://github.com/helm/helm/releases
[go-template]:https://pkg.go.dev/text/template
[helm-template-guide]:https://helm.sh/docs/chart_template_guide/getting_started/
