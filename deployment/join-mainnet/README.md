# Prerequisites
Install Docker following the installation guide for Linux OS: [https://docs.docker.com/engine/installation/](https://docs.docker.com/engine/installation/)
* [CentOS](https://docs.docker.com/install/linux/docker-ce/centos) 
* [Ubuntu](https://docs.docker.com/install/linux/docker-ce/ubuntu)

Install docker compose
* [Docker compose](https://docs.docker.com/compose/install/)

Build dependencies
* [Network](https://docs.u2u.xyz/network/build-dependencies)

### Buiding image
```
cd <path>/go-u2u
make NET=mainnet u2u-image
```
### Download snapshot and rename file to `mainnet.g` into <path>/go-u2u if need to fast sync 
```
https://go-u2u-mainnet-snap.s3.ap-southeast-1.amazonaws.com/6321132-full.g
```
### Running node
``` 
docker compose up -d
```

### Logging
````
docker logs -f --tail 10 u2u
