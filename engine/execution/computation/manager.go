package computation

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/engine/execution/computation/computer"
	"github.com/onflow/flow-go/engine/execution/computation/computer/uploader"
	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/flow-go/fvm/programs"
	reusableRuntime "github.com/onflow/flow-go/fvm/runtime"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/executiondatasync/provider"
	"github.com/onflow/flow-go/module/mempool/entity"
	"github.com/onflow/flow-go/module/trace"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/utils/debug"
	"github.com/onflow/flow-go/utils/logging"
)

const (
	DefaultScriptLogThreshold       = 1 * time.Second
	DefaultScriptExecutionTimeLimit = 10 * time.Second

	MaxScriptErrorMessageSize = 1000 // 1000 chars

	ReusableCadenceRuntimePoolSize = 1000
)

var uploadEnabled = true

func SetUploaderEnabled(enabled bool) {
	uploadEnabled = enabled

	log.Info().Msgf("changed uploadEnabled to %v", enabled)
}

func GetUploaderEnabled() bool {
	return uploadEnabled
}

type ComputationManager interface {
	ExecuteScript(context.Context, []byte, [][]byte, *flow.Header, state.View) ([]byte, error)
	ComputeBlock(
		ctx context.Context,
		block *entity.ExecutableBlock,
		view state.View,
	) (*execution.ComputationResult, error)
	GetAccount(addr flow.Address, header *flow.Header, view state.View) (*flow.Account, error)
}

type ComputationConfig struct {
	CadenceTracing           bool
	ExtensiveTracing         bool
	ProgramsCacheSize        uint
	ScriptLogThreshold       time.Duration
	ScriptExecutionTimeLimit time.Duration

	// When NewCustomVirtualMachine is nil, the manager will create a standard
	// fvm virtual machine via fvm.NewVM.  Otherwise, the manager
	// will create a virtual machine using this function.
	//
	// Note that this is primarily used for testing.
	NewCustomVirtualMachine func() computer.VirtualMachine
}

// Manager manages computation and execution
type Manager struct {
	log                      zerolog.Logger
	tracer                   module.Tracer
	metrics                  module.ExecutionMetrics
	me                       module.Local
	protoState               protocol.State
	vm                       computer.VirtualMachine
	vmCtx                    fvm.Context
	blockComputer            computer.BlockComputer
	programsCache            *programs.ChainPrograms
	scriptLogThreshold       time.Duration
	scriptExecutionTimeLimit time.Duration
	uploaders                []uploader.Uploader
	rngLock                  *sync.Mutex
	rng                      *rand.Rand
}

func New(
	logger zerolog.Logger,
	metrics module.ExecutionMetrics,
	tracer module.Tracer,
	me module.Local,
	protoState protocol.State,
	vmCtx fvm.Context,
	committer computer.ViewCommitter,
	uploaders []uploader.Uploader,
	executionDataProvider *provider.Provider,
	params ComputationConfig,
) (*Manager, error) {
	log := logger.With().Str("engine", "computation").Logger()

	var vm computer.VirtualMachine
	if params.NewCustomVirtualMachine != nil {
		vm = params.NewCustomVirtualMachine()
	} else {
		vm = fvm.NewVM()
	}

	options := []fvm.Option{
		fvm.WithReusableCadenceRuntimePool(
			reusableRuntime.NewReusableCadenceRuntimePool(
				ReusableCadenceRuntimePoolSize,
				runtime.Config{
					TracingEnabled: params.CadenceTracing,
				})),
	}
	if params.ExtensiveTracing {
		options = append(options, fvm.WithExtensiveTracing())
	}

	vmCtx = fvm.NewContextFromParent(vmCtx, options...)

	blockComputer, err := computer.NewBlockComputer(
		vm,
		vmCtx,
		metrics,
		tracer,
		log.With().Str("component", "block_computer").Logger(),
		committer,
		executionDataProvider,
	)

	if err != nil {
		return nil, fmt.Errorf("cannot create block computer: %w", err)
	}

	programsCache, err := programs.NewChainPrograms(params.ProgramsCacheSize)
	if err != nil {
		return nil, fmt.Errorf("cannot create programs cache: %w", err)
	}

	e := Manager{
		log:                      log,
		tracer:                   tracer,
		metrics:                  metrics,
		me:                       me,
		protoState:               protoState,
		vm:                       vm,
		vmCtx:                    vmCtx,
		blockComputer:            blockComputer,
		programsCache:            programsCache,
		scriptLogThreshold:       params.ScriptLogThreshold,
		scriptExecutionTimeLimit: params.ScriptExecutionTimeLimit,
		uploaders:                uploaders,
		rngLock:                  &sync.Mutex{},
		rng:                      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return &e, nil
}

func (e *Manager) VM() computer.VirtualMachine {
	return e.vm
}

func (e *Manager) ExecuteScript(
	ctx context.Context,
	code []byte,
	arguments [][]byte,
	blockHeader *flow.Header,
	view state.View,
) ([]byte, error) {

	startedAt := time.Now()
	memAllocBefore := debug.GetHeapAllocsBytes()

	// allocate a random ID to be able to track this script when its done,
	// scripts might not be unique so we use this extra tracker to follow their logs
	// TODO: this is a temporary measure, we could remove this in the future
	if e.log.Debug().Enabled() {
		e.rngLock.Lock()
		trackerID := e.rng.Uint32()
		e.rngLock.Unlock()

		trackedLogger := e.log.With().Hex("script_hex", code).Uint32("trackerID", trackerID).Logger()
		trackedLogger.Debug().Msg("script is sent for execution")
		defer func() {
			trackedLogger.Debug().Msg("script execution is complete")
		}()
	}

	requestCtx, cancel := context.WithTimeout(ctx, e.scriptExecutionTimeLimit)
	defer cancel()

	script := fvm.NewScriptWithContextAndArgs(code, requestCtx, arguments...)
	blockCtx := fvm.NewContextFromParent(
		e.vmCtx,
		fvm.WithBlockHeader(blockHeader),
		fvm.WithBlockPrograms(
			e.programsCache.NewBlockProgramsForScript(blockHeader.ID())))

	err := func() (err error) {

		start := time.Now()

		defer func() {

			prepareLog := func() *zerolog.Event {

				args := make([]string, 0, len(arguments))
				for _, a := range arguments {
					args = append(args, hex.EncodeToString(a))
				}
				return e.log.Error().
					Hex("script_hex", code).
					Str("args", strings.Join(args, ","))
			}

			elapsed := time.Since(start)

			if r := recover(); r != nil {
				prepareLog().
					Interface("recovered", r).
					Msg("script execution caused runtime panic")

				err = fmt.Errorf("cadence runtime error: %s", r)
				return
			}
			if elapsed >= e.scriptLogThreshold {
				prepareLog().
					Dur("duration", elapsed).
					Msg("script execution exceeded threshold")
			}
		}()

		return e.vm.RunV2(blockCtx, script, view)
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to execute script (internal error): %w", err)
	}

	if script.Err != nil {
		scriptErrMsg := script.Err.Error()
		if len(scriptErrMsg) > MaxScriptErrorMessageSize {
			split := int(MaxScriptErrorMessageSize/2) - 1
			var sb strings.Builder
			sb.WriteString(scriptErrMsg[:split])
			sb.WriteString(" ... ")
			sb.WriteString(scriptErrMsg[len(scriptErrMsg)-split:])
			scriptErrMsg = sb.String()
		}

		return nil, fmt.Errorf("failed to execute script at block (%s): %s", blockHeader.ID(), scriptErrMsg)
	}

	encodedValue, err := jsoncdc.Encode(script.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode runtime value: %w", err)
	}

	memAllocAfter := debug.GetHeapAllocsBytes()
	e.metrics.ExecutionScriptExecuted(time.Since(startedAt), script.GasUsed, memAllocAfter-memAllocBefore, script.MemoryEstimate)

	return encodedValue, nil
}

func (e *Manager) ComputeBlock(
	ctx context.Context,
	block *entity.ExecutableBlock,
	view state.View,
) (*execution.ComputationResult, error) {

	e.log.Debug().
		Hex("block_id", logging.Entity(block.Block)).
		Msg("received complete block")

	blockPrograms := e.programsCache.GetOrCreateBlockPrograms(
		block.ID(),
		block.ParentID())

	result, err := e.blockComputer.ExecuteBlock(ctx, block, view, blockPrograms)
	if err != nil {
		e.log.Error().
			Hex("block_id", logging.Entity(block.Block)).
			Msg("failed to compute block result")

		return nil, fmt.Errorf("failed to execute block: %w", err)
	}

	e.log.Debug().Hex("block_id", logging.Entity(block.Block)).Msg("block result computed")

	e.log.Debug().Hex("block_id", logging.Entity(block.Block)).Msg("programs cache updated")

	if uploadEnabled {
		var group errgroup.Group

		for _, uploader := range e.uploaders {
			uploader := uploader

			group.Go(func() error {
				span, _ := e.tracer.StartSpanFromContext(ctx, trace.EXEUploadCollections)
				defer span.End()

				return uploader.Upload(result)
			})
		}

		err = group.Wait()

		if err != nil {
			return nil, fmt.Errorf("failed to upload block result: %w", err)
		}
	}

	e.log.Debug().
		Hex("block_id", logging.Entity(result.ExecutableBlock.Block)).
		Msg("computed block result")

	return result, nil
}

func (e *Manager) GetAccount(address flow.Address, blockHeader *flow.Header, view state.View) (*flow.Account, error) {
	blockCtx := fvm.NewContextFromParent(
		e.vmCtx,
		fvm.WithBlockHeader(blockHeader),
		fvm.WithBlockPrograms(
			e.programsCache.NewBlockProgramsForScript(blockHeader.ID())))

	account, err := e.vm.GetAccountV2(blockCtx, address, view)
	if err != nil {
		return nil, fmt.Errorf("failed to get account (%s) at block (%s): %w", address.String(), blockHeader.ID(), err)
	}

	return account, nil
}
