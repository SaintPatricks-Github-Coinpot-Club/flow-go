package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
)

type ExecutionCollector struct {
	tracer                              module.Tracer
	stateReadsPerBlock                  prometheus.Histogram
	totalExecutedBlocksCounter          prometheus.Counter
	totalExecutedCollectionsCounter     prometheus.Counter
	totalExecutedTransactionsCounter    prometheus.Counter
	totalExecutedScriptsCounter         prometheus.Counter
	totalFailedTransactionsCounter      prometheus.Counter
	lastExecutedBlockHeightGauge        prometheus.Gauge
	stateStorageDiskTotal               prometheus.Gauge
	storageStateCommitment              prometheus.Gauge
	forestApproxMemorySize              prometheus.Gauge
	forestNumberOfTrees                 prometheus.Gauge
	latestTrieRegCount                  prometheus.Gauge
	latestTrieRegCountDiff              prometheus.Gauge
	latestTrieRegSize                   prometheus.Gauge
	latestTrieRegSizeDiff               prometheus.Gauge
	latestTrieMaxDepthTouched           prometheus.Gauge
	updated                             prometheus.Counter
	proofSize                           prometheus.Gauge
	updatedValuesNumber                 prometheus.Counter
	updatedValuesSize                   prometheus.Gauge
	updatedDuration                     prometheus.Histogram
	updatedDurationPerValue             prometheus.Histogram
	readValuesNumber                    prometheus.Counter
	readValuesSize                      prometheus.Gauge
	readDuration                        prometheus.Histogram
	readDurationPerValue                prometheus.Histogram
	blockComputationUsed                prometheus.Histogram
	blockExecutionTime                  prometheus.Histogram
	blockTransactionCounts              prometheus.Histogram
	blockCollectionCounts               prometheus.Histogram
	collectionComputationUsed           prometheus.Histogram
	collectionExecutionTime             prometheus.Histogram
	collectionTransactionCounts         prometheus.Histogram
	collectionRequestSent               prometheus.Counter
	collectionRequestRetried            prometheus.Counter
	transactionParseTime                prometheus.Histogram
	transactionCheckTime                prometheus.Histogram
	transactionInterpretTime            prometheus.Histogram
	transactionExecutionTime            prometheus.Histogram
	transactionMemoryUsage              prometheus.Histogram
	transactionMemoryEstimate           prometheus.Histogram
	transactionMemoryDifference         prometheus.Histogram
	transactionComputationUsed          prometheus.Histogram
	transactionEmittedEvents            prometheus.Histogram
	scriptExecutionTime                 prometheus.Histogram
	scriptComputationUsed               prometheus.Histogram
	scriptMemoryUsage                   prometheus.Histogram
	scriptMemoryEstimate                prometheus.Histogram
	scriptMemoryDifference              prometheus.Histogram
	numberOfAccounts                    prometheus.Gauge
	chunkDataPackRequestProcessedTotal  prometheus.Counter
	stateSyncActive                     prometheus.Gauge
	blockDataUploadsInProgress          prometheus.Gauge
	blockDataUploadsDuration            prometheus.Histogram
	maxCollectionHeight                 prometheus.Gauge
	computationResultUploadedCount      prometheus.Counter
	computationResultUploadRetriedCount prometheus.Counter
}

func NewExecutionCollector(tracer module.Tracer) *ExecutionCollector {

	forestApproxMemorySize := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "forest_approx_memory_size",
		Help:      "an approximate size of in-memory forest in bytes",
	})

	forestNumberOfTrees := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "forest_number_of_trees",
		Help:      "the number of trees in memory",
	})

	latestTrieRegCount := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "latest_trie_reg_count",
		Help:      "the number of allocated registers (latest created trie)",
	})

	latestTrieRegCountDiff := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "latest_trie_reg_count_diff",
		Help:      "the difference between number of unique register allocated of the latest created trie and parent trie",
	})

	latestTrieRegSize := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "latest_trie_reg_size",
		Help:      "the size of allocated registers (latest created trie)",
	})

	latestTrieRegSizeDiff := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "latest_trie_reg_size_diff",
		Help:      "the difference between size of unique register allocated of the latest created trie and parent trie",
	})

	latestTrieMaxDepthTouched := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "latest_trie_max_depth_touched",
		Help:      "the maximum depth touched of the latest created trie",
	})

	updatedCount := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "updates_counted",
		Help:      "the number of updates",
	})

	proofSize := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "average_proof_size",
		Help:      "the average size of a single generated proof in bytes",
	})

	updatedValuesNumber := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "update_values_number",
		Help:      "the total number of values updated",
	})

	updatedValuesSize := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "update_values_size",
		Help:      "the total size of values for single update in bytes",
	})

	updatedDuration := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "update_duration",
		Help:      "the duration of update operation",
		Buckets:   []float64{0.05, 0.2, 0.5, 1, 2, 5},
	})

	updatedDurationPerValue := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "update_duration_per_value",
		Help:      "the duration of update operation per value",
		Buckets:   []float64{0.05, 0.2, 0.5, 1, 2, 5},
	})

	readValuesNumber := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "read_values_number",
		Help:      "the total number of values read",
	})

	readValuesSize := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "read_values_size",
		Help:      "the total size of values for single read in bytes",
	})

	readDuration := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "read_duration",
		Help:      "the duration of read operation",
		Buckets:   []float64{0.05, 0.2, 0.5, 1, 2, 5},
	})

	readDurationPerValue := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemMTrie,
		Name:      "read_duration_per_value",
		Help:      "the duration of read operation per value",
		Buckets:   []float64{0.05, 0.2, 0.5, 1, 2, 5},
	})

	blockExecutionTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "block_execution_time_milliseconds",
		Help:      "the total time spent on block execution in milliseconds",
		Buckets:   []float64{100, 500, 1000, 1500, 2000, 2500, 3000, 6000},
	})

	blockComputationUsed := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "block_computation_used",
		Help:      "the total amount of computation used by a block",
		Buckets:   []float64{1000, 10000, 100000, 500000, 1000000, 5000000, 10000000},
	})

	blockTransactionCounts := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "block_transaction_counts",
		Help:      "the total number of transactions per block",
		Buckets:   prometheus.ExponentialBuckets(4, 2, 10),
	})

	blockCollectionCounts := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "block_collection_counts",
		Help:      "the total number of collections per block",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 8),
	})

	collectionExecutionTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "collection_execution_time_milliseconds",
		Help:      "the total time spent on collection execution in milliseconds",
		Buckets:   []float64{100, 200, 500, 1000, 1500, 2000},
	})

	collectionComputationUsed := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "collection_computation_used",
		Help:      "the total amount of computation used by a collection",
		Buckets:   []float64{1000, 10000, 50000, 100000, 500000, 1000000},
	})

	collectionTransactionCounts := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "collection_transaction_counts",
		Help:      "the total number of transactions per collection",
		Buckets:   prometheus.ExponentialBuckets(4, 2, 8),
	})

	collectionRequestsSent := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemIngestion,
		Name:      "collection_requests_sent",
		Help:      "the number of collection requests sent",
	})

	collectionRequestsRetries := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemIngestion,
		Name:      "collection_requests_retries",
		Help:      "the number of collection requests retried",
	})

	transactionParseTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_parse_time_nanoseconds",
		Help:      "the parse time for a transaction in nanoseconds",
		Buckets:   prometheus.ExponentialBuckets(10, 10, 8),
	})

	transactionCheckTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_check_time_nanoseconds",
		Help:      "the checking time for a transaction in nanoseconds",
		Buckets:   prometheus.ExponentialBuckets(10, 10, 8),
	})

	transactionInterpretTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_interpret_time_nanoseconds",
		Help:      "the interpretation time for a transaction in nanoseconds",
		Buckets:   prometheus.ExponentialBuckets(10, 10, 8),
	})

	transactionExecutionTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_execution_time_milliseconds",
		Help:      "the total time spent on transaction execution in milliseconds",
		Buckets:   []float64{2, 4, 8, 16, 32, 64, 100, 250, 500},
	})

	transactionComputationUsed := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_computation_used",
		Help:      "the total amount of computation used by a transaction",
		Buckets:   []float64{50, 100, 500, 1000, 5000, 10000},
	})

	transactionMemoryUsage := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_memory_usage",
		Help:      "the total amount of memory allocated by a transaction",
		Buckets:   []float64{100_000, 1_000_000, 10_000_000, 50_000_000, 100_000_000, 500_000_000, 1_000_000_000},
	})

	transactionMemoryEstimate := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_memory_estimate",
		Help:      "the estimated memory used by a transaction",
		Buckets:   []float64{1_000_000, 10_000_000, 100_000_000, 1_000_000_000, 5_000_000_000, 10_000_000_000, 50_000_000_000, 100_000_000_000},
	})

	transactionMemoryDifference := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_memory_difference",
		Help:      "the difference in actual memory usage and estimate for a transaction",
		Buckets:   []float64{-1, 0, 10_000_000, 100_000_000, 1_000_000_000},
	})

	transactionEmittedEvents := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "transaction_emitted_events",
		Help:      "the total number of events emitted by a transaction",
	})

	scriptExecutionTime := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "script_execution_time_milliseconds",
		Help:      "the total time spent on script execution in milliseconds",
		Buckets:   []float64{2, 4, 8, 16, 32, 64, 100, 250, 500},
	})

	scriptComputationUsed := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "script_computation_used",
		Help:      "the total amount of computation used by an script",
		Buckets:   []float64{50, 100, 500, 1000, 5000, 10000},
	})

	scriptMemoryUsage := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "script_memory_usage",
		Help:      "the total amount of memory allocated by a script",
		Buckets:   []float64{100_000, 1_000_000, 10_000_000, 50_000_000, 100_000_000, 500_000_000, 1_000_000_000},
	})

	scriptMemoryEstimate := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "script_memory_estimate",
		Help:      "the estimated memory used by a script",
		Buckets:   []float64{1_000_000, 10_000_000, 100_000_000, 1_000_000_000, 5_000_000_000, 10_000_000_000, 50_000_000_000, 100_000_000_000},
	})

	scriptMemoryDifference := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemRuntime,
		Name:      "script_memory_difference",
		Help:      "the difference in actual memory usage and estimate for a script",
		Buckets:   []float64{-1, 0, 10_000_000, 100_000_000, 1_000_000_000},
	})

	chunkDataPackRequestProcessedTotal := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemProvider,
		Name:      "chunk_data_packs_requested_total",
		Help:      "the total number of chunk data pack requests processed by provider engine",
	})

	blockDataUploadsInProgress := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemBlockDataUploader,
		Name:      "block_data_upload_in_progress",
		Help:      "number of concurrently running Block Data upload operations",
	})

	blockDataUploadsDuration := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemBlockDataUploader,
		Name:      "block_data_upload_duration_ms",
		Help:      "the duration of update upload operation",
		Buckets:   []float64{1, 100, 500, 1000, 2000},
	})

	computationResultUploadedCount := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemProvider,
		Name:      "computation_result_uploaded_count",
		Help:      "the total count of computation result uploaded",
	})

	computationResultUploadRetriedCount := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespaceExecution,
		Subsystem: subsystemProvider,
		Name:      "computation_result_upload_retried_count",
		Help:      "the total count of computation result upload retried",
	})

	ec := &ExecutionCollector{
		tracer: tracer,

		forestApproxMemorySize:              forestApproxMemorySize,
		forestNumberOfTrees:                 forestNumberOfTrees,
		latestTrieRegCount:                  latestTrieRegCount,
		latestTrieRegCountDiff:              latestTrieRegCountDiff,
		latestTrieRegSize:                   latestTrieRegSize,
		latestTrieRegSizeDiff:               latestTrieRegSizeDiff,
		latestTrieMaxDepthTouched:           latestTrieMaxDepthTouched,
		updated:                             updatedCount,
		proofSize:                           proofSize,
		updatedValuesNumber:                 updatedValuesNumber,
		updatedValuesSize:                   updatedValuesSize,
		updatedDuration:                     updatedDuration,
		updatedDurationPerValue:             updatedDurationPerValue,
		readValuesNumber:                    readValuesNumber,
		readValuesSize:                      readValuesSize,
		readDuration:                        readDuration,
		readDurationPerValue:                readDurationPerValue,
		blockExecutionTime:                  blockExecutionTime,
		blockComputationUsed:                blockComputationUsed,
		blockTransactionCounts:              blockTransactionCounts,
		blockCollectionCounts:               blockCollectionCounts,
		collectionExecutionTime:             collectionExecutionTime,
		collectionComputationUsed:           collectionComputationUsed,
		collectionTransactionCounts:         collectionTransactionCounts,
		collectionRequestSent:               collectionRequestsSent,
		collectionRequestRetried:            collectionRequestsRetries,
		transactionParseTime:                transactionParseTime,
		transactionCheckTime:                transactionCheckTime,
		transactionInterpretTime:            transactionInterpretTime,
		transactionExecutionTime:            transactionExecutionTime,
		transactionComputationUsed:          transactionComputationUsed,
		transactionMemoryUsage:              transactionMemoryUsage,
		transactionMemoryEstimate:           transactionMemoryEstimate,
		transactionMemoryDifference:         transactionMemoryDifference,
		transactionEmittedEvents:            transactionEmittedEvents,
		scriptExecutionTime:                 scriptExecutionTime,
		scriptComputationUsed:               scriptComputationUsed,
		scriptMemoryUsage:                   scriptMemoryUsage,
		scriptMemoryEstimate:                scriptMemoryEstimate,
		scriptMemoryDifference:              scriptMemoryDifference,
		chunkDataPackRequestProcessedTotal:  chunkDataPackRequestProcessedTotal,
		blockDataUploadsInProgress:          blockDataUploadsInProgress,
		blockDataUploadsDuration:            blockDataUploadsDuration,
		computationResultUploadedCount:      computationResultUploadedCount,
		computationResultUploadRetriedCount: computationResultUploadRetriedCount,

		stateReadsPerBlock: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Buckets:   []float64{5, 10, 50, 100, 500},
			Name:      "block_state_reads",
			Help:      "count of state access/read operations performed per block",
		}),

		totalExecutedBlocksCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "total_executed_blocks",
			Help:      "the total number of blocks that have been executed",
		}),

		totalExecutedCollectionsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "total_executed_collections",
			Help:      "the total number of collections that have been executed",
		}),

		totalExecutedTransactionsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "total_executed_transactions",
			Help:      "the total number of transactions that have been executed",
		}),

		totalFailedTransactionsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "total_failed_transactions",
			Help:      "the total number of transactions that has failed when executed",
		}),

		totalExecutedScriptsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "total_executed_scripts",
			Help:      "the total number of scripts that have been executed",
		}),

		lastExecutedBlockHeightGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "last_executed_block_height",
			Help:      "the last height that was executed",
		}),

		stateStorageDiskTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemStateStorage,
			Name:      "data_size_bytes",
			Help:      "the execution state size on disk in bytes",
		}),

		storageStateCommitment: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemStateStorage,
			Name:      "commitment_size_bytes",
			Help:      "the storage size of a state commitment in bytes",
		}),

		stateSyncActive: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemIngestion,
			Name:      "state_sync_active",
			Help:      "indicates if the state sync is active",
		}),

		numberOfAccounts: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceExecution,
			Subsystem: subsystemRuntime,
			Name:      "number_of_accounts",
			Help:      "the number of existing accounts on the network",
		}),

		maxCollectionHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "max_collection_height",
			Namespace: namespaceExecution,
			Subsystem: subsystemIngestion,
			Help:      "gauge to track the maximum block height of collections received",
		}),
	}

	return ec
}

// StartBlockReceivedToExecuted starts a span to trace the duration of a block
// from being received for execution to execution being finished
func (ec *ExecutionCollector) StartBlockReceivedToExecuted(blockID flow.Identifier) {
}

// FinishBlockReceivedToExecuted finishes a span to trace the duration of a block
// from being received for execution to execution being finished
func (ec *ExecutionCollector) FinishBlockReceivedToExecuted(blockID flow.Identifier) {
}

// ExecutionBlockExecuted reports computation and total time spent on a block computation
func (ec *ExecutionCollector) ExecutionBlockExecuted(dur time.Duration, compUsed uint64, txCounts int, colCounts int) {
	ec.totalExecutedBlocksCounter.Inc()
	ec.blockExecutionTime.Observe(float64(dur.Milliseconds()))
	ec.blockComputationUsed.Observe(float64(compUsed))
	ec.blockTransactionCounts.Observe(float64(txCounts))
	ec.blockCollectionCounts.Observe(float64(colCounts))
}

// ExecutionCollectionExecuted reports computation and total time spent on a block computation
func (ec *ExecutionCollector) ExecutionCollectionExecuted(dur time.Duration, compUsed uint64, txCounts int) {
	ec.totalExecutedCollectionsCounter.Inc()
	ec.collectionExecutionTime.Observe(float64(dur.Milliseconds()))
	ec.collectionComputationUsed.Observe(float64(compUsed))
	ec.collectionTransactionCounts.Observe(float64(txCounts))
}

// TransactionExecuted reports the time and computation spent executing a single transaction
func (ec *ExecutionCollector) ExecutionTransactionExecuted(
	dur time.Duration,
	compUsed, memoryUsed, memoryEstimated uint64,
	eventCounts int,
	failed bool,
) {
	ec.totalExecutedTransactionsCounter.Inc()
	ec.transactionExecutionTime.Observe(float64(dur.Milliseconds()))
	ec.transactionComputationUsed.Observe(float64(compUsed))
	ec.transactionMemoryUsage.Observe(float64(memoryUsed))
	ec.transactionMemoryEstimate.Observe(float64(memoryEstimated))
	ec.transactionMemoryDifference.Observe(float64(memoryEstimated) - float64(memoryUsed))
	ec.transactionEmittedEvents.Observe(float64(eventCounts))
	if failed {
		ec.totalFailedTransactionsCounter.Inc()
	}
}

// ScriptExecuted reports the time spent executing a single script
func (ec *ExecutionCollector) ExecutionScriptExecuted(dur time.Duration, compUsed, memoryUsed, memoryEstimated uint64) {
	ec.totalExecutedScriptsCounter.Inc()
	ec.scriptExecutionTime.Observe(float64(dur.Milliseconds()))
	ec.scriptComputationUsed.Observe(float64(compUsed))
	ec.scriptMemoryUsage.Observe(float64(memoryUsed))
	ec.scriptMemoryEstimate.Observe(float64(memoryEstimated))
	ec.scriptMemoryDifference.Observe(float64(memoryEstimated) - float64(memoryUsed))
}

// ExecutionStateReadsPerBlock reports number of state access/read operations per block
func (ec *ExecutionCollector) ExecutionStateReadsPerBlock(reads uint64) {
	ec.stateReadsPerBlock.Observe(float64(reads))
}

// ExecutionStateStorageDiskTotal reports the total storage size of the execution state on disk in bytes
func (ec *ExecutionCollector) ExecutionStateStorageDiskTotal(bytes int64) {
	ec.stateStorageDiskTotal.Set(float64(bytes))
}

// ExecutionStorageStateCommitment reports the storage size of a state commitment
func (ec *ExecutionCollector) ExecutionStorageStateCommitment(bytes int64) {
	ec.storageStateCommitment.Set(float64(bytes))
}

// ExecutionLastExecutedBlockHeight reports last executed block height
func (ec *ExecutionCollector) ExecutionLastExecutedBlockHeight(height uint64) {
	ec.lastExecutedBlockHeightGauge.Set(float64(height))
}

// ForestApproxMemorySize records approximate memory usage of forest (all in-memory trees)
func (ec *ExecutionCollector) ForestApproxMemorySize(bytes uint64) {
	ec.forestApproxMemorySize.Set(float64(bytes))
}

// ForestNumberOfTrees current number of trees in a forest (in memory)
func (ec *ExecutionCollector) ForestNumberOfTrees(number uint64) {
	ec.forestNumberOfTrees.Set(float64(number))
}

// LatestTrieRegCount records the number of unique register allocated (the lastest created trie)
func (ec *ExecutionCollector) LatestTrieRegCount(number uint64) {
	ec.latestTrieRegCount.Set(float64(number))
}

// LatestTrieRegCountDiff records the difference between the number of unique register allocated of the latest created trie and parent trie
func (ec *ExecutionCollector) LatestTrieRegCountDiff(number int64) {
	ec.latestTrieRegCountDiff.Set(float64(number))
}

// LatestTrieRegSize records the size of unique register allocated (the lastest created trie)
func (ec *ExecutionCollector) LatestTrieRegSize(size uint64) {
	ec.latestTrieRegSize.Set(float64(size))
}

// LatestTrieRegSizeDiff records the difference between the size of unique register allocated of the latest created trie and parent trie
func (ec *ExecutionCollector) LatestTrieRegSizeDiff(size int64) {
	ec.latestTrieRegSizeDiff.Set(float64(size))
}

// LatestTrieMaxDepthTouched records the maximum depth touched of the last created trie
func (ec *ExecutionCollector) LatestTrieMaxDepthTouched(maxDepth uint16) {
	ec.latestTrieMaxDepthTouched.Set(float64(maxDepth))
}

// UpdateCount increase a counter of performed updates
func (ec *ExecutionCollector) UpdateCount() {
	ec.updated.Inc()
}

// ProofSize records a proof size
func (ec *ExecutionCollector) ProofSize(bytes uint32) {
	ec.proofSize.Set(float64(bytes))
}

// UpdateValuesNumber accumulates number of updated values
func (ec *ExecutionCollector) UpdateValuesNumber(number uint64) {
	ec.updatedValuesNumber.Add(float64(number))
}

// UpdateValuesSize total size (in bytes) of updates values
func (ec *ExecutionCollector) UpdateValuesSize(bytes uint64) {
	ec.updatedValuesSize.Set(float64(bytes))
}

// UpdateDuration records absolute time for the update of a trie
func (ec *ExecutionCollector) UpdateDuration(duration time.Duration) {
	ec.updatedDuration.Observe(duration.Seconds())
}

// UpdateDurationPerItem records update time for single value (total duration / number of updated values)
func (ec *ExecutionCollector) UpdateDurationPerItem(duration time.Duration) {
	ec.updatedDurationPerValue.Observe(duration.Seconds())
}

// ReadValuesNumber accumulates number of read values
func (ec *ExecutionCollector) ReadValuesNumber(number uint64) {
	ec.readValuesNumber.Add(float64(number))
}

// ReadValuesSize total size (in bytes) of read values
func (ec *ExecutionCollector) ReadValuesSize(bytes uint64) {
	ec.readValuesSize.Set(float64(bytes))
}

// ReadDuration records absolute time for the read from a trie
func (ec *ExecutionCollector) ReadDuration(duration time.Duration) {
	ec.readDuration.Observe(duration.Seconds())
}

// ReadDurationPerItem records read time for single value (total duration / number of read values)
func (ec *ExecutionCollector) ReadDurationPerItem(duration time.Duration) {
	ec.readDurationPerValue.Observe(duration.Seconds())
}

func (ec *ExecutionCollector) ExecutionCollectionRequestSent() {
	ec.collectionRequestSent.Inc()
}

func (ec *ExecutionCollector) ExecutionCollectionRequestRetried() {
	ec.collectionRequestRetried.Inc()
}

func (ec *ExecutionCollector) ExecutionBlockDataUploadStarted() {
	ec.blockDataUploadsInProgress.Inc()
}

func (ec *ExecutionCollector) ExecutionBlockDataUploadFinished(dur time.Duration) {
	ec.blockDataUploadsInProgress.Dec()
	ec.blockDataUploadsDuration.Observe(float64(dur.Milliseconds()))
}

// TransactionParsed reports the time spent parsing a single transaction
func (ec *ExecutionCollector) RuntimeTransactionParsed(dur time.Duration) {
	ec.transactionParseTime.Observe(float64(dur))
}

// TransactionChecked reports the time spent checking a single transaction
func (ec *ExecutionCollector) RuntimeTransactionChecked(dur time.Duration) {
	ec.transactionCheckTime.Observe(float64(dur))
}

// TransactionInterpreted reports the time spent interpreting a single transaction
func (ec *ExecutionCollector) RuntimeTransactionInterpreted(dur time.Duration) {
	ec.transactionInterpretTime.Observe(float64(dur))
}

// ChunkDataPackRequestProcessed is executed every time a chunk data pack request is picked up for processing at execution node.
// It increases the request processed counter by one.
func (ec *ExecutionCollector) ChunkDataPackRequestProcessed() {
	ec.chunkDataPackRequestProcessedTotal.Inc()
}

func (ec *ExecutionCollector) ExecutionSync(syncing bool) {
	if syncing {
		ec.stateSyncActive.Set(float64(1))
		return
	}
	ec.stateSyncActive.Set(float64(0))
}

func (ec *ExecutionCollector) RuntimeSetNumberOfAccounts(count uint64) {
	ec.numberOfAccounts.Set(float64(count))
}

func (ec *ExecutionCollector) UpdateCollectionMaxHeight(height uint64) {
	ec.maxCollectionHeight.Set(float64(height))
}

func (ec *ExecutionCollector) ExecutionComputationResultUploaded() {
	ec.computationResultUploadedCount.Inc()
}

func (ec *ExecutionCollector) ExecutionComputationResultUploadRetried() {
	ec.computationResultUploadRetriedCount.Inc()
}
