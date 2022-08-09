# Docker Compose v2.9.0.1

A docker compose enhanced tool. 

> Base on [docker/compose v2.9.0](https://github.com/docker/compose), Follow official updates unscheduled.

---

Additional support: 

- HOOKs： Executing commands before/after creating containers
  - Shell 
  - Command
  - [Golang+ script](https://github.com/goplus/gop
  )(via [interpreter](https://github.com/goplus/igop)) 

- Copy file/folder from the image to the local filesystem.


## TOC

- [Install](#Install)
- [Usage](#Usage)
  - [Copy from image](#Copy-from-image)
  - [Hooks](#Hooks)

## Install

Copy the [release](https://github.com/fly-studio/docker-compose/releases) to 
```
/usr/libexec/docker/cli-plugins/docker-compose 
```

or (`ln -s` is recommended)

```
/usr/bin/docker-compose
```

## Usage

### Copy from image

```
docker compose [OPTIONS] cpi [SERVICE] [PATH_IN_IMAGE:LOCAL_PATH...] [--follow-link]
```

Copy a file/folder from the image of the SERVICE to the local filesystem

|                               |                                                                           |
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

### Hooks

```
docker compose [OPTIONS] deploy [SERVICE...] [OPTIONS_OF_UP] [--pull always] [--hook]
```

Creating and starting containers with HOOKs, the usage is similar to [docker compose up](docs/reference/compose_up.md).

|                 | Values                       | Default |                                                                                                                    |
|-----------------|------------------------------|---------|--------------------------------------------------------------------------------------------------------------------|
| [OPTIONS]       |                              |         | The options of [docker compose --help](docs/reference/compose.md#Options)                                          |
| [SERVICE...]    |                              |         | The list of services that you want to `up`                                                                         |
| [OPTIONS_OF_UP] |                              |         | The options of [docker compose up --help](docs/reference/compose_up.md#Options)                                    |
| --pull          | missing<br/>always<br/>never | missing | Pull the images before `up`. <br/> Reuse by the option of `--build` (build the images before starting containers.) |
| --hook          |                              | false   | Enable executing commands before/after `up`                                                                        | 

docker-compose.yml

|               | Types |                   |           |
|---------------|-------|-------------------|-----------|
| x-hooks       |       | Global<br/>Scoped | the hooks |
| · pre-deploy  | Array |                   | command   |
| · post-deploy | Array |                   | command   |

#### Examples

- Hooks
  - Copy the config files from the image of 'nginx'
  - Create a config file to /local/nginx/conf.d/vhosts.conf
- Mount the path of config to container of 'nginx'

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
        - ["sh", "-c", "echo 'include conf.d/vhosts/*.conf' > /local/nginx/conf.d/vhosts.conf"]  
      post-deploy:
        - ["echo", "scoped post-deploy"]
```

```
$ docker compose deploy --pull always --hook
```

#### Execution sequence

1. Pull images
2. Global _pre-deploy_ 
3. _pre-deploy_ of each service of `[SERVICE...]`
4. Up
5. _post-deploy_ of each service of `[SERVICE...]`
6. Global _post-deploy_

#### Relative path, Working directory

```
$ docker -f /a/b/docker-compose.yml deploy --hook
```

1. Working directory is the directory of `docker-compose.yaml`, eg: `/a/b/`

2. Path in the command is relative to the directory of `docker-compose.yaml`,
   - `- ["sh", "-c", "echo 'xxx' >> scripts/main.txt"`, the real path of `scripts/main.txt` is `/a/b/scripts/main.txt`


#### Command ADVANCED usage

 - Execute a Go+ script file, a Golang project
 - Execute Shell from `x-key`
 - Execute Golang script from `x-key`
 - Custom arguments

  see [Command of Hooks](docs/hooks-command.md)