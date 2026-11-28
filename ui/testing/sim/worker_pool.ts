// A worker pool with no workers: the stats a recompute asks for never arrive, so a test's store
// holds only what the test writes. Reached lazily, as `vi.mock` is hoisted above the imports.
export const idleWorkerPool = () => ({
	generateRequestId: () => 'test-request',
	WorkerPool: class {
		setNumWorkers() {}
		getNumWorkers = () => 1;
		isWasm = () => Promise.resolve(false);
		computeStats = () => new Promise(() => {});
	},
});
