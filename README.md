# rdma-dummy-device-plugin

This device plugin is a dummy plugin. It 

## Build
To build simply run:
```
go install
```

## Building Docker
Builds a local docker image with the label `rdma-dummy-dp:latest`. Exports a tar file of the docker image labelled `rdma-dummy-dp.tar` into the local directoy which you can copy into other machines and load in with the command `docker load < rdma-dummy-dp.tar`.
```
VERSION=latest ./build_docker.sh
```

## Launch on Kubernetes
*NOTE* If you change the image than you need to update the yaml's image tag before launching it on Kubernetes.

Simply run the following command on your master node in Kubernetes:
```
kubectl create -f rdma-dummy-dp-ds.yaml
```

## Usage in Kubernetes YAML
To include it in the yaml file, make sure to add the following under the container spec:
```
resources:
    limits:
        rdma-sriov/vf: 1
```