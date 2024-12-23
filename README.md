# Prysm exercise

## Context

You are developing software that relies on Ethereum Beacon Nodes. One of these nodes
seems to be malfunctioning. To help track its performance, we need to add logging to the
node. We suspect that Prysm is not verifying attestations correctly.

## Goal

Your task is to modify Prysm's code by adding a structure that:

- Counts successfully verified attestations.
- Counts failed attestations and records the reason for each failure.
- Outputs a summary of the collected data at the end of each epoch.

## Solution

### 1. Scope definition

Prysm is made of several components such as: beacon-node, validator, slasher. Here the context clearly state that we
are working with beacon-nodes so we will focus our attention to this component only.
