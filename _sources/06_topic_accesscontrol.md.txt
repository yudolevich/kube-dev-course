# Access Control
![](slides/img/acctl-overview.svg)

## Authentication

Аутентификация в Kubernetes – это процесс проверки подлинности идентификационных данных пользователя,
который запрашивает доступ к ресурсам кластера. В Kubernetes есть несколько методов аутентификации,
включая:
* Файл сертификата клиента: при этом методе пользователь использует сертификат и закрытый ключ для
  аутентификации.
* Использование токенов: при этом методе пользователь использует токен доступа для аутентификации.
* Прокси аутентификация: при этом методе пользователь производит аутентификацию через прокси сервер,
  который добавляет информацию о пользователе в HTTP заголовки.

Аутентификацию в кластере можно пройти как обычный пользователь, либо как сервисный аккаунт, иначе
запрос будет считаться анонимным.

### Certificate
Аутентификация пользователей в Kubernetes с помощью сертификата - это один из способов подтверждения
личности пользователя. При использовании этого метода, пользователь использует свой клиентский сертификат
для подключения к Kubernetes API серверу.

Для того чтобы использовать этот метод аутентификации, необходимо создать сертификат и приватный ключ для
каждого пользователя и добавить их в соответствующие места в кластере. Как правило, это делается следующим
образом:
* Создание сертификатов и ключей для пользователей.
* Создание объекта CertificateSigningRequest (CSR) для каждого сертификата пользователя. CSR является
  запросом на подпись сертификата удостоверяющим центром (CA).
* Утверждение CSR для каждого пользователя, чтобы получить сертификаты, которые затем могут использоваться
  для аутентификации.

При использовании сертификатов для аутентификации, можно использовать Role-Based Access Control (RBAC)
для управления доступом пользователей к ресурсам Kubernetes в соответствии с определенными правилами.

### Token
Другой способ аутентификации в кластере с использованием токена - уникальной строки, идентифицирующей
пользователя, делающего запрос в API Kubernetes. В зависимости от конфигурации API токен может
генерироваться различными способами: [csv файл][static-token], [jwt sa][sa-token], [jwt oidc][oidc].

При каждом запросе токен передается в HTTP заголовке, таким образом идентифицируя пользователя:
```console
Authorization: Bearer 31ada4fd-adec-460c-809a-9e56ceb75269
```

Пример аутентификации с использованием [OpenID Connect][oidc]:
```{mermaid}
sequenceDiagram
  participant user as User
  participant idp as Identity Provider
  participant kube as Kubectl
  participant api as API Server
  user ->> idp: 1. Login to IdP
  activate idp
  idp -->> user: 2. Provide access_token,<br>id_token, and refresh_token
  deactivate idp
  activate user
  user ->> kube: 3. Call Kubectl<br>with --token being the id_token<br>OR add tokens to .kube/config
  deactivate user
  activate kube
  kube ->> api: 4. Authorization: Bearer...
  deactivate kube
  activate api
  api ->> api: 5. Is JWT signature valid?
  api ->> api: 6. Has the JWT expired? (iat+exp)
  api ->> api: 7. User authorized?
  api -->> kube: 8. Authorized: Perform<br>action and return result
  deactivate api
  activate kube
  kube --x user: 9. Return result
  deactivate kube
```

## Authorization
Авторизация в Kubernetes – это процесс контроля доступа к ресурсам API сервера на основе правил,
определяющих, какие действия могут выполнять пользователи, группы пользователей или сервисные аккаунты.
В Kubernetes авторизация осуществляется на уровне API сервера.

Для реализации авторизации Kubernetes использует модуль авторизации, который работает внутри API
сервера и проверяет права доступа к каждому запросу. Каждый запрос, поступающий на API сервер,
проходит через ряд проверок, в том числе аутентификацию и авторизацию.

Авторизация в Kubernetes может быть реализована с помощью различных механизмов, включая:
* RBAC (Role-Based Access Control) – это механизм контроля доступа на основе ролей. RBAC позволяет
  определять роли и разрешения для групп пользователей или сервисных аккаунтов. RBAC в Kubernetes
  позволяет гибко настраивать доступ к ресурсам и действиям в кластере.
* ABAC (Attribute-Based Access Control) – это механизм контроля доступа на основе атрибутов.
  ABAC позволяет определять разрешения на основе значений атрибутов, таких как имя пользователя,
  группа, IP-адрес и т.д. Однако использование ABAC не рекомендуется, так как он менее гибок и
  безопасен по сравнению с RBAC.
* Webhook – это механизм авторизации на основе внешнего приложения, которое может принимать решения о
  доступе на основе дополнительных данных, например, информации о состоянии сети или времени.
  Webhook используется для реализации более сложных политик авторизации.
* Node – это механизм авторизации, который позволяет назначать разрешения на уровне узла.
  Node Authorization используется для разграничения доступа к файловой системе и сетевым ресурсам
  на узлах.

### RBAC
В Kubernetes для управления доступом к ресурсам используется механизм Role-Based Access Control (RBAC).
RBAC позволяет определить, какие пользователи или группы пользователей имеют доступ к конкретным
ресурсам и какие действия они могут выполнять с этими ресурсами. RBAC состоит из нескольких компонентов:
Role, ClusterRole, RoleBinding и ClusterRoleBinding.

#### Role
Role - это ресурс Kubernetes, который определяет набор разрешений на выполнение операций с объектами
внутри одного неймспейса. К примеру, мы можем определить Role, которая позволяет пользователю только
просматривать секреты в неймспейсе.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-viewer
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]
```

В этом манифесте Role secret-viewer разрешено выполнять операции чтения (get, watch, list) с ресурсами
типа secrets.

#### ClusterRole
ClusterRole - это ресурс Kubernetes, который определяет набор разрешений на выполнение операций с
объектами во всем кластере. К примеру, мы можем определить ClusterRole, который позволяет пользователю
только просматривать все поды в кластере.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pod-viewer
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
```

В этом манифесте ClusterRole pod-viewer разрешено выполнять операции чтения (get, watch, list)
с ресурсами типа pods.

#### RoleBinding
RoleBinding - это ресурс Kubernetes, который связывает объект Role с пользователем или группой
пользователей в конкретном неймспейсе. К примеру, мы можем создать RoleBinding, которая связывает Role
secret-viewer с группой пользователей developers в неймспейсе my-namespace.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: secret-viewer-binding
  namespace: my-namespace
subjects:
- kind: Group
  name: developers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: secret-viewer
  apiGroup: rbac.authorization.k8s.io
```

В этом манифесте RoleBinding secret-viewer-binding связывает Role secret-viewer с группой пользователей
developers в неймспейсе my-namespace.

#### ClusterRoleBinding
ClusterRoleBinding - это ресурс Kubernetes, который связывает объект ClusterRole с пользователем или
группой пользователей во всем кластере. К примеру, мы можем создать ClusterRoleBinding, которая
связывает ClusterRole pod-viewer с конкретным пользователем.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pod-viewer-binding
subjects:
- kind: User
  name: jane
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: pod-viewer
  apiGroup: rbac.authorization.k8s.io
```

В этом манифесте ClusterRoleBinding pod-viewer-binding связывает ClusterRole pod-viewer с пользователем
jane выдавая права во всем кластере.



[static-token]:https://kubernetes.io/docs/reference/access-authn-authz/authentication/#static-token-file
[sa-token]:https://kubernetes.io/docs/reference/access-authn-authz/authentication/#service-account-tokens
[oidc]:https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens
