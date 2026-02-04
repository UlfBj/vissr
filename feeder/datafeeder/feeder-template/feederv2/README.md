**(C) 2023 Ford Motor Company**<br>

# VISSR feeder templte version 2

This version introduced a feature that the server issues Set request data to the feeder over a US channel.
This eliminates the need for the feeder to continuosly poll the state storage for Set data written to it,
thus removing a significant latency issue.

The interface towards the VISSv2 server consists of two parts, one for reading data and one for writing data. When data is to be written to the server, the feeder writes to the statestorage from which the server then can read it. Data coming from the server is read over a Unix domain socket connection, where the feeder acts as a server.
The VISSv2 server, which on this Unix connection acts as the client, reads a "feeder registration" file at startup that provides the socket address and
the root node name of the tree that the feeder manages (VSS currently only defines one tree, but that may change).

This feeder version does not scale the data value and the information needed for that is not present in the map and scale instructions read at startup.
A solution for this is in the planning.

Version 2 supports both signal name mapping and scaling,
and it relies on the <a href="https://github.com/COVESA/vissr/tree/master/tools/DomainConversionTool">Domain Conversion Tool</a>
(DCT) to create the mapping and scaling instructions that it uses.

The feeder expects the statestorage to be implemented using a Redis database, the details of this can be found at
<a href="https://github.com/COVESA/ccs-components/tree/master/statestorage">COVESA CCS components</a>.
Memcached DB is also supported.

## Version 2 mapping and scaling
This version requires that the Domain Conversion Tool is used to create the feeder conversion instructions that the feeder reads at startup.
The two fiels that it reads are:
* VssVehicle.cvt
* VssVehicleScaling.json

The first file contains the primary conversion instructions, aving the format of an array of structs, each with the following format:<br>
```
struct {
	MapIndex uint16
	Name string
	Type int8
	Datatype int8
	ConvertIndex uint16
}
```
Each struct element contains data associated to one signal of any of the two domains.
The MapIndex links it logically to another element in the array, which is the corresponding signal of the other domain tat it is mapped to.
This array can then by the feeder be used to search for signal names (Name element in the struct) of any of the two domains, 
and as the array is sorted on the struct Name element, a binary search algorithm can be used.
The scaling operation is controlled by the ConvertIndex of the struct.
An index of zero is interpreted by the feeder as no scaling needed (one-to-one),
any other number, except 65535 that indicates a DCT mapping error, is set to point to an element of the scaling array that the feeder read from the file VssVehicleScaling.json at startup.
A string element in this array is after being addressed by the ConvertIndex interpreted as either a JSON object containing a list of key-value pairs, or a JSON number array.<br>
In the case of a JSON object, each key-value pair represents the associated scaling values of the "allowed" values from repsective domain.<br>
In the case of a JSON number array, the two elements of the array represents the A and B coefficients of the equation y = A*x + B (or y = (x-B)/A in the other direction).
The struct Datatype can be used to reformat if needed after a linear conversion that is always calculated using float64.
