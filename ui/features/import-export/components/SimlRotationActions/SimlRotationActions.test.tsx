import { APLRotation } from '@generated/proto/apl';
import { Spec } from '@generated/proto/common';
import { SimlPrintResult } from '@generated/proto/ui';
import { renderWithHost } from '@testing/react';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// The exporter dialog reaches the addon importer, which asks GitHub for its version when its module loads.
// Answered offline here, in `vi.hoisted` so it runs before the imports, no request is left open when the window
// closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

const printSiml = vi.hoisted(() => vi.fn());
vi.mock('../../siml_client', async importOriginal => ({ ...(await importOriginal<typeof import('../../siml_client')>()), printSiml }));

import { SIML_ROTATION_EXPORTER } from '../../exporters';
import { SIML_ROTATION_IMPORTER } from '../../importers';
import { SimlRotationActions } from './SimlRotationActions';

const rotation = APLRotation.fromJson({ type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }] });
const player = { getSpec: () => Spec.SpecHunter, getResolvedAplRotation: () => rotation };

const renderActions = (waitForInit: () => Promise<void>) => renderWithHost(<SimlRotationActions />, { player, sim: { waitForInit } });

describe('SimlRotationActions', () => {
	it('keeps both buttons disabled until the sim is ready', () => {
		renderActions(() => new Promise(() => {}));
		expect(screen.getByTestId('siml-rotation-export').hasAttribute('disabled')).toBe(true);
		expect(screen.getByTestId('siml-rotation-import').hasAttribute('disabled')).toBe(true);
	});

	it('exports the rotation as SIML', async () => {
		renderActions(() => Promise.resolve());
		const exportButton = screen.getByTestId('siml-rotation-export');
		await waitFor(() => expect(exportButton.hasAttribute('disabled')).toBe(false));
		printSiml.mockResolvedValue(SimlPrintResult.create({ text: 'siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n' }));
		fireEvent.click(exportButton);
		expect(screen.getByTestId('sim-dialog-title').textContent).toBe(SIML_ROTATION_EXPORTER.title);
		await waitFor(() =>
			expect(screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value).toBe('siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n'),
		);
	});

	it('opens the rotation importer', async () => {
		renderActions(() => Promise.resolve());
		const importButton = screen.getByTestId('siml-rotation-import');
		await waitFor(() => expect(importButton.hasAttribute('disabled')).toBe(false));
		fireEvent.click(importButton);
		expect(screen.getByTestId('sim-dialog-title').textContent).toBe(SIML_ROTATION_IMPORTER.title);
	});
});
