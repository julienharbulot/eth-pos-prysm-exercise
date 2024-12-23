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
        ├── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L229">Service.validatePendingAtts()</a>
        └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/pending_attestations_queue.go#L91">Service.processAttestations()</a>
            ├── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_aggregate_proof.go#L141"><b></b>validateAggregatedAtt()</b></a>
            └── <a href="https://github.com/prysmaticlabs/prysm/blob/develop/beacon-chain/sync/validate_beacon_attestation.go#L254"><b></b>validateUnaggregatedAttWithState()</b></a>
</pre>

To confirm we did not miss any important aspect of the code related to attestation validation, we can search for relevant keywords throughout the codebase: [in the operations service](https://github.com/search?q=repo%3Aprysmaticlabs%2Fprysm+path%3Abeacon-chain%2Foperations+%22validate%22&type=code) and [in the sync service](https://github.com/search?q=repo%3Aprysmaticlabs%2Fprysm+path%3Abeacon-chain%2Fsync+%22validate%22&type=code).

### 3. Code inspection: reporting code

Now that we know where the relevant information is located, we need to find a way to extract it and report it.
