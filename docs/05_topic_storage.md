# Storage

Хоть и kubernetes очень удобен для разворачивания и управления нагрузкой, которая не имеет
состояния(stateless), он также предоставляет механизмы для хранения данных и работы приложений, которым
необходимо хранить в себе некоторое состояние(stateful). Данные механизмы позволяют не терять данные
между перезапусками контейнеров([Volumes][]), а также динамически выделять хранилище([StorageClass][sc],
[PersistentVolumeClaim][pvc], [PersistentVolume][pv]) и управлять снимками
хранилища([VolumeSnapshotClass][vsc], [VolumeSnapshot][vs]).

## Volumes
Для предотвращения потери данных между перезапусками контейнера в описании пода есть параметр
[.spec.volumes][volumes], в котором можно указать описание тома, где будут храниться данные между
перезапусками. Также тома можно смонтировать одновременно сразу к нескольким контейнерам, тем самым
обеспечив связь между ними.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pd
spec:
  containers:
  - image: registry.k8s.io/test-webserver
    name: test-container
    volumeMounts:
    - mountPath: /my-nfs-data
      name: test-volume
  volumes:
  - name: test-volume
    nfs:
      server: my-nfs-server.example.com
      path: /my-nfs-volume
      readOnly: true
```

В данном примере показан [volume типа nfs][volume-nfs], который монтируется в конкретный контейнер
через параметр `volumeMounts`. Связь между `volume` пода и `volumeMounts` контейнера производится через
параметр `name`. Volume поддерживает множество различных типов томов: configMap, secret, emptyDir,
fc, iscsi, nfs, cephfs, rbd и другие. Помимо встроенных типов есть возможность добавление в кластер
дополнительных типов с помощью [Container Storage Interface (CSI)][csi].

## Ephemeral Volumes
Не всем приложениям нужно хранить данные между пересозданием пода, например данные временных файлов,
кэша, логов и т.д. Также приложению могут потребоваться данные из api kubernetes, которые можно
смонтировать внутрь пода. Для этого подойдут эфемерные тома, которые создаются только на время жизни пода.

### EmptyDir
Это пустая директория, которая создается и монтируется в под после его создания и назначения на ноду.
Данный тип удобно использовать для временного хранения файлов, которые нужны только во время работы
приложения, либо для обмена данными между контейнерами в поде.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pd
spec:
  containers:
  - image: registry.k8s.io/test-webserver
    name: test-container
    volumeMounts:
    - mountPath: /cache
      name: cache-volume
  volumes:
  - name: cache-volume
    emptyDir:
      medium: Memory
      sizeLimit: 500Mi
```

По умолчанию размер тома не ограничивается(только размером файловой системы на ноде), но может быть
ограничен дополнительным параметром `sizeLimit` и через [resources][eph-resources]. Через параметр
`medium: Memory` можно указать хранение данных не на файловой системе, а в оперативной памяти.

### ConfigMap
[ConfigMap][cm] в Kubernetes предназначен для хранения конфигурационных данных, которые могут быть
использованы внутри приложения, запущенного в контейнере. ConfigMap Volume позволяет использовать
данные [ConfigMap][cm] в виде файлов внутри контейнера. Для использования ConfigMap Volume необходимо
сначала создать [ConfigMap][cm], который будет содержать необходимые конфигурационные данные.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: log-config
data:
  log_level: debug
```

Затем можно создать Volume в поде, используя данные из [ConfigMap][cm].

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: configmap-pod
spec:
  containers:
    - name: test
      image: busybox:1.28
      volumeMounts:
        - name: config-vol
          mountPath: /etc/config
  volumes:
    - name: config-vol
      configMap:
        name: log-config
        items:
          - key: log_level
            path: log_level
        optional: false
```

### Secret
Также в Kubernetes существует ресурс [Secret][], который используется для хранения конфиденциальных
данных, таких как пароли, ключи API и сертификаты.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  key: dmFsdWU=
type: Opaque
```

[Secret][] может быть использован в качестве тома в Pod.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  containers:
  - name: mypod
    image: redis
    volumeMounts:
    - name: foo
      mountPath: "/etc/foo"
      readOnly: true
  volumes:
  - name: foo
    secret:
      secretName: mysecret
      optional: true
```

### DownwardAPI
В Kubernetes существует возможность для контейнеров получать информацию об окружении, в котором они
запущены, с помощью механизма Downward API. Downward API позволяет контейнерам получить информацию о:
* лейблах пода
* полях пода
* имени и уникальном идентификаторе пода
* имени и уникальном идентификаторе неймспейса

Для использования Downward API можно использовать volume типа downwardAPI, который включает в себя файлы,
содержащие запрашиваемую информацию.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: downwardapi-example
  labels:
    app: myapp
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - name: downwardapi
      mountPath: /etc/podinfo
      readOnly: true
  volumes:
  - name: downwardapi
    downwardAPI:
      items:
      - path: "metadata/labels"
        fieldRef:
          fieldPath: metadata.labels
      - path: "mem_limit"
        resourceFieldRef:
          containerName: client-container
          resource: limits.memory
          divisor: 1Mi
```

### Projected
Projected volume в Kubernetes позволяет объединять в одном volume несколько источников данных, таких как
ConfigMap, Secret, Downward API и ServiceAccountToken, и предоставлять их как один файловый системный
раздел в контейнере. Это может быть полезно, если контейнеру требуется доступ к нескольким ресурсам
конфигурации или сервисным данным.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: projected-example
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - name: config
      mountPath: /etc/config
      readOnly: true
  volumes:
  - name: config
    projected:
      sources:
      - configMap:
          name: myconfigmap
      - secret:
          name: mysecret
```

## Persistent Volumes
[Persistent Volumes (PV)][pv] в Kubernetes - это абстракция над физическим хранилищем данных, которая
позволяет абстрагировать хранилище от приложений, которые его используют. [PV][] - это ресурс Kubernetes,
который описывает хранилище данных, такое как диск, NFS-шару или том в облаке, и его характеристики,
такие как доступность, емкость и режим доступа. [PV][] можно создать заранее, до того, как к нему
обратится приложение, а потом использовать его как ресурс для подов. Когда приложение запрашивает [PV][],
Kubernetes выбирает наиболее подходящее свободное хранилище на основе его параметров.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0003
spec:
  capacity:
    storage: 5Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: slow
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /tmp
    server: 172.17.0.2
```

[Persistent Volume Claim (PVC)][pvc] - это способ запроса ресурса [PV][] в Kubernetes. Когда [PVC][]
создается, Kubernetes выбирает [PV][], который соответствует критериям [PVC][] (размер,
доступность и т.д.), и резервирует его для использования.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: myclaim
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 1Gi
  storageClassName: slow
  selector:
    matchLabels:
      release: "stable"
    matchExpressions:
      - {key: environment, operator: In, values: [dev]}
```

Поды могут использовать [PVC][], чтобы монтировать том в свои контейнеры.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  containers:
    - name: myfrontend
      image: nginx
      volumeMounts:
      - mountPath: "/var/www/html"
        name: mypd
  volumes:
    - name: mypd
      persistentVolumeClaim:
        claimName: myclaim
```

## Storage Classes
[StorageClass][sc] - это объект, который определяет класс хранилища для [PersistentVolume (PV)][pv].
[StorageClass][sc] определяет способ, которым [PV][] будет создан, конфигурирован и управляться.
Он также определяет тип хранилища, такой как NFS, iSCSI, Ceph, AWS EBS, GCP PD и т.д., и параметры
хранилища, такие как размер, скорость, доступность и т.д.

[StorageClass][sc] может использоваться для динамического выделения [PV][]. Это означает, что при
создании пода, если нужный [PV][] не существует, [StorageClass][sc] может автоматически создать его
для пода. Например, если поду требуется [PersistentVolumeClaim (PVC)][pvc] определенного размера,
и нет доступного [PV][], [StorageClass][sc] может создать новый [PV][] нужного размера и назначить
его [PVC][]. Это может быть очень удобно для разработчиков, которые не хотят создавать и управлять
[PV][] вручную.

Кроме того, разные приложения могут использовать разные классы хранилищ, а [StorageClass][sc]
обеспечивает способ управления и мониторинга этих классов в едином месте.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
reclaimPolicy: Retain
allowVolumeExpansion: true
mountOptions:
  - debug
volumeBindingMode: Immediate
```





[volumes]:https://kubernetes.io/docs/concepts/storage/volumes/
[volume-nfs]:https://kubernetes.io/docs/concepts/storage/volumes/#nfs
[sc]:https://kubernetes.io/docs/concepts/storage/storage-classes/
[pv]:https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistent-volumes
[pvc]:https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims
[vsc]:https://kubernetes.io/docs/concepts/storage/volume-snapshot-classes/
[vs]:https://kubernetes.io/docs/concepts/storage/volume-snapshots/
[csi]:https://kubernetes.io/docs/concepts/storage/volumes/#csi
[eph-resources]:https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#setting-requests-and-limits-for-local-ephemeral-storage
[cm]:https://kubernetes.io/docs/concepts/configuration/configmap/
[secret]:https://kubernetes.io/docs/concepts/configuration/secret/
[downwardapi]:https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/
