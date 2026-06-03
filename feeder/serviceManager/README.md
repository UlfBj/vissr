**(C) 2026 Ford Motor Company**<br>

# VISSR service manager

The service manager controls which services that shall be started at system startup.
The service manager can be started up to only read the file containing the list of services of which one or more shall be started by it,
before it terminates.\
It can also be started up to display a UI which enables the addition or removal of services of the dynamic type.
Services of the static type cannot be modified by the UI.
A service is on the list represented by the following data:
* Name: This must be the path to the service as defined by the service tree,
or the path to a service group.
* Searchpath to the service binary file.
* CLI parameters: ny CLI parameters that the service shall be provided with at its startup.
* PID: The PID that the runtime assigned to the service at startup. Should be zero if the service is not executing.
* Service type: Denotes whether the service is of static or dynamic type.
* Service status: Denotes whether the service has active or passive status.
If active status then the service shall be automatically started at system startup.
If status is set to passive then the service will not be started by the service manager.
The service status can be changed via the service manager UI for service having a dynamic service type.

If the service manager is started up having the CLI parameter set that leads to that the UI is displayed,
then the UI will offer the following services:
* Add a service of dynamic type to the list of services.
* Remove a service of dynamic type from the list of services.
* Change the service status of a dynamic service on the list of services.
* Startup a dynamic service that is not already started.
* Termination of a dynamic service that has been started.
* Termination of the service manager.
