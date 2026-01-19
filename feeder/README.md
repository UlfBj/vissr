**(C) 2023 Ford Motor Company**<br>

# VISSR feeders

A feeder is a SwC that are deployed on the south side of the state storage. It has hree main tasks:
* Realizing the interface towards the state storage and the server.
* Converting data between the two domains - the vehicle native data format and the VSS data format.
* Realizing the interface towards the underlying vehicle system.

Several feeders have been developed over the years, such as the Remotive Labs feeder found in the feeder-rl directory,
or the External Vehicle Interface Client feeder found in the feeder-evic directory.
There are also a few versions of a feeder template where the interface towards the underlying system is not implemented,
instead it has a rudimentary vehicle data simulator.

The interactions between the feeder and the state storage and server has evolved over time,
this has been done without ensuring backwrds compatibility with previous version.
So to use any feeder but the latest version of the feeder template it i necessary to either back up to an older server version that is compatible with the feeder
or to update the feeder to the latest server interactions.

The main feature changes between different versions are:
* Version 1: All interactions between server and feeder go via the state storage.
* Version 2: Data from Set requests are sent via UDS from server to feeder.
* Version 3: The feeder issues "read triggers" to the server for certain subscription types after it has written the data to the state storage.
* Version 4: A model supporting multiple feeders is introduced where feeders register with the server when starting up.
Dynamic deregistration is also supported.

Feeders are built and run as separate executables with the commands below issued in the respective directory.<br>
$ go build<br>
$ ./name-of-executable

