## Storage

```{image} img/pv.svg
:width: 200px
```

### Storage

```{revealjs-fragments}
* Volumes
* StorageClass
* PersistentVolumeClaim
* PersistentVolume
* VolumeSnapshotClass
* VolumeSnapshot
```

### Volumes

```{image} img/vol.svg
:width: 500px
```

### Volumes

```{revealjs-code-block} yaml
---
data-line-numbers: 14|16|17-19|15|9-13
---
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
      readOnly: true
      subPath: /my-nfs-data/path
  volumes:
  - name: test-volume
    nfs:
      server: my-nfs-server.example.com
      path: /my-nfs-volume
      readOnly: true
```

### Ephemeral Volumes

### EmptyDir

```{revealjs-code-block} yaml
---
data-line-numbers: 14|15|16
---
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

### ConfigMap

```{image} img/cm.svg
:width: 500px
```

### ConfigMap

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|5-7
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: log-config
data:
  log_level: debug
  path: /data
```

### ConfigMap

```{revealjs-code-block} yaml
---
data-line-numbers: 14|15|16|17-18|19
---
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

```{image} img/secret.svg
:width: 500px
```

### Secret

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|5-7
---
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  key: dmFsdWU=
type: Opaque
```

### Secret

```{revealjs-code-block} yaml
---
data-line-numbers: 14|15|16|17
---
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

```{revealjs-code-block} yaml
---
data-line-numbers: 16-17|19-21|22-26
---
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

```{revealjs-fragments}
* ConfigMap
* Secret
* Downward API
* ServiceAccountToken
```

### Projected

```{revealjs-code-block} yaml
---
data-line-numbers: 15-20
---
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

### Persistent Volumes

```{image} img/pv-l.svg
:width: 500px
```

### Persistent Volumes

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|6-7|8|9-10|11|12|13-15|16-18
---
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

### Persistent Volume Claim

```{image} img/pvc.svg
:width: 500px
```

### Persistent Volume Claim

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|6-7|8|9-11|12|13-17
---
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

### Persistent Volume Claim

```{revealjs-code-block} yaml
---
data-line-numbers: 9-15
---
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

### Storage Classes

```{image} img/sc.svg
:width: 500px
```

### Storage Classes

```{revealjs-code-block} yaml
---
data-line-numbers: 1-2|5-7|8|9|10-11|12
---
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
