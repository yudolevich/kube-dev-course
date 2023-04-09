# LoadBalance

В данном практическом занятии предлагаю опробовать работу [Ingress][] используя
[Nginx Ingress Controller][nginx-ic] и посмотреть на процесс балансировки трафика в момент обновления
версии приложения.

## Kind

Для корректной работы [Ingress Controller][ic] в kind необходимы дополнительные настройки кластера,
для этого проще всего создать новый кластер с дополнительным блоком конфигурации,
который позволит получить доступ к подам [Ingress Controller][ic] с нашей машины.

Можно создать новый кластер с другим именем, либо сначала удалить старый:
```console
$ kind delete cluster
```

Дополнительная конфигурация выставляет наружу порты 80 и 443, а также добавляет лейбл к ноде.
```bash
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

## Nginx Ingress

Для установки [Nginx Ingress Controller][nginx-ic] достаточно выполнить команду:
```console
$ k apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

После которой в неймспейсе **ingress-nginx** создадутся необходимые ресурсы и запустится под контроллера.
```console
$ k get po -n ingress-nginx
NAME                                        READY   STATUS      RESTARTS     AGE
ingress-nginx-admission-create-hkt6j        0/1     Completed   0            5h57m
ingress-nginx-admission-patch-k2wmt         0/1     Completed   0            5h57m
ingress-nginx-controller-69dfcc796b-zm5pt   1/1     Running     1 (5h ago)   5h57m
```

```{warning}
Существует как минимум две реализации Ingress Controller с использованием Nginx:
[от компании Nginx Inc][nginx-off] и [от kubernetes сообщества][nginx-ic].
В данном руководстве используется реализация от сообщества!
```

## Application

Для проверки работы [Ingress][] сделаем простое приложение, которое будет отвечать на HTTP запросы:
* GET / - будет возвращать HTTP код 200 и сообщение "ok"
* GET /version - будет возвращать HTTP код 200 и сообщение с текущей версией приложения,
  а в лог также выводить количество запросов, которое было обработано по данному пути

Вы можете написать его на своем любимом языке, в качестве примера приведу простую реализацию на Golang:
```golang
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	Version  string
	reqCount uint
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok\n")
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		fmt.Printf("GET /version - %s, count - %d\n", Version, reqCount)
		io.WriteString(w, fmt.Sprintf("version: %s\n", Version))
	})

	fmt.Println("start server")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("error serve http: %s\n", err)
		os.Exit(1)
	}
}
```

Версию приложения предлагаю указывать при сборке образа в Dockerfile:
```dockerfile
FROM golang:1.20-alpine as build

ARG BUILD_VERSION
WORKDIR /build

COPY . ./
RUN GOOS=linux GOARCH=amd64 go build \
  -ldflags="-X main.Version=${BUILD_VERSION}" \
  -a -o app main.go

FROM alpine:3.17

COPY --from=build /build/app /app

CMD ["/app"]
```

В данном примере версия приложения будет указана в итоговом бинарном файле, переданная через
аргумент `BUILD_VERSION`.

```console
$ docker build --build-arg BUILD_VERSION=1.0 -t app:1.0 .
Sending build context to Docker daemon  5.632kB
Step 1/8 : FROM golang:1.20-alpine as build
 ---> 818ca3531f99
Step 2/8 : ARG BUILD_VERSION
 ---> Using cache
 ---> 9a3dd54335f8
Step 3/8 : WORKDIR /build
 ---> Using cache
 ---> 9c28030abe9e
Step 4/8 : COPY . ./
 ---> 3c00293d730e
Step 5/8 : RUN GOOS=linux GOARCH=amd64 go build   -ldflags="-X main.Version=${BUILD_VERSION}"   -a -o app main.go
 ---> Running in dfb517e386a7
Removing intermediate container dfb517e386a7
 ---> 42fe6c7a47c5
Step 6/8 : FROM alpine:3.17
 ---> b2aa39c304c2
Step 7/8 : COPY --from=build /build/app /app
 ---> 68d71f56dd53
Step 8/8 : CMD ["/app"]
 ---> Running in 448b41e615d3
Removing intermediate container 448b41e615d3
 ---> 8f797e1b1ccb
Successfully built 8f797e1b1ccb
Successfully tagged app:1.0
```

Также соберем версию 2.0:
```console
$ docker build --build-arg BUILD_VERSION=2.0 -t app:2.0 .
...
Successfully built 5b5feec48cb7
Successfully tagged app:2.0
```

И загрузим в наш кластер:
```console
$ kind load docker-image app:1.0
Image: "" with ID "sha256:8f797e1b1ccb95ef25b846714a5162ec80aee72b102c78f7d188e52bf119a8c1" not yet present on node "kind-control-plane", loading...
$ kind load docker-image app:2.0
Image: "" with ID "sha256:5b5feec48cb7b016637584224e59c64c5cb8e107226384f5bdc3c587da5b0fb9" not yet present on node "kind-control-plane", loading...
```

## Deploy

Создадим ресурс Deployment с нашим приложением в трех репликах:
```console
$ k create deployment app --image=app:1.0 --replicas 3
deployment.apps/app created
```

Создадим ресурс Service, указав его порт и порт, который слушает приложение:
```console
$ k expose deployment app --port 80 --target-port 8080
service/app exposed
```

Создадим ресурс Ingress, указав в качестве хоста - localhost и сервис app с портом 80:
```console
$ k create ingress app --rule=localhost/*=app:80
ingress.networking.k8s.io/app created
```

Убедимся, что все ресурсы созданы и поды запущены:
```console
$ k get deploy,svc,ingress,pod
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/app   3/3     3            3           17m

NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/app          ClusterIP   10.96.219.251   <none>        80/TCP    14m
service/kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP   6h57m

NAME                            CLASS    HOSTS       ADDRESS     PORTS   AGE
ingress.networking.k8s.io/app   <none>   localhost   localhost   80      5m5s

NAME                       READY   STATUS    RESTARTS   AGE
pod/app-64885bbddf-5f5g7   1/1     Running   0          17m
pod/app-64885bbddf-gtnt6   1/1     Running   0          17m
pod/app-64885bbddf-mdnhs   1/1     Running   0          17m
```

Теперь доступ к приложению возможен через localhost:
```console
$ curl localhost
ok
$ curl localhost/version
version: 1.0
```

## LoadBalance

Убедимся, что трафик распределяется на все поды с приложением. Для этого удалим старые поды и убедимся,
что новые запущены:
```console
$ k delete po -l app=app
pod "app-64885bbddf-5f5g7" deleted
pod "app-64885bbddf-gtnt6" deleted
pod "app-64885bbddf-mdnhs" deleted
$ k get po
NAME                   READY   STATUS    RESTARTS   AGE
app-64885bbddf-886x2   1/1     Running   0          10s
app-64885bbddf-9rrh9   1/1     Running   0          10s
app-64885bbddf-lfjdz   1/1     Running   0          10s
```

Сделаем в цикле 100 запросов к нашему Ingress и посмотрим в логах каждого пода,  сколько запросов
он обработал:
```console
$ for i in {1..100};do curl localhost/version -so /dev/null;done
$ k logs -l app=app --tail=1
GET /version - 1.0, count - 34
GET /version - 1.0, count - 33
GET /version - 1.0, count - 33
```

Как видно из вывода трафик равномерно распределился между подами. [Nginx Ingress Controller][nginx-ic]
по-умолчанию использует алгоритм балансировки *round robin*, изменить это поведение можно с помощью
[аннотаций на ресурсе Ingress][lb-alg].

## RollingUpdate

Посмотрим, как в процессе обновлении будут выглядеть ответы от приложения.

Для начала увеличим количество реплик, к примеру, до 20, чтобы процесс обновления занял более
продолжительное время:
```console
$ k scale deploy/app --replicas 20
deployment.apps/app scaled
```

Для наблюдения за процессом обновления в отдельном окне терминала запустим цикл, который раз в секунду
будет делать запрос к нашему приложению:
```console
$ while sleep 1;do echo -n "$(date '+%H:%M:%S') ";curl localhost/version;done
```

И обновим образ на версию 2.0:
```console
$ k set image deploy/app app=app:2.0
deployment.apps/app image updated
```

В нашем цикле будет видно как постепенно версия 1.0 сменяется на 2.0:
```console
21:07:01 version: 1.0
21:07:02 version: 1.0
21:07:03 version: 1.0
21:07:04 version: 1.0
21:07:05 version: 1.0
21:07:06 version: 1.0
21:07:07 version: 1.0
21:07:08 version: 2.0
21:07:09 version: 1.0
21:07:10 version: 2.0
21:07:11 version: 1.0
21:07:12 version: 2.0
21:07:13 version: 2.0
21:07:14 version: 2.0
21:07:20 version: 2.0
21:07:21 version: 2.0
```

Пока все поды полностью не обновятся и останется только версия 2.0.

Попробуем также откатиться обратно:
```console
$ k rollout undo deploy/app
deployment.apps/app rolled back
```

В нашем цикле мы также увидим процесс смены версии в обратную сторону:
```console
21:19:14 version: 2.0
21:19:15 version: 2.0
21:19:16 version: 2.0
21:19:17 version: 2.0
21:19:23 version: 1.0
21:19:29 version: 2.0
21:19:30 version: 2.0
21:19:31 version: 1.0
21:19:37 version: 1.0
21:19:38 version: 1.0
21:19:39 version: 1.0
```

Все примеры, используемые здесь, можно также посмотреть на [github][example].

[ingress]:https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[ic]:https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[nginx-ic]:https://kubernetes.github.io/ingress-nginx/
[nginx-off]:https://www.nginx.com/products/nginx-ingress-controller/
[lb-alg]:https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#load-balance
[example]:https://github.com/yudolevich/kube-dev-course/tree/main/examples/netbalance/
