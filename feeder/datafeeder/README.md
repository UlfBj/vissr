**(C) 2023 Ford Motor Company**<br>

# VISSR datafeeders

A datafeeder is a SwC that are deployed on the south side of the state storage. It has the following main tasks:
* Realizing the interface towards the state storage and the server.
* Converting data between the two domains - the vehicle native data format and the VSS data format.
* Realizing the interface towards the underlying vehicle system.

Several datafeeders have been developed over the years, such as the Remotive Labs feeder found in the feeder-rl directory,
or the External Vehicle Interface Client feeder found in the feeder-evic directory.
There are also a few versions of a feeder template where the interface towards the underlying system is not implemented,
instead it has a rudimentary vehicle data simulator.

The interactions between the datafeeder and the state storage and server has evolved over time,
this has been done without ensuring backwards compatibility with previous version.
So to use any feeder but the latest version of the feeder template it i necessary to either back up to an older server version that is compatible with the feeder
or to update the feeder to the latest server interactions.

The main feature changes between different versions are:
* Version 1: All interactions between server and feeder go via the state storage.
* Version 2: Data from Set requests are sent via UDS from server to feeder.
* Version 3: The feeder issues "read triggers" to the server for certain subscription types after it has written the data to the state storage.
* Version 4: A model supporting multiple feeders is introduced where feeders register with the server when starting up.
Dynamic deregistration is also supported.

Datafeeders are built and run as separate executables with the commands below issued in the respective directory.<br>
$ go build<br>
$ ./name-of-executable

## Trigger channel protocol between server and data feeder
The main reason for the trigger channel is to eliminate the latency that the polling of the statestorage otherwise would lead to.
This was first adressed in the template version 2 where data to be written to the underlying vehicle system was directly sent to the feeder over the trigger channel.
In template version three this was extended so that the feeder can be instructed by the server that for certain signals it wants to get a trigger when the feeder has written the signal to the statestorage.
The protocol to realize this is shown below.
The server issues the following request to the feeder for data that shall be written to a vehicle signal.
```
{”action”: ”set”, "data": {"path":"x", "dp":{"value":"y", "ts":"z"}}}
```
The server may decide for certain signals that it has received a client subscribe request on to request the feeder to
send it a trigger event after writing data for that signal in the statestorage.
This request shall have the following format:
```
{”action”: ”subscribe”, ”path”: [”p1”, ..., ”pN”]}
```
and the events that the feeder then may issue to the server has the following format:
```
{”action”: ”subscription”, ”path”: ”p”}
```
When the server receives a client unsubscribe request,
if it has issued a previous subscribe request to the feeder the it shall issue an unsubscribe request to the feeder to stop further trigger messages.
The unsubsribe request shall have the followign format:
```
{”action”: ”unsubscribe”, ”path”: [”p1”, ..., ”pN”]} // message from fcommmgr to feeder
```
All request messages, but not the event messages shall get a response message with the following format:
```
{”action”: ”subscribe”, ”status”: “ok/nok”}
```
Status shall be "ok" for successful processing of the request and "nok" for unsuccessful processing.
