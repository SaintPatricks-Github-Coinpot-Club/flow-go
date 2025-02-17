package fvm_test

import (
	"bytes"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/flow-go/fvm/programs"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/fvm/utils"
	"github.com/onflow/flow-go/model/flow"
)

func TestSafetyCheck(t *testing.T) {

	t.Run("parsing error in transaction", func(t *testing.T) {

		buffer := &bytes.Buffer{}
		log := zerolog.New(buffer)
		txInvoker := fvm.NewTransactionInvoker()

		code := `X`

		proc := fvm.Transaction(&flow.TransactionBody{Script: []byte(code)}, 0)

		view := utils.NewSimpleView()
		context := fvm.NewContext(fvm.WithLogger(log))

		txnState := state.NewTransactionState(
			view,
			state.DefaultParameters().
				WithMaxKeySizeAllowed(context.MaxStateKeySize).
				WithMaxValueSizeAllowed(context.MaxStateValueSize).
				WithMaxInteractionSizeAllowed(context.MaxStateInteractionSize),
		)

		blockPrograms := programs.NewEmptyBlockPrograms()
		txnPrograms, err := blockPrograms.NewTransactionPrograms(0, 0)
		require.NoError(t, err)

		err = txInvoker.Process(context, proc, txnState, txnPrograms)
		require.Error(t, err)

		require.NotContains(t, buffer.String(), "programs")
		require.NotContains(t, buffer.String(), "codes")

	})

	t.Run("checking error in transaction", func(t *testing.T) {

		buffer := &bytes.Buffer{}
		log := zerolog.New(buffer)
		txInvoker := fvm.NewTransactionInvoker()

		code := `transaction(arg: X) { }`

		proc := fvm.Transaction(&flow.TransactionBody{Script: []byte(code)}, 0)

		view := utils.NewSimpleView()
		context := fvm.NewContext(fvm.WithLogger(log))

		txnState := state.NewTransactionState(
			view,
			state.DefaultParameters().
				WithMaxKeySizeAllowed(context.MaxStateKeySize).
				WithMaxValueSizeAllowed(context.MaxStateValueSize).
				WithMaxInteractionSizeAllowed(context.MaxStateInteractionSize),
		)

		blockPrograms := programs.NewEmptyBlockPrograms()
		txnPrograms, err := blockPrograms.NewTransactionPrograms(0, 0)
		require.NoError(t, err)

		err = txInvoker.Process(context, proc, txnState, txnPrograms)
		require.Error(t, err)

		require.NotContains(t, buffer.String(), "programs")
		require.NotContains(t, buffer.String(), "codes")
	})
}

type ErrorReturningRuntime struct {
	TxErrors []error
}

func (e *ErrorReturningRuntime) NewScriptExecutor(script runtime.Script, context runtime.Context) runtime.Executor {
	panic("NewScriptExecutor not expected")
}

func (e *ErrorReturningRuntime) NewTransactionExecutor(script runtime.Script, context runtime.Context) runtime.Executor {
	panic("NewTransactionExecutor not expected")
}

func (e *ErrorReturningRuntime) NewContractFunctionExecutor(contractLocation common.AddressLocation, functionName string, arguments []cadence.Value, argumentTypes []sema.Type, context runtime.Context) runtime.Executor {
	panic("NewContractFunctionExecutor not expected")
}

func (e *ErrorReturningRuntime) SetInvalidatedResourceValidationEnabled(_ bool) {
	panic("SetInvalidatedResourceValidationEnabled not expected")
}

func (e *ErrorReturningRuntime) SetResourceOwnerChangeHandlerEnabled(_ bool) {
	panic("SetResourceOwnerChangeHandlerEnabled not expected")
}

var _ runtime.Runtime = &ErrorReturningRuntime{}

func (e *ErrorReturningRuntime) ExecuteTransaction(_ runtime.Script, _ runtime.Context) error {
	if len(e.TxErrors) == 0 {
		panic("no tx errors left")
	}

	errToReturn := e.TxErrors[0]
	e.TxErrors = e.TxErrors[1:]
	return errToReturn
}

func (*ErrorReturningRuntime) ExecuteScript(_ runtime.Script, _ runtime.Context) (cadence.Value, error) {
	panic("ExecuteScript not expected")
}

func (*ErrorReturningRuntime) ParseAndCheckProgram(_ []byte, _ runtime.Context) (*interpreter.Program, error) {
	panic("ParseAndCheckProgram not expected")
}

func (*ErrorReturningRuntime) SetCoverageReport(_ *runtime.CoverageReport) {
	panic("not used coverage")
}

func (*ErrorReturningRuntime) SetContractUpdateValidationEnabled(_ bool) {
	panic("SetContractUpdateValidationEnabled not expected")
}

func (*ErrorReturningRuntime) SetAtreeValidationEnabled(_ bool) {
	panic("SetAtreeValidationEnabled not expected")
}

func (e *ErrorReturningRuntime) ReadStored(_ common.Address, _ cadence.Path, _ runtime.Context) (cadence.Value, error) {
	return nil, nil
}

func (e *ErrorReturningRuntime) ReadLinked(_ common.Address, _ cadence.Path, _ runtime.Context) (cadence.Value, error) {
	panic("ReadLinked not expected")
}

func (e *ErrorReturningRuntime) InvokeContractFunction(_ common.AddressLocation, _ string, _ []cadence.Value, _ []sema.Type, _ runtime.Context) (cadence.Value, error) {
	panic("InvokeContractFunction not expected")
}

func (e *ErrorReturningRuntime) SetTracingEnabled(_ bool) {
	panic("SetTracingEnabled not expected")
}

func (*ErrorReturningRuntime) SetDebugger(_ *interpreter.Debugger) {
	panic("SetDebugger not expected")
}

func (ErrorReturningRuntime) Storage(runtime.Context) (*runtime.Storage, *interpreter.Interpreter, error) {
	panic("Storage not expected")
}
