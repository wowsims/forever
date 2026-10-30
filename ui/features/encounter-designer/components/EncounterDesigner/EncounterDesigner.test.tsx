import { EncounterPicker } from '@features/encounter/components/EncounterPicker';
import { JSON_IMPORTER } from '@features/import-export/importers/json';
import { Player as PlayerProto } from '@generated/proto/api';
import {
	AddWave,
	DesignDowntime,
	Encounter as EncounterProto,
	EncounterDesign,
	HealthKeyframe,
	MobType,
	PresetEncounter,
	PresetTarget,
	Stat,
	Target,
} from '@generated/proto/common';
import { IndividualSimSettings, UIDatabase } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { CREATURE_ARMOR_BY_LEVEL } from '@sim/proto/sim_tables_auto_gen';
import { Encounter } from '@sim/raid/encounter';
import { MY_ENCOUNTERS_STORAGE_KEY } from '@sim/state/storage_keys';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { memoryEnv } from '@testing/sim/memory_env';
import { loadSettings, STORAGE_KEY } from '@testing/sim/stat_filters';
import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// The add of a preset encounter's wave, which no encounter lists as a target.
const DYNAMIC_ADD = Target.create({ id: 99001, name: 'Dynamic Add', level: 63, mobType: MobType.MobTypeMechanical });
DYNAMIC_ADD.stats = Array.from({ length: Math.max(...Object.values(Stat).filter(value => typeof value === 'number')) + 1 }, () => 0);
DYNAMIC_ADD.stats[Stat.StatHealth] = 5000;

const ADD = PresetTarget.create({
	path: 'Molten Core/Son of Flame',
	target: Target.create({ id: 12143, name: 'Son of Flame', level: 62, mobType: MobType.MobTypeElemental }),
});

// The fields, buttons and lanes are found by the names a player reads, so the suite loads the English strings.
beforeAll(() => {
	serveDatabase(
		UIDatabase.create({
			encounters: [
				PresetEncounter.create({ path: 'Molten Core/Ragnaros', targets: [ADD] }),
				PresetEncounter.create({
					path: 'Default/Dynamic Adds',
					targets: [PresetTarget.create({ path: 'Default/Dynamic Boss', target: Target.create({ id: 99000, name: 'Dynamic Boss' }) })],
					design: EncounterDesign.create({ addWaves: [AddWave.create({ name: 'Add', adds: [DYNAMIC_ADD] })] }),
				}),
			],
		}),
	);
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});
beforeEach(() => window.localStorage.clear());

// The host files My encounters under a key of its spec, as the sim page does.
const specKey = (spec: string) => (part: string) => `__${spec}__${part}`;
const myEncountersKey = (spec = 'warrior') => specKey(spec)(MY_ENCOUNTERS_STORAGE_KEY);
const stored = (spec?: string): Record<string, unknown> => JSON.parse(window.localStorage.getItem(myEncountersKey(spec)) ?? '{}');
const storeEncounters = (encounters: Array<EncounterProto>, spec?: string) =>
	window.localStorage.setItem(myEncountersKey(spec), JSON.stringify(Object.fromEntries(encounters.map(proto => [proto.name, EncounterProto.toJson(proto)]))));

const renderPicker = ({ sim, player }: Awaited<ReturnType<typeof makeIndividualSim>>, spec = 'warrior') =>
	renderWithHost(<EncounterPicker showExecuteProportion={false} />, { sim, player, getStorageKey: specKey(spec), individualConfig: { defaults: {} } });

const setup = async ({ targets = 1, design = EncounterDesign.create(), spec = 'warrior' } = {}) => {
	const built = await makeIndividualSim();
	const { sim } = built;
	sim.encounter.setTargets(Array.from({ length: targets }, (_, index) => Target.create({ id: index + 1, name: `Target ${index + 1}` })));
	sim.encounter.setDesign(design);
	const view = renderPicker(built, spec);
	return { encounter: sim.encounter, baseSeconds: sim.encounter.getDuration(), host: built.host, view };
};

const openDesigner = async () => {
	await act(async () => fireEvent.click(screen.getByRole('button', { name: 'Designer' })));
	return screen.getByTestId('encounter-designer');
};

const lane = (name: string) => screen.getByRole('group', { name });

const add = (laneName: string, event: string) => act(() => void fireEvent.click(within(lane(laneName)).getByRole('button', { name: `Add ${event}` })));

const press = (bar: HTMLElement, key: string, shiftKey = false, ctrlKey = false) => act(() => void fireEvent.keyDown(bar, { key, shiftKey, ctrlKey }));

const clickButton = (name: string) => act(async () => fireEvent.click(screen.getByRole('button', { name })));

const clickFirst = (name: string) => act(async () => fireEvent.click(screen.getAllByRole('button', { name })[0]));

const commit = (input: HTMLElement, value: string) => act(() => void fireEvent.change(input, { target: { value } }));

// The first target is the one the picker never removes.
const removeSecondTarget = async () => {
	await clickButton('Targets');
	const menu = screen.getAllByRole('button', { name: 'Target menu' })[1];
	await act(async () => {
		menu.focus();
		fireEvent.click(menu);
	});
	await clickButton('Delete Target');
};

const downtimeOn = (targetIndex: number) =>
	EncounterDesign.create({ downtimes: [DesignDowntime.create({ targetIndex, start: { at: 0.2, jitter: 0 }, end: { at: 0.3, jitter: 0 } })] });

describe('EncounterDesigner', () => {
	it("offers the add of a preset encounter's wave as a template, and edits the add's stats in the saved design", async () => {
		{
			const { encounter } = await setup();
			const dialog = await openDesigner();

			await add('Add waves', 'Wave');
			await clickButton('Edit add');
			const addDialog = screen.getByTestId('encounter-designer-add-dialog');
			expect(within(addDialog).getByLabelText('Add template').textContent).toContain('Default/Dynamic Add');
			expect((within(addDialog).getByLabelText('Health') as HTMLInputElement).value).toBe('5000');
			await commit(within(addDialog).getByLabelText('Health'), '7000');
			expect(within(addDialog).getByLabelText('Add template').textContent).toContain('Custom');
			await clickButton('Done');

			expect(within(dialog).queryByTestId('encounter-designer-add-dialog')).toBeNull();
			await clickButton('Save');
			const [wave] = encounter.getDesign().addWaves;
			expect(wave.adds).toHaveLength(1);
			expect(wave.adds[0].name).toBe('Dynamic Add');
			expect(wave.adds[0].stats[Stat.StatHealth]).toBe(7000);
		}
	});

	const pickLevel = async (addDialog: HTMLElement, level: number) => {
		await userEvent.click(within(addDialog).getByLabelText('Level'));
		await userEvent.click(screen.getByRole('option', { name: String(level) }));
	};

	const armorField = (addDialog: HTMLElement) => (within(addDialog).getByLabelText('Armor') as HTMLInputElement).value;

	it('edits a new add in a dialog over the designer: its level sets its armor, Escape hands focus back, and Save keeps the edits', async () => {
		const { encounter } = await setup();
		await openDesigner();
		await add('Add waves', 'Wave');

		await clickButton('Add an add');
		let addDialog = screen.getByTestId('encounter-designer-add-dialog');
		expect(within(addDialog).getByTestId('sim-dialog-title').textContent).toBe('Dynamic Add');
		await pickLevel(addDialog, 61);
		expect(CREATURE_ARMOR_BY_LEVEL[61]).toBeGreaterThan(0);
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[61]));
		await commit(within(addDialog).getByLabelText('Health'), '9000');
		await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
		await waitFor(() => expect(screen.queryByTestId('encounter-designer-add-dialog')).toBeNull());
		expect(screen.getByTestId('encounter-designer')).toBeTruthy();

		const edit = screen.getAllByRole('button', { name: 'Edit add' })[1];
		await act(async () => {
			edit.focus();
			fireEvent.click(edit);
		});
		addDialog = screen.getByTestId('encounter-designer-add-dialog');
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[61]));
		await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
		await waitFor(() => expect(screen.queryByTestId('encounter-designer-add-dialog')).toBeNull());
		expect(document.activeElement).toBe(edit);

		await clickButton('Save');
		const [first, second] = encounter.getDesign().addWaves[0].adds;
		expect(first).toEqual(DYNAMIC_ADD);
		expect(second.level).toBe(61);
		expect(second.stats[Stat.StatArmor]).toBe(CREATURE_ARMOR_BY_LEVEL[61]);
		expect(second.stats[Stat.StatHealth]).toBe(9000);
	});

	it("sets the armor of the Dynamic Adds preset's add to the game files' creature armor of the level picked", async () => {
		const { encounter } = await setup();
		await userEvent.click(screen.getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		await openDesigner();

		await act(async () => fireEvent.click(within(lane('Add waves')).getAllByRole('slider')[0]));
		await clickButton('Edit add');
		const addDialog = screen.getByTestId('encounter-designer-add-dialog');
		expect(within(addDialog).getByLabelText('Level').textContent).toContain('63');
		await pickLevel(addDialog, 60);
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[60]));

		await clickButton('Done');
		await clickButton('Save');
		const [savedAdd] = encounter.getDesign().addWaves[0].adds;
		expect(savedAdd.level).toBe(60);
		expect(savedAdd.stats[Stat.StatArmor]).toBe(CREATURE_ARMOR_BY_LEVEL[60]);
	});

	it('adds a wave in the draft, moves and edits it, and saves it to the store as fractions', async () => {
		const { encounter, baseSeconds } = await setup();
		await openDesigner();

		await add('Add waves', 'Wave');
		const bar = within(lane('Add waves')).getByRole('slider');
		expect(bar.hasAttribute('data-selected')).toBe(true);
		expect(encounter.getDesign().addWaves).toEqual([]);

		await press(bar, 'ArrowRight', true);
		await press(bar, 'ArrowRight');
		await clickButton('Add an add');
		await clickButton('Done');
		await clickFirst('Duplicate add');
		await clickFirst('Remove add');
		expect(encounter.getDesign().addWaves).toEqual([]);

		await clickButton('Save');
		expect(screen.queryByTestId('encounter-designer')).toBeNull();
		const [wave] = encounter.getDesign().addWaves;
		expect(wave.spawn!.at).toBeCloseTo(0.4 + 6 / baseSeconds, 9);
		expect(wave.lifetime).toBeCloseTo(0.1, 9);
		expect(wave.adds).toEqual([DYNAMIC_ADD, DYNAMIC_ADD]);
	});

	describe('event bars', () => {
		// happy-dom has no pointer capture, which a drag needs to start.
		beforeEach(() => {
			vi.spyOn(Element.prototype, 'setPointerCapture').mockImplementation(() => {});
			vi.spyOn(Element.prototype, 'hasPointerCapture').mockReturnValue(false);
		});

		const bars = () => within(lane('Boss')).getAllByRole('slider');
		const startOf = (bar: HTMLElement) => Number(bar.getAttribute('aria-valuenow'));

		// At the default zoom a pixel is a hundredth of a second.
		const drag = (bar: HTMLElement, dx: number, { altKey = false, escape = false } = {}) => {
			const pointer = { pointerId: 1, pointerType: 'mouse', button: 0, altKey };
			act(() => void fireEvent.pointerDown(bar, { ...pointer, clientX: 100 }));
			act(() => void fireEvent.pointerMove(bar, { ...pointer, clientX: 100 + dx }));
			if (escape) act(() => void fireEvent.keyDown(bar, { key: 'Escape' }));
			act(() => void fireEvent.pointerUp(bar, { ...pointer, clientX: 100 + dx }));
		};

		const withDowntime = async () => {
			const { encounter } = await setup();
			const dialog = await openDesigner();
			await add('Boss', 'Downtime');
			return { encounter, dialog, first: startOf(bars()[0]) };
		};

		it('moves an event with a drag, copies it with an Alt drag, and drops either on Escape', async () => {
			const { first } = await withDowntime();

			drag(bars()[0], 300);
			expect(bars().map(startOf)).toEqual([first + 3]);

			drag(bars()[0], 600, { altKey: true });
			expect(
				bars()
					.map(startOf)
					.sort((a, b) => a - b),
			).toEqual([first + 3, first + 9]);

			drag(bars()[0], 500, { escape: true });
			drag(bars()[0], 500, { altKey: true, escape: true });
			expect(bars()).toHaveLength(2);
			expect(
				bars()
					.map(startOf)
					.sort((a, b) => a - b),
			).toEqual([first + 3, first + 9]);
		});

		it('pans the timeline with a drag on empty lane space, and leaves the bars where they are', async () => {
			const { dialog, first } = await withDowntime();
			const timeline = within(dialog).getByRole('region', { name: 'Encounter timeline' });
			timeline.scrollLeft = 500;
			drag(lane('Boss'), -200);
			expect(timeline.scrollLeft).toBe(700);
			expect(bars().map(startOf)).toEqual([first]);
			expect(bars()[0].hasAttribute('data-selected')).toBe(true);

			drag(bars()[0], 300);
			expect(timeline.scrollLeft).toBe(700);
			expect(bars().map(startOf)).toEqual([first + 3]);
		});

		it('moves only the keyframe at the end of a boss health segment, and leaves the implied start alone', async () => {
			await setup();
			await openDesigner();
			await add('Boss', 'Boss health');
			const text = () => bars()[0].getAttribute('aria-valuetext') ?? '';
			expect(text()).toContain('from 0s to 72s');

			drag(bars()[0], 300);
			expect(text()).toContain('from 0s to 75s');

			// The first segment starts at the implied 0 s, which has no keyframe to move.
			const before = text();
			act(() => void fireEvent.keyDown(bars()[0], { key: 'ArrowLeft', altKey: true, shiftKey: true }));
			expect(text()).toBe(before);
		});

		it('duplicates on Ctrl+D one step later and focuses the copy', async () => {
			const { first } = await withDowntime();
			await press(bars()[0], 'd', false, true);
			expect(bars()).toHaveLength(2);
			expect(startOf(document.activeElement as HTMLElement)).toBe(first + 1);
		});

		it('renames inline: Enter keeps the text, Escape drops it, and a click elsewhere keeps focus where it went', async () => {
			const { dialog } = await withDowntime();
			const label = () => within(bars()[0]).getByText('Downtime 1');

			act(() => void fireEvent.doubleClick(label()));
			let input = within(dialog).getByRole('textbox', { name: 'Event name' });
			expect(bars()[0].contains(input)).toBe(false);
			await commit(input, 'Phase 2');
			await act(() => void fireEvent.keyDown(input, { key: 'Enter' }));
			expect(bars()[0].getAttribute('aria-label')).toBe('Phase 2');
			expect(document.activeElement).toBe(bars()[0]);

			await press(bars()[0], 'F2');
			input = within(lane('Boss')).getByRole('textbox', { name: 'Event name' });
			await commit(input, 'Dropped');
			await act(() => void fireEvent.keyDown(input, { key: 'Escape' }));
			expect(bars()[0].getAttribute('aria-label')).toBe('Phase 2');

			await press(bars()[0], 'F2');
			input = within(lane('Boss')).getByRole('textbox', { name: 'Event name' });
			await commit(input, 'Phase 3');
			const panelName = within(screen.getByRole('region', { name: 'Phase 2' })).getByLabelText('Name');
			await act(async () => {
				panelName.focus();
				fireEvent.blur(input);
			});
			expect(bars()[0].getAttribute('aria-label')).toBe('Phase 3');
			expect(document.activeElement).toBe(panelName);
		});

		it('renames from the context menu, and the menu hands focus back to the bar on Escape', async () => {
			const { encounter } = await withDowntime();
			const bar = bars()[0];

			act(() => bar.focus());
			await act(async () => void fireEvent.contextMenu(bar, { clientX: 10, clientY: 10 }));
			await screen.findByRole('menuitem', { name: 'Duplicate' });
			await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
			await waitFor(() => expect(screen.queryByRole('menuitem', { name: 'Duplicate' })).toBeNull());
			expect(document.activeElement).toBe(bar);

			await act(async () => void fireEvent.contextMenu(bar, { clientX: 10, clientY: 10 }));
			await act(async () => fireEvent.click(await screen.findByRole('menuitem', { name: 'Rename' })));
			const input = within(lane('Boss')).getByRole('textbox', { name: 'Event name' });
			await commit(input, 'From the menu');
			await act(() => void fireEvent.keyDown(input, { key: 'Enter' }));
			await clickButton('Save');
			expect(encounter.getDesign().downtimes[0].name).toBe('From the menu');
		});

		const keyframe = (at: number, healthPercent: number) => HealthKeyframe.create({ at, healthPercent });
		const bossDowntime = (start: number, end: number) =>
			DesignDowntime.create({ targetIndex: 0, start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
		const openMenu = async (bar: HTMLElement) => act(async () => void fireEvent.contextMenu(bar, { clientX: 10, clientY: 10 }));

		it('holds the boss health across a boss downtime from its context menu, in one change', async () => {
			const { encounter } = await setup({ design: EncounterDesign.create({ downtimes: [bossDowntime(0.4, 0.5)], bossHealth: [keyframe(0.8, 60)] }) });
			await openDesigner();
			const downtime = bars().find(bar => bar.getAttribute('aria-label') === 'Downtime 1')!;
			await openMenu(downtime);
			await act(async () => fireEvent.click(await screen.findByRole('menuitem', { name: 'Hold boss health here' })));
			expect(within(lane('Boss')).getByText('Hold 80 %')).toBeTruthy();
			await clickButton('Save');
			expect(encounter.getDesign().bossHealth.map(entry => [entry.at, entry.healthPercent])).toEqual([
				[0.8, 60],
				[0.4, 80],
				[0.5, 80],
			]);
		});

		it('disables the hold while a keyframe lies inside the downtime, and says which', async () => {
			await setup({ design: EncounterDesign.create({ downtimes: [bossDowntime(0.4, 0.5)], bossHealth: [keyframe(0.45, 80)] }) });
			await openDesigner();
			await openMenu(bars().find(bar => bar.getAttribute('aria-label') === 'Downtime 1')!);
			const item = await screen.findByRole('menuitem', { name: 'Hold boss health here' });
			expect(item.hasAttribute('data-disabled')).toBe(true);
			expect(item.getAttribute('data-tip')).toBe('Boss health 1 at 81s is in the way');
		});

		it('offers the hold of a boss downtime only for the boss', async () => {
			await setup({
				targets: 2,
				design: EncounterDesign.create({
					downtimes: [DesignDowntime.create({ targetIndex: 1, start: { at: 0.4, jitter: 0 }, end: { at: 0.5, jitter: 0 } })],
				}),
			});
			await openDesigner();
			await openMenu(bars()[0]);
			await screen.findByRole('menuitem', { name: 'Duplicate' });
			expect(screen.queryByRole('menuitem', { name: 'Hold boss health here' })).toBeNull();
		});

		it('cuts a flat part off a boss health segment from its context menu', async () => {
			await setup({ design: EncounterDesign.create({ bossHealth: [keyframe(0.5, 50)] }) });
			await openDesigner();
			await openMenu(bars()[0]);
			await act(async () => fireEvent.click(await screen.findByRole('menuitem', { name: 'Hold health' })));
			expect(bars()).toHaveLength(2);
			expect(within(lane('Boss')).getByText('Hold 100 %')).toBeTruthy();
		});

		it('opens no menu on empty lane space', async () => {
			await withDowntime();
			await act(
				async () => void fireEvent.contextMenu(within(lane('Boss')).getByRole('slider').parentElement!.firstElementChild!, { clientX: 5, clientY: 5 }),
			);
			expect(screen.queryByRole('menuitem')).toBeNull();
		});
	});

	it('focuses the first add button on open, and moves focus into the event panel on Enter, not on a click', async () => {
		await setup();
		await openDesigner();
		expect(document.activeElement).toBe(within(lane('Boss')).getByRole('button', { name: 'Add Downtime' }));

		await add('Boss', 'Downtime');
		const bar = within(lane('Boss')).getByRole('slider');
		const panel = screen.getByRole('region', { name: 'Downtime 1' });
		act(() => {
			bar.focus();
			fireEvent.click(bar);
		});
		expect(document.activeElement).toBe(bar);
		await press(bar, 'Enter');
		expect(document.activeElement).toBe(panel);
	});

	it('puts back the duration, the targets and the design on Cancel, and starts the next opening from the store', async () => {
		const design = downtimeOn(2);
		const { encounter, baseSeconds } = await setup({ targets: 3, design });
		const targets = encounter.getTargets();
		const dialog = await openDesigner();

		await commit(within(dialog).getByLabelText('Duration'), String(baseSeconds + 60));
		await removeSecondTarget();
		await add('Boss', 'Downtime');
		expect(encounter.getDuration()).toBe(baseSeconds + 60);
		expect(encounter.getTargets()).toHaveLength(2);

		await clickButton('Cancel');
		expect(screen.queryByTestId('encounter-designer')).toBeNull();
		expect(encounter.getDuration()).toBe(baseSeconds);
		expect(encounter.getTargets()).toBe(targets);
		expect(EncounterDesign.equals(encounter.getDesign(), design)).toBe(true);

		await openDesigner();
		expect(within(lane('Boss')).getAllByRole('slider')).toHaveLength(1);
	});

	it('follows a removed target in the draft the way the store does', async () => {
		const { encounter } = await setup({ targets: 3, design: downtimeOn(2) });
		await openDesigner();

		await removeSecondTarget();
		await clickButton('Save');
		expect(encounter.getDesign().downtimes.map(downtime => downtime.targetIndex)).toEqual([1]);
	});

	it('lists an issue and marks its bar, and still saves the design', async () => {
		const { encounter } = await setup();
		const dialog = await openDesigner();
		await add('Movement', 'Move');
		await add('Movement', 'Move');

		const [first, second] = within(lane('Movement')).getAllByRole('slider');
		expect([first.hasAttribute('data-invalid'), second.hasAttribute('data-invalid')]).toEqual([false, true]);
		const issues = within(dialog).getByRole('region', { name: 'Issues' });
		expect(issues.textContent).toContain(english.settings_tab.encounter.designer.validation.overlap);

		await clickButton('Save');
		expect(encounter.getDesign().moves).toHaveLength(2);
	});

	it('names in seconds the part of a boss downtime no add is up for', async () => {
		const { baseSeconds } = await setup({
			design: EncounterDesign.create({
				downtimes: [DesignDowntime.create({ targetIndex: 0, start: { at: 0.25, jitter: 0 }, end: { at: 0.5, jitter: 0 } })],
				addWaves: [AddWave.create({ adds: [DYNAMIC_ADD], spawn: { at: 0.375, jitter: 0 }, lifetime: 0.25 })],
			}),
		});
		const dialog = await openDesigner();
		const issues = within(dialog).getByRole('region', { name: 'Issues' });
		expect(issues.textContent).toContain(`Every target is down from ${baseSeconds * 0.25}s to ${baseSeconds * 0.375}s, so the player idles.`);
	});
});

describe('My encounters', () => {
	const own = (name: string, duration: number) =>
		EncounterProto.create({
			...Encounter.defaultEncounterProto(),
			name,
			duration,
			design: EncounterDesign.create({
				downtimes: [DesignDowntime.create({ targetIndex: 0, start: { at: 0.2, jitter: 0 }, end: { at: 0.3, jitter: 0 } })],
			}),
		});

	const choice = () => screen.getByLabelText('Encounter');
	const pick = async (name: string) => {
		await userEvent.click(choice());
		await userEvent.click(screen.getByRole('option', { name }));
	};
	const readList = async <T,>(read: () => T, field = choice()) => {
		await userEvent.click(field);
		const value = read();
		await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
		return value;
	};
	const ownGroup = () => screen.queryAllByRole('group').find(group => within(group).queryByText('My encounters'));
	// The options the list files under My encounters, or none when it has no such group.
	const listedOwn = (field?: HTMLElement) =>
		readList(() => {
			const group = ownGroup();
			return group
				? within(group)
						.getAllByRole('option')
						.map(option => option.textContent)
				: [];
		}, field);
	const checked = () =>
		readList(() =>
			screen
				.getAllByRole('option')
				.filter(option => option.getAttribute('aria-selected') === 'true')
				.map(option => option.textContent),
		);
	const nameField = (dialog: HTMLElement) => within(dialog).getByLabelText('Name') as HTMLInputElement;

	it('starts a named blank encounter from the +, and Save adds it to My encounters and selects it', async () => {
		const { encounter } = await setup();
		await pick('Default/Dynamic Adds');

		await clickButton('New encounter');
		const dialog = screen.getByTestId('encounter-designer');
		expect(nameField(dialog).value).toBe('New encounter');
		expect(within(lane('Add waves')).queryAllByRole('slider')).toEqual([]);
		expect(encounter.getTargets()).toEqual([Encounter.defaultTargetProto()]);
		expect(within(dialog).queryByRole('button', { name: 'Save as copy' })).toBeNull();

		await commit(nameField(dialog), 'Ragnaros practice');
		await add('Boss', 'Downtime');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['Ragnaros practice']);
		expect(EncounterProto.fromJson(stored()['Ragnaros practice'] as never).design!.downtimes).toHaveLength(1);
		expect(choice().textContent).toBe('Ragnaros practice');
		expect(await listedOwn()).toEqual(['Ragnaros practice']);
	});

	it('puts back what was selected on Cancel of a new encounter, and stores nothing', async () => {
		const { encounter } = await setup();
		await pick('Default/Dynamic Adds');

		await clickButton('New encounter');
		await clickButton('Cancel');

		expect(stored()).toEqual({});
		expect(encounter.getName()).toBe('');
		expect(choice().textContent).toBe('Default/Dynamic Adds');
	});

	it('loads an entry picked from the list, and the designer opens on it again after another preset', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();

		await pick('Two adds');
		expect(encounter.getDuration()).toBe(200);
		await pick('Default/Dynamic Adds');
		expect(encounter.getName()).toBe('');
		await pick('Two adds');

		const dialog = await openDesigner();
		expect(nameField(dialog).value).toBe('Two adds');
		expect(within(lane('Boss')).getAllByRole('slider')).toHaveLength(1);
		expect(within(dialog).getByRole('button', { name: 'Save as copy' })).toBeTruthy();
		expect(within(dialog).getByRole('button', { name: 'Delete' })).toBeTruthy();
	});

	it('marks the picked entry edited once a field outside the designer changes it', async () => {
		storeEncounters([own('Two adds', 200)]);
		await setup();
		await pick('Two adds');
		expect(choice().textContent).toBe('Two adds');

		await commit(screen.getByLabelText('Duration'), '240');
		expect(choice().textContent).toBe('Two adds (edited)');
	});

	it('updates the entry in place on Save, its new name included, and keeps its place in the list', async () => {
		storeEncounters([own('First', 200), own('Second', 300)]);
		await setup();
		await pick('First');

		const dialog = await openDesigner();
		await commit(nameField(dialog), 'First, renamed');
		await commit(within(dialog).getByLabelText('Duration'), '210');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['First, renamed', 'Second']);
		expect(EncounterProto.fromJson(stored()['First, renamed'] as never).duration).toBe(210);
		expect(choice().textContent).toBe('First, renamed');
	});

	it('adds a new entry under a free name on Save as copy, selects it, and leaves the original as it was', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();
		await pick('Two adds');

		const dialog = await openDesigner();
		await commit(within(dialog).getByLabelText('Duration'), '260');
		await clickButton('Save as copy');

		expect(Object.keys(stored())).toEqual(['Two adds', 'Two adds (2)']);
		expect(EncounterProto.fromJson(stored()['Two adds'] as never).duration).toBe(200);
		expect(EncounterProto.fromJson(stored()['Two adds (2)'] as never).duration).toBe(260);
		expect(encounter.getName()).toBe('Two adds (2)');
		expect(choice().textContent).toBe('Two adds (2)');
	});

	it('deletes the entry after a confirm inside the designer, and falls back to the default encounter', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();
		await pick('Two adds');

		await openDesigner();
		await clickButton('Delete');
		const confirm = screen.getByTestId('encounter-designer-delete-confirm');
		expect(confirm.textContent).toContain('Delete Two adds from My encounters?');
		expect(stored()).not.toEqual({});
		await act(async () => fireEvent.click(within(confirm).getByRole('button', { name: 'Delete' })));

		expect(stored()).toEqual({});
		expect(screen.queryByTestId('encounter-designer')).toBeNull();
		expect(EncounterProto.equals(encounter.toProto(), Encounter.defaultEncounterProto())).toBe(true);
		expect(await listedOwn()).toEqual([]);
	});

	it('keeps a preset out of My encounters on Save, and adds a named encounter from an import', async () => {
		const { encounter } = await setup();
		await pick('Default/Dynamic Adds');
		await openDesigner();
		await clickButton('Save');
		expect(stored()).toEqual({});

		act(() => encounter.fromProto(own('Shared with me', 150)));
		expect(choice().textContent).toBe('Shared with me');
		expect(await listedOwn()).toEqual([]);
		const dialog = await openDesigner();
		expect(nameField(dialog).value).toBe('Shared with me');
		await clickButton('Save');
		expect(Object.keys(stored())).toEqual(['Shared with me']);
	});

	it("lists only the spec's own entries", async () => {
		storeEncounters([own('Two adds', 200)], 'mage');
		await setup();
		expect(await listedOwn()).toEqual([]);
	});

	it('lists My encounters in the Advanced dialog too, loads the entry picked there, and marks it edited the same way', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();

		await clickButton('Advanced');
		const dialog = screen.getByTestId('advanced-encounter-picker-modal');
		const field = () => within(dialog).getByLabelText('Encounter');
		expect(await listedOwn(field())).toEqual(['Two adds']);
		await userEvent.click(field());
		await userEvent.click(within(ownGroup()!).getByRole('option', { name: 'Two adds' }));

		expect(encounter.getDuration()).toBe(200);
		expect(field().textContent).toBe('Two adds');
		await commit(within(dialog).getByLabelText('Duration'), '230');
		expect(field().textContent).toBe('Two adds (edited)');
	});

	it('reads an import under the name of an entry as imported, and Save adds it under a free name and leaves the entry as it was', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter, host } = await setup();
		await pick('Two adds');
		expect(await checked()).toEqual(['Two adds']);

		const shared = IndividualSimSettings.create({ player: PlayerProto.create(), encounter: { ...own('Two adds', 150), myEncounter: true } });
		await act(() => JSON_IMPORTER.onImport(host, IndividualSimSettings.toJsonString(shared)));
		expect(encounter.getDuration()).toBe(150);
		expect(choice().textContent).toBe('Two adds (imported)');
		expect(await checked()).toEqual([]);

		const dialog = await openDesigner();
		expect(within(dialog).queryByRole('button', { name: 'Delete' })).toBeNull();
		await clickButton('Save');
		expect(Object.keys(stored())).toEqual(['Two adds', 'Two adds (2)']);
		expect(EncounterProto.fromJson(stored()['Two adds'] as never).duration).toBe(200);
		expect(EncounterProto.fromJson(stored()['Two adds (2)'] as never).duration).toBe(150);
		expect(choice().textContent).toBe('Two adds (2)');

		await pick('Two adds');
		expect(choice().textContent).toBe('Two adds');
		expect(await checked()).toEqual(['Two adds']);
	});

	it('keeps the picked entry selected through a reload of the autosaved settings', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { host, view } = await setup();
		await pick('Two adds');
		await commit(screen.getByLabelText('Duration'), '230');
		view.unmount();

		// What the page autosaves, loaded into a fresh page.
		const env = memoryEnv();
		env.storage.setItem(STORAGE_KEY, IndividualSimSettings.toJsonString(host.toProto()));
		const reloaded = await makeIndividualSim({ env });
		loadSettings(reloaded);
		renderPicker(reloaded);

		expect(reloaded.sim.encounter.getDuration()).toBe(230);
		expect(choice().textContent).toBe('Two adds (edited)');
		expect(await checked()).toEqual(['Two adds']);
	});
});
