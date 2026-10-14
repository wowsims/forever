// The addon importer's one piece of behaviour that is pure enough to pin here is the version probe
// it fires at module load; the rest of `onImport` is proto plumbing over a live `Database`.
import { describe, expect, it, vi } from 'vitest';

// `getWSEVersion()` runs at module evaluation, so importing this module fires a request. Stubbed
// in `vi.hoisted`, which runs before the imports.
const fetchStub = vi.hoisted(() => {
	const stub = vi.fn((_url: string) => Promise.reject(new Error('offline')));
	(globalThis as unknown as { fetch: unknown }).fetch = stub;
	return stub;
});

import { ADDON_IMPORTER } from './addon';

describe('ADDON_IMPORTER', () => {
	// The failure is swallowed by design: no GitHub, no version warning, and the import still runs.
	it('asks GitHub for the addon version once, at module load', () => {
		expect(fetchStub).toHaveBeenCalledTimes(1);
		expect(fetchStub.mock.calls[0][0]).toBe('https://api.github.com/repos/wowsims/exporter/releases/latest');
	});

	it('rejects a payload that is not an addon export', async () => {
		await expect(ADDON_IMPORTER.onImport({} as never, 'not json')).rejects.toThrow('import.addon.error_invalid_export');
	});
});
