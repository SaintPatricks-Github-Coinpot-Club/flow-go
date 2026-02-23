# Design: Propagate ComputationUsed Through Access TransactionResult API

## Problem

`ComputationUsed` is available in `flow.LightTransactionResult` (stored locally on the access node)
and was recently added to `model/access/transaction_result.go`, but it is not propagated through
to gRPC or REST responses.

## Scope

Only the **local (indexed) provider** path is in scope. The EN provider cannot be updated because
the execution node gRPC response proto does not include per-transaction computation used.

## Changes

### 1. `engine/access/rpc/backend/transactions/provider/local.go`

The `TransactionResult` method (lookup by block ID + tx ID) does not set `ComputationUsed` in
its return value. The `txResult` variable is already a `LightTransactionResult` with
`ComputationUsed` available — just add `ComputationUsed: txResult.ComputationUsed`.

The other two methods (`TransactionResultByIndex`, `TransactionResultsByBlockID`) are already
fixed on this branch.

### 2. `engine/common/rpc/convert/transaction_result.go`

- `TransactionResultToMessage`: add `ComputationUsage: result.ComputationUsed` so the field is
  included in the gRPC response. The protobuf field `computation_usage` (field 10, type uint64)
  already exists in `access.TransactionResultResponse`.
- `MessageToTransactionResult`: add `ComputationUsed: message.ComputationUsage` for round-trip
  consistency (used when proxying results from other access nodes).

### 3. `engine/access/rest/common/models/transaction.go`

`TransactionResult.Build` hardcodes `t.ComputationUsed = util.FromUint(uint64(0))` with a TODO.
Replace with `t.ComputationUsed = util.FromUint(txr.ComputationUsed)`.

## Non-changes

- No proto changes needed — `ComputationUsage` (field 10) already exists in
  `access.TransactionResultResponse`.
- EN provider is not updated — the EN response proto has no per-transaction computation field.
