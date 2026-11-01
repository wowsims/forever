import { SavedRotation } from '@features/apl/components/SavedRotation';
import { storedSavedData } from '@features/bulk/model/builds/saved_refs';
import { SavedEncounter } from '@features/encounter/components/SavedEncounter';
import { SavedGear } from '@features/gear/components/SavedGear';
import { useSavedGear } from '@features/gear/hooks/useSavedGear';
import { APLRotation } from '@generated/proto/apl';
import { Encounter, EquipmentSpec, ItemSpec, Spec } from '@generated/proto/common';
import { SavedGearSet } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { renderReady } from '@testing/saved_panel';
import { fakeHost } from '@testing/sim/fake_host';
import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import character from '../../../../../assets/locales/en/character.json';
import english from '../../../../../assets/locales/en/translation.json';
import { decodeSavedData, encodeSavedData, MAX_TEXT_LENGTH, type SavedDataExport } from '../../model/transfer_text';

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

const download = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/utils/dom', async importOriginal => ({ ...(await importOriginal<typeof import('@ui-kit/utils/dom')>()), downloadString: download }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent, trackPageView: () => {} }));

const GEAR_KEY = 'test-savedGearSets';
const ROTATION_KEY = 'test-savedRotations';
const ENCOUNTER_KEY = 'test-savedEncounters';
const HOSTILE_NAME = '<img src=x onerror=alert(1)>';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
	i18n.addResourceBundle('en', 'character', character, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
	i18n.removeResourceBundle('en', 'character');
});

beforeEach(() => {
	window.localStorage.clear();
	download.mockClear();
	trackEvent.mockClear();
});

// Stands in for the database's one throw: an item with no slot left.
const lookupEquipmentSpec = (spec: EquipmentSpec) => {
	if (spec.items.length > 16) throw new Error('No slots left');
	return { asSpec: () => spec };
};

const hostParts = ({ powerTools = true } = {}) => {
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	return {
		player: {
			getSpec: () => Spec.SpecDpsWarrior,
			getGear: () => ({ asSpec: () => EquipmentSpec.create() }),
			getBonusStats: () => ({ toProto: () => undefined }),
			getResolvedAplRotation: () => APLRotation.create(),
			setAplRotation: vi.fn(),
		},
		sim: {
			store,
			waitForInit: () => Promise.resolve(),
			db: { lookupEquipmentSpec },
			env: { storage: window.localStorage },
			encounter: { toProto: () => Encounter.create() },
		},
		individualConfig: {
			presets: {
				gear: [{ name: 'Preset Set', gear: EquipmentSpec.create() }],
				rotations: [],
				encounters: [],
			},
		},
		getSavedGearStorageKey: () => GEAR_KEY,
		getSavedRotationStorageKey: () => ROTATION_KEY,
		getSavedEncounterStorageKey: () => ENCOUNTER_KEY,
	};
};

const gearJson = (itemId: number) => SavedGearSet.toJson(SavedGearSet.create({ gear: EquipmentSpec.create({ items: [ItemSpec.create({ id: itemId })] }) }));
const store = (key: string, entries: Record<string, unknown>) => window.localStorage.setItem(key, JSON.stringify(entries));
const stored = (key: string) => JSON.parse(window.localStorage.getItem(key) ?? '{}');
const exportOf = (entries: SavedDataExport['entries'], extra: Partial<SavedDataExport> = {}) =>
	encodeSavedData({ type: 'gear', spec: 'SpecDpsWarrior', entries, ...extra });

const renderPanel = (ui: ReactElement = <SavedGear />, parts = hostParts()) => renderReady(ui, parts);

const openFromMenu = async (item: 'export' | 'import') => {
	fireEvent.click(screen.getByTestId('saved-data-transfer-button'));
	fireEvent.click(await screen.findByTestId(`saved-data-transfer-${item}`));
	await screen.findByTestId(item === 'export' ? 'exporter' : 'importer');
};

const exportText = () => screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value;
const exportEntries = async () => (await decodeSavedData(exportText())).entries;
const copyButton = () => screen.getByTestId<HTMLButtonElement>('copy-button');
const downloadButton = () => screen.getByRole<HTMLButtonElement>('button', { name: i18n.t('export.json.download_button') });
const tickBox = (name: string) => screen.getByRole<HTMLElement>('checkbox', { name });
const isTicked = (name: string) => tickBox(name).getAttribute('aria-checked') === 'true';
const importButton = () => screen.getByTestId<HTMLButtonElement>('import-button');
const problems = () => screen.findByTestId('importer-problems');

const paste = async (text: string) => {
	await act(async () => {
		fireEvent.change(screen.getByTestId('importer-textarea'), { target: { value: text } });
	});
};
const pastePreview = async (text: string) => {
	await paste(text);
	await screen.findByTestId('saved-data-tick-list');
};
const importTicked = async () => {
	await act(async () => {
		fireEvent.click(importButton());
	});
};

describe('SavedDataTransfer', () => {
	it('shows the cog only with power tools on', async () => {
		const { unmount } = await renderPanel(<SavedGear />, hostParts({ powerTools: false }));
		expect(screen.queryByTestId('saved-data-transfer-button')).toBeNull();
		unmount();

		await renderPanel();

		const header = screen.getByTestId('content-block-header');
		expect(within(header).getByTestId('saved-data-transfer-button')).toBeTruthy();
	});

	describe('export', () => {
		it('lists only the saved entries, every one ticked, and prints them all', async () => {
			store(GEAR_KEY, { Raid: gearJson(1), Dungeon: gearJson(2) });
			await renderPanel();

			await openFromMenu('export');

			const dialog = within(screen.getByTestId('exporter'));
			expect(dialog.getAllByRole('checkbox').map(box => box.getAttribute('aria-checked'))).toEqual(['true', 'true']);
			expect(dialog.queryByText('Preset Set')).toBeNull();
			await waitFor(() => expect(exportText()).toMatch(/^wowsims-saved:v1:/));
			const decoded = await decodeSavedData(exportText());
			expect(decoded.type).toBe('gear');
			expect(decoded.spec).toBe('SpecDpsWarrior');
			expect(decoded.entries).toEqual({ Raid: gearJson(1), Dungeon: gearJson(2) });
		});

		it('prints what stays ticked and offers nothing with nothing ticked', async () => {
			store(GEAR_KEY, { Raid: gearJson(1), Dungeon: gearJson(2) });
			await renderPanel();
			await openFromMenu('export');
			await waitFor(() => expect(copyButton().disabled).toBe(false));

			fireEvent.click(tickBox('Raid'));
			await waitFor(async () => expect(Object.keys(await exportEntries())).toEqual(['Dungeon']));
			expect(downloadButton().disabled).toBe(false);

			fireEvent.click(tickBox('Dungeon'));
			await waitFor(() => expect(exportText()).toBe(''));
			expect(copyButton().disabled).toBe(true);
			expect(downloadButton().disabled).toBe(true);
		});

		it('downloads the same text it shows', async () => {
			store(GEAR_KEY, { Raid: gearJson(1) });
			await renderPanel();
			await openFromMenu('export');
			await waitFor(() => expect(downloadButton().disabled).toBe(false));

			fireEvent.click(downloadButton());

			expect(download).toHaveBeenCalledWith(exportText(), 'wowsims-saved-gear.txt', 'text/plain');
		});

		it('starts ticked again when it opens again', async () => {
			store(GEAR_KEY, { Raid: gearJson(1) });
			await renderPanel();
			await openFromMenu('export');
			fireEvent.click(tickBox('Raid'));
			fireEvent.click(screen.getByRole('button', { name: i18n.t('common.close') }));
			await waitFor(() => expect(screen.queryByTestId('exporter')).toBeNull());

			await openFromMenu('export');

			expect(isTicked('Raid')).toBe(true);
		});

		it('turns the menu item off and says why when nothing is saved, so the dialog never opens empty', async () => {
			await renderPanel();

			fireEvent.click(screen.getByTestId('saved-data-transfer-button'));
			const item = await screen.findByTestId('saved-data-transfer-export');

			expect(item.textContent).toBe('No saved gear sets to export');
			expect(item.hasAttribute('data-disabled')).toBe(true);
			fireEvent.click(item);
			expect(screen.queryByTestId('exporter')).toBeNull();
			expect(trackEvent).not.toHaveBeenCalled();
		});

		it('turns the menu item off when only an unreadable entry is stored', async () => {
			store(GEAR_KEY, { Broken: 'nope' });
			vi.spyOn(console, 'warn').mockImplementation(() => {});
			await renderPanel();

			fireEvent.click(screen.getByTestId('saved-data-transfer-button'));

			expect((await screen.findByTestId('saved-data-transfer-export')).hasAttribute('data-disabled')).toBe(true);
		});
	});

	describe('import', () => {
		it('saves several entries at once and shows them', async () => {
			store(GEAR_KEY, { Kept: gearJson(9) });
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ One: gearJson(1), Two: gearJson(2), Three: gearJson(3) }));
			expect(importButton().textContent).toBe('Import 3');
			await importTicked();

			await waitFor(() => expect(screen.queryByTestId('importer')).toBeNull());
			expect(stored(GEAR_KEY)).toEqual({ Kept: gearJson(9), One: gearJson(1), Two: gearJson(2), Three: gearJson(3) });
			const chips = within(screen.getByTestId('saved-data-custom'));
			expect(['Kept', 'One', 'Two', 'Three'].map(name => chips.getByRole('button', { name }).textContent)).toEqual(['Kept', 'One', 'Two', 'Three']);
			expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'import', label: expect.any(String) });
		});

		it('reads the same text from a file', async () => {
			await renderPanel();
			await openFromMenu('import');

			const file = new File([await exportOf({ Filed: gearJson(1) })], 'gear.txt');
			await act(async () => {
				fireEvent.change(screen.getByTestId('importer-upload-input'), { target: { files: [file] } });
			});

			await screen.findByTestId('saved-data-tick-list');
			expect(isTicked('Filed')).toBe(true);
		});

		it('refuses a file past the size cap without reading it', async () => {
			await renderPanel();
			await openFromMenu('import');

			const file = new File(['x'], 'huge.bin');
			Object.defineProperty(file, 'size', { value: MAX_TEXT_LENGTH + 1 });
			const read = vi.spyOn(file, 'text');
			await act(async () => {
				fireEvent.change(screen.getByTestId('importer-upload-input'), { target: { files: [file] } });
			});

			expect((await problems()).textContent).toBe('This file is too large to import.');
			expect(read).not.toHaveBeenCalled();
		});

		it('saves a name in use under the next free number, and leaves identical contents unticked', async () => {
			store(GEAR_KEY, { Taken: gearJson(1), 'Taken (2)': gearJson(2), Same: gearJson(3) });
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ Taken: gearJson(7), Same: gearJson(3), 'Preset Set': gearJson(8) }));

			const notes = screen.getAllByTestId('saved-data-tick-note').map(note => note.textContent);
			expect(notes).toEqual(['Saves as "Taken (3)"', 'Already saved', 'Saves as "Preset Set (2)"']);
			expect([isTicked('Taken'), isTicked('Same'), isTicked('Preset Set')]).toEqual([true, false, true]);
			expect(importButton().textContent).toBe('Import 2');

			await importTicked();

			await waitFor(() => expect(stored(GEAR_KEY)['Taken (3)']).toEqual(gearJson(7)));
			expect(stored(GEAR_KEY).Taken).toEqual(gearJson(1));
			expect(stored(GEAR_KEY)['Preset Set (2)']).toEqual(gearJson(8));
			expect(Object.keys(stored(GEAR_KEY))).toHaveLength(5);
		});

		it('saves nothing new when the same file is imported twice, a clash that went to (2) included', async () => {
			store(GEAR_KEY, { Taken: gearJson(1) });
			await renderPanel();
			const text = await exportOf({ Taken: gearJson(7), Fresh: gearJson(8) });
			await openFromMenu('import');
			await pastePreview(text);
			await importTicked();
			await waitFor(() => expect(stored(GEAR_KEY)['Taken (2)']).toEqual(gearJson(7)));
			await waitFor(() => expect(screen.queryByTestId('importer')).toBeNull());
			const afterFirst = stored(GEAR_KEY);

			await openFromMenu('import');
			await pastePreview(text);

			expect(screen.getAllByTestId('saved-data-tick-note').map(note => note.textContent)).toEqual(['Already saved as "Taken (2)"', 'Already saved']);
			expect([isTicked('Taken'), isTicked('Fresh')]).toEqual([false, false]);
			expect(importButton().textContent).toBe('Import 0');
			expect(importButton().disabled).toBe(true);
			expect(stored(GEAR_KEY)).toEqual(afterFirst);
		});

		it('notes the saved name when it differs, and saves a ticked copy under a free name', async () => {
			store(GEAR_KEY, { 'Crit test (2)': gearJson(1) });
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ 'Crit test': gearJson(1) }));

			expect(screen.getByTestId('saved-data-tick-note').textContent).toBe('Already saved as "Crit test (2)"');
			expect(isTicked('Crit test')).toBe(false);

			fireEvent.click(tickBox('Crit test'));
			expect(screen.getByTestId('saved-data-tick-note').textContent).toBe('Saves as "Crit test"');
			await importTicked();

			await waitFor(() => expect(Object.keys(stored(GEAR_KEY))).toEqual(['Crit test (2)', 'Crit test']));
		});

		it('keeps both when an entry the codec rejects holds the name', async () => {
			vi.spyOn(console, 'warn').mockImplementation(() => {});
			store(GEAR_KEY, { X: 'nope' });
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ X: gearJson(1) }));

			expect(screen.getByTestId('saved-data-tick-note').textContent).toBe('Saves as "X (2)"');
			await importTicked();

			await waitFor(() => expect(stored(GEAR_KEY)['X (2)']).toEqual(gearJson(1)));
			expect(stored(GEAR_KEY).X).toBe('nope');
		});

		it('saves only the ticked entries and cannot import none', async () => {
			await renderPanel();
			await openFromMenu('import');
			expect(importButton().disabled).toBe(true);
			await pastePreview(await exportOf({ One: gearJson(1), Two: gearJson(2) }));

			fireEvent.click(tickBox('One'));
			expect(importButton().textContent).toBe('Import 1');
			fireEvent.click(tickBox('Two'));
			expect(importButton().textContent).toBe('Import 0');
			expect(importButton().disabled).toBe(true);

			fireEvent.click(tickBox('Two'));
			await importTicked();

			await waitFor(() => expect(Object.keys(stored(GEAR_KEY))).toEqual(['Two']));
		});

		it('imports gear and encounters from another spec', async () => {
			const parts = hostParts();
			const { unmount } = await renderPanel(<SavedGear />, parts);
			await openFromMenu('import');
			await pastePreview(await exportOf({ Theirs: gearJson(1) }, { spec: 'SpecMage' }));
			expect(isTicked('Theirs')).toBe(true);
			unmount();

			await renderPanel(<SavedEncounter />, parts);
			await openFromMenu('import');
			await pastePreview(await exportOf({ Boss: { encounter: Encounter.toJson(Encounter.create()) } }, { type: 'encounter', spec: 'SpecMage' }));
			expect(isTicked('Boss')).toBe(true);
		});

		it('refuses a spec-locked type from another spec, naming the spec', async () => {
			await renderPanel(<SavedRotation />);
			await openFromMenu('import');

			await paste(await exportOf({ Theirs: {} }, { type: 'rotation', spec: 'SpecProtectionWarrior' }));

			expect((await problems()).textContent).toBe('This export is for Protection Warrior.');
			expect(screen.queryByTestId('saved-data-tick-list')).toBeNull();
			expect(importButton().disabled).toBe(true);
		});

		it('refuses a spec name it does not know, the prototype chain included', async () => {
			await renderPanel(<SavedRotation />);
			await openFromMenu('import');

			await paste(await exportOf({ Theirs: {} }, { type: 'rotation', spec: 'constructor' }));

			expect((await problems()).textContent).toBe('This export is for a spec this site does not know.');
		});

		it('refuses the wrong type, naming what it holds and the panel asked for', async () => {
			await renderPanel();
			await openFromMenu('import');

			await paste(await exportOf({ Mine: { talentsString: '1' } }, { type: 'talents' }));

			expect((await problems()).textContent).toBe('This export holds talents, not gear sets.');
			expect(importButton().disabled).toBe(true);
		});

		it.each([
			['unreadable text', () => 'hello', 'This text is not a saved data export, or part of it is missing. Nothing was saved.'],
			[
				'an unknown format version',
				async () => (await exportOf({ A: gearJson(1) })).replace('v1', 'v9'),
				'This export uses a format this site does not know. Nothing was saved.',
			],
			['an export with no entries', () => exportOf({}), 'This export holds no entries.'],
			['entries that all fail the codec', () => exportOf({ A: 'nope', B: 5 }), 'None of the entries in this export could be read. Nothing was saved.'],
		])('refuses %s and saves nothing', async (_, text, message) => {
			await renderPanel();
			await openFromMenu('import');

			await paste(await text());

			expect((await problems()).textContent).toBe(message);
			expect(importButton().disabled).toBe(true);
			expect(window.localStorage.getItem(GEAR_KEY)).toBeNull();
		});

		it('lists the good entries and counts the unreadable ones', async () => {
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ Good: gearJson(1), Bad: 'nope', Worse: 5 }));

			expect(screen.getByTestId('saved-data-import-unreadable').textContent).toBe('2 entries could not be read and are left out.');
			expect(within(screen.getByTestId('saved-data-tick-list')).getAllByRole('checkbox')).toHaveLength(1);
		});

		it('counts gear the database cannot equip as unreadable, since the Builds tab looks it up while it renders', async () => {
			const items = Array.from({ length: 17 }, (_, id) => ItemSpec.create({ id }));
			const overfull = SavedGearSet.toJson(SavedGearSet.create({ gear: EquipmentSpec.create({ items }) }));
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ Good: gearJson(1), Overfull: overfull }));

			expect(screen.getByTestId('saved-data-import-unreadable').textContent).toBe('1 entry could not be read and is left out.');
		});

		it('counts an encounter entry with no encounter as unreadable', async () => {
			await renderPanel(<SavedEncounter />);
			await openFromMenu('import');

			await paste(await exportOf({ Empty: {} }, { type: 'encounter' }));

			expect((await problems()).textContent).toBe('None of the entries in this export could be read. Nothing was saved.');
		});

		it('keeps the dialog open and says so when storage refuses the write', async () => {
			await renderPanel();
			await openFromMenu('import');
			await pastePreview(await exportOf({ One: gearJson(1) }));

			const full = {
				getItem: () => null,
				setItem: () => {
					throw new Error('QuotaExceededError');
				},
			};
			const refused = vi.spyOn(window, 'localStorage', 'get').mockReturnValue(full as unknown as Storage);
			await importTicked();
			refused.mockRestore();

			expect((await problems()).textContent).toContain('nothing was saved');
			expect(screen.getByTestId('importer')).toBeTruthy();
			expect(window.localStorage.getItem(GEAR_KEY)).toBeNull();
		});

		it('drops the text and the preview when it closes', async () => {
			await renderPanel();
			await openFromMenu('import');
			await pastePreview(await exportOf({ One: gearJson(1) }));

			fireEvent.click(screen.getByRole('button', { name: i18n.t('common.close') }));
			await waitFor(() => expect(screen.queryByTestId('importer')).toBeNull());
			await openFromMenu('import');

			expect(screen.getByTestId<HTMLTextAreaElement>('importer-textarea').value).toBe('');
			expect(screen.queryByTestId('saved-data-tick-list')).toBeNull();
		});
	});

	describe('names from other people', () => {
		it('show as text in the preview, the chip and the delete confirmation', async () => {
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ [HOSTILE_NAME]: gearJson(1) }));

			const list = screen.getByTestId('saved-data-tick-list');
			expect(list.textContent).toContain(HOSTILE_NAME);
			expect(list.querySelector('img')).toBeNull();

			await importTicked();
			await waitFor(() => expect(stored(GEAR_KEY)[HOSTILE_NAME]).toBeDefined());

			const chip = await within(screen.getByTestId('saved-data-custom')).findByRole('button', { name: HOSTILE_NAME });
			expect(chip.textContent).toBe(HOSTILE_NAME);
			fireEvent.click(screen.getByTestId('saved-data-set-delete'));
			expect((await screen.findByText(new RegExp(`'${HOSTILE_NAME.replace(/[()]/g, '\\$&')}'`))).textContent).toContain(HOSTILE_NAME);
			expect(document.querySelector('img[src="x"]')).toBeNull();
		});

		it('keep replacement patterns as text in a confirmation built from a template', async () => {
			store(ROTATION_KEY, { '$& $1 $`': {} });
			await renderPanel(<SavedRotation />);

			fireEvent.click(screen.getByTestId('saved-data-set-delete'));

			expect(await screen.findByText("Delete saved rotation '$& $1 $`'?")).toBeTruthy();
		});

		it('keep translation patterns as text', async () => {
			const name = '$t(common.help) {{count}}';
			store(GEAR_KEY, { [name]: gearJson(1) });
			await renderPanel();
			await openFromMenu('import');

			await pastePreview(await exportOf({ [name]: gearJson(2) }));

			expect(screen.getByTestId('saved-data-tick-note').textContent).toBe(`Saves as "${name} (2)"`);
		});

		it('are trimmed, and __proto__ never reaches the record', async () => {
			await renderPanel();
			await openFromMenu('import');

			const entries = JSON.parse(`{"__proto__":${JSON.stringify(gearJson(1))},"  Spaced  ":${JSON.stringify(gearJson(2))}}`);
			await pastePreview(await exportOf(entries));

			expect(screen.getByTestId('saved-data-import-unreadable').textContent).toContain('1 entry');
			await importTicked();

			await waitFor(() => expect(Object.keys(stored(GEAR_KEY))).toEqual(['Spaced']));
			expect(({} as Record<string, unknown>).id).toBeUndefined();
		});
	});

	describe('imported entries', () => {
		it('reach the other readers of the slot without a reload, as the Builds tab reads them', async () => {
			const parts = hostParts();
			let builds: string[] = [];
			const BuildsReader = () => {
				builds = useSavedGear().entries.map(({ name }) => name);
				return null;
			};
			await renderPanel(
				<>
					<SavedGear />
					<BuildsReader />
				</>,
				parts,
			);
			await openFromMenu('import');
			await pastePreview(await exportOf({ One: gearJson(1), Two: gearJson(2) }));

			await importTicked();

			await waitFor(() => expect(builds).toEqual(['One', 'Two']));
			const atRunStart = storedSavedData(fakeHost(parts), GEAR_KEY, SavedGearSet, { ignoreUnknownFields: true });
			expect(atRunStart.map(({ name }) => name)).toEqual(['One', 'Two']);
		});

		it('stay and load with power tools off', async () => {
			store(GEAR_KEY, { Imported: gearJson(1) });
			await renderPanel(<SavedGear />, hostParts({ powerTools: false }));

			expect(within(screen.getByTestId('saved-data-custom')).getByRole('button', { name: 'Imported' })).toBeTruthy();
		});
	});
});
