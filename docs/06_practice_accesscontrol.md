# Access Control
В данном практическом занятии рассмотрим процесс аутентификации с помощью сертификата и разделение
прав с [RBAC][].

## Authn
Для того, чтобы получить сертификат, которому будет доверять api kubernetes, необходимо сгенерировать
запрос на подпись - [Certificate Signing Request(csr)] и отправить на подтверждение в api. В данном
примере для генерации csr и ключа я воспользуюсь утилитой openssl, в Linux и MacOS данная утилита
обычно есть по умолчанию, для Windows [список источников для загрузки есть тут][openssl-bin].

### Generate csr and key
Генерацию ключа и [csr][] можно произвести одной командой:
```console
$ openssl req -new -newkey rsa:2048 -nodes -subj '/CN=Alex' -keyout key.pem -out csr.pem
Generating a 2048 bit RSA private key
.................................................+++++
...............................................+++++
writing new private key to 'key.pem'
-----
```
В данной команде используются опции:
* `req -new` - создание нового csr
* `-newkey rsa:2048` - генерация нового RSA ключа длиной 2048 бит
* `-nodes` - не использовать пароль для ключа
* `-subj '/CN=<name>'` - указание Subject для сертификата, где `name` - ваше имя
* `-keyout <path>` - путь и имя для сохранения ключа
* `-out <path>` - путь и имя для сохранения csr

После выполнения данной команды в директории, в которой вы находитесь, создадутся файлы
`key.pem` и `csr.pem`.

### kind: CertificateSigningRequest
Для возможности подписать наш csr внутрикластерным сертификатом удостоверяющего центра необходимо
создать объект [CertificateSigningRequest][csr-api].
```bash
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: myuser
spec:
  request: $(base64 < csr.pem)
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 86400  # one day
  usages:
  - client auth
EOF
```

В данном ресурсе указаны поля:
* `request` - содержимое файла `csr.pem` в base64 формате
* `signerName` - кем будет подписан сертификат
* `expirationSeconds` - время до истечения срока жизни сертификата
* `usage` - опции использования, в нашем случае `client auth`

После создания ресурса необходимо подтвердить выписывание сертификата для данного csr:
```console
$ kubectl certificate approve myuser
certificatesigningrequest.certificates.k8s.io/myuser approved
```

Когда ручное подтверждение было выполнено контроллер выпишет сертификат и добавить его в поле
`.status.certificate` в base64 формате нашего ресурса CertificateSigningRequest. Сохраним его в файл
`crt.pem` декодировав из base64:
```console
$ kubectl get csr myuser -o jsonpath='{.status.certificate}' | base64 -d > crt.pem
```

### kubectl credentials
Теперь у нас есть необходимые файлы для аутентификации в кластере - ключ `key.pem` и подписанный
сертификат `crt.pem`, осталось внести изменения в конфигурацию kubectl, чтобы воспользоваться ими.

Создаем пользовательскую конфигурацию `myuser` указав сертификат и ключ:
```console
$ kubectl config set-credentials myuser --client-certificate crt.pem --client-key key.pem --embed-certs
User "myuser" set.
```

Модифицируем текущий контекст, указав созданного пользователя, а также можем убедиться, что в
конфигурации применились наши изменения:
```console
$ kubectl config set-context --current --user myuser
Context "kind-kind" modified.
$ kubectl config view
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://192.168.56.2:6443
  name: kind-kind
contexts:
- context:
    cluster: kind-kind
    user: myuser
  name: kind-kind
current-context: kind-kind
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
    client-certificate-data: DATA+OMITTED
    client-key-data: DATA+OMITTED
- name: myuser
  user:
    client-certificate-data: DATA+OMITTED
    client-key-data: DATA+OMITTED
```

Как видно в конфигурации добавился новый пользователь - `myuser`, а также он применился к текущему
контексту в блоке `contexts`.

Можно попробовать сделать любой запрос в кластер:
```console
$ kubectl get pods
Error from server (Forbidden): pods is forbidden: User "Alex" cannot list resource "pods" in API group "" in the namespace "default"
```
Так как наш новый пользователь не имеет никаких прав, то мы получаем данное сообщение.

## RBAC Authz
Для выдачи прав нашему пользователю воспользуемся режимом авторизации
[Role-based access control (RBAC)][rbac].

### Role
Набор прав в неймспейсе можно задать с помощью ресурса `Role`:
```bash
$ kubectl config set-context --current --user kind-kind # вернем пользователя по-умолчанию
Context "kind-kind" modified.
$ kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
EOF
```
Как видно из правил, данная роль позволяет просматривать поды в неймспейсе default.

### Cluster Role
Набор прав на уровне всего кластера можно задать с помощью ресурса `ClusterRole`:
```bash
$ kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cm-reader
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "watch", "list"]
EOF
```
Данная роль создается на уровне кластера(cluster-wide) и дает права на чтение ресурса configmaps.

### RoleBinding
Чтобы присвоить права нашему пользователю в конкретном неймспейсе необходимо создать объект
`RoleBinding`:
```bash
$ kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-pods
  namespace: default
subjects:
- kind: User
  name: Alex
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io
EOF
```
Данный объект связывает созданную ранее роль `pod-reader` с нашим пользователем. Теперь можем
переключиться на него и проверить:
```console
$ kubectl config set-context --current --user myuser
Context "kind-kind" modified.
$ kubectl get pods
NAME        READY   STATUS    RESTARTS        AGE
pv-pod-sc   1/1     Running   3 (3h14m ago)   15d
```
Как видно теперь пользователь обладает правами указанными в объекте `Role` `pod-reader`.

### RoleBinding with Cluster Role
В объекте `RoleBinding` также можно ссылаться на `ClusterRole`, при этом нет необходимости создавать
данную роль в каждом неймспейсе, таким образом выдавая одни и те же права в различных неймспейсах.

```bash
$ kubectl config set-context --current --user kind-kind
$ kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-cm
  namespace: kube-system
subjects:
- kind: User
  name: Alex
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cm-reader
  apiGroup: rbac.authorization.k8s.io
EOF
```

```console
$ kubectl config set-context --current --user myuser
$ kubectl get cm
Error from server (Forbidden): configmaps is forbidden: User "Alex" cannot list resource "configmaps" in API group "" in the namespace "default"
$ kubectl get cm -n kube-system
NAME                                 DATA   AGE
coredns                              1      29d
extension-apiserver-authentication   6      29d
kube-proxy                           2      29d
kube-root-ca.crt                     1      29d
kubeadm-config                       1      29d
kubelet-config                       1      29d
```

Как видно пользователь получил права в конкретном неймспейсе `kube-system`.

### ClusterRoleBinding
Для получения прав во всем кластере, а не только в конкретных неймспейсах, необходимо воспользоваться
`ClusterRoleBinding`. Он позволяет связать пользователя с `ClusterRole` и выдать права на все
неймспейсы, либо права для cluster-wide объектов(например ноды).

```bash
$ kubectl config set-context --current --user kind-kind
$ kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: read-cm
subjects:
- kind: User
  name: Alex
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cm-reader
  apiGroup: rbac.authorization.k8s.io
EOF
```

```console
$ kubectl config set-context --current --user myuser
Context "kind-kind" modified.
$ kubectl get cm -A
NAMESPACE            NAME                                 DATA   AGE
default              kube-root-ca.crt                     1      29d
ingress-nginx        ingress-nginx-controller             1      29d
ingress-nginx        kube-root-ca.crt                     1      29d
kube-node-lease      kube-root-ca.crt                     1      29d
kube-public          cluster-info                         1      29d
kube-public          kube-root-ca.crt                     1      29d
kube-system          coredns                              1      29d
kube-system          extension-apiserver-authentication   6      29d
kube-system          kube-proxy                           2      29d
kube-system          kube-root-ca.crt                     1      29d
kube-system          kubeadm-config                       1      29d
kube-system          kubelet-config                       1      29d
local-path-storage   kube-root-ca.crt                     1      29d
local-path-storage   local-path-config                    4      29d
```

Как видно пользователь получил возможность чтения configmaps во всем кластере.


[csr]:https://en.wikipedia.org/wiki/Certificate_signing_request
[rbac]:https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[openssl]:https://www.openssl.org/
[openssl-bin]:https://wiki.openssl.org/index.php/Binaries
[csr-api]:https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/
