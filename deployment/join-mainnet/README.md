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

### Running node
``` 
docker compose up -d
```

### Logging
````
docker logs -f --tail 10 u2u
