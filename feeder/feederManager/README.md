**(C) 2026 Ford Motor Company**<br>

# VISSR feeder manager

The feeder manager controls which datafeeders and servicefeeders that shall be started at system startup.
The feeder manager can be started up to only read the file containing the list of feeders of which one or more shall be started by it,
before it terminates.\
It can also be started up to display a UI which enables the addition or removal of feeders of the dynamic type.
Feeders of the static type cannot be modified by the UI.
A feeder is on the list represented by the following data:
* Name: A short but preferrably descriptive name.
* Searchpath to the feeder binary file.
* CLI parameters: ny CLI parameters that the feeder shall be provided with at its startup.
* PID: The PID that the runtime assigned to the feeder at startup. Should be zero if the feeder is not executing.
* Feeder type: Denotes whether the feeder is of static or dynamic type.
* Feeder status: Denotes whether the feeder has active or passive status.
If active status then the feeder shall be automatically started at system startup.
If status is set to passive then the feeder will not be started by the feeder manager.
The feeder status can be changed via the feeder manager UI for feeders having a dynamic feeder type.

If the feeder manager is started up having the CLI parameter set that leads to that the UI is displayed,
then the UI will offer the following services:
* Add a feeder of dynamic type to the list of feeders.
* Remove a feeder of dynamic type from the list of feeders.
* Change the feeder status of a dynamic feeder on the list of feeders.
* Startup a dynamic feeder that is not already started.
* Termination of a dynamic feeder that has been started.
* Termination of the feeder manager.
