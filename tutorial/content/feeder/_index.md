---
title: "VISSR Feeders"
---

A feeder is a Sw component that needs to implement three tasks:
* Implement an interface to the data storage
* Implement an interface to the underlying vehicle interface
* Translate data from the format used in the "VSS domain" to te format used in the "Vehicle domain".

The SW architecture shown in figure 1 reflects the division of the three tasks in that the translation (map & scale) is done in the min process,
which spawns two threads that implement the respective interface task.
The architecture shown handle all its communication with the server via the state storage.
This leads to a polling paradigm and thus a potential latency and performance weakness.
This architecture is therefore not found on the master branch, but available on the datastore-poll branch.
There is still a feeder design for this in the feeder-template/feederv1 directory.
![Feeder Sw architecture, unoptimized polling](/vissr/images/feeder-sw-design-v1.jpg?width=50pc)
* Figure 1. Feeder software architecture unoptimized polling

An improved architecture that eliminates the mentioned weaknesses for data flowing in the direction from the server to the feeder (i. e client write requests)
is shown in figure 2. For write requests the server communicates directly over an IPC channel with the feeder, thus removing the need for the feeder to poll
the state storage to find new write requests.
![Feeder Sw architecture, optimized polling](/vissr/images/feeder-sw-design-v2.jpg?width=50pc)
* Figure 2. Feeder software architecture optimized polling

A feeder implementing this solution can be found in the feeder-template/feederv2 directory.
This feeder can be configured to either use an SQLite, or a Redis state storage interface, please see the Datastore chapter for details.

However, the solution implementd in the feederv2 template does not support that the server also can replace the polling with a more effective event based solution.
For this the feeder implementation in the feeder-template/feederv3 directory needs to be used.

The server is able via the interface to detect whether a feeder implements version 2 or 3 of the interface.
In case of version 2 it keeps poling of the data store, while for version 3 it relies on event signalling from the feeder instead.
For more details of this interface, see the [VISSR Server:Feeder interface](/vissr/server/#feeder-interface) chapter.

The feeder translation task is divided into a mapping of the signal name, and a possible scaling of the value.
The instructions for how to do this is encoded into one or more configuration files that the feeder reads at startup.
There are two versions of the feeder instructions, being used in template feederv1 and feederv2, respectively.

In version 1 this file only contains a signal name mapping, while version 2 supports also scaling instructions.
For more about the telpate feeders, see the README in [feeder-template](https://github.com/covesa/vissr/tree/master/feeder/feeder-template) directory.

An OEM wanting to deploy the VISSv2 tech stack needs to implement the Vehicle interface of the feeder, e. g. to implement a CAN bus interface.
In the template feeders the Vehicle interface contains a minimal signal simulator that generates random values for the signals it is configured to support.
All  of that code should be removed and replaced by the code implementing the new interface client.

Besides the feeder templates there is also an [rl-feeder](https://github.com/covesa/vissr/tree/master/feeder/feeder-rl)
where the Vehicle interface is implemented to connect to a RemotiveLabs broker.
[RemotiveLabs](https://remotivelabs.com/) has a public cloud version of its broker that can be used to replay trip data available in VSS format.

There is also an External Vehicle Interface Client [EVIC](https://github.com/covesa/vissr/tree/master/feeder/feeder-evic)
feeder that enables the interface client to be implemented in a separate executable.

## Simulated vehicle data sources
The feederv2 template contains two different simulation mechanisms that are selected via the command line configuration parameters.

The one configured by "internal" uses the conversion instructions data as input for which signals to simulate, which are then randomly selected and set to random values.

The other configured by "vssjson" tries to read the file "tripdata.json", which must have a format as found in the example file. s seen in that file it contains an array of signal path names, and for each signal it contains an array of datapoints, i.e. timestamps and values. The data points are replayed at a constant frequency of 1 Hz. To change the frequency the time.Sleep input in the code must be changed and recompiled. This sets the rquirement on the data point arrays that their values must have been captured at this frequency, or recalculated to this frequency. Each data point array must have the same length. The simulator waps around and starts again from the beginning after reaching the end.

The signal names must be VSS paths as they are not processed by the conversion engine.
Extending the model to instead expect "vehicle domain signals" (like CAN signal data) should be a simple coding exercise for anyone preferring that.
