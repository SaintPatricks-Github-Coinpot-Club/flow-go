package fvm

import (
	otelTrace "go.opentelemetry.io/otel/trace"

	"github.com/onflow/flow-go/fvm/environment"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/model/flow"
)

func NewTransactionEnvironment(
	ctx Context,
	txnState *state.TransactionState,
	programs environment.TransactionPrograms,
	tx *flow.TransactionBody,
	txIndex uint32,
	traceSpan otelTrace.Span,
	eventHandlerOptions ...handler.FlowEventHandlerOption,
) environment.Environment {
	ctx.RootSpan = traceSpan
	ctx.TxIndex = txIndex
	ctx.TxId = tx.ID()
	ctx.TxBody = tx

	return environment.NewTransactionEnvironment(
		ctx.EnvironmentParams,
		txnState,
		programs)
}
