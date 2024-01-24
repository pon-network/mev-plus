# MEV Plus

## Table of Contents

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Build](#building-mev-plus)
- [Workflow](#mev-plus-workflow-overview)
- [Usage](#usage)
- [Adding Custom Modules to MEV Plus](#adding-custom-modules-to-mev-plus)
- [Pre-packaged Modules](#pre-packaged-modules)
- [Key Features enabled through Pre-packaged Modules](#key-features-enabled-through-pre-packaged-modules)
- [Contributing](#contributing)
- [License](#license)

## Introduction

MEV Plus is a side car ecosystem where consensus layer validators can communicate with outside applications such as PBS, execution layer etc. It can be thought of as the next evolution of validator proxy software introducing exciting on-chain and off-chain possibilities. This open-source solution offers modularity and exceptional scalability to accommodate a wide range of applications where each application is known as a module or plug-in.

MEV-plus's first flagship application enables validators to leverage a marketplace of block builders and relayers. These builders are responsible for crafting blocks that include transaction orderflow, along with a fee for the validator who proposes the block. The separation of roles between proposers and builders fosters healthy competition, decentralization, and enhances censorship resistance within the Ethereum ecosystem. It's a significant step towards a more robust and efficient network.

## Prerequisites

Requires [Go 1.20+](https://go.dev/doc/install).

## Installation

Install the latest MEV-plus release with go install:

```shell
go install github.com/pon-network/mev-plus@latest
go run mevplus.go --help
```

## Building MEV-Plus

Before proceeding, make sure you have the latest release of MEV-Plus.

```shell
# Clone the MEV-Plus repository, which includes ongoing merged PRs for future releases.
git clone https://github.com/pon-network/mev-plus
cd mev-plus

# Build the most recent version of MEV-Plus into a binary for your system
go build -o mevPlus mevPlus.go

# You may copy this binary to your local bin directory for easy access
cp mevPlus /usr/local/bin

# To ensure that MEV-Plus is ready to start and know what commands are available to run/configure MEV-Plus using the built binary:
mevPlus --help

# Alternatively, you may run mevPlus.go directly instead of building it into a binary. You can perform the same steps as above, but instead of running the binary, run the go file directly:
go run mevPlus.go --help

```

## MEV Plus Workflow Overview

The MEV Plus project operates with a clear workflow centered around four key APIs: handleStatus, handleRegisterValidator, handleGetHeader, and handleGetPayload. These APIs are integral to the system's functionality and are accessible through the Builder API module.

### Core Module: The Heart of MEV Plus

At the core of MEV Plus lies the "Core" module, which acts as the central nervous system of the entire project. Core serves as a communication hub, orchestrating interactions between various modules. Other modules are registered within Core, allowing seamless communication and interaction.

### Builder API: Orchestrating RPC Calls

The Builder API module leverages the capabilities of Core to initiate RPC calls. These calls are directed towards the Block Aggregator module, facilitating communication and data exchange. Builder API is instrumental in handling requests and obtaining responses between MEV Plus and the connected Consensus Client.

### Block Aggregator: Your Gateway to Blocks

The Block Aggregator module plays a vital role in MEV Plus. It serves as the gateway to blocks, managing the retrieval of headers and payloads. It connects with the Relay module to obtain the necessary data. Moreover, the Block Aggregator takes charge of storing multiple blocks, thus facilitating efficient block management.

### Relay Module: Your External Gateway

The Relay module serves as the external gateway for MEV Plus. It is responsible for handling external HTTP calls, specifically connecting with relays. This module ensures reliable communication and data exchange with connected relays, enabling seamless interaction with external sources.

In summary, MEV Plus is a well-orchestrated project with clear communication pathways between modules. The Core module acts as the linchpin, while the Builder API, Block Aggregator, and Relay modules each fulfill their unique roles, ensuring the smooth operation and functionality of the entire system. This cohesive workflow promotes efficiency, reliability, and effective data management within MEV Plus.

![MEV-Plus overview](./docs/Flowchart.png?raw=true)

## Usage

Versatile Usage:
MEV-Plus allows seamless integration with multiple beacon nodes and validators, offering flexibility and efficiency.

Configuration Essentials:
In addition to deploying MEV-Plus within your local network, it's crucial to configure the following components:

Beacon Nodes: Each beacon node must be individually tailored to connect with MEV-plus. The configuration specifics may vary depending on the chosen consensus client.

Validators: For optimal performance, validators should set up a preferred relay selection. Please exercise caution and ensure that you only connect to trusted relays to maintain network security.

These configurations are essential to ensure a smooth and secure operation of MEV-Plus within your environment.

## Adding Custom Modules to MEV Plus

MEV Plus is designed to be highly extensible, allowing developers to integrate custom modules to enhance its capabilities. When adding a custom module, there are specific guidelines and requirements to follow for seamless integration. These guidelines ensure that your module is consistent with the structure and functionality of MEV Plus.

### Naming Convention

First and foremost, the name of your custom module must adhere to certain naming conventions:

- **Unique Module Name**: The name of your module should be unique within MEV Plus. It's essential to prevent naming conflicts with existing modules.

- **Avoid Underscores and Spaces**: Module names should not contain underscores or spaces. Instead, use a naming convention that aligns with Go's standard naming guidelines. This ensures clarity and consistency across the codebase.

### Implementation of the Module

When adding a custom module to MEV Plus, you must implement specific components to make it compatible with the system:

- **Command Line Interface (CLI) Commands**: Your module should include CLI commands to manage and control its functionality. These commands allow users to interact with your module seamlessly.

- **New Commands and Flags**: You need to provide new commands and flags that are specific to your module's operations. These commands and flags will be used by users to configure and execute actions within your module.

### Required Service Functions

To ensure that your custom module can be seamlessly integrated into the MEV Plus ecosystem, it should contain a set of mandatory service functions:

- **Name()**: This function is responsible for returning the name of your module. It helps identify your module within MEV Plus.

- **Start()**: The Start function is used to initiate and launch your module's functionality. It kicks off the core processes associated with your module.

- **Stop()**: When called, the Stop function should gracefully halt and terminate your module's operations. It ensures that your module can be shut down without disrupting other parts of the system.

- **ConnectCore()**: This function is vital for establishing a connection with MEV Plus's core components. It allows your module to communicate and collaborate with other parts of the system seamlessly.

- **Configure()**: The Configure function is responsible for setting up your module based on module-specific flags and configurations. It ensures that your module can be customized to suit different use cases and scenarios.

### Installing Modules

To install additional modules, import the module into your MEV Plus `moduleList.go` file and instantiate the module **SERVICE** and the module's CLI **COMMAND** functions in the respective service and command lists variables within the [moduleList.go](moduleList/moduleList.go) file.

Build your modified MEV Plus binary and run it with the required flags.

By adhering to these guidelines and incorporating these components into your custom module, you can extend MEV Plus's functionality according to your unique requirements while ensuring compatibility and consistency with the overall structure of the platform. This modularity allows developers to tailor MEV Plus to specific use cases, making it a flexible and adaptable solution for various applications.

## Pre-packaged Modules

MEV Plus comes with a set of pre-packaged modules that can be used to enhance the platform's functionality. These modules are designed to be highly extensible, allowing developers to customize them to suit their specific needs. The following modules are currently available within MEV Plus:

- Builder API: [builder-api](modules/builder-api/)
- Block Aggregator: [block-aggregator](modules/block-aggregator/)
- External Validator Proxy: [external-validator-proxy](modules/external-validator-proxy/)
- Relay: [relay](modules/relay/)
- K2: [k2](moduleList/moduleList.go#L25)

## Key Features enabled through Pre-packaged Modules

- **Customizable Auction Window** (by means of [block-aggregator](modules/block-aggregator/)): MEV Plus allows you to configure an auction time window, during which all requests for headers are paused. This feature can be tailored to your specific needs, with the option to set the window duration to zero for instantaneous requests.

- **Fairness and Competition** (by means of [block-aggregator](modules/block-aggregator/)): The auction mechanism fosters healthy competition among block builders and proposers. It introduces a separation of roles where builders craft blocks that include transaction orderflow, and a bounty bid is awarded to the validator who proposes the block. This separation of responsibilities enhances decentralization and censorship resistance within the Ethereum ecosystem.

- **PBS Operations** (by means of [external-validator-proxy](modules/external-validator-proxy/) and/or [relay](modules/relay/) through [builder-api](modules/builder-api/)): MEV Plus allows you to perform PBS operations, including block building and relaying. These operations are facilitated through the Builder API module, which leverages the capabilities of the Relay module to communicate with external sources. Of such external sources is the option to use the [external-validator-proxy](modules/external-validator-proxy/) module to connect to existing validator proxy softwares that are already connected/configured to a PBS pipeline. This would allow MEV Plus to act as a side car to the existing validator proxy software. All PBS requests are forwarded to and from the connected external validator address(es) `-externalValidatorProxy.address`. This module allows for a maximum of two proxy softwares (comma-separated addresses) to be served through MEV Plus, allowing for the use of all MEV Plus features and modules while maintaining the existing PBS pipeline. This can also be run in conjunction with the [relay](modules/relay/) module to allow for the use of the MEV Plus relay functionality to aggregate block sources from both the connected external validator proxy softwares and the MEV Plus relay module.

- **Validator Restaking** (by means of [k2](moduleList/moduleList.go#L25)): MEV Plus allows you to restake your validator. This feature is facilitated through the K2 module, which is responsible for managing validator restaking operations. It allows you to restake your validator seamlessly, through your nodes' builder API registration requests to MEV Plus. This feature additionally facilitates the on-chain registration of your validator in a proposer registry to track your validator's equivalent performance on the execution layer from the consensus layer across many protocols.

## Contributing

We welcome contributions to MEV Plus. If you'd like to contribute, please follow these guidelines:

1. Fork the project.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them with clear and concise messages.
4. Push your changes to your fork.
5. Create a pull request to the main repository.

## License

This project is licensed under MIT - see the [LICENSE.md](LICENSE.md) file for details.
