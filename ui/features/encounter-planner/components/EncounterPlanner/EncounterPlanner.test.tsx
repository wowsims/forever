import { EncounterPicker } from '@features/encounter/components/EncounterPicker';
import { JSON_IMPORTER } from '@features/import-export/importers/json';
import { Player as PlayerProto } from '@generated/proto/api';
import {
	AddWave,
	Encounter as EncounterProto,
	EncounterPlan,
	HealthKeyframe,
	MobType,
	PlanBossAbility_Kind,
	PlanDowntime,
	PresetEncounter,
	PresetTarget,
	Stat,
	Target,
} from '@generated/proto/common';
import { IndividualSimSettings, UIDatabase } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { CURRENT_API_VERSION } from '@sim/constants/other';
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
					plan: EncounterPlan.create({ addWaves: [AddWave.create({ name: 'Add', adds: [DYNAMIC_ADD] })] }),
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

const renderPicker = ({ sim, player }: Awaited<ReturnType<typeof makeIndividualSim>>, spec = 'warrior', execute = false) =>
	renderWithHost(<EncounterPicker showExecuteProportion={execute} />, { sim, player, getStorageKey: specKey(spec), individualConfig: { defaults: {} } });

const setup = async ({ targets = 1, plan = EncounterPlan.create(), spec = 'warrior', execute = false } = {}) => {
	const built = await makeIndividualSim();
	const { sim } = built;
	sim.encounter.setTargets(Array.from({ length: targets }, (_, index) => Target.create({ id: index + 1, name: `Target ${index + 1}` })));
	sim.encounter.setPlan(plan);
	const view = renderPicker(built, spec, execute);
	return { encounter: sim.encounter, baseSeconds: sim.encounter.getDuration(), host: built.host, view };
};

const openPlanner = async () => {
	await act(async () => fireEvent.click(screen.getByRole('button', { name: 'Encounter planner' })));
	return screen.getByTestId('encounter-planner');
};

const lane = (name: string) => screen.getByRole('group', { name });

const add = (laneName: string, event: string) => act(() => void fireEvent.click(within(lane(laneName)).getByRole('button', { name: `Add ${event}` })));

const press = (bar: HTMLElement, key: string, shiftKey = false, ctrlKey = false) => act(() => void fireEvent.keyDown(bar, { key, shiftKey, ctrlKey }));

const clickButton = (name: string) => act(async () => fireEvent.click(screen.getByRole('button', { name })));

const clickFirst = (name: string) => act(async () => fireEvent.click(screen.getAllByRole('button', { name })[0]));

const commit = (input: HTMLElement, value: string) => act(() => void fireEvent.change(input, { target: { value } }));

const selectTarget = (name: string) => act(async () => fireEvent.click(within(lane('Targets')).getByRole('button', { name })));

const targetPanel = (name: string) => screen.getByRole('region', { name });

// The first target is the one the planner never removes.
const removeSecondTarget = async () => {
	await selectTarget('Target 2');
	await act(async () => fireEvent.click(within(targetPanel('Target 2')).getByRole('button', { name: 'Remove' })));
};

const downtimeOn = (targetIndex: number) =>
	EncounterPlan.create({ downtimes: [PlanDowntime.create({ targetIndex, start: { at: 0.2, jitter: 0 }, end: { at: 0.3, jitter: 0 } })] });

describe('EncounterPlanner', () => {
	it("offers the add of a preset encounter's wave as a template, and edits the add's stats in the saved plan", async () => {
		{
			const { encounter } = await setup();
			const dialog = await openPlanner();

			await add('Add waves', 'Wave');
			await clickButton('Edit add');
			const addDialog = screen.getByTestId('encounter-planner-add-dialog');
			expect(within(addDialog).getByLabelText('Add template').textContent).toContain('Default/Dynamic Add');
			expect((within(addDialog).getByLabelText('Health') as HTMLInputElement).value).toBe('5000');
			await commit(within(addDialog).getByLabelText('Health'), '7000');
			expect(within(addDialog).getByLabelText('Add template').textContent).toContain('Custom');
			await clickButton('Done');

			expect(within(dialog).queryByTestId('encounter-planner-add-dialog')).toBeNull();
			await clickButton('Save');
			const [wave] = encounter.getPlan().addWaves;
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

	it('edits a new add in a dialog over the planner: its level sets its armor, Escape hands focus back, and Save keeps the edits', async () => {
		const { encounter } = await setup();
		await openPlanner();
		await add('Add waves', 'Wave');

		await clickButton('Add an add');
		let addDialog = screen.getByTestId('encounter-planner-add-dialog');
		expect(within(addDialog).getByTestId('sim-dialog-title').textContent).toBe('Dynamic Add');
		await pickLevel(addDialog, 61);
		expect(CREATURE_ARMOR_BY_LEVEL[61]).toBeGreaterThan(0);
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[61]));
		await commit(within(addDialog).getByLabelText('Health'), '9000');
		await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
		await waitFor(() => expect(screen.queryByTestId('encounter-planner-add-dialog')).toBeNull());
		expect(screen.getByTestId('encounter-planner')).toBeTruthy();

		const edit = screen.getAllByRole('button', { name: 'Edit add' })[1];
		await act(async () => {
			edit.focus();
			fireEvent.click(edit);
		});
		addDialog = screen.getByTestId('encounter-planner-add-dialog');
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[61]));
		await act(async () => void fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' }));
		await waitFor(() => expect(screen.queryByTestId('encounter-planner-add-dialog')).toBeNull());
		expect(document.activeElement).toBe(edit);

		await clickButton('Save');
		const [first, second] = encounter.getPlan().addWaves[0].adds;
		expect(first).toEqual(DYNAMIC_ADD);
		expect(second.level).toBe(61);
		expect(second.stats[Stat.StatArmor]).toBe(CREATURE_ARMOR_BY_LEVEL[61]);
		expect(second.stats[Stat.StatHealth]).toBe(9000);
	});

	it("sets the armor of the Dynamic Adds preset's add to the game files' creature armor of the level picked", async () => {
		const { encounter } = await setup();
		await userEvent.click(screen.getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		await openPlanner();

		await act(async () => fireEvent.click(within(lane('Add waves')).getAllByRole('slider')[0]));
		await clickButton('Edit add');
		const addDialog = screen.getByTestId('encounter-planner-add-dialog');
		expect(within(addDialog).getByLabelText('Level').textContent).toContain('63');
		await pickLevel(addDialog, 60);
		expect(armorField(addDialog)).toBe(String(CREATURE_ARMOR_BY_LEVEL[60]));

		await clickButton('Done');
		await clickButton('Save');
		const [savedAdd] = encounter.getPlan().addWaves[0].adds;
		expect(savedAdd.level).toBe(60);
		expect(savedAdd.stats[Stat.StatArmor]).toBe(CREATURE_ARMOR_BY_LEVEL[60]);
	});

	it('adds a wave in the draft, moves and edits it, and saves it to the store as fractions', async () => {
		const { encounter, baseSeconds } = await setup();
		await openPlanner();

		await add('Add waves', 'Wave');
		const bar = within(lane('Add waves')).getByRole('slider');
		expect(bar.hasAttribute('data-selected')).toBe(true);
		expect(encounter.getPlan().addWaves).toEqual([]);

		await press(bar, 'ArrowRight', true);
		await press(bar, 'ArrowRight');
		await clickButton('Add an add');
		await clickButton('Done');
		await clickFirst('Duplicate add');
		await clickFirst('Remove add');
		expect(encounter.getPlan().addWaves).toEqual([]);

		await clickButton('Save');
		expect(screen.queryByTestId('encounter-planner')).toBeNull();
		const [wave] = encounter.getPlan().addWaves;
		expect(wave.spawn!.at).toBeCloseTo(0.4 + 6 / baseSeconds, 9);
		expect(wave.lifetime).toBeCloseTo(0.1, 9);
		expect(wave.adds).toEqual([DYNAMIC_ADD, DYNAMIC_ADD]);
	});

	it("applies the raid's debuffs to a new wave's adds, edits their delay in seconds, and hides delay and jitter once unticked", async () => {
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await add('Add waves', 'Wave');

		const apply = within(dialog).getByRole('checkbox', { name: 'Raid debuffs' });
		expect(apply.getAttribute('aria-checked')).toBe('true');
		await commit(within(dialog).getByLabelText('Debuff delay (s)'), '3');
		expect(within(dialog).getByLabelText('Debuff jitter (s)')).toBeTruthy();

		await act(async () => fireEvent.click(apply));
		expect(within(dialog).queryByLabelText('Debuff delay (s)')).toBeNull();
		expect(within(dialog).queryByLabelText('Debuff jitter (s)')).toBeNull();

		await clickButton('Save');
		const [wave] = encounter.getPlan().addWaves;
		expect([wave.holdDebuffs, wave.debuffDelay, wave.debuffJitter]).toEqual([true, 3, 0]);
	});

	describe('event bars', () => {
		// happy-dom has no pointer capture, which a drag needs to start.
		beforeEach(() => {
			vi.spyOn(Element.prototype, 'setPointerCapture').mockImplementation(() => {});
			vi.spyOn(Element.prototype, 'hasPointerCapture').mockReturnValue(false);
		});

		const bars = () => within(lane('Target 1')).getAllByRole('slider');
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
			const dialog = await openPlanner();
			await add('Target 1', 'Downtime');
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
			drag(lane('Target 1'), -200);
			expect(timeline.scrollLeft).toBe(700);
			expect(bars().map(startOf)).toEqual([first]);
			expect(bars()[0].hasAttribute('data-selected')).toBe(true);

			drag(bars()[0], 300);
			expect(timeline.scrollLeft).toBe(700);
			expect(bars().map(startOf)).toEqual([first + 3]);
		});

		it('moves only the keyframe at the end of a boss health segment, and leaves the implied start alone', async () => {
			await setup();
			await openPlanner();
			await add('Target 1', 'Boss health');
			const text = () => bars()[0].getAttribute('aria-valuetext') ?? '';
			expect(text()).toContain('from 0s to 72s');

			drag(bars()[0], 300);
			expect(text()).toContain('from 0s to 75s');

			// The first segment starts at the implied 0 s, which has no keyframe to move.
			const before = text();
			act(() => void fireEvent.keyDown(bars()[0], { key: 'ArrowLeft', altKey: true, shiftKey: true }));
			expect(text()).toBe(before);
		});

		const healthKeyframes = (...points: Array<[number, number]>) =>
			EncounterPlan.create({ bossHealth: points.map(([at, healthPercent]) => HealthKeyframe.create({ at, healthPercent })) });

		it('drags a boss health segment by its body as far left as its neighbour lets its keyframe go', async () => {
			const { encounter, baseSeconds } = await setup({ plan: healthKeyframes([0.5, 50], [0.6, 40]) });
			await openPlanner();

			// A drag the whole fight to the left, far past the segment's own length.
			drag(bars()[1], -(baseSeconds * 100));
			await clickButton('Save');
			const [before, moved] = encounter.getPlan().bossHealth;
			expect(moved.at).toBeCloseTo(0.5 + 1 / baseSeconds, 9);
			expect(before.at).toBe(0.5);
		});

		it('leaves the neighbour of a boss health segment alone while a copy of the segment is dragged', async () => {
			await setup({ plan: healthKeyframes([0.5, 50], [0.75, 25]) });
			await openPlanner();
			const [first, second] = bars();
			const before = second.getAttribute('aria-valuetext');
			const pointer = { pointerId: 1, pointerType: 'mouse', button: 0, altKey: true };

			act(() => void fireEvent.pointerDown(first, { ...pointer, clientX: 100 }));
			act(() => void fireEvent.pointerMove(first, { ...pointer, clientX: 400 }));
			expect(second.getAttribute('aria-valuetext')).toBe(before);
			act(() => void fireEvent.pointerUp(first, { ...pointer, clientX: 400 }));
		});

		it('moves a move whose walk runs past the fight end to the right on Right arrow', async () => {
			const { baseSeconds } = await setup();
			await openPlanner();
			await add('Movement', 'Move');
			const bar = within(lane('Movement')).getByRole('slider');
			await commit(screen.getByLabelText('Yards'), String(baseSeconds * 7));
			const start = Number(bar.getAttribute('aria-valuenow'));

			await press(bar, 'ArrowRight');
			expect(Number(bar.getAttribute('aria-valuenow'))).toBe(start + 1);
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
			input = within(lane('Target 1')).getByRole('textbox', { name: 'Event name' });
			await commit(input, 'Dropped');
			await act(() => void fireEvent.keyDown(input, { key: 'Escape' }));
			expect(bars()[0].getAttribute('aria-label')).toBe('Phase 2');

			await press(bars()[0], 'F2');
			input = within(lane('Target 1')).getByRole('textbox', { name: 'Event name' });
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
			const input = within(lane('Target 1')).getByRole('textbox', { name: 'Event name' });
			await commit(input, 'From the menu');
			await act(() => void fireEvent.keyDown(input, { key: 'Enter' }));
			await clickButton('Save');
			expect(encounter.getPlan().downtimes[0].name).toBe('From the menu');
		});

		const keyframe = (at: number, healthPercent: number) => HealthKeyframe.create({ at, healthPercent });
		const bossDowntime = (start: number, end: number) =>
			PlanDowntime.create({ targetIndex: 0, start: { at: start, jitter: 0 }, end: { at: end, jitter: 0 } });
		const openMenu = async (bar: HTMLElement) => act(async () => void fireEvent.contextMenu(bar, { clientX: 10, clientY: 10 }));

		it('holds the boss health across a boss downtime from its context menu, in one change', async () => {
			const { encounter } = await setup({ plan: EncounterPlan.create({ downtimes: [bossDowntime(0.4, 0.5)], bossHealth: [keyframe(0.8, 60)] }) });
			await openPlanner();
			const downtime = bars().find(bar => bar.getAttribute('aria-label') === 'Downtime 1')!;
			await openMenu(downtime);
			await act(async () => fireEvent.click(await screen.findByRole('menuitem', { name: 'Hold boss health here' })));
			expect(within(lane('Target 1')).getByText('Hold 80 %')).toBeTruthy();
			await clickButton('Save');
			expect(encounter.getPlan().bossHealth.map(entry => [entry.at, entry.healthPercent])).toEqual([
				[0.8, 60],
				[0.4, 80],
				[0.5, 80],
			]);
		});

		it('disables the hold while a keyframe lies inside the downtime, and says which', async () => {
			await setup({ plan: EncounterPlan.create({ downtimes: [bossDowntime(0.4, 0.5)], bossHealth: [keyframe(0.45, 80)] }) });
			await openPlanner();
			await openMenu(bars().find(bar => bar.getAttribute('aria-label') === 'Downtime 1')!);
			const item = await screen.findByRole('menuitem', { name: 'Hold boss health here' });
			expect(item.hasAttribute('data-disabled')).toBe(true);
			expect(item.getAttribute('data-tip')).toBe('Boss health 1 at 81s is in the way');
		});

		it('offers the hold of a boss downtime only for the boss', async () => {
			await setup({
				targets: 2,
				plan: EncounterPlan.create({
					downtimes: [PlanDowntime.create({ targetIndex: 1, start: { at: 0.4, jitter: 0 }, end: { at: 0.5, jitter: 0 } })],
				}),
			});
			await openPlanner();
			await openMenu(within(lane('Target 2')).getByRole('slider'));
			await screen.findByRole('menuitem', { name: 'Duplicate' });
			expect(screen.queryByRole('menuitem', { name: 'Hold boss health here' })).toBeNull();
		});

		it('cuts a flat part off a boss health segment from its context menu', async () => {
			await setup({ plan: EncounterPlan.create({ bossHealth: [keyframe(0.5, 50)] }) });
			await openPlanner();
			await openMenu(bars()[0]);
			await act(async () => fireEvent.click(await screen.findByRole('menuitem', { name: 'Hold health' })));
			expect(bars()).toHaveLength(2);
			expect(within(lane('Target 1')).getByText('Hold 100 %')).toBeTruthy();
		});

		it('opens no menu on empty lane space', async () => {
			await withDowntime();
			await act(
				async () =>
					void fireEvent.contextMenu(within(lane('Target 1')).getByRole('slider').parentElement!.firstElementChild!, { clientX: 5, clientY: 5 }),
			);
			expect(screen.queryByRole('menuitem')).toBeNull();
		});
	});

	it('focuses the first add button on open, and moves focus into the event panel on Enter, not on a click', async () => {
		await setup();
		await openPlanner();
		expect(document.activeElement).toBe(within(lane('Target 1')).getByRole('button', { name: 'Add Downtime' }));

		await add('Target 1', 'Downtime');
		const bar = within(lane('Target 1')).getByRole('slider');
		const panel = screen.getByRole('region', { name: 'Downtime 1' });
		act(() => {
			bar.focus();
			fireEvent.click(bar);
		});
		expect(document.activeElement).toBe(bar);
		await press(bar, 'Enter');
		expect(document.activeElement).toBe(panel);
	});

	it('turns a new boss ability into a pulse with working defaults, its own fields and no target, and saves it', async () => {
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await add('Abilities', 'Ability');
		expect(within(dialog).queryByLabelText('Target')).not.toBeNull();

		await userEvent.click(within(dialog).getByLabelText('Kind'));
		await userEvent.click(screen.getByRole('option', { name: 'Pulse' }));

		expect(within(dialog).queryByLabelText('Target')).toBeNull();
		expect(Number((within(dialog).getByLabelText('Tick every (s)') as HTMLInputElement).value)).toBe(1);
		expect(Number((within(dialog).getByLabelText('Max damage') as HTMLInputElement).value)).toBe(700);
		expect(within(dialog).getByRole('region', { name: 'Issues' }).textContent).toContain('No issues.');

		await commit(within(dialog).getByLabelText('Tick every (s)'), '0');
		expect(within(dialog).getByRole('region', { name: 'Issues' }).textContent).toContain(
			english.settings_tab.encounter.planner.validation.no_tick_interval,
		);
		await commit(within(dialog).getByLabelText('Tick every (s)'), '2');
		await clickButton('Save');
		expect(encounter.getPlan().bossAbilities[0]).toMatchObject({
			kind: PlanBossAbility_Kind.Pulse,
			tickInterval: 2,
			tickJitter: 0.2,
			minDamage: 500,
			resistible: true,
			avoidable: false,
		});
	});

	it('puts back the duration, the targets and the plan on Cancel, and starts the next opening from the store', async () => {
		const plan = downtimeOn(2);
		const { encounter, baseSeconds } = await setup({ targets: 3, plan });
		const targets = encounter.getTargets();
		const dialog = await openPlanner();

		await commit(within(dialog).getByLabelText('Duration'), String(baseSeconds + 60));
		await removeSecondTarget();
		await add('Target 1', 'Downtime');
		expect(encounter.getDuration()).toBe(baseSeconds + 60);
		expect(encounter.getTargets()).toHaveLength(2);

		await clickButton('Cancel');
		expect(screen.queryByTestId('encounter-planner')).toBeNull();
		expect(encounter.getDuration()).toBe(baseSeconds);
		expect(encounter.getTargets()).toBe(targets);
		expect(EncounterPlan.equals(encounter.getPlan(), plan)).toBe(true);

		await openPlanner();
		expect(within(lane('Target 3')).getAllByRole('slider')).toHaveLength(1);
	});

	it('follows a removed target in the draft the way the store does', async () => {
		const { encounter } = await setup({ targets: 3, plan: downtimeOn(2) });
		await openPlanner();

		await removeSecondTarget();
		expect(
			within(lane('Targets'))
				.getAllByRole('button', { name: /^Target \d$/ })
				.map(bar => bar.textContent),
		).toEqual(['Target 1', 'Target 3']);
		expect(screen.queryByRole('region', { name: 'Target 2' })).toBeNull();
		await clickButton('Save');
		expect(encounter.getPlan().downtimes.map(downtime => downtime.targetIndex)).toEqual([1]);
	});

	describe('execute proportions', () => {
		const panel = () => screen.getByTestId('encounter-picker-root');

		it("shows a simple fight's execute fields on the Encounter panel until a saved plan makes it a fight with a plan", async () => {
			const { encounter } = await setup({ execute: true });
			expect(within(panel()).queryByLabelText('Execute Duration 20 (%)')).not.toBeNull();

			const dialog = await openPlanner();
			const field = within(dialog).getByLabelText('Execute Duration 20 (%)');
			await commit(field, '25');
			expect(encounter.getExecuteProportion20()).toBe(0.25);

			await add('Target 1', 'Downtime');
			expect(within(panel()).queryByLabelText('Execute Duration 20 (%)')).not.toBeNull();
			await clickButton('Save');
			expect(within(panel()).queryByLabelText('Execute Duration 20 (%)')).toBeNull();
			expect(encounter.getExecuteProportion20()).toBe(0.25);

			const reopened = await openPlanner();
			expect(within(reopened).queryByLabelText('Execute Duration 20 (%)')).not.toBeNull();
		});

		it('keeps the execute fields on the Encounter panel after Cancel, and puts back the proportions changed in the header', async () => {
			const { encounter } = await setup({ execute: true });
			const proportion = encounter.getExecuteProportion20();
			const dialog = await openPlanner();
			await commit(within(dialog).getByLabelText('Execute Duration 20 (%)'), '25');
			await add('Target 1', 'Downtime');
			await clickButton('Cancel');

			expect(within(panel()).queryByLabelText('Execute Duration 20 (%)')).not.toBeNull();
			expect(encounter.getExecuteProportion20()).toBe(proportion);
		});

		it('leaves the execute fields out for a spec that does not read them', async () => {
			await setup();
			const dialog = await openPlanner();
			expect(within(dialog).queryByLabelText('Execute Duration 20 (%)')).toBeNull();
		});
	});

	describe('target rows', () => {
		const swingSpeed = (panel: HTMLElement) => within(panel).getByLabelText('Swing Speed') as HTMLInputElement;
		const rows = () =>
			within(lane('Targets'))
				.getAllByRole('group')
				.map(row => row.getAttribute('aria-label'));
		const barsOn = (row: HTMLElement) =>
			within(row)
				.queryAllByRole('slider')
				.map(bar => bar.getAttribute('aria-label'));
		const onTarget = (targetIndex: number, start: number) =>
			PlanDowntime.create({ targetIndex, start: { at: start, jitter: 0 }, end: { at: start + 0.1, jitter: 0 } });

		it('puts each downtime on its target row and the boss health on the first', async () => {
			await setup({
				targets: 2,
				plan: EncounterPlan.create({
					downtimes: [onTarget(0, 0.2), onTarget(1, 0.4)],
					bossHealth: [HealthKeyframe.create({ at: 0.5, healthPercent: 50 })],
				}),
			});
			await openPlanner();

			expect(rows()).toEqual(['Target 1', 'Target 2']);
			expect(barsOn(lane('Target 1'))).toEqual(['Downtime 1', 'Boss health 1']);
			expect(barsOn(lane('Target 2'))).toEqual(['Downtime 2']);
			expect(within(lane('Target 1')).queryByRole('button', { name: 'Add Boss health' })).not.toBeNull();
			expect(within(lane('Target 2')).queryByRole('button', { name: 'Add Boss health' })).toBeNull();
		});

		it('names the abilities lane Abilities', async () => {
			await setup({ targets: 2 });
			await openPlanner();

			expect(lane('Abilities')).toBeTruthy();
			expect(screen.queryByRole('group', { name: 'Boss abilities' })).toBeNull();
			expect(screen.queryByRole('group', { name: 'Boss' })).toBeNull();
		});

		it("moves focus with ArrowDown and ArrowUp among the bars of the pressed bar's row only", async () => {
			await setup({
				targets: 2,
				plan: EncounterPlan.create({ downtimes: [onTarget(0, 0.2), onTarget(0, 0.4), onTarget(1, 0.6)] }),
			});
			await openPlanner();
			const [first, second] = within(lane('Target 1')).getAllByRole('slider');
			const other = within(lane('Target 2')).getByRole('slider');

			act(() => first.focus());
			press(first, 'ArrowDown');
			expect(document.activeElement).toBe(second);
			press(second, 'ArrowDown');
			expect(document.activeElement).toBe(second);
			press(second, 'ArrowUp');
			expect(document.activeElement).toBe(first);
			press(first, 'ArrowUp');
			expect(document.activeElement).toBe(first);
			expect(document.activeElement).not.toBe(other);
		});

		it('adds a downtime on the target of the row whose + Downtime was pressed', async () => {
			const { encounter } = await setup({ targets: 2 });
			await openPlanner();

			await add('Target 2', 'Downtime');
			expect(barsOn(lane('Target 2'))).toEqual(['Downtime 1']);
			expect(barsOn(lane('Target 1'))).toEqual([]);
			await clickButton('Save');
			expect(encounter.getPlan().downtimes.map(downtime => downtime.targetIndex)).toEqual([1]);
		});

		it("opens a target's settings from its row label, and moves focus into them on Enter", async () => {
			await setup({ targets: 2 });
			await openPlanner();
			const label = within(lane('Target 2')).getByRole('button', { name: 'Target 2' });

			await act(async () => fireEvent.click(label));
			expect(label.getAttribute('aria-pressed')).toBe('true');
			const panel = targetPanel('Target 2');
			expect(document.activeElement).not.toBe(panel);

			await press(label, 'Enter');
			expect(document.activeElement).toBe(panel);
		});

		it("keeps a downtime on its target's row as the target moves up and is copied", async () => {
			await setup({ targets: 2, plan: downtimeOn(1) });
			await openPlanner();

			await selectTarget('Target 2');
			await act(async () => fireEvent.click(within(targetPanel('Target 2')).getByRole('button', { name: 'Move up' })));
			expect(rows()).toEqual(['Target 2', 'Target 1']);
			expect(barsOn(within(lane('Targets')).getAllByRole('group')[0])).toEqual(['Downtime 1']);

			await act(async () => fireEvent.click(within(targetPanel('Target 2')).getByRole('button', { name: 'Duplicate' })));
			expect(rows()).toEqual(['Target 2', 'Target 1', 'Target 2 (2)']);
			expect(
				within(lane('Targets'))
					.getAllByRole('group')
					.map(row => barsOn(row).length),
			).toEqual([1, 0, 0]);
		});

		it("keeps the downtime of a removed target on the first target's row, marked, and selectable", async () => {
			await setup({ targets: 2, plan: downtimeOn(1) });
			await openPlanner();
			await removeSecondTarget();

			const bar = within(lane('Target 1')).getByRole('slider', { name: 'Downtime 1' });
			expect(bar.hasAttribute('data-invalid')).toBe(true);
			await act(async () => fireEvent.click(bar));
			expect(bar.hasAttribute('data-selected')).toBe(true);
			expect(screen.getByRole('region', { name: 'Downtime 1' })).toBeTruthy();
		});

		it("edits a target's swing speed in the store as it changes, and Cancel puts it back", async () => {
			const { encounter } = await setup({ targets: 2 });
			act(() => encounter.sim.setShowThreatMetrics(true));
			const before = encounter.getTargets();
			await openPlanner();

			await selectTarget('Target 2');
			const panel = targetPanel('Target 2');
			expect(within(lane('Targets')).getByRole('button', { name: 'Target 2' }).getAttribute('aria-pressed')).toBe('true');
			await commit(swingSpeed(panel), '1.4');
			expect(encounter.getTargets()[1].swingSpeed).toBe(1.4);
			expect(Number(swingSpeed(panel).value)).toBe(1.4);

			await clickButton('Cancel');
			expect(encounter.getTargets()).toBe(before);
			expect(encounter.getTargets()[1].swingSpeed).toBe(0);
		});

		it("shows a target's swing speed only while threat metrics are on, and an add's either way", async () => {
			const { encounter } = await setup();
			act(() => encounter.sim.setShowThreatMetrics(false));
			await openPlanner();
			await add('Add waves', 'Wave');
			await clickButton('Edit add');
			expect(within(screen.getByTestId('encounter-planner-add-dialog')).queryByLabelText('Swing Speed')).not.toBeNull();
			await clickButton('Done');

			await selectTarget('Target 1');
			expect(within(targetPanel('Target 1')).queryByLabelText('Swing Speed')).toBeNull();

			act(() => encounter.sim.setShowThreatMetrics(true));
			expect(within(targetPanel('Target 1')).queryByLabelText('Swing Speed')).not.toBeNull();
		});

		it('offers no Remove and no Move up for the first target', async () => {
			await setup({ targets: 2 });
			await openPlanner();

			await selectTarget('Target 1');
			const panel = targetPanel('Target 1');
			expect(within(panel).queryByRole('button', { name: 'Remove' })).toBeNull();
			expect((within(panel).getByRole('button', { name: 'Move up' }) as HTMLButtonElement).disabled).toBe(true);
			expect((within(panel).getByRole('button', { name: 'Move down' }) as HTMLButtonElement).disabled).toBe(false);
		});

		it("adds a target from the lane's +, moves it up and copies it under a free name, and the plan keeps naming the target it named", async () => {
			const { encounter } = await setup({ targets: 2, plan: downtimeOn(1) });
			await openPlanner();

			await act(async () => fireEvent.click(within(lane('Targets')).getByRole('button', { name: 'Add Target' })));
			expect(encounter.getTargets()).toHaveLength(3);
			expect(encounter.getTargets()[2]).toEqual(Encounter.defaultTargetProto());
			const added = targetPanel('Raid Target');

			await act(async () => fireEvent.click(within(added).getByRole('button', { name: 'Move up' })));
			expect(encounter.getTargets().map(target => target.name)).toEqual(['Target 1', 'Raid Target', 'Target 2']);
			expect(targetPanel('Raid Target')).toBeTruthy();

			await act(async () => fireEvent.click(within(targetPanel('Raid Target')).getByRole('button', { name: 'Duplicate' })));
			expect(encounter.getTargets().map(target => target.name)).toEqual(['Target 1', 'Raid Target', 'Target 2', 'Raid Target (2)']);
			expect(encounter.getTargets()[3]).not.toBe(encounter.getTargets()[1]);
			expect(within(lane('Targets')).getByRole('button', { name: 'Raid Target (2)' }).getAttribute('aria-pressed')).toBe('true');

			await act(async () => fireEvent.click(within(lane('Targets')).getByRole('button', { name: 'Add Target' })));
			expect(encounter.getTargets()[4].name).toBe('Raid Target (3)');

			await clickButton('Save');
			expect(encounter.getPlan().downtimes.map(downtime => downtime.targetIndex)).toEqual([2]);
		});

		it("renames a target from its Name field, and its row follows while its NPC stays the preset's", async () => {
			const { encounter } = await setup();
			act(() => encounter.setTargets([Target.clone(ADD.target!)]));
			await openPlanner();
			await selectTarget('Son of Flame');
			await act(async () => fireEvent.click(within(targetPanel('Son of Flame')).getByRole('button', { name: 'Duplicate' })));
			expect(within(targetPanel('Son of Flame (2)')).getByLabelText('NPC').textContent).toContain('Molten Core/Son of Flame');

			await commit(within(targetPanel('Son of Flame (2)')).getByLabelText('Name') as HTMLInputElement, 'Second flame');
			expect(encounter.getTargets()[1].name).toBe('Second flame');
			expect(within(lane('Targets')).getByRole('button', { name: 'Second flame' }).getAttribute('aria-pressed')).toBe('true');
			expect(within(targetPanel('Second flame')).getByLabelText('NPC').textContent).toContain('Molten Core/Son of Flame');
		});

		// The fields read the target the panel shows, so a new selection or a move with no store write between shows the right one.
		it('shows the NPC and the stats of the target selected, after a switch and after a move', async () => {
			const { encounter } = await setup({ targets: 2 });
			act(() => encounter.applyPresetTarget(encounter.sim.db.getPresetTarget('Molten Core/Son of Flame')!, 0));
			await openPlanner();
			const npc = (panel: HTMLElement) => within(panel).getByLabelText('NPC').textContent;
			const level = (panel: HTMLElement) => within(panel).getByLabelText('Level').textContent;

			await selectTarget('Son of Flame');
			expect(npc(targetPanel('Son of Flame'))).toContain('Molten Core/Son of Flame');
			expect(level(targetPanel('Son of Flame'))).toContain('62');

			await selectTarget('Target 2');
			expect(npc(targetPanel('Target 2'))).toContain('Custom');
			expect(level(targetPanel('Target 2'))).not.toContain('62');

			await act(async () => fireEvent.click(within(targetPanel('Target 2')).getByRole('button', { name: 'Move up' })));
			expect(encounter.getTargets().map(target => target.name)).toEqual(['Target 2', 'Son of Flame']);
			expect(npc(targetPanel('Target 2'))).toContain('Custom');

			await act(async () => fireEvent.click(within(targetPanel('Target 2')).getByRole('button', { name: 'Move down' })));
			await selectTarget('Son of Flame');
			expect(npc(targetPanel('Son of Flame'))).toContain('Molten Core/Son of Flame');
		});

		it("loads a preset NPC into the target picked from the panel's NPC field", async () => {
			const { encounter } = await setup({ targets: 2 });
			await openPlanner();

			await selectTarget('Target 2');
			await userEvent.click(within(targetPanel('Target 2')).getByLabelText('NPC'));
			await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Boss' }));

			expect(encounter.getTargets()[1]).toEqual(encounter.sim.db.getPresetTarget('Default/Dynamic Boss')!.target);
			expect(within(targetPanel('Dynamic Boss')).getByLabelText('NPC').textContent).toContain('Default/Dynamic Boss');
		});
	});

	it('lists an issue and marks its bar, and still saves the plan', async () => {
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await add('Movement', 'Move');
		await add('Movement', 'Move');

		const [first, second] = within(lane('Movement')).getAllByRole('slider');
		expect([first.hasAttribute('data-invalid'), second.hasAttribute('data-invalid')]).toEqual([false, true]);
		const issues = within(dialog).getByRole('region', { name: 'Issues' });
		expect(issues.textContent).toContain(english.settings_tab.encounter.planner.validation.overlap);

		await clickButton('Save');
		expect(encounter.getPlan().moves).toHaveLength(2);
	});

	it('names in seconds the part of a boss downtime no add is up for', async () => {
		const { baseSeconds } = await setup({
			plan: EncounterPlan.create({
				downtimes: [PlanDowntime.create({ targetIndex: 0, start: { at: 0.25, jitter: 0 }, end: { at: 0.5, jitter: 0 } })],
				addWaves: [AddWave.create({ adds: [DYNAMIC_ADD], spawn: { at: 0.375, jitter: 0 }, lifetime: 0.25 })],
			}),
		});
		const dialog = await openPlanner();
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
			plan: EncounterPlan.create({
				downtimes: [PlanDowntime.create({ targetIndex: 0, start: { at: 0.2, jitter: 0 }, end: { at: 0.3, jitter: 0 } })],
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
		const dialog = screen.getByTestId('encounter-planner');
		expect(nameField(dialog).value).toBe('New encounter');
		expect(within(lane('Add waves')).queryAllByRole('slider')).toEqual([]);
		expect(encounter.getTargets()).toEqual([Encounter.defaultTargetProto()]);
		expect(within(dialog).queryByRole('button', { name: 'Save as copy' })).toBeNull();

		await commit(nameField(dialog), 'Ragnaros practice');
		await add('Raid Target', 'Downtime');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['Ragnaros practice']);
		expect(EncounterProto.fromJson(stored()['Ragnaros practice'] as never).plan!.downtimes).toHaveLength(1);
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

	it('loads an entry picked from the list, and the planner opens on it again after another preset', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();

		await pick('Two adds');
		expect(encounter.getDuration()).toBe(200);
		await pick('Default/Dynamic Adds');
		expect(encounter.getName()).toBe('');
		await pick('Two adds');

		const dialog = await openPlanner();
		expect(nameField(dialog).value).toBe('Two adds');
		expect(within(lane('Raid Target')).getAllByRole('slider')).toHaveLength(1);
		expect(within(dialog).getByRole('button', { name: 'Save as copy' })).toBeTruthy();
		expect(within(dialog).getByRole('button', { name: 'Delete' })).toBeTruthy();
	});

	it('reads an entry saved at an older api version as unedited', async () => {
		storeEncounters([{ ...own('Two adds', 200), apiVersion: CURRENT_API_VERSION - 1 }]);
		await setup();
		await pick('Two adds');
		expect(choice().textContent).toBe('Two adds');
	});

	it('marks the picked entry edited once a field outside the planner changes it', async () => {
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

		const dialog = await openPlanner();
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

		const dialog = await openPlanner();
		await commit(within(dialog).getByLabelText('Duration'), '260');
		await clickButton('Save as copy');

		expect(Object.keys(stored())).toEqual(['Two adds', 'Two adds (2)']);
		expect(EncounterProto.fromJson(stored()['Two adds'] as never).duration).toBe(200);
		expect(EncounterProto.fromJson(stored()['Two adds (2)'] as never).duration).toBe(260);
		expect(encounter.getName()).toBe('Two adds (2)');
		expect(choice().textContent).toBe('Two adds (2)');
	});

	it('deletes the entry after a confirm inside the planner, and falls back to the default encounter', async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter } = await setup();
		await pick('Two adds');

		await openPlanner();
		await clickButton('Delete');
		const confirm = screen.getByTestId('encounter-planner-delete-confirm');
		expect(confirm.textContent).toContain('Delete Two adds from My encounters?');
		expect(stored()).not.toEqual({});
		await act(async () => fireEvent.click(within(confirm).getByRole('button', { name: 'Delete' })));

		expect(stored()).toEqual({});
		expect(screen.queryByTestId('encounter-planner')).toBeNull();
		expect(EncounterProto.equals(encounter.toProto(), Encounter.defaultEncounterProto())).toBe(true);
		expect(await listedOwn()).toEqual([]);
	});

	it('adds a preset with a wave to My encounters as New encounter on Save, and a named encounter from an import', async () => {
		const { encounter } = await setup();
		await pick('Default/Dynamic Adds');
		await openPlanner();
		await clickButton('Save');
		expect(Object.keys(stored())).toEqual(['New encounter']);
		expect(choice().textContent).toBe('New encounter');

		act(() => encounter.fromProto(own('Shared with me', 150)));
		expect(choice().textContent).toBe('Shared with me');
		expect(await listedOwn()).toEqual(['New encounter']);
		const dialog = await openPlanner();
		expect(nameField(dialog).value).toBe('Shared with me');
		await clickButton('Save');
		expect(Object.keys(stored())).toEqual(['New encounter', 'Shared with me']);
	});

	it("lists only the spec's own entries", async () => {
		storeEncounters([own('Two adds', 200)], 'mage');
		await setup();
		expect(await listedOwn()).toEqual([]);
	});

	it("lists My encounters in the planner's header too, loads the entry picked there into the timeline, and Cancel puts the encounter back", async () => {
		storeEncounters([own('Two adds', 200)]);
		const { encounter, baseSeconds } = await setup();
		const before = encounter.toProto();

		const dialog = await openPlanner();
		const field = () => within(dialog).getByLabelText('Encounter');
		expect(await listedOwn(field())).toEqual(['Two adds']);
		await userEvent.click(field());
		await userEvent.click(within(ownGroup()!).getByRole('option', { name: 'Two adds' }));

		expect(encounter.getDuration()).toBe(200);
		expect(field().textContent).toBe('Two adds');
		expect(within(lane('Raid Target')).getAllByRole('slider')).toHaveLength(1);
		await commit(within(dialog).getByLabelText('Duration'), '230');
		expect(field().textContent).toBe('Two adds (edited)');

		await clickButton('Cancel');
		expect(EncounterProto.equals(encounter.toProto(), before)).toBe(true);
		expect(encounter.getDuration()).toBe(baseSeconds);
		expect(encounter.isMyEncounter()).toBe(false);
	});

	it("replaces the draft's lanes with the plan of a preset picked in the header, and Cancel puts back the plan from before", async () => {
		const { encounter } = await setup({ targets: 2 });
		const before = encounter.toProto();
		const dialog = await openPlanner();
		await add('Target 1', 'Downtime');

		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		expect(within(lane('Dynamic Boss')).queryAllByRole('slider')).toEqual([]);
		expect(within(lane('Add waves')).getAllByRole('slider')).toHaveLength(1);
		expect(within(lane('Targets')).getAllByRole('button', { name: 'Dynamic Boss' })).toHaveLength(1);

		await clickButton('Cancel');
		expect(EncounterProto.equals(encounter.toProto(), before)).toBe(true);
		expect(encounter.getPlan().addWaves).toEqual([]);
	});

	it('keeps the draft when Custom is picked in the header, which loads nothing', async () => {
		await setup();
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		await add('Dynamic Boss', 'Downtime');

		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Custom' }));
		expect(within(lane('Dynamic Boss')).getAllByRole('slider')).toHaveLength(1);
		expect(within(lane('Add waves')).getAllByRole('slider')).toHaveLength(1);
	});

	it('updates the entry picked in the header on Save, and leaves the one the planner opened on', async () => {
		storeEncounters([own('First', 200), own('Second', 230)]);
		await setup();
		await pick('First');
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Second' }));
		await clickButton('Save');

		expect(Object.keys(stored()).sort()).toEqual(['First', 'Second']);
		expect(choice().textContent).toBe('Second');
	});

	it('stores a preset with a wave picked in the header as New encounter on Save, and leaves the entry the planner opened on as it was', async () => {
		storeEncounters([own('First', 200)]);
		const { encounter } = await setup();
		await pick('First');
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['First', 'New encounter']);
		expect(EncounterProto.fromJson(stored()['First'] as never).duration).toBe(200);
		expect(encounter.isMyEncounter()).toBe(true);
		expect(encounter.getPlan().addWaves.map(wave => wave.name)).toEqual(['Add']);
		expect(choice().textContent).toBe('New encounter');
	});

	it('deletes the entry picked in the header, and leaves the one the planner opened on', async () => {
		storeEncounters([own('First', 200), own('Second', 230)]);
		await setup();
		await pick('First');
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Second' }));
		await clickButton('Delete');
		const confirm = screen.getByTestId('encounter-planner-delete-confirm');
		expect(confirm.textContent).toContain('Delete Second from My encounters?');
		await act(async () => fireEvent.click(within(confirm).getByRole('button', { name: 'Delete' })));

		expect(Object.keys(stored())).toEqual(['First']);
		expect(screen.queryByTestId('encounter-planner')).toBeNull();
	});

	it('keeps the target a downtime names when an entry with more targets is picked in the header', async () => {
		storeEncounters([EncounterProto.create({ ...Encounter.defaultEncounterProto(3), name: 'Three targets', plan: downtimeOn(2) })]);
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Three targets' }));
		await clickButton('Save');

		expect(encounter.getPlan().downtimes.map(downtime => downtime.targetIndex)).toEqual([2]);
	});

	it('saves the plan of a preset picked in the header as New encounter', async () => {
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await userEvent.click(within(dialog).getByLabelText('Encounter'));
		await userEvent.click(screen.getByRole('option', { name: 'Default/Dynamic Adds' }));
		await clickButton('Save');

		expect(encounter.getPlan().addWaves.map(wave => wave.name)).toEqual(['Add']);
		expect(choice().textContent).toBe('New encounter');
	});

	it('stores a preset opened in the planner under the name typed once a downtime is added, lists it and selects it', async () => {
		const { encounter } = await setup();
		await pick('Default/Dynamic Adds');
		const dialog = await openPlanner();
		await commit(nameField(dialog), 'Adds practice');
		await add('Dynamic Boss', 'Downtime');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['Adds practice']);
		expect(EncounterProto.fromJson(stored()['Adds practice'] as never).plan!.downtimes).toHaveLength(1);
		expect(encounter.getName()).toBe('Adds practice');
		expect(await listedOwn()).toEqual(['Adds practice']);
		expect(choice().textContent).toBe('Adds practice');
		expect(await checked()).toEqual(['Adds practice']);
	});

	it('stores the default fight as New encounter once a downtime is added to it', async () => {
		await setup();
		await openPlanner();
		await add('Target 1', 'Downtime');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['New encounter']);
		expect(await listedOwn()).toEqual(['New encounter']);
		expect(choice().textContent).toBe('New encounter');
	});

	it('keeps the default fight out of My encounters when only a target field changes, and puts the change into the sim', async () => {
		const { encounter } = await setup();
		act(() => encounter.sim.setShowThreatMetrics(true));
		await openPlanner();
		await selectTarget('Target 1');
		await commit(within(targetPanel('Target 1')).getByLabelText('Swing Speed'), '1.4');
		await clickButton('Save');

		expect(stored()).toEqual({});
		expect(encounter.getPlan()).toEqual(EncounterPlan.create());
		expect(encounter.getTargets()[0].swingSpeed).toBe(1.4);
		expect(await listedOwn()).toEqual([]);
	});

	it('stores the default fight under the name typed even with no event added', async () => {
		const { encounter } = await setup();
		const dialog = await openPlanner();
		await commit(nameField(dialog), 'Just a name');
		await clickButton('Save');

		expect(Object.keys(stored())).toEqual(['Just a name']);
		expect(encounter.hasPlan()).toBe(false);
		expect(await listedOwn()).toEqual(['Just a name']);
		expect(choice().textContent).toBe('Just a name');
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

		const dialog = await openPlanner();
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
