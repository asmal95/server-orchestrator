This service provides the ability to manage the docker configuration on your server using a telegram bot.

## How To

To get a list of deployments sends the `/deployments` command. It will display the list of deployments.
To see details click on interested deployment.

To create a new deployment send the `/new_deployment` command and follow the bot's commands.

## Launch

Because the service is used to manage your docker container, it should be run on the host machine.

### Configuration

Before starting the service please provide the following `config.yaml` file next to the executable:
```yaml
bot:
  name: <bot_name>
  token: <bot_token>
  debug: <is_debug>
docker-orchestrator:
  config-location: <json_config_file>
  synchronization-interval: <synchronization_interval>
  network-name: <network_name>
admin-chat-id: <admin_chat_id>
super-user-id: <super_user_id>
```
where:
* `bot_name` - the parameter specifies telegram bot name.
* `bot_token` - the parameter specifies telegram token. **Required**
* `is_debug` - the parameter specifies is the debug enabled for the telegram bot. Default value is `false`.
* `json_config_file` - the parameter specifies location for the configuration file in the `json` format. **Required**
* `synchronization_interval` - the parameter specifies the duration between synchronization of the managed deployments with actual docker containers. The duration is a sequence of decimal numbers, each with a unit suffix, such as `30s` or `1m30s`. Valid time units are `ms`, `s`, `m`, `h`. The default value is `30s`.
* `network_name` - the parameter specifies the network name for your containers in the docker.
* `admin_chat_id` - the parameter specifies the admin chat id, where the bot will respond to your messages.
* `super_user_id` - the parameter specifies the super user id, to whom the bot will respond in the private chat.

Also you can override the configuration with deployment variables:
* `TG_BOT_NAME` - env variable for `bot_name`.
* `TG_BOT_TOKEN` - env variable for `bot_token`.
* `TG_BOT_DEBUG` - env variable for `is_debug`.
* `DOCKER_ORCHESTRATOR_CONFIG_LOCATION` - env variable for `json_config_file`.
* `DOCKER_ORCHESTRATOR_SYNCHRONIZATION_INTERVAL` - env variable for `synchronization_interval`.
* `DOCKER_ORCHESTRATOR_NETWORK_NAME` - env variable for `network_name`.
* `ADMIN_CHAT_ID` - env variable for `admin_chat_id`.
* `SUPER_USER_ID` - env variable for `super_user_id`.

### Start and Stop

You can use the follow `bash` script to run the service:
```shell
#!/bin/bash

nohup ./server-orchestrator > log.txt &
```
Or use the `start.sh` file.

To stop the service you can use the `stop.sh` file. The file will find the process by name and terminate it.
### Build

To build the service you can use the standard `go build` command: 
```shell
go build
```

It will produce executable `server-orchestrator` file.

If you need to make the `go.sub` file, please use the `go mod tidy` command:
```shell
go mod tidy
```

Or use the `build.sh` file.