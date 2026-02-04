**(C) 2025 Ford Motor Company**<br>

# VISSR feeder template version 3

This version extends the UDS interface between server and feeder to enable feeders to issue "read triggers" to the server.
The main reason for this is to eliminate the need for the server to poll the state storage for asynchronous subscription data written be the feeder.

The server must issue a request to the feeder containing the paths for which it wants to receive a "read trigger" event.
This request shall have the following format:
```
{”action”: ”subscribe”, ”path”: [”p1”, ..., ”pN”]}
```
In case the feeder then later writes data to this path in the state storage it shall issue the following "read trigger" request to the server:
```
{”action”: ”subscription”, ”path”: ”p”} // notification from feeder
```
The server can terminate this action by the feeder by issuing an unsubscribe request:
```
{”action”: ”unsubscribe”, ”path”: [”p1”, ..., ”pN”]}
```
The server can also issue the Set requests inherited from feeder-template v3.
```
{”action”: ”set”, "data": {"path":"x", "dp":{"value":"y", "ts":"z"}}}
```
This interface also support default value update requests:
```
{"action": "update", "default": [{"path": "x", "value": "y"},..., {}]}
```
The feeder will then update the state storage with this default value.
This is particularly meant to be used for attribute values that is typically not written to again by the feeder.
The server can at startup be provided with the CLI parameter "-d" in which case it reads the default values found in the vehicle tree and then issues an update request.

Responses to all these requests are either a status OK response or a Status NOK response as shown below.
The action value xxx shall be set to the same ation values as in the request.
```
{”action”: ”xxx”, ”status”: “ok/nok”}  // response from feeder (optional for nok case)
```

As mentioned the server issues a request to inform the feeder for which paths it shall issue a read trigger.
The server shall not use this to send all signal paths of the tree(s) it manages, which could lead to performance issues.
Instead it is expected to select a smaller set of signals that is sparsely updated, but for which a minimized latency is required.
An example of such a signal could be "Vehicle.Cabin.Door.Row1.DriverSide.IsLocked" that typically changes status infrequently,
but for which a minimized latency is desireable.
