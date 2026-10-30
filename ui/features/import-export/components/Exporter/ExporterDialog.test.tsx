import i18n from '@i18n/config';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

// The exporters reach the addon importer, which asks GitHub for its version when its module loads.
// Answered offline here, in `vi.hoisted` so it runs before the imports, no request is left open
// when the window closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

import { renderWithHost } from '@testing/react';

import { EXPORT_CATEGORY_OPTIONS } from '../../exporters';
import { exporterDialog } from './ExporterDialog';

const DOWNLOAD = i18n.t('export.json.download_button');

const getData = () => '{"player":{}}';
const TestExporterDialog = exporterDialog({ title: 'Test Export', getData });
const FullExporterDialog = exporterDialog({ title: 'Full Export', allowDownload: true, selectCategories: true, getData });

describe('exporterDialog', () => {
	it('opens the exporter under the definition title with its data', () => {
		renderWithHost(<TestExporterDialog open onOpenChange={() => {}} />);
		expect(screen.getByTestId('sim-dialog-title').textContent).toBe('Test Export');
		expect(screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value).toBe('{"player":{}}');
	});

	it('leaves the category pickers and the download button out unless the definition asks', () => {
		renderWithHost(<TestExporterDialog open onOpenChange={() => {}} />);
		expect(screen.queryAllByRole('checkbox')).toHaveLength(0);
		expect(screen.queryByRole('button', { name: DOWNLOAD })).toBeNull();
	});

	it('shows the category pickers and the download button when the definition asks', () => {
		renderWithHost(<FullExporterDialog open onOpenChange={() => {}} />);
		expect(screen.getAllByRole('checkbox')).toHaveLength(EXPORT_CATEGORY_OPTIONS.length);
		expect(screen.getByRole('button', { name: DOWNLOAD })).toBeTruthy();
	});

	it('copies the export text from its copy button', () => {
		const writeText = vi.fn(() => Promise.resolve());
		const own = Object.getOwnPropertyDescriptor(navigator, 'clipboard');
		Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
		vi.useFakeTimers();
		try {
			renderWithHost(<TestExporterDialog open onOpenChange={() => {}} />);
			fireEvent.click(screen.getByTestId('copy-button'));
			expect(writeText).toHaveBeenCalledWith('{"player":{}}');
		} finally {
			vi.useRealTimers();
			if (own) Object.defineProperty(navigator, 'clipboard', own);
			else Reflect.deleteProperty(navigator, 'clipboard');
		}
	});
});

describe('an exporter whose text comes later', () => {
	it('offers Copy and Download only once its promise resolves', async () => {
		let resolve: (text: string) => void = () => {};
		const AsyncExporterDialog = exporterDialog({ title: 'Async Export', allowDownload: true, getData: () => new Promise<string>(r => (resolve = r)) });
		renderWithHost(<AsyncExporterDialog open onOpenChange={() => {}} />);
		const copy = screen.getByTestId<HTMLButtonElement>('copy-button');
		const download = screen.getByRole<HTMLButtonElement>('button', { name: DOWNLOAD });
		expect([copy.disabled, download.disabled]).toEqual([true, true]);

		resolve('siml 1\n');
		await waitFor(() => expect([copy.disabled, download.disabled]).toEqual([false, false]));
	});

	it('shows the text once its promise resolves', async () => {
		const AsyncExporterDialog = exporterDialog({ title: 'Async Export', getData: () => Promise.resolve('siml 1\n') });
		renderWithHost(<AsyncExporterDialog open onOpenChange={() => {}} />);
		await waitFor(() => expect(screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value).toBe('siml 1\n'));
	});

	it('reports a rejected promise as an error toast', async () => {
		toasts.length = 0;
		const FailingExporterDialog = exporterDialog({ title: 'Failing Export', getData: () => Promise.reject(new Error('the chunk did not load')) });
		renderWithHost(<FailingExporterDialog open onOpenChange={() => {}} />);
		await waitFor(() => expect(toasts).toEqual([{ variant: 'error', body: i18n.t('export.error', { message: 'the chunk did not load' }) }]));
		expect(screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value).toBe('');
		expect(screen.getByTestId<HTMLButtonElement>('copy-button').disabled).toBe(true);
	});
});
