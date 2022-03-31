# splunk-mqtt
SPLUNK-MQTT is a simple go binary used to connect to a MQTT broker and stream Messages directly into Splunk HEC

## Features

- [x] subscribe multiple topics at mqtt broker
- [x] send events to Splunk HEC
- [x] output received messages to console
- [ ] use tls to connect to mqtt broker
- [ ] caching

## config
splunk-mqtt tries to open config.yaml in the current working directory.
environment variables take precedence over config.yaml variables
|  config.yaml | ENVIRONMENT  |  Default |  Required | Description |
|---|---|---|---|---|
| broker  | BROKER  |  "" | yes  | MQTT Broker URL. tcp://192.168.1.1:1883  |
| mqtt_username  | MQTT_USERNAME  |  "" | no  | MQTT Username |
| mqtt_password  | MQTTT_PASSWORD  |  "" | no  | MQTT Passowrd  |
| hec_url  | HEC_URL  |  "" | yes | Splunk HEC URL. https://hec.splunk.com   |
| hec_token  | HEC_TOKEN  | ""  |  yes | Splunk HEC token  |
| client_id  | CLIENT_ID  | ""  |  no | MQTT Client ID  |
| write_to_console  |  WRITE_TO_CONSOLE |  false |  no | Write received MQTT messages to console  |
| write_to_splunk  |  WRITE_TO_SPLUNK | false | no  | Write received MQTT messaged to splunk  |
| topics  | TOPICS  |  "" |  yes |  List of MQTT Topics to subscribe |
| insecure_skip_verify  | INSECURE_SKIP_VERIFY  | false  | no  | Skip TLS Verification  |

### yaml example
```
---
broker: tcp://192.168.1.1:1883
mqtt_username: splunk
mqtt_password: abcde12345
hec_url : https://hec.splunk.com
hec_token: xxxxxxx-xxxx-xxxx-xxxx-073155c4c54e
client_id: mqtt_subscribe
write_to_console: true
write_to_splunk: true
topics:
  - tele/+/SENSOR
  - tele/some/SENSOR
insecure_skip_verify: false

```
### environment example
```
export BROKER=tcp://192.168.1.1:1883
export HEC_URL=https://hec.splunk.com
export HEC_TOKEN=xxxxxxx-xxxx-xxxx-xxxx-073155c4c54e
export MQTT_USERNAME=splunk
export MQTT_PASSWORD=abcde12345
export CLIENT_ID=mqtt_subscribe
export WRITE_TO_CONSOLE=true
export WRITE_TO_SPLUNK=true
export TOPICS=tele/+/SENSOR,tele/some/SENSOR
export INSECURE_SKIP_VERIFY=false
```
## run
```
$ ./splunk-mqtt
MQTT Broker:  tcp://192.168.1.1:1883
MQTT Username:  
Splunk HEC URL:  https://hec.splunk.com
Connection is up
connection established
subscribed to:  tele/+/SENSOR
subscribed to:  tele/some/SENSOR
```
## run docker
```
docker run \
-e BROKER=tcp://192.168.1.1:1883 \
-e HEC_URL=https://hec.splunk.com \
-e HEC_TOKEN=xxxxxxx-xxxx-xxxx-xxxx-073155c4c54e \
-e TOPICS=tele/+/SENSOR,tele/some/SENSOR \
-e WRITE_TO_CONSOLE=true \
-e WRITE_TO_SPLUNK=true \
asauerwein/splunk-mqtt:0.1.0-alpha
```

## Licenses of dependencies
- github.com/eclipse/paho.mqtt.golang [Eclipse Public License - v 1.0](https://github.com/eclipse/paho.mqtt.golang/blob/master/LICENSE)
- github.com/jhop310/splunk-hec-go [Apache License 2.0](https://github.com/jhop310/splunk-hec-go/blob/master/LICENSE)
- github.com/kelseyhightower/envconfig [MIT License](https://github.com/kelseyhightower/envconfig/blob/master/LICENSE)
