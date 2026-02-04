**(C) 2026 Ford Motor Company**<br>

# VISSR service feeder template version 1

The servicefeeder template version 1 exposes the same interface towards the server / state storage as the datafeeder template version 4,
i. e. it register/deregister with the server thereby supporting scenarios with multiple feeders of both data and service type.
For more information about the registration interface see the [data feder]) README.
The only difference is that this feeder shall use the infotype "Service" in the registration request:
```
Reg-request:{"action": "reg", "name": "xxxx", "infotype": "Service"}
```
The servicefeeder Sw design shown below has a similar design as the datafeeder except for that the map and scale functionality is removed
as this responsibility is deferred to the service end points.
![Servicefeeder Sw design](serviccefeeder-sw-design.jpg?raw=true)<br>
*Fig 1. Servicefeeder software design

Not shown in the figure is the service end point registration interface that is used by the service end points to register/de-register with the servicefeeder.
This interface has a similar pattern to the registration interface where feeders rgister/de-register with the server, see below.
A service end point registers by issuing the following request to the servicefeeder:
```
Reg-request:{"action": "reg", "servicename": "a.b.c"}
```
where a.b.c is the path in the service tree. This path can point to a separate service node or to a service group branch node.
In the latter case the service end point then is responsible for the execution of all the services of the service group.\
If successful the server responds with
```
Reg-response:{"action": "reg"", "servicename": "a.b.c", "sockfile": "/x/y/z.sock"}
```
where the sockfile shall be used by the service end point to set up a UDS server listening to requests from the servicefeeder.

A service end point can also at any point de-register with the servicefeeder by issuing the request:
```
Dereg-request:{"action": "dereg", "servicename": "a.b.c"}
```
which if successful get the following response:
```
Dereg-response:{"action": "dereg", "servicename": "a.b.c"}
```
Non-successful requests gets the following error response from the server:
```
Error-response:{"action": "error"}
```
The communication over the UDS channel between the service feeder and a service end point has the following structure.


The server does not hold information on which Invoke or Cancel requests that shall be sent to which feeder that is responsible for forwarding it to the underlying vehicle system.
It therefore sends the request to all registered service feeders, and then it is the responsibility of the feeders to evaluate which requests it shall act on, or just drop.

The current logic used by the servicefeederv1 supports any tree/branch to be defined as the feeder scope,
but it does not support the exclusion of parts of the tree/branch.
This will be added at some later point.
