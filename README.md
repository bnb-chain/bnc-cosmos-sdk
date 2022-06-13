# Why we create this repo

This repo is forked from [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

The BNB Beacon Chain leverages cosmos-sdk to fast build a dApp running with tendermint. As the app becomes more and more complex, the original cosmos-sdk can hardly fit all requirements. 
We changed a lot to the copied sdk, but it makes the future integration harder and harder. So we decided to fork cosmos-sdk and add features onto it.

## Key Features

1. **Native Cross Chain Support**. Cross-chain communication is the key foundation to allow the community to take advantage of the BNB Beacon Chain and BNB Smart Chain dual chain structure.
2. **Staking**. Staking and reward logic should be built into the blockchain, and automatically executed as the blocking happens. Cosmos Hub, who shares the same Tendermint consensus and libraries with BNB Beacon Chain, works in this way. In order to keep the compatibility and reuse the good foundation of BC, the staking logic of BSC is implemented on BC. The BSC validator set is determined by its staking and delegation logic, via a staking module built on BC for BSC, and propagated every day UTC 00:00 from BC to BSC via Cross-Chain communication.
3. **Rewarding**. Both the validator update and reward distribution happen every day around UTC 00:00. This is to save the cost of frequent staking updates and block reward distribution. This cost can be significant, as the blocking reward is collected on BSC and distributed on BC to BSC validators and delegators. 
4. **Slashing**. Slashing is part of the on-chain governance, to ensure the malicious or negative behaviors are punished. BSC slash can be submitted by anyone. The transaction submission requires slash evidence and cost fees but also brings a larger reward when it is successful. So far there are two slashable cases: Double Sign and Inavailability.
5. **ParamHub && Governance**. There are many system parameters to control the behavior of the BNB Beacon Chain and BNB Smart Chain, e.g. slash amount, cross-chain transfer fees. All these parameters will be determined by BSC and BC Validator Set together through a proposal-vote process based on their staking. Such the process will be carried on cosmos sdk.
6. **Performance Improvement** Parallelization, dedicated cache, priority lock and many other program skills are applied to improvement the capacity of BNB Beacon Chain.

## Quick Start

See the [Cosmos Docs](https://cosmos.network/docs/) and [Getting started with the SDK](https://cosmos.network/docs/sdk/core/intro.html).

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to bnc-cosmos-sdk, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. 

Please make sure your contributions adhere to our coding guidelines:

- Code must adhere to the official Go formatting guidelines (i.e. uses gofmt).
- Code must be documented adhering to the official Go [commentary guidelines](https://go.dev/doc/effective_go#commentary).
- Pull requests need to be based on and opened against the master branch.
Commit messages should be prefixed with the working progress.
E.g. "\[WIP\] make trace configs optional", "\[R4R\] make trace configs optional". 
