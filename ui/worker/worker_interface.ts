import type { SimRequest, WorkerReceiveMessage, WorkerSendMessage } from './types';

export type HandlerProgressCallback = (outputData: Uint8Array) => void;
export type HandlerFunction = (data: Uint8Array, progress: HandlerProgressCallback, id: string, msg: SimRequest) => Uint8Array | Promise<Uint8Array>;
export type Handlers = Record<SimRequest, HandlerFunction>;

/**
 * Communication with the UI.
 */
export class WorkerInterface {
	private _workerId = '';
	private readonly handlers: Handlers;

	constructor(handlers: Handlers) {
		this.handlers = handlers;

		addEventListener('message', async ({ data }: MessageEvent<WorkerReceiveMessage>) => {
			// Handled by the wasm boot listener in sim_worker.ts, not by the request API.
			if (data.msg === 'wasmModule') {
				return;
			}
			const { id, msg, inputData } = data;

			if (msg === 'setID') {
				this._workerId = id;
				this.postMessage({ msg: 'idConfirm' });
				return;
			}

			const handlerFunc = this.handlers?.[msg];

			if (!handlerFunc) {
				console.error(`Request msg: ${msg}, id: ${id}, is not handled!`);
				this.postMessage({ msg, id, outputData: new Uint8Array() });
				return;
			}

			const progressCallback: HandlerProgressCallback = prog => {
				this.postMessage({
					msg: 'progress',
					id: `${id}progress`,
					outputData: prog,
				});
			};

			try {
				const outputData = await handlerFunc(inputData, progressCallback, id, msg);
				// A Go runtime that has exited, after a panic no handler recovers, answers every call with
				// undefined: this request is lost and so is every later one on this worker.
				if (!(outputData instanceof Uint8Array)) {
					const error = `The sim worker answered ${msg} with no result; its sim runtime has stopped.`;
					console.error(error);
					this.postMessage({ msg, id, outputData: new Uint8Array(), error, stopped: true });
					return;
				}
				this.postMessage({ msg, id, outputData });
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : String(error);
				console.error(errorMessage);
				// wasm_exec.js throws this for every call after the Go runtime has exited.
				const stopped = errorMessage.includes('Go program has already exited');
				this.postMessage({ msg, id, outputData: new Uint8Array(), error: errorMessage, stopped });
			}
		});
	}

	private postMessage(m: WorkerSendMessage) {
		postMessage(m);
	}

	get workerId() {
		return this._workerId;
	}

	/**
	 * Tell UI that the worker is ready.
	 * @param isWasm true if worker is using wasm.
	 */
	ready(isWasm: boolean) {
		this.postMessage({ msg: 'ready', outputData: new Uint8Array([+isWasm]) });
	}
}
