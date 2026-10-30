// What this pins is the shell's behaviour, which is everything the four importers share: what gets
// handed to `onImport`, and what happens to the dialog when it resolves and when it rejects.
import { renderWithHost } from '@testing/react';
import { act, fireEvent } from '@testing-library/react';
import { useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ImportProblems } from '../../importers/types';
import { Importer, type ImporterProps } from './Importer';

const trackPageView = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackPageView }));

// What matters here is only that a rejected import reports a toast, and with the message the
// thrown error carried.
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

const HOST_PARTS = { rootElem: document.createElement('div') };
const theHost = expect.objectContaining(HOST_PARTS);
const aSignal = expect.any(AbortSignal);

const renderImporter = (props: Partial<Parameters<typeof Importer>[0]> = {}) => {
	const onImport = props.onImport ?? vi.fn().mockResolvedValue(undefined);
	const onOpenChange = props.onOpenChange ?? vi.fn();
	renderWithHost(
		<Importer open title="JSON Import" allowFileUpload {...props} onImport={onImport} onOpenChange={onOpenChange}>
			<p>how to</p>
		</Importer>,
		HOST_PARTS,
	);
	return { onImport, onOpenChange };
};

// The importer as the menus hold it, mounted while closed; what it gives opens the dialog again.
const renderReopenable = (onImport: ImporterProps['onImport']) => {
	let setOpen: (open: boolean) => void = () => {};
	const Menu = () => {
		const [open, set] = useState(true);
		setOpen = set;
		return (
			<Importer open={open} onOpenChange={set} title="SIML Import" allowFileUpload onImport={onImport}>
				<p>how to</p>
			</Importer>
		);
	};
	renderWithHost(<Menu />, HOST_PARTS);
	return () => act(async () => setOpen(true));
};

const closeDialog = () =>
	act(async () => {
		fireEvent.click(document.querySelector('[data-testid="sim-dialog-close"]')!);
	});

const upload = async (name: string, text: string) => {
	const input = document.querySelector<HTMLInputElement>('[data-testid="importer-upload-input"]')!;
	await act(async () => {
		fireEvent.change(input, { target: { files: [{ name, text: () => Promise.resolve(text) }] } });
	});
};

const clickImport = async () => {
	await act(async () => {
		fireEvent.click(document.querySelector('[data-testid="import-button"]')!);
	});
};

describe('Importer', () => {
	beforeEach(() => {
		toasts.length = 0;
		trackPageView.mockClear();
	});

	it('hands the textarea contents and the host to onImport, with no file name for typed text, then closes', async () => {
		const { onImport, onOpenChange } = renderImporter();
		const textarea = document.querySelector('textarea')!;
		fireEvent.input(textarea, { target: { value: '{"player":{}}' } });

		await clickImport();

		expect(onImport).toHaveBeenCalledWith(theHost, '{"player":{}}', undefined, aSignal);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('reports a rejected import as an error toast and leaves the dialog open', async () => {
		const { onOpenChange } = renderImporter({ onImport: vi.fn().mockRejectedValue(new Error('Could not parse Class!')) });

		await clickImport();

		expect(toasts).toEqual([{ variant: 'error', body: 'import.error Could not parse Class!' }]);
		expect(onOpenChange).not.toHaveBeenCalled();
	});

	// Problems point into the text, so they stay beside it rather than in a toast that goes away.
	it('lists the problems of a refused import under the text, until the next import', async () => {
		const onImport = vi.fn().mockRejectedValueOnce(new ImportProblems('The text does not read:', ['5:8 no spell is named x', '6:1 no y']));
		renderImporter({ onImport });

		await clickImport();
		const problems = document.querySelector('[data-testid="importer-problems"]')!;
		expect(problems.textContent).toBe('The text does not read:5:8 no spell is named x6:1 no y');
		expect(Array.from(problems.querySelectorAll('li'), li => li.textContent)).toEqual(['5:8 no spell is named x', '6:1 no y']);
		expect(toasts).toEqual([]);

		await clickImport();
		expect(document.querySelector('[data-testid="importer-problems"]')).toBeNull();
	});

	it('reports a page view under the importer title on open', () => {
		renderImporter();
		expect(trackPageView).toHaveBeenCalledWith('JSON Import', '/import/json-import');
	});

	it('does not report a page view while closed', () => {
		renderImporter({ open: false });
		expect(trackPageView).not.toHaveBeenCalled();
	});

	it('an upload replaces text the user has already typed', async () => {
		const { onImport } = renderImporter();
		const textarea = document.querySelector('textarea')!;
		fireEvent.input(textarea, { target: { value: 'typed by hand' } });

		await upload('setup.siml', 'from the file');

		await clickImport();
		expect(onImport).toHaveBeenCalledWith(theHost, 'from the file', 'setup.siml', aSignal);
	});

	it('imports an uploaded file with its line breaks as line feeds', async () => {
		const { onImport } = renderImporter();
		await upload('setup.siml', 'siml 1\r\nspec hunter\rrace orc');

		await clickImport();
		expect(onImport).toHaveBeenCalledWith(theHost, 'siml 1\nspec hunter\nrace orc', 'setup.siml', aSignal);
	});

	// A typo fixed in an uploaded p1.gear.siml still reads as a gear file.
	it('keeps the uploaded file name while the text is edited', async () => {
		const { onImport } = renderImporter();
		await upload('p1.gear.siml', 'siml 1\nhead x');
		const textarea = document.querySelector('textarea')!;
		fireEvent.input(textarea, { target: { value: 'siml 1\nhead y' } });
		textarea.setSelectionRange(7, 11);
		fireEvent.paste(textarea);

		await clickImport();
		expect(onImport).toHaveBeenCalledWith(theHost, 'siml 1\nhead y', 'p1.gear.siml', aSignal);
	});

	it('hands no file name on once the uploaded text is cleared', async () => {
		const { onImport } = renderImporter();
		await upload('rotation.apl.siml', 'from the file');
		const textarea = document.querySelector('textarea')!;
		fireEvent.input(textarea, { target: { value: '' } });
		fireEvent.input(textarea, { target: { value: 'typed' } });

		await clickImport();
		expect(onImport).toHaveBeenCalledWith(theHost, 'typed', undefined, aSignal);
	});

	it('hands no file name on for text pasted over the whole upload', async () => {
		const { onImport } = renderImporter();
		await upload('rotation.apl.siml', 'from the file');
		const textarea = document.querySelector('textarea')!;
		textarea.select();
		fireEvent.paste(textarea);
		fireEvent.input(textarea, { target: { value: 'pasted' } });

		await clickImport();
		expect(onImport).toHaveBeenCalledWith(theHost, 'pasted', undefined, aSignal);
	});

	it('drops the problems and the uploaded file name when the dialog closes', async () => {
		const onImport = vi.fn().mockRejectedValueOnce(new ImportProblems('The text does not read:', ['5:8 no spell is named x']));
		const reopen = renderReopenable(onImport);
		await upload('p1.gear.siml', 'siml 1');
		await clickImport();
		expect(document.querySelector('[data-testid="importer-problems"]')).not.toBeNull();

		await closeDialog();
		await reopen();
		expect(document.querySelector('[data-testid="importer-problems"]')).toBeNull();
		fireEvent.input(document.querySelector('textarea')!, { target: { value: 'typed' } });
		await clickImport();
		expect(onImport).toHaveBeenLastCalledWith(theHost, 'typed', undefined, aSignal);
	});

	it('drops the uploaded file name when an import closes the dialog', async () => {
		const onImport = vi.fn().mockResolvedValue(undefined);
		const reopen = renderReopenable(onImport);
		await upload('p1.gear.siml', 'siml 1');
		await clickImport();
		expect(onImport).toHaveBeenLastCalledWith(theHost, 'siml 1', 'p1.gear.siml', aSignal);

		await reopen();
		fireEvent.input(document.querySelector('textarea')!, { target: { value: 'typed' } });
		await clickImport();
		expect(onImport).toHaveBeenLastCalledWith(theHost, 'typed', undefined, aSignal);
	});

	it('takes no second click while an import has not ended, and takes one once it has', async () => {
		let end: () => void = () => {};
		const onImport = vi.fn(() => new Promise<void>(resolve => (end = resolve)));
		renderImporter({ onImport, onOpenChange: vi.fn() });
		const button = document.querySelector('[data-testid="import-button"]')!;

		await clickImport();
		expect(button.hasAttribute('disabled')).toBe(true);
		expect(button.getAttribute('aria-busy')).toBe('true');
		await clickImport();
		expect(onImport).toHaveBeenCalledTimes(1);

		await act(async () => end());
		expect(button.hasAttribute('disabled')).toBe(false);
		await clickImport();
		expect(onImport).toHaveBeenCalledTimes(2);
	});

	// A read that ends after the dialog closed would load into a page the user has gone on with.
	it('aborts the import that has not ended when the dialog closes, and reports nothing of it', async () => {
		let fail: (error: Error) => void = () => {};
		const onImport = vi.fn((..._args: Parameters<ImporterProps['onImport']>) => new Promise<void>((_resolve, reject) => (fail = reject)));
		const reopen = renderReopenable(onImport);

		await clickImport();
		const signal = onImport.mock.calls[0][3]!;
		await closeDialog();
		expect(signal.aborted).toBe(true);

		await act(async () => fail(new ImportProblems('The text does not read:', ['5:8 no spell is named x'])));
		await reopen();
		expect(document.querySelector('[data-testid="importer-problems"]')).toBeNull();
		expect(document.querySelector('[data-testid="import-button"]')!.hasAttribute('disabled')).toBe(false);
		expect(toasts).toEqual([]);
	});

	it('aborts the import that has not ended when the importer unmounts', async () => {
		const onImport = vi.fn((..._args: Parameters<ImporterProps['onImport']>) => new Promise<void>(() => {}));
		const { unmount } = renderWithHost(
			<Importer open title="SIML Import" onImport={onImport} onOpenChange={vi.fn()}>
				<p>how to</p>
			</Importer>,
			HOST_PARTS,
		);

		await clickImport();
		unmount();
		expect(onImport.mock.calls[0][3]!.aborted).toBe(true);
	});

	it('ignores a cancelled file picker', async () => {
		renderImporter();
		const upload = document.querySelector<HTMLInputElement>('[data-testid="importer-upload-input"]')!;
		await act(async () => {
			fireEvent.change(upload, { target: { files: [] } });
		});
		expect(document.querySelector('textarea')!.value).toBe('');
	});

	it('labels the upload input, and shows the label only when uploading is allowed', () => {
		renderImporter();
		const upload = document.querySelector<HTMLInputElement>('[data-testid="importer-upload-input"]')!;
		expect(document.querySelector('[data-testid="upload-button"]')!.getAttribute('for')).toBe(upload.id);
		// Derived from the title, because `keepMounted` puts every importer in the page at once.
		expect(upload.id).toBe('upload-input-json-import');
	});

	it('renders the hidden file input even when uploading is not allowed, as vanilla did', () => {
		renderImporter({ allowFileUpload: false });
		expect(document.querySelector('[data-testid="upload-button"]')).toBe(null);
		expect(document.querySelector('[data-testid="importer-upload-input"]')).not.toBe(null);
	});

	it('puts the description above the textarea', () => {
		renderImporter();
		const body = document.querySelector('[data-testid="sim-dialog-body"] > div')!;
		const [description, box] = Array.from(body.children);
		expect(description.textContent).toBe('how to');
		expect(box.contains(document.querySelector('[data-testid="importer-textarea"]'))).toBe(true);
	});
});
