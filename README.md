# Docker Compose v2.9.0.2

A docker compose enhanced tool. 

> Base on [docker/compose v2.9.0](https://github.com/docker/compose), Follow official updates unscheduled.

---

## ✨ Features: 

- HOOKs: Executing commands 
  - Hook `docker compose up`
  - Hook `docker compose down`
  - COMMAND
    - CLI
    - Shell 
    - [Golang+ script](https://github.com/goplus/gop
    )(via [interpreter](https://github.com/goplus/igop)) 

- Copy file/folder from the image to the local filesystem.


## TOC

- [Build](#build)
- [Install](#install)
- [Usage](#usage)
  - [up Hooks](#up-hooks)
  - [down Hooks](#down-hooks)
  - [Copy from image](#copy-from-image)

## ⌛ Build

```
$ cd /this/project/cmd
$ go build -ldflags "-X github.com/docker/compose/v2/internal.Version=2.9.0.2" -o docker-compose
```

## Install

Copy the [release](https://github.com/fly-studio/docker-compose/releases) to 
```
$ wget https://github.com/fly-studio/docker-compose/releases/download/v2.9.0.2/docker-compose -O /usr/libexec/docker/cli-plugins/docker-compose 
$ chmod +x /usr/libexec/docker/cli-plugins/docker-compose 
```

And copy or symlink to `/usr/bin`
```
$ cp /usr/libexec/docker/cli-plugins/docker-compose /usr/bin/docker-compose
```

## ⚡ Usage

### up Hooks

```
docker compose [OPTIONS] deploy [SERVICE...] [OPTIONS_OF_UP] [--pull always] [--hook]
```

Creating and starting containers with HOOKs, the usage is similar to [docker compose up](docs/reference/compose_up.md).

| Name            | Default | Description                                                                    |
|-----------------|---------|--------------------------------------------------------------------------------|
| [OPTIONS]       |         | The options of [docker compose --help](docs/reference/compose.md#Options)      |
| [SERVICE...]    |         | The list of services that you want to `up`                                     |
| [OPTIONS_OF_UP] |         | The options of [docker compose up --help](docs/reference/compose_up.md#Options) |
| --pull missing \| always \| never  | missing | Pull the images before `up`. <br/> Reuse by the option of `--build` (build the images before starting containers.) |
| --hook          | false   | Enable executing commands before/after `up`                                    | 

docker-compose.yml

| Name          | Types | Description                       |
|---------------|-------|-----------------------------------|
| x-hooks       |       |                                   |
| - pre-deploy  | Array | CLI, igo-key, igo-path, shell-key |
| - post-deploy | Array | CLI, igo-key, igo-path, shell-key |

#### Examples

- Hooks
  - Copy the config files from the image of 'nginx'
  - Create a config file to /local/nginx/conf.d/vhosts.conf
- Mount the path of config to container of 'nginx'

docker-compose.yml
```
x-hooks:  # Global
  pre-deploy:
    - ["echo", "global pre-deploy"]  
  post-deploy:
    - ["echo", "global post-deploy"]
    
services:
  nginx:
    image: nginx:latest
    volumes:
      - /local/nginx/:/etc/nginx/
    x-hooks:   # Scoped
      pre-deploy:
        - ["mkdir", "-p", "/local/nginx/conf.d/vhosts"]
        - ["docker", "compose", "cpi", "nginx", "/etc/nginx/:/local/"]
        - ["sh", "-c", "echo 'include conf.d/vhosts/*.conf;' > /local/nginx/conf.d/vhosts.conf"]  
      post-deploy:
        - ["echo", "scoped post-deploy"]
```

Deploy
```
$ docker-compose deploy --pull always --hook -d
```

#### Execution sequence

1. Pull images
2. Global _pre-deploy_ 
3. _pre-deploy_ of each service of _[SERVICE...]_
4. Up
5. _post-deploy_ of each service of _[SERVICE...]_
6. Global _post-deploy_

#### Relative path, Working directory

```
$ docker -f /a/b/docker-compose.yml deploy --hook
```

1. Working directory is the directory of `docker-compose.yaml`, eg: `/a/b/`

2. Path in the command is relative to the directory of `docker-compose.yaml`
  ```
  - ["sh", "-c", "echo 'xxx' >> scripts/main.txt"
  ```
  the real path of `scripts/main.txt` is `/a/b/scripts/main.txt`


#### Command ADVANCED usage

##### Go

Execute a Go+ script file, a Golang project
```
["igo-path", "/path/to/file.go"]
```

Execute Golang scripts from _x-key_
```
["igo-key", "x-key"]
```
##### Shell

Execute Shell from _x-key_
```
["shell-key", "x-key"]
```

Execute Shell file
```
["sh", "/path/to/file.sh"]
```

##### Custom arguments

Specify any arguments to _igo-key, igo-path, shell-key_

```
["igo-key", "x-key", "--argument1", "value1", "--argument2"]
```

See `examples/docker-compose.yaml`

### down Hooks

```
docker compose [OPTIONS] undeploy [SERVICE...] [OPTIONS_OF_DOWN] [--hook]
```

Stopping containers with HOOKs, the usage is similar to [docker compose down](docs/reference/compose_down.md).

| Name              | Default | Description                                                                         |
|-------------------|---------|-------------------------------------------------------------------------------------|
| [OPTIONS]         |         | The options of [docker compose --help](docs/reference/compose.md#Options)           |
| [SERVICE...]      |         | The list of services that you want to `down`                                        |
| [OPTIONS_OF_DOWN] |         | The options of [docker compose down --help](docs/reference/compose_down.md#Options) |
| --hook            | false   | Enable executing commands before/after `down`                                       | 

docker-compose.yml

| Name            | Types | Description                       |
|-----------------|-------|-----------------------------------|
| x-hooks         |       |                                   |
| - pre-undeploy  | Array | CLI, igo-key, igo-path, shell-key |
| - post-undeploy | Array | CLI, igo-key, igo-path, shell-key |

#### Execution sequence

1. Global _pre-undeploy_
2. _pre-undeploy_ of each service of _[SERVICE...]_
3. Down
4. _post-undeploy_ of each service of _[SERVICE...]_
5. Global _post-undeploy_

#### Examples

docker-compose.yml
```
services:
  nginx:
    x-hooks:
      post-undepoly:
        - ["rm", "-rf", "/local/nginx"]
```

Undeploy
```
$ docker-compose undeploy nginx --hook
```

### Copy from image

```
docker compose [OPTIONS] cpi [SERVICE] [PATH_IN_IMAGE:LOCAL_PATH...] [--follow-link]
```

Copy a file/folder from the image of the SERVICE to the local filesystem

| Name                          | Description                                                               |
|-------------------------------|---------------------------------------------------------------------------|
| [OPTIONS]                     | The options of [docker compose --help](docs/reference/compose.md#Options) |
| [SERVICE]                     | The service name that you want to copy from                               |
| [PATH_IN_IMAGE:LOCAL_PATH...] | Array                                                                     |
| · PATH_IN_IMAGE               | The source path in the image of the `[SERVICE]`                           |
| · LOCAL_PATH                  | The destination path of local filesystem                                  |
| --follow-link <br/>-L         | Always follow symbol link in `[PATH_IN_IMAGE]`                            |  


#### LOCAL_PATH

- Can be a DIRECTORY when `PATH_IN_IMAGE` is a file or directory
- Can be a FILE when `PATH_IN_IMAGE` is a file
- **The base directory of `LOCAL_PATH` MUST exist**

| PATH_IN_IMAGE | LOCAL_PATH folder | LOCAL_PATH file |
|---------------|-------------------|-----------------|
| folder        | √                 | ×               |
| file          | √                 | √               |

#### Examples

```
$ mkdir -p /local/nginx/  ## path must exist ##
$ docker compose -f "/a/b/docker-compose.yaml" cpi nginx \
  /etc/nginx/conf:/local/nginx/ \ 
  /etc/resolve.conf:/local/resolve.conf
```

