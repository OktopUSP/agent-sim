<p align="center">
    <img src="https://github.com/OktopUSP/agent-sim/assets/83298718/c9a6347d-4576-4ae1-a1ba-7a0c28883855"/>
</p>

# Introduction
This repository aims to promote the development of an agent TR-369 simulator to be used as an example of the protocol behavior in a embedded device. Our project is made based on [obuspa](https://github.com/BroadbandForum/obuspa).

## Features
- Load tests
- Controller development without specific hardware
- Simulate real-world operations
- Try TR-369 without buying devices
- Exploit controller vulnerabilities

### Usage
This simulator works by loading the Obusba project inside a docker container. There are several possible configurations you can customize.

**Requisites to run:**
- Docker
- GO

Before running the simulator make sure you download all go modules:

```shell
agent@sim:/home/apps/agent-sim$ go mod download
agent@sim:/home/apps/GitHub/agent-sim$ go mod tidy
```

It’s possible to configure flags either through the command line or by using the **.env** file in the project root folder. Below are the details for each available configuration flag:

**Options**
 
| CMD Argument    | ENV              | Default     | Description                                |
|-----------------|------------------|-------------|--------------------------------------------|
| `-sim_number`   | `SIM_NUM`        | `1`         | Number of simulated devices                |
| `-num_to_start_ids` | `NUM_TO_START_IDS` | `0`   | From where to start your IDs               |
| `-protocol`     | `MTP`            | `""`        | MTP to use (mqtt, stomp, websockets)       |
| `-mqtt_addr`    | `MQTT_ADDR`      | `localhost` | Address of the mqtt broker                 |
| `-mqtt_port`    | `MQTT_PORT`      | `1883`      | Port of the mqtt broker                    |
| `-mqtt_user`    | `MQTT_USER`      | `""`        | Mqtt user                                  |
| `-mqtt_passwd`  | `MQTT_PASSWD`    | `""`        | Mqtt password                              |
| `-mqtt_ssl`     | `MQTT_SSL`       | `false`     | Mqtt with tls/ssl                          |
| `-ws_addr`      | `WS_ADDR`        | `localhost` | Address of the websockets server           |
| `-ws_port`      | `WS_PORT`        | `8080`      | Port of the websockets server              |
| `-ws_route`     | `WS_ROUTE`       | `/ws/agent` | Route of the websockets server             |
| `-ws_ssl`       | `WS_SSL`         | `false`     | Websockets with tls/ssl                    |
| `-path`         | `PATH`           | `""`        | Folder path to save configurations         |
| `-imgpath`      | `DOCKERFILE_PATH`| `""`        | Path to Dockerfile                         |
| `-prefix`       | `PREFIX`         | `oktopus`   | Prefix of device id                        |

The application processes flags in the following order of priority:
```txt
1º - Flag through command line.
2º - Env variables.
3º - Default flag value.
```

**Command Line Examples**

Running a simulated MQTT client:
```shell
go run cmd/main.go -path=/home/apps/agent-sim/configs/ -protocol=mqtt  -mqtt_addr=192.168.10.159 -mqtt_port=1883
```

If the command succeeds, a Docker container will be built and run:

```shell
2025/09/09 15:54:19 Loaded variables from '.env'
2025/09/09 15:54:19 main.go:50: Starting Oktopus TR-369 Agent Simulator Version: 0.0.1
{"stream":"Step 1/21 : FROM ubuntu AS build-env"}
{"stream":"\n"}
{"stream":" ---\u003e 802541663949\n"}
{"stream":"Step 2/21 : RUN apt update \u0026\u0026 apt -y install         build-essential         libssl-dev         libcurl4-openssl-dev        libsqlite3-dev         libz-dev         autoconf         automake         libtool         libmosquitto-dev         pkg-config         git         cmake         make     \u0026\u0026 apt clean"}
{"aux":{"ID":"sha256:0d679f9e9582039599dafbca24a3a153415ae22244fe5ad0fe99aecf53fb3f1f"}}
{"stream":"Successfully built 0d679f9e9582\n"}
{"stream":"Successfully tagged obuspa:latest\n"}
2025/09/09 15:54:20 mqtt.go:29: Create new agent(s) with mqtt protocol
2025/09/09 15:54:20 mqtt.go:30: Mqtt client config: {Addr:192.168.10.159 Port:1883 User: Pass: Ssl:false}
2025/09/09 15:54:20 mqtt.go:47: Device: oktopus-0
2025/09/09 15:54:20 container.go:125: Container oktopus-0-mqtt started
```

Container **oktopus-0-mqtt** was created and will simulate a mqtt client connecting to host 192.168.10.159 on port 1883. 

Running 100 simulated mqtt clients:
```shell
go run cmd/main.go -sim_number=100 -path=/home/apps/agent-sim/configs/ -protocol=mqtt  -mqtt_addr=192.168.10.159 -mqtt_port=1883
```
100 docker containers will be created in this case. You can check them by using **docker ps** command:
```shell
docker ps
CONTAINER ID   IMAGE                           COMMAND                  CREATED          STATUS          PORTS                                                                                            NAMES
0a845c29eea7   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 13 seconds                                                                                                    oktopus-70-mqtt
ac41a997c6b3   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 11 seconds                                                                                                    oktopus-82-mqtt
999d5c888e49   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 13 seconds                                                                                                    oktopus-19-mqtt
e81bf0d80e67   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 14 seconds                                                                                                    oktopus-7-mqtt
3e3230a7a2b4   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 14 seconds                                                                                                    oktopus-9-mqtt
416621cc1af3   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 11 seconds                                                                                                    oktopus-81-mqtt
ef6127c163ae   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 10 seconds                                                                                                    oktopus-11-mqtt
6cae0db74dd6   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 13 seconds                                                                                                    oktopus-62-mqtt
451f8b65332a   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 10 seconds                                                                                                    oktopus-66-mqtt
55b11d721707   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 9 seconds                                                                                                     oktopus-67-mqtt
088fda26ed68   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 16 seconds                                                                                                    oktopus-87-mqtt
61f93f6dfd13   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 11 seconds                                                                                                    oktopus-80-mqtt
67288fc7e2d8   obuspa:latest                   "obuspa -p -v 4 -r /…"   19 seconds ago   Up 14 seconds                                                                                                    oktopus-64-mqtt
```

To stop and remove all containers type **ctrl+c**:
```shell
^C2025/09/09 16:04:50 mqtt.go:160: Deleted docker mqtt container: 1c93a54521dac6f1d811cd506e65d15b6b7870a0363d32516db74ff163484583
2025/09/09 16:04:50 mqtt.go:160: Deleted docker mqtt container: 5c936835a39ffc5be37d861041e808c52b9cfe1cb182198c404a0b6e2da576e4
2025/09/09 16:04:50 mqtt.go:160: Deleted docker mqtt container: 7c0844585f3b4e9cfa387c334ebd0cbedfb9ba55e30573030dff8ea885232515
2025/09/09 16:04:50 main.go:121: (⌐■_■) Agent simulator is out!
```

After running the agent simulator you can view them at the Devices page from the Oktopus controller:

![alt text](/img/devices.png)