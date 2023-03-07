# Работа с Pod

Для практики с подами нам также понадобится установленный `kind` и утилита `kubectl`, для удобства
можно стереть все данные от предыдущей работы и пересоздать кластер командами:
```console
$ kind delete cluster
$ kind create cluster
```

Либо можно удалить все оставшиеся поды в неймспейсе командой:
```console
$ k delete pod --all
pod "nginx" deleted
```

```{note}
Для удобства и компактности для команды `kubectl` будет использоваться псевдоним `k`.
Как его сделать описано в предыдущем практическом занятии.
```

Так как большинство полей в поде неизменяемые, то в этом занятии будет использоваться команда
`kubectl create` для создания пода, а вместо изменения под будет удаляться командой `kubectl delete`.

## Создание
Создать под можно командой `kubectl create -f -`, которая будет ожидать ввода из потока стандартного
ввода вместо файла, таким образом можно просто вставить текст в терминал, перевести на новую строку
и нажать `Ctrl+d`. Также можно воспользоваться синтаксисом [heredoc][] в шеле, чтобы передать на
стандартный ввод текст между разделителями, в данном случае `EOF`:
```bash
$ k create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: nginx
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
EOF
pod/nginx created
```

## Получение информации
Получить подробную информацию о поде в удобном формате можно командой `kubectl describe pod`:
```console
$ kubectl describe po nginx
Name:             nginx
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-worker/172.18.0.3
Start Time:       Mon, 06 Mar 2023 21:02:20 +0300
Labels:           run=nginx
Annotations:      <none>
Status:           Running
IP:               10.244.2.5
IPs:
  IP:  10.244.2.5
Containers:
  nginx:
    Container ID:   containerd://883f111c6fb037bc1991cf2bde34a91e4d3e8238e93aba1d54415243bd4f28c0
    Image:          nginx
    Image ID:       docker.io/library/nginx@sha256:aa0afebbb3cfa473099a62c4b32e9b3fb73ed23f2a75a65ce1d4b4f55a5c2ef2
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Mon, 06 Mar 2023 21:02:22 +0300
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-7jkrp (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  kube-api-access-7jkrp:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  27m   default-scheduler  Successfully assigned default/nginx to kind-worker
  Normal  Pulling    27m   kubelet            Pulling image "nginx"
  Normal  Pulled     27m   kubelet            Successfully pulled image "nginx" in 1.094512433s
  Normal  Created    27m   kubelet            Created container nginx
  Normal  Started    27m   kubelet            Started container nginx
```

Также информацию можно получить в различных форматах с помощью опции `-o` команды `kubectl get`.

Вывод только имен `name`:
```console
$ k get po -o name
pod/nginx
```

Расширенный вывод `wide`:
```console
$ k get po nginx -o wide
NAME    READY   STATUS    RESTARTS   AGE   IP           NODE          NOMINATED NODE   READINESS GATES
nginx   1/1     Running   0          31m   10.244.2.5   kind-worker   <none>           <none>
```

Формат `yaml/json`:
```console
$ k get po nginx -o yaml | head -15
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2023-03-06T18:02:20Z"
  labels:
    run: nginx
  name: nginx
  namespace: default
  resourceVersion: "19432"
  uid: ce208149-7f16-4226-b50a-f916854cbca3
spec:
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx
```

Шаблон `jsonpath`:
```console
$ k get po nginx -o jsonpath='{.spec.containers[0].image}'
nginx
```

Шаблон `go-template`:
```console
$ k get pod nginx -o go-template='{{range .spec.containers}}{{.image}}{{"\n"}}{{end}}'
nginx
```

[Заданный формат столбцов][custom-columns]:
```console
$ kubectl get pod nginx -o custom-columns=NAME:.metadata.name
NAME
nginx
```

Также очень популярная в использовании утилита [jq][] для форматирования вывода в виде `json`:
```console
$ k get po nginx -o json | jq '.status.conditions[] | select(.type == "ContainersReady") | .status'
"True"
```

## Логи
Получить логи контейнера можно командой `kubectl logs`.

```console
$ k logs nginx
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
/docker-entrypoint.sh: Configuration complete; ready for start up
2023/03/06 18:02:22 [notice] 1#1: using the "epoll" event method
2023/03/06 18:02:22 [notice] 1#1: nginx/1.23.3
2023/03/06 18:02:22 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2023/03/06 18:02:22 [notice] 1#1: OS: Linux 4.19.130-boot2docker
2023/03/06 18:02:22 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2023/03/06 18:02:22 [notice] 1#1: start worker processes
2023/03/06 18:02:22 [notice] 1#1: start worker process 35
2023/03/06 18:02:22 [notice] 1#1: start worker process 36
2023/03/06 18:02:22 [notice] 1#1: start worker process 37
2023/03/06 18:02:22 [notice] 1#1: start worker process 38
127.0.0.1 - - [06/Mar/2023:18:27:10 +0000] "GET / HTTP/1.1" 200 615 "-" "curl/7.79.1" "-"
```

Чтобы вывести только последние несколько строк, можно воспользоваться опцией `--tail=<n>`, где
`<n>` - количество последних строк.
```console
$ k logs nginx --tail=1
127.0.0.1 - - [06/Mar/2023:18:27:10 +0000] "GET / HTTP/1.1" 200 615 "-" "curl/7.79.1" "-"
```

Также можно ограничить вывод по времени опциями `--since=<t>`, где `<t>` - относительное время,
например `5m` - за последние 5 минут. А также опцией `--since-time=<f>`, где `<f>` - время
в формате [RFC3339][].
```console
$ k logs nginx --since-time='2023-03-06T18:10:00Z'
127.0.0.1 - - [06/Mar/2023:18:27:10 +0000] "GET / HTTP/1.1" 200 615 "-" "curl/7.79.1" "-"
```

Если процесс в контейнере не указывает временные отметки в логах, то можно их добавить опцией
`--timestamps`.
```console
$ k logs nginx --timestamps
2023-03-06T18:02:22.892400062Z /docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
2023-03-06T18:02:22.892484132Z /docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
2023-03-06T18:02:22.899665882Z /docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
2023-03-06T18:02:22.912301591Z 10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
2023-03-06T18:02:22.922720728Z 10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
2023-03-06T18:02:22.924250668Z /docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
2023-03-06T18:02:22.930897003Z /docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
2023-03-06T18:02:22.934077876Z /docker-entrypoint.sh: Configuration complete; ready for start up
2023-03-06T18:02:22.940752160Z 2023/03/06 18:02:22 [notice] 1#1: using the "epoll" event method
2023-03-06T18:02:22.940780609Z 2023/03/06 18:02:22 [notice] 1#1: nginx/1.23.3
2023-03-06T18:02:22.940791331Z 2023/03/06 18:02:22 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2023-03-06T18:02:22.940797071Z 2023/03/06 18:02:22 [notice] 1#1: OS: Linux 4.19.130-boot2docker
2023-03-06T18:02:22.940800309Z 2023/03/06 18:02:22 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2023-03-06T18:02:22.941055633Z 2023/03/06 18:02:22 [notice] 1#1: start worker processes
2023-03-06T18:02:22.941072933Z 2023/03/06 18:02:22 [notice] 1#1: start worker process 35
2023-03-06T18:02:22.941076973Z 2023/03/06 18:02:22 [notice] 1#1: start worker process 36
2023-03-06T18:02:22.941466310Z 2023/03/06 18:02:22 [notice] 1#1: start worker process 37
2023-03-06T18:02:22.941796944Z 2023/03/06 18:02:22 [notice] 1#1: start worker process 38
2023-03-06T18:27:10.252486288Z 127.0.0.1 - - [06/Mar/2023:18:27:10 +0000] "GET / HTTP/1.1" 200 615 "-" "curl/7.79.1" "-"
```

Создадим под с несколькими контейнерами:
```bash
$ k delete pod --all
$ k create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: nginx
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
  - image: busybox
    name: date
    command: [/bin/sh, -c, 'while sleep 10;do date;done']
EOF
```

Если контейнеров в поде несколько, то можно воспользоваться опцией `-c` или `--container` для
указания конкретного контейнера. Для получения логов всех контейнеров есть опция `--all-containers`.
Для того, чтобы было понятно какой именно контейнер вывел строку, можно добавить опцию `--prefix`.
```console
$ k logs nginx -c date
Mon Mar  6 19:34:49 UTC 2023
$ k logs nginx --all-containers --prefix --tail=5
[pod/nginx/date] Mon Mar  6 19:34:49 UTC 2023
[pod/nginx/date] Mon Mar  6 19:35:02 UTC 2023
[pod/nginx/date] Mon Mar  6 19:35:19 UTC 2023
[pod/nginx/nginx] 2023/03/06 19:34:37 [notice] 1#1: start worker processes
[pod/nginx/nginx] 2023/03/06 19:34:37 [notice] 1#1: start worker process 35
[pod/nginx/nginx] 2023/03/06 19:34:37 [notice] 1#1: start worker process 36
[pod/nginx/nginx] 2023/03/06 19:34:37 [notice] 1#1: start worker process 37
[pod/nginx/nginx] 2023/03/06 19:34:37 [notice] 1#1: start worker process 38

```

Для постоянного отслеживания новых сообщений в логе есть опция `-f` или `--follow`.
```console
$ k logs nginx -c date -f
Mon Mar  6 19:34:49 UTC 2023
Mon Mar  6 19:35:02 UTC 2023
Mon Mar  6 19:35:19 UTC 2023
Mon Mar  6 19:35:35 UTC 2023
Mon Mar  6 19:35:49 UTC 2023
Mon Mar  6 19:36:02 UTC 2023
```

## Проброс портов
Есть несколько способов отправлять сетевые запросы внутрь пода, наиболее простой - это проброс
портов на localhost командой `kubectl port-forward`.
```console
$ k port-forward nginx 8080:80 &
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```
Команда запускается с символом `&` в конце для того, чтобы она продолжила работу в фоновом режиме и
отдала нам управление терминалом. Из ее вывода видно, что запросы с адреса `127.0.0.1:8080` будут
перенаправляться в контейнер на порт `80`. Можно использовать команду `curl`, которая обычно есть
по умолчанию, для отправки HTTP запроса в контейнер.
```console
$ curl localhost:8080
Handling connection for 8080
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```
Как видно nginx нам ответил своей стандартной страницей приветствия.

## Запуск команд в контейнере
Запустить команду в контейнере можно с помощью `kubectl exec`, если контейнеров несколько, то
конкретный можно указать опцией `-c` или `--container`.
```console
$ k exec -c nginx nginx -- hostname
nginx
```
Для запуска в контейнере командной оболочки в интерактивном режиме следует добавить опции
`-i` - для возможности ввода через стандартный ввод и `-t` для выделения устройства псевдотерминала.
```console
$ k exec -it -c nginx nginx -- /bin/bash
root@nginx:/# pwd
/
root@nginx:/# echo 'hello' > /usr/share/nginx/html/index.html
root@nginx:/# exit
exit
$ curl localhost:8080
Handling connection for 8080
hello
```
Здесь видно как можно изменить файл `index.html` изнутри контейнера, который отдает nginx.

## Копирование файлов контейнера
Копировать файлы контейнера можно командой `kubectl cp`, для этого в контейнере должна быть
утилита `tar`. По сути это удобная обертка над командой `kubectl exec`. Конкретный контейнер
также можно указать опцией `-c` или `--container`.
```console
$ k cp -c nginx nginx:/usr/share/nginx/html/index.html ./index.html
tar: Removing leading `/' from member names
$ cat ./index.html
hello
$ echo 'world' > ./index.html
$ k cp ./index.html -c nginx nginx:/usr/share/nginx/html/index.html
$ curl localhost:8080
Handling connection for 8080
world
```

## Использование ConfigMap/Secret
Для передачи параметров внутрь контейнера есть удобный способ с использованием ресурсов
**ConfigMap** и **Secret**. Сами объекты представляют из себя структуру данных в виде ключ-значение.
В поде их можно использовать смонтировав в файловую систему как файл или в переменных среды.
```bash
$ k create -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: html
data:
  index.html: hello world
EOF
$ k create -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: envs
stringData:
  ENV_NAME: ENV_VALUE
EOF
$ k delete pod --all
$ k create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    envFrom:
    - secretRef:
        name: envs
    volumeMounts:
    - mountPath: /usr/share/nginx/html
      name: html
  volumes:
  - name: html
    configMap:
      name: html
EOF
```

После пересоздания пода проброс портов нужно сделать заново и можно убедиться, что данные из
**ConfigMap** и **Secret** попали в под.
```console
$ k port-forward nginx 8080:80 &
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```
```console
$ curl localhost:8080
Handling connection for 8080
hello world
$ k exec nginx -- /bin/sh -c 'echo $ENV_NAME'
ENV_VALUE
```


[heredoc]:https://en.wikipedia.org/wiki/Here_document
[custom-columns]:https://kubernetes.io/docs/reference/kubectl/#custom-columns
[jq]:https://stedolan.github.io/jq/
[RFC3339]:https://www.rfc-editor.org/rfc/rfc3339
