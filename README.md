This service provides the ability to manage the docker configuration on your server using a telegram bot.

## Launch

Because the service is used to manage your docker container, it should be run on the host machine.

To start the service please provide the following `config.yaml` file:
```yaml
bot:
  name: <bot_name>
  token: <bot_token>
  debug: <is_debug>
docker-orchestrator:
  config-location: <json_config_file>
```
where:
* `bot_name` - the parameter specifies telegram bot name.
* `bot_token` - the parameter specifies telegram token. **Required**
* `is_debug` - the parameter specifies is the debug enabled for the telegram bot. Default value is `false`.
* `json_config_file` - the parameter specifies location for the configuration file in the `json` format. **Required**

Also you can override the configuration with deployment variables:
* `TG_BOT_NAME` - env variable for `bot_name`.
* `TG_BOT_TOKEN` - env variable for `bot_token`.
* `TG_BOT_DEBUG` - env variable for `is_debug`.
* `DOCKER_ORCHESTRATOR_CONFIG_LOCATION` - env variable for `json_config_file`.

You can use the follow `bash` script to run the service:
```shell
#!/bin/bash

export TG_BOT_NAME=<bot_name>
export TG_BOT_TOKEN=<bot_token>
export DOCKER_ORCHESTRATOR_CONFIG_LOCATION=deployment_configs.json

nohup bash ./server-orchestrator > service-log.txt &
```

## Build
