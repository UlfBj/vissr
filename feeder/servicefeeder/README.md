**(C) 2026 Ford Motor Company**<br>

# VISSR service feeder

## Service feeder registration with server
The servicefeeder exposes the same interfaces towards the server / state storage as the datafeeder template version 4,
i. e. it registers/deregisters with the server thereby supporting scenarios with multiple feeders of both data and service type.
For more information about the registration interface see the [data feeder v4]() README.
The only difference is that this feeder shall use the infotype "Service" in the registration request:
```
Reg-request:{"action": "reg", "name": "xxxx", "infotype": "Service"}
```
The servicefeeder Sw design shown below has a similar design as the datafeeder except for that the map and scale functionality is removed
as this responsibility is deferred to the service end points.
![Servicefeeder Sw design](serviccefeeder-sw-design.jpg?raw=true)<br>
*Fig 1. Servicefeeder software design

## Service registration with the service feeder
Not shown in the figure is the service end point registration interface that is used by the service end points to register/de-register with the servicefeeder.
This interface has a similar pattern to the registration interface where feeders register/de-register with the server, see below.
A service end point registers by issuing the following request to the servicefeeder:
```
Reg-request:{"action": "reg", "servicepath": "["path1", .., "pathN"]"}
```
where servicepath is an array of the paths in the service tree that the service manages.
The service end point is responsible for the execution of all the services provided in the array.\
If successful the server responds with
```
Reg-response:{"action": "reg"", "servicepath": "["path1", .., "pathN"]", "sockfile": "/x/y/z.sock"}
```
where the sockfile shall be used by the service end point to set up a UDS server listening to requests from the servicefeeder.

A service end point can also at any point de-register with the servicefeeder by issuing the request:
```
Dereg-request:{"action": "dereg", "servicepath": "["path1", .., "pathN"]"}
```
which if successful get the following response:
```
Dereg-response:{"action": "dereg", "servicepath": "["path1", .., "pathN"]"}
```
Non-successful requests gets the following error response from the server:
```
Error-response:{"action": "error"}
```
The service must not de-register all the services it previously registered.

The server does not hold information on which Invoke or Cancel requests that shall be sent to which service feeder that is responsible for forwarding it to the underlying vehicle system.
It therefore sends the request to all registered service feeders, and then it is the responsibility of the feeders to evaluate which requests it shall act on, or just drop.
