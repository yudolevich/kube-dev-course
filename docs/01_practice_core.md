# Основы работы с kubernetes

## kind
Для взаимодействия с кластером kubernetes нам в первую учередь необходим сам кластер.
Для локальной работы и обучения предлагаю воспользовать открытым проектом
[kind(Kubernetes in Docker)][kind], который позволяет запустить локальный кластер в Docker контейнере.

### Установка
Варианты установки описаны на сайте [kind.sigs.k8s.io][kind-install], можно воспользоваться
пакетным менеджером или достаточно просто скачать и положить в каталог из переменной `PATH`
исполняемый файл для своей ОС и архитектуры со страницы [releases][kind-releases].

Вариант для Linux:
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.17.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Создание кластера

Для создания кластера достаточно запустить команду:
```bash
kind create cluster
```

Посмотреть созданные кластера можно командой:
```bash
kind get clusters
```

## kubectl
Для взаимодействия с кластером нам понадобится утилита [kubectl][kubectl-overview] - это инструмент
командной строки для управления кластерами Kubernetes.

### Установка
Варианты установки описаны на сайте [kubernetes][kubectl-install] для каждой ОС. Здесь также достаточно
скачать исполняемый файл и положить в каталог из переменной `PATH`.

Вариант для Linux:
```bash
curl -Lo ./kubectl \
  "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```



[kind]:https://kind.sigs.k8s.io/
[kind-install]:https://kind.sigs.k8s.io/docs/user/quick-start/#installation
[kind-releases]:https://github.com/kubernetes-sigs/kind/releases
[kubectl-overview]:https://kubernetes.io/ru/docs/reference/kubectl/overview/
[kubectl-install]:https://kubernetes.io/docs/tasks/tools/#kubectl
