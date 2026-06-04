---
title: "COVESA Vehicle Information Service Specification version 3.2 Reference Implementation Tutorial"
---
## COVESA Vehicle Information Service Specification ver 3.2 Reference Implementation Tutorial

The COVESA VISSv3.2 specification development is not yet started at the [COVESA VISS Github](https://github.com/COVESA/vehicle-information-service-specification).
The set of new specification features are unknown but the assumption done here is that it will contain support for the HIM Service profile.
Support for the HIM Data profile was added in [VISSRv3.1](https://github.com/COVESA/vissr/tree/v1.0).

The final VISSR development related to the VISSv3.0 specification is available on the [v0.9 tag](https://github.com/COVESA/vissr/tree/v0.9),
thereafter (potential) VISSv3.1 related features are introduced om the master branch.\
Before starting to merge VISSv3.2 features to the master branch a tag will be created that marks the end of the VISSv3.1 development.
The VISSv3.2 related development will until then be done at the v3.2 branch.

This documentation covers the VISSRv3.2 implementation, which is backwards compatible with the VISSRv3.1 implementation.
The new features are listed below.

### VISSRv3.2 new features
* Supports the [COVESA HIM](https://github.com/COVESA/hierarchical_information_model) Service profile.
* Introduces the "service feeder" type, and assigns the type "data feeder" to the legacy feeders.
* Introduces the FeederManager and the ServiceManager SwCs.
These two SwCs manages the lifetime of feeders and services respectively.
* Introduces the Service SwC(s).
These SwCs are the "end point" for service requests that control the logic that executes a service.
* The server supports both the HIM information types Data and Service.
* The server supports the new "Service specific" protocol messages as defined in the VISSv3.2 specification (yet to be written, a so speculative format is used).

### VISSRv3.2 tech stack
The figure below shows the VISSRv3.2 tech stack architecture.
The architecture inherits the general structure from previous generation,
the new parts are found in the "feeder framework" south of the state storage.

![VISSRv3.2 tech stack](/vissr/images/VISSRv3.2-tech-stack.jpg?width=40pc)

Also found on this repo are implementations of other components that are needed to realize a communication tech stack that reaches from clients through the server and to the underlying vehicle system interface.

These software components (SwCs) can be categorized as follows:
* server
* clients
* data storage
* feeders
* tools

The tutorial describes each SwC category in a separate chapter.
It also contains a few Proof of concept (POC) examples, and information about installing,
building and running Golang based SwCs, a Docker containerization, and about some peripheral components.
