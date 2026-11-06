import { afterEach, describe, expect, it, vi } from 'vitest';

import type { WorkerReceiveMessage, WorkerSendMessage } from './types';
import { WorkerInterface } from './worker_interface';

// The worker's global message listener and postMessage, captured so a request can be sent in and its
// answer read back.
const setupWorkerGlobals = () => {
	let listener: ((event: { data: WorkerReceiveMessage }) => void) | undefined;
	const posted: WorkerSendMessage[] = [];
	vi.stubGlobal('addEventListener', (_type: string, l: typeof listener) => (listener = l));
	vi.stubGlobal('postMessage', (message: WorkerSendMessage) => posted.push(message));
	return { send: (message: WorkerReceiveMessage) => listener?.({ data: message }), posted };
};

const flush = () => new Promise(resolve => setTimeout(resolve, 0));

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('WorkerInterface', () => {
	it('reports a handler that answers with no result as a stopped runtime', async () => {
		const { send, posted } = setupWorkerGlobals();
		vi.spyOn(console, 'error').mockImplementation(() => {});
		new WorkerInterface({ computeStats: () => undefined as unknown as Uint8Array } as unknown as ConstructorParameters<typeof WorkerInterface>[0]);

		send({ msg: 'computeStats', id: 'computeStats-1', inputData: new Uint8Array() } as WorkerReceiveMessage);
		await flush();

		expect(posted).toHaveLength(1);
		expect(posted[0]).toMatchObject({ msg: 'computeStats', id: 'computeStats-1', stopped: true });
		expect(posted[0].error).toContain('no result');
	});

	it('reports a handler that throws because the Go runtime has exited as a stopped runtime', async () => {
		const { send, posted } = setupWorkerGlobals();
		vi.spyOn(console, 'error').mockImplementation(() => {});
		const handlers = {
			computeStats: () => {
				throw new Error('Go program has already exited');
			},
		} as unknown as ConstructorParameters<typeof WorkerInterface>[0];
		new WorkerInterface(handlers);

		send({ msg: 'computeStats', id: 'computeStats-3', inputData: new Uint8Array() } as WorkerReceiveMessage);
		await flush();

		expect(posted[0]).toMatchObject({ id: 'computeStats-3', error: 'Go program has already exited', stopped: true });
	});

	it('answers a handler result as it is', async () => {
		const { send, posted } = setupWorkerGlobals();
		new WorkerInterface({ computeStats: () => new Uint8Array([7]) } as unknown as ConstructorParameters<typeof WorkerInterface>[0]);

		send({ msg: 'computeStats', id: 'computeStats-2', inputData: new Uint8Array() } as WorkerReceiveMessage);
		await flush();

		expect(posted).toEqual([{ msg: 'computeStats', id: 'computeStats-2', outputData: new Uint8Array([7]) }]);
	});
});
