# Prysm exercise

## Assignment

You are developing software that relies on Ethereum Beacon Nodes. One of these nodes
seems to be malfunctioning. To help track its performance, we need to add logging to the
node. We suspect that Prysm is not verifying attestations correctly.

Your task is to modify Prysm's code by adding a structure that:

- Counts successfully verified attestations.
- Counts failed attestations and records the reason for each failure.
- Outputs a summary of the collected data at the end of each epoch.

## Solution

### 1. Scope definition

Prysm is made of several components such as: beacon-node, validator, slasher. Here the assignment clearly state that we
are working with beacon-nodes so we will focus our attention to this component only.

According to the [official documentation](https://docs.prylabs.network/docs/how-prysm-works/beacon-node), the beacon-node is partitioned into several services. It seems that the [sync service](https://docs.prylabs.network/docs/how-prysm-works/beacon-node#sync-service) is where we will need to work so we will start our investigation there.

Also, since a Beacon node receives attestations either directly from Validators or from other beacon-nodes through the Gossip network, we will keep an eye out from two different possible sources of incoming attestations.

### 2. Code inspection: validation code

Looking at [how the prysm binary is built](https://docs.prylabs.network/docs/advanced/proof-of-stake-devnet):

```bash
go build -o=../beacon-chain ./cmd/beacon-chain
```

We look for the main function in the `cmd/beacon-chain` directory: [cmd/beacon-chain/main.go:221](https://github.com/prysmaticlabs/prysm/blob/96b31a9f64a8f8b3909b11171ce3c2dab877cfc7/cmd/beacon-chain/main.go#L221).

From there, we can follow the execution tree down to the validation code as follows:

<pre>
<a href="https://github.com/prysmaticlabs/prysm/blob/96b31a9f64a8f8b3909b11171ce3c2dab877cfc7/cmd/beacon-chain/main.go#L221">main()</a>
└── <a href="https://github.com/prysmaticlabs/prysm/blob/96b31a9f64a8f8b3909b11171ce3c2dab877cfc7/cmd/beacon-chain/main.go#L256">startNode()</a>
    └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/node/node.go#L127">Node.New()</a>
        └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/node/node.go#L301">Node.registerServices()</a>
            └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/node/node.go#L794">Node.registerSyncService()</a>
                └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/service.go#L222">Service.Start()</a>
<a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/service.go#L222">Service.Start()</a>
├── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/service.go#L329">Service.StartTasksPostInitialSync()</a>
│   └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/subscriber.go#L81">registerSubscriber()</a>
│       └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_beacon_attestation.go#L38"><b></b>validateComitteeIndexBeaconAttestation()</b></a>
│           └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_beacon_attestation.go#L128">savePendingAtt()</a>
└── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L29">Service.processPendingAttsQueue()</a>
    └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L45">Service.processPendingAtts()</a>
        ├── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L245">Service.validatePendingAtts()</a>
        └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L92">Service.processAttestations()</a>
            ├── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_aggregate_proof.go#L155"><b></b>validateAggregatedAtt()</b></a>
            └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_beacon_attestation.go#L254"><b></b>validateUnaggregatedAttWithState()</b></a>
</pre>

To confirm we did not miss any important aspect of the code related to attestation validation, we can search for relevant keywords throughout the codebase: [in the operations service](https://github.com/search?q=repo%3Aprysmaticlabs%2Fprysm+path%3Abeacon-chain%2Foperations+%22validate%22&type=code) and [in the sync service](https://github.com/search?q=repo%3Aprysmaticlabs%2Fprysm+path%3Abeacon-chain%2Fsync+%22validate%22&type=code).

### 3. Code inspection: reporting code

Now that we know where the relevant information is located, we need to find a way to extract it and report it.

In the `validateComitteeIndexBeaconAttestation` method, there is [an interesting broadcast](https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_beacon_attestation.go#L76) on the `s.cfg.attestationNotifier` member variable:

```go
// Broadcast the unaggregated attestation on a feed to notify other services in the beacon node
// of a received unaggregated attestation.
 s.cfg.attestationNotifier.OperationFeed().Send(&feed.Event{
  Type: operation.UnaggregatedAttReceived,
  Data: &operation.UnAggregatedAttReceivedData{
   Attestation: att,
  },
 })
```

Which could be a hint at how to properly implementing the reporting we need. However, this Notifier is not used in `validateAggregatedAtt` (nor in `validateUnaggregatedAttWithState`). And since this `attestationNotifier` is only used to notify about valid attestations and uses types defined [in the core package](https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/core/feed/operation/events.go#L9), so it wouldn't be appropriate to extend it to notify about invalid attestations.

### 4. Implementation

Since we have a 4 hours time constraint and this is only an assignment, we will directly choose the obvious solution: add a member variable in the service to collect the data and lazy report the data once per epoch. Please see relevant commits in this repository.

To make the report useful, we save the number of attestation that is ok (`nOk uint64`) the number of attestations that is rejected (`nErr uint64`) and for each individual error message, the number of times we have encountered this message (`map[string]uint64`).

To achieve a smaller diff and make the code easier to maintain, I decided to decorate the validation functions with this pattern:

```go
func validationFunction(...) {
    v, err := validationFunctionImpl(...)
    updateReport(v, err)
    return v, err
}
```

Possible alternatives that would have required more work were:

1. call the update method every time a validation function is called. This is less robust: in this alternative, everytime the validation function is called we have to remember to call the update function too.

```go
func someCallingCode() {
  v, err := validationFunction(...)
  updateReport(v, err)
  if (err != nil) {
    // other client logic
  }
}
```

2. call the update method in the validation function directly. This has two drawbacks: (a) the validation function shoould ideally be a pure function without side effect to make the code simpler, and (b) each validatiaon function has multiple exit point which would have required multiple call to the update function and is error prone in case another exit point is introduced later and the update function is forgotten.

```go
func validationFunction(...) {
  if (a) {
    updateReport(v_a, err_a)
    return v_a, err_a
  } else {
    updateReport(v_not_a, err_not_a)
    return v_not_a, err_not_a
  }
}
```

### 5. Possible ameliorations

From a purely architectural point of view, it would be better to move the reporting to a dedicated module or service. Currently I did not find a service that would be a good fit for this data collection so I opted for a simpler implementation inside the `sync` service itself.

In theory the `uint64` counter variables could overflow so having a reset logic at some point would be preferable. In practice, it is unlikely that the variables will overflow before the node is restarted. As a gauge of how big uint64 are here is a count of how many validataion per seconds we can receive if this capacity were to overflow in a year: `uint64.max / seconds_in_a_year = 584,942,417,355`.

### 6. Running the code locally

Use the configuration files in the `devnet` directory and follow instruction in the second part of [this guide](https://docs.prylabs.network/docs/advanced/proof-of-stake-devnet)

You can build our custom beacon code as follows:

```shell
go1.22.10 build -o devnet/beacon-chain.exe ./prysm/cmd/beacon-chain
go1.22.10 build -o devnet/validator.exe ./prysm/cmd/validator
go1.22.10 build -o devnet/prysmctl.exe ./prysm/cmd/prysmctl
```

And run the components as follows:

```shell
./geth --http --http.api eth,net,web3 --ws --ws.api eth,net,web3 --authrpc.jwtsecret jwt.hex --datadir data/gethdata --nodiscover --syncmode full --allow-insecure-unlock --unlock 0x123463a4b065722e99115d6c222f267d9cabb524 --password password.txt

./beacon-chain --datadir data/beacondata --min-sync-peers 0 --genesis-state genesis.ssz --bootstrap-node= --interop-eth1data-votes --chain-config-file config.yml --contract-deployment-block 0 --chain-id 1 --accept-terms-of-use --jwt-secret jwt.hex --suggested-fee-recipient 0x123463a4B065722E99115D6c222f267d9cABb524 --minimum-peers-per-subnet 0 --enable-debug-rpc-endpoints --execution-endpoint http://127.0.0.1:8545 --verbosity=debug

./validator --datadir data/validatordata --accept-terms-of-use --interop-num-validators 64 --chain-config-file
 config.yml
```
