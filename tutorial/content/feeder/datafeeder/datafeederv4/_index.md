---
title: "VISSR Data Feederv4"
---

The supported set of feederv4 CLI commands is the same as for feederv3 except for the addition of the -c CLI parameter.
The following part of this section therefore only describes the difference to feederv3, please see the feederv3 section for other features supported by feederv4.

If the data feederv4 is started with the CLI parameter -h the following is presented:
```
usage: print [-h|--help] [-c|--configfile "<value>"] [-m|--mapfile "<value>"]
             [-s|--scldatafile "<value>"] [-t|--tripdatafile "<value>"]
             [--logfile] [--loglevel (trace|debug|info|warn|error|fatal|panic)]
             [-i|--simsource (vssjson|internal)] [-d|--statestorage
             (sqlite|redis|memcache|none)] [-f|--dbfile "<value>"]

             Data feeder template version 4

Arguments:

  -h  --help          Print help information
  -c  --configfile    Feeder configuration filename. Default: feederConfig.json
  -m  --mapfile       VSS-Vehicle mapping data filename. Default:
                      VssVehicle.cvt
  -s  --scldatafile   VSS-Vehicle scaling data filename. Default:
                      VssVehicleScaling.json
  -t  --tripdatafile  Filename for simulated trip data. Default: tripdata.json
      --logfile       outputs to logfile in ./logs folder
      --loglevel      changes log output level. Default: info
  -i  --simsource     Simulator source must be either vssjson, or internal.
                      Default: internal
  -d  --statestorage  Statestorage must be either sqlite, redis, memcache, or
                      none. Default: redis
  -f  --dbfile        statestorage database filename. Default:
                      ../../server/vissv2server/serviceMgr/statestorage.db
```
* The -c command is used together with a searchpath to the the feeder configuration file to be read by the feeder at startup.
If the -c command is not applied then the feeder will by default read the file feederConfig.json.

For the other command parameter please see [feederv3](/vissr/feeder/datafeeder/datafeederv3).

## Feeder configuration file
The configuration file is JSON formatted file that contains the following feeder configuration parameters:
* Name: The feeder name shall be a short, descriptive name. It must be different from the name of all other feeders that registers with the server.
* InfoType: The information type must have the value "Data".
It is has no functional meaning currently but it will have if in later version other informartion types also will be supported.
* Scope: The scope parameter is an array of paths that define the scope of the set of vehicle signals that the feeder manages.
The path shall be the name of the root node of a tree if the feeder manages the entire tree.
It can be a path to a branch of a tree, or even to a single node in a tree.

The feeder shall support all the signals that its Scope configuration defines.
The paths o all other signals shall be ignored and silently dropped.

## Feeder registration
The feederv4 implements a registration protocol that is also implemented by the server.
This enables multiple feeders to register with the server.
For more information please see the [feederv4 README](https://github.com/COVESA/vissr/tree/master/feeder/datafeeder/feeder-template/feederv4).
