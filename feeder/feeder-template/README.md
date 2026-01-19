# Feeder templates
The feeder template is designed to work together with the VISSv2 server and a state storage as shown in the figure below.
![Software architecture](feeder-sw-arch.jpg?raw=true)<br>
*Fig 1. Software architecture overview

The feeder has two interface clients, each running on a thread, one exercises the interface towards the server, and the other the underlying vehicle interface.
When either client reads data from he interface it forwards it to the Map & Scale component.
This component uses mapping instructions that it reads from a config file at startup to swwitch the name of the data to the name of the other domain, and scales the value if necessary.
When that is done it forwards the mapped data to the other interface client that then uses the interface to send it into the domain.

The feeder template has gone through several version updates, see details for its featue support in the respective directories.

The vehicle interface client exercises the vehicle interface. This interface can be an interface towards a CAN bus, a Flexray bus, etc., and the details of it is OEM proprietary.
Therefore this feeder implements a simulation of a data exchange over the interface, code that will have to be replaced by the OEM before deployment.

The feeder templates are meant to be used as a starting point for development of feeders for different southbound/vehicle interfaces.
The interface client on the southbound side is all that needs to be modified to support the new interface,
all other functionality in the feeder is already implemented.

It is the hope that feeders supporting new vehicle interfaces are upstreamed to the repo.
