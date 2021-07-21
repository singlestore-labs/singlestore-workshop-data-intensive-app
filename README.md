# Workshop: Building data-intensive apps with SingleStore, Redpanda, and Golang

This repo provides a starting point for building a data-intensive application
using SingleStore, Vectorized Redpanda, and Golang. SingleStore is a scale-out
relational database built for data-intensive workloads. Redpanda is a Kafka API
compatible streaming platform for mission-critical workloads created by the team
at Vectorized.

Optionally, the rest of this readme will walk you through building a simple
application starting from the code here. If you follow the readme, you will
accomplish the following tasks:

1. Prepare your environment
2. Write a digital-twin (data simulation)
3. Define a SingleStore schema
4. Load the simulation data using Pipelines
5. Expose business logic via an HTTP API
6. Visualize your data

# 1. Prepare your environment

Before we can start writing code, we need to make sure that your environment is
setup and ready to go.

1. Make sure you clone this git repository to your machine
   ```bash
   git clone https://github.com/singlestore-labs/singlestore-workshop-data-intensive-app.git
   ```

2. This project uses docker & docker-compose
   - [docker][docker]
   - [docker-compose][docker-compose]

> ❗ **Note:** If you are following this tutorial on Windows or Mac OSX you may
> need to increase the amount of RAM and CPU made available to docker. You can
> do this from the docker configuration on your machine. More information:
> [Mac OSX documentation][docker-ramcpu-osx],
> [Windows documentation][docker-ramcpu-win]

> ❗ **Note 2:** If you are following this tutorial on a Mac with the new M1
> chipset, it is unlikely to work. SingleStore does not yet support running on
> M1. We are actively fixing this, but in the meantime please follow along using
> a Linux virtual machine.

3. A code editor that you are comfortable using, if you aren't sure pick one of
   the options below:
   - [Visual Studio Code][vscode]

     > _Recommended:_ this repository is setup with VSCode launch configurations
     > as well as [SQL Tools][sqltools] support.

   - [Jetbrains Goland][jetbrains-go]

## Test that your environment is working

Before proceeding, please run the `make` in the root of this repository to
verify that your environment is working as expected.

```bash
$ make
(... lots of docker output here ...)
   Name                  Command               State                                                                     Ports                                                                   
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
grafana       /run.sh                          Up      0.0.0.0:3000->3000/tcp,:::3000->3000/tcp                                                                                                  
prometheus    /bin/prometheus --config.f ...   Up      0.0.0.0:9090->9090/tcp,:::9090->9090/tcp                                                                                                  
redpanda      /bin/bash -c rpk config se ...   Up      0.0.0.0:29092->29092/tcp,:::29092->29092/tcp, 0.0.0.0:9092->9092/tcp,:::9092->9092/tcp, 0.0.0.0:9093->9093/tcp,:::9093->9093/tcp, 9644/tcp
simulator     ./simulator --config confi ...   Up                                                                                                                                                
singlestore   /startup                         Up      0.0.0.0:3306->3306/tcp,:::3306->3306/tcp, 3307/tcp, 0.0.0.0:8080->8080/tcp,:::8080->8080/tcp
```

If the command fails or doesn't end with the above status report, you may need
to start debugging your environment. If you are completely stumped please file
an issue against this repository and someone will try to help you out.

If the command succeeds, then you are good to go! Lets check out the various
services before we continue:

| service            | url                   | port | username | password |
| ------------------ | --------------------- | ---- | -------- | -------- |
| SingleStore Studio | http://localhost:8080 | 8080 | root     | root     |
| Grafana Dashboards | http://localhost:3000 | 3000 | root     | root     |
| Prometheus         | http://localhost:9090 | 9090 |          |          |

You can open the three urls directly in your browser and login using the
provided username & password.

In addition, you can also connect directly to SingleStore DB using any MySQL
compatible client or [VSCode SQL Tools][sqltools]. For example:

```
$ mariadb -u root -h 127.0.0.1 -proot
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 28
Server version: 5.5.58 MemSQL source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> select "hello world";
+-------------+
| hello world |
+-------------+
| hello world |
+-------------+
1 row in set (0.001 sec)
```

# 2. Write a digital-twin

Now that we have our environment setup

<!-- Link index -->

[docker-compose]: https://docs.docker.com/compose/install/
[docker-ramcpu-osx]: https://docs.docker.com/docker-for-mac/#resources
[docker-ramcpu-win]: https://docs.docker.com/docker-for-windows/#resources
[docker]: https://docs.docker.com/get-docker/
[jetbrains-go]: https://www.jetbrains.com/go/
[sqltools]: https://marketplace.visualstudio.com/items?itemName=mtxr.sqltools
[vscode]: https://code.visualstudio.com/
