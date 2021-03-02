This service provides the ability to manage the docker configuration on your server using a telegram bot.

## Launch

Because the service is used to manage your docker container, it should be run on the host machine.

To start the service please provide the following environment variables:
* `TG_BOT_NAME` - the parameter specifies telegram bot name.
* `TG_BOT_TOKEN` - the parameter specifies telegram token. **Required**
* `TG_BOT_DEBUG` - the parameter specifies is the debug enabled for the telegram bot. Default value is `false`.
* `DOCKER_ORCHESTRATOR_CONFIG_LOCATION` - the parameter specifies location for the configuration file in the `json` format. Default value is `data/deployment_configs.json`.

You can use the follow `bash` script to run the service:
```shell
#!/bin/bash

export TG_BOT_NAME=<bot_name>
export TG_BOT_TOKEN=<bot_token>
export DOCKER_ORCHESTRATOR_CONFIG_LOCATION=deployment_configs.json

nohup bash ./server-orchestrator > service-log.txt &
```

## Build
