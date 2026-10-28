import { AddWave, DesignDowntime, EncounterDesign, MobType, PresetEncounter, PresetTarget, Stat, Target } from '@generated/proto/common';
import { UIDatabase } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { EncounterPicker } from '../EncounterPicker';

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

const setup = async ({ targets = 1, design = EncounterDesign.create() } = {}) => {
	const { sim, player } = await makeIndividualSim();
	sim.encounter.setTargets(Array.from({ length: targets }, (_, index) => Target.create({ id: index + 1, name: `Target ${index + 1}` })));
	sim.encounter.setDesign(design);
	renderWithHost(<EncounterPicker showExecuteProportion={false} />, { sim, player });
	return { encounter: sim.encounter, baseSeconds: sim.encounter.getDuration() };
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
			await clickButton('Add stats');
			expect(within(dialog).getByLabelText('Add template').textContent).toContain('Default/Dynamic Add');
			expect((within(dialog).getByLabelText('Health') as HTMLInputElement).value).toBe('5000');
			await commit(within(dialog).getByLabelText('Health'), '7000');
			expect(within(dialog).getByLabelText('Add template').textContent).toContain('Custom');

			await clickButton('Save');
			const [wave] = encounter.getDesign().addWaves;
			expect(wave.adds).toHaveLength(1);
			expect(wave.adds[0].name).toBe('Dynamic Add');
			expect(wave.adds[0].stats[Stat.StatHealth]).toBe(7000);
		}
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
});
