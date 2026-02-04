**(C) 2023 Ford Motor Company**<br>

# VISSR feeder template version 1

This version used the state storage for all data transfer between the server and the feeder.

This provided a loose coupling between the two SwCs, but also that all data retrieval had to be done using polling,
leading to significant latencies.

Version 1 is limited to support only the signal name mapping, and no scaling.
This version was used used mainly for initial testing.

## Version 1 mapping
The VehicleVssMapData.json contains an example of the signal name mapping. 
The format must be an array of JSON objects, where each object contains two key-value pairs holding the signal names that are mapped.
This format
```
[{"vssdata":"vssname1","vehicledata":"vehiclename1"}, ..., {"vssdata":"vssnameN","vehicledata":"vehiclenameN"}]
```
must be used to represent the mapping.
Manual editing of this file is needed to update the mapping.
The feeder reads this file a startup.



