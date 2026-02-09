---
title: "VISSR Feeders"
---

With the support of both the HIM Data profile and the HIM Service profile this is reflected in that there are now two types of feeders:
* Data feeders: This feeder type implements the flow of vehicle signal data between the HIM Data domain and the underlying vehicle data domain.
This is the legacy feeder used in earlier genrations of VISSR.
* Service feeders: This feeder type implements the flow of vehicle service data between the HIM Service domain and the service implementation.
The service implementation contains the logic for the service execution and it interacts with the underlying vehicle data domain.
This interaction can be directly with the the underlying vehicle data domain or it can be by interaction with the VISSR client interface,
then leaving the direct interaction to the VISSR data feeders.
The latter alternative leads to that the service becomes OEM agnostic and can be deployed at all vehicles that deploys the VISSR tech stack.

Feeders are configured with one of two deployment types:
* Static deployment type: This type is automatically started/terminated at each end of an ignition cycle.
* Dynamic deployment type: This type is automatically started/terminated at each end of an ignition cycle,
but it can also be started and terminated by the Feeder Manager at any point during an ignition cycle.
The Feeder Manager can also change the configuration that controls whether it shall be automatically started/terminated or not.


### Data feeders
For more info see the [Data feeders](/vissr/feeder/datafeeder/) section.

### Service feeders
For more info see the [Service feeders](/vissr/feeder/servicefeeder/) section.

## Feeder manager
The Feeder manager is the tool that controls the life time of feeders - starting of feeders and termination of feeders.\
The tool can be used both "programmatically" by other SwCs and also "manually" by a user via a UI.
When started without any CLI parameter the tool displays the UI, but in all cases where a CLI parameter is used the tool acts headless (display no UI).
Except for the case when started with the CLI parameter -h, then the following is displayed.
```
$ ./feederManager -h
usage: print [-h|--help] [-s|--start] [-b|--build] [-t|--terminate]
             [-f|--feederfile "<value>"]

             Feeder manager

Arguments:

  -h  --help        Print help information
  -s  --start       start all feeders
  -b  --build       build active feeders
  -t  --terminate   terminate running feeders
  -f  --feederfile  Feeder-list filename. Default: feederList.json
```
As can be seen above the following actions can be requested:
* -s: Start all feeders.
* -t: Terminate all running feeders.
* -b: Build all active feeders.
* -f: Select which file containing a list of feeders that the feeder manager can control.

The -f parameter can be combined with any of the other CLI parameters, the others can only be combined with the -f CLI parameter.
If more than one of the -s/-t/-b parameters are to be submitted to the tool then it has to be done by multiple calls to the tool.
The tool terminates directly after executing the requested service.

## Feeder list
The file containing the list of feeders that the feeder manager has the ability to control is a JSON formatted file containing the following data items:
* Name of feeder. This should be a short but descriptive name. It has no functional purpose but may be helpful if logs have to be read.
* Searchpath to the feeder binary file (executable). The searchpath should be relative to the directory of the feeder manager.
* A string with eventual CLI parameters that should be added to the command line starting the feeder.
* Feeder deployment type. This can be either "static" or "dynamic". The latter type can be started/terminated at any point in time via the feeder manaer UI,
which is not the case for the former type. Both types are automatically started/terminated at each end of the vehicle ignition cycle.
* Feeder activation status. This can be either "active" or "passive". 
The former leads to that the feeder manager starts/terminates the feeder at ignition start/end.
The latter leads to that the feeder manager does not start/terminate the feeder at ignition start/end.

## Feeder manager UI
When the feeder is started without any CLI parameter it displays the following table:
```
Select one of the following numbers:
1: Display feederlist summary
2: Add feeder to list
3: Remove feeder from list
4: Start feeder on list
5: Terminate running feeder on list
6: Change feeder status
0: Exit feederManager

Option number selected:
```
Entering one of the numbers 0 through 6 leads to:
* 0: The feeder manager terminates.
* 1: A summary of the feeders on the list is displayed.
* 2: A dialogue is run that asks the user to input for the feeder parameters described above.
A new feeder on the feeder list is then added to the list.
* 3: The user is asked about which feeder on the list that shall be removed, which then leads to its removal.
* 4: The user is asked about which feeder on the list that shall be started, which then leads to the feeder being started.
Only not already running feeders can be started.
* 5: The user is asked about which feeder on the list that shall be terminated, which then leads to the feeder being terminated.
Only running feeders can be terminated.
* 6: The user is asked about which feeder on the list that shall have its activation status changed.
The status is the switched to the other of the two possible states "active" or "passive".

The feeder deployment type cannot be changed by the tool.
The reason for this is to prevent the passivation of feeders that are critical to the vehicle functionlity and safety.
A "manual" modification of the content in the feeder list file is possible if there is a real need to change it.
