import { afterEach, describe, expect, it, vi } from 'vitest';

import type { WorkerReceiveMessage, WorkerSendMessage } from './types';
import { type Handlers, WorkerInterface } from './worker_interface';

// Starts a WorkerInterface whose computeStats is `computeStats`, sends it one request and gives back what it posts.
const answerTo = async (computeStats: () => unknown, runtimeStopped?: () => boolean) => {
	let listener: ((event: { data: WorkerReceiveMessage }) => Promise<void>) | undefined;
	const posted: WorkerSendMessage[] = [];
	vi.stubGlobal('addEventListener', (_type: string, l: typeof listener) => (listener = l));
	vi.stubGlobal('postMessage', (message: WorkerSendMessage) => posted.push(message));
	new WorkerInterface({ computeStats } as unknown as Handlers, runtimeStopped);
	await listener?.({ data: { msg: 'computeStats', id: 'computeStats-1', inputData: new Uint8Array() } as WorkerReceiveMessage });
	return posted;
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('WorkerInterface', () => {
	it('answers a handler result as it is', async () => {
		expect(await answerTo(() => new Uint8Array([7]))).toEqual([{ msg: 'computeStats', id: 'computeStats-1', outputData: new Uint8Array([7]) }]);
	});

	it('answers an async handler that sends its result as progress as it is, on a runtime that runs', async () => {
		expect(
			await answerTo(
				() => undefined,
				() => false,
			),
		).toEqual([{ msg: 'computeStats', id: 'computeStats-1', outputData: undefined }]);
	});

	it('reports a call during which the runtime stopped as an error from a stopped runtime', async () => {
		vi.spyOn(console, 'error').mockImplementation(() => {});
		const [answer] = await answerTo(
			() => undefined,
			() => true,
		);
		expect(answer).toMatchObject({ id: 'computeStats-1', stopped: true });
		expect(answer.error).toContain('stopped');
	});

	it('reports a throw from a stopped runtime as an error from a stopped runtime', async () => {
		vi.spyOn(console, 'error').mockImplementation(() => {});
		const thrown = () => {
			throw new Error('Go program has already exited');
		};
		expect(await answerTo(thrown, () => true)).toMatchObject([{ id: 'computeStats-1', error: 'Go program has already exited', stopped: true }]);
	});

	it('reports a throw from a runtime that runs as an error only', async () => {
		vi.spyOn(console, 'error').mockImplementation(() => {});
		const thrown = () => {
			throw new Error('bad request');
		};
		expect(await answerTo(thrown, () => false)).toMatchObject([{ id: 'computeStats-1', error: 'bad request', stopped: false }]);
	});
});
