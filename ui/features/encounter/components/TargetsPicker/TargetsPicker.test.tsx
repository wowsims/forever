import { CreatureFamily, InputType, MobType, PresetTarget, Stat, Target as TargetProto, TargetInput } from '@generated/proto/common';
import { CREATURE_ARMOR_BY_LEVEL } from '@sim/proto/sim_tables_auto_gen';
import type { Encounter } from '@sim/raid/encounter';
import { act, createEvent, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TargetsPicker } from './TargetsPicker';

const source = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	return {
		listeners,
		subscribe: (onChange: () => void) => {
			listeners.add(onChange);
			return () => listeners.delete(onChange);
		},
		notify: () => Array.from(listeners).forEach(listener => listener()),
	};
});
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(source.subscribe));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@i18n/localization', () => ({
	translateCreatureFamily: (value: number) => `family-${value}`,
	translateItemLabel: (label: string) => label,
	translateMobType: (value: number) => `mob-${value}`,
	translateSpellSchool: (value: number) => `school-${value}`,
	translateStat: (value: number) => `stat-${value}`,
	translateTargetInputLabel: (label: string) => label,
	translateTargetInputTooltip: (_label: string, tooltip: string) => tooltip,
}));
const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const defaultTarget = () => TargetProto.create({ level: 63, swingSpeed: 2 });

/** Stands in for the `Encounter` facade: the seven members the target pickers reach for. */
class FakeEncounter {
	constructor(
		public targets: Array<TargetProto> = [defaultTarget()],
		public presetTargets: Array<PresetTarget> = [],
	) {}
	readonly sim = {
		db: { getAllPresetTargets: () => this.presetTargets },
		getShowDamageMetrics: () => true,
		getShowThreatMetrics: () => true,
		getShowHealingMetrics: () => true,
	};
	getTargets() {
		return this.targets;
	}
	getTarget(index: number) {
		return this.targets[index];
	}
	setTargets(next: Array<TargetProto>) {
		this.targets = next;
		source.notify();
	}
	modifyTarget(index: number, apply: (target: TargetProto) => void) {
		const draft = TargetProto.clone(this.targets[index]);
		apply(draft);
		this.targets = this.targets.map((target, at) => (at === index ? draft : target));
		source.notify();
	}
	applyPresetTarget() {
		/* no preset targets in this fixture */
	}
}

const mount = (encounter: FakeEncounter) => render(<TargetsPicker encounter={encounter as unknown as Encounter} />);

const root = () => document.querySelector('[data-testid="targets-picker"]') as HTMLElement;
const targetRoots = () => [...document.querySelectorAll<HTMLElement>('[data-testid="target-picker-root"]')];
const items = () => [...root().querySelectorAll<HTMLElement>('[data-testid="targets-picker-item"]')];
const idsIn = (element: Element) => [...element.querySelectorAll('[id]')].map(node => node.id);
const openMenu = (index: number) => {
	const trigger = items()[index].querySelector('[data-testid="tree-item-actions"]') as HTMLButtonElement;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	return document.querySelector('[data-testid="tree-item-actions-popup"][data-open]') as HTMLElement;
};

// happy-dom lays nothing out, so each row gets a 20px box, `index` rows from the top.
const place = (element: HTMLElement, index: number) => vi.spyOn(element, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, index * 20, 20, 20));
const grab = (index: number) => {
	fireEvent.mouseDown(items()[index].querySelector('[data-testid="targets-picker-handle"]')!);
	fireEvent.dragStart(items()[index], { dataTransfer: new DataTransfer() });
};
// happy-dom's DragEvent is a plain Event, which drops `clientY` from its init.
const dragEvent = (type: 'dragOver' | 'drop', target: HTMLElement, y: number) => {
	const event = createEvent[type](target, { dataTransfer: new DataTransfer() });
	Object.defineProperty(event, 'clientY', { value: y });
	return fireEvent(target, event);
};

it('shows Immune to Bleeds without threat metrics', () => {
	const encounter = new FakeEncounter();
	encounter.sim.getShowThreatMetrics = () => false;
	mount(encounter);

	expect(idsIn(root()).some(id => id.endsWith('bleed-immune'))).toBe(true);
	expect(idsIn(root()).some(id => id.endsWith('parry-haste'))).toBe(false);
});

beforeEach(() => {
	source.listeners.clear();
	trackEvent.mockClear();
});

describe('TargetsPicker', () => {
	it('renders one tree row per target, each titled by its position, and follows the list on a write', () => {
		const encounter = new FakeEncounter();
		mount(encounter);

		expect(items()).toHaveLength(1);
		expect(targetRoots()).toHaveLength(1);
		expect(items()[0].querySelector('h6')!.textContent).toBe('settings_tab.encounter.target 1');
		const sections = [...targetRoots()[0].querySelectorAll('[data-testid="target-picker-section"]')];
		expect(sections).toHaveLength(3);

		act(() => encounter.setTargets([defaultTarget(), defaultTarget()]));
		expect(items()).toHaveLength(2);
		expect(targetRoots()).toHaveLength(2);
		expect(items()[1].querySelector('h6')!.textContent).toBe('settings_tab.encounter.target 2');
	});

	it('withholds delete from the first target, so the list never runs empty', () => {
		const encounter = new FakeEncounter([defaultTarget(), defaultTarget()]);
		mount(encounter);

		const menu = openMenu(0);
		expect(menu.querySelector('[data-testid="targets-picker-copy"]')).not.toBeNull();
		expect(menu.querySelector('[data-testid="targets-picker-delete"]')).toBeNull();
	});

	it('gives every target its own picker ids and removes the second through the menu', () => {
		const encounter = new FakeEncounter([defaultTarget(), defaultTarget()]);
		mount(encounter);

		const [first, second] = targetRoots().map(idsIn);
		expect(first.length).toBeGreaterThan(10);
		expect(first).toEqual(expect.arrayContaining(['target-0-picker-npc', 'target-0-picker-level', 'target-0-picker-spell-school']));
		expect(second).toEqual(expect.arrayContaining(['target-1-picker-npc', 'target-1-picker-level', 'target-1-picker-spell-school']));
		expect(first.filter(id => second.includes(id))).toEqual([]);

		const deleteSecond = openMenu(1).querySelector('[data-testid="targets-picker-delete"]');
		expect(deleteSecond).not.toBeNull();
		act(() => void fireEvent.click(deleteSecond!));

		expect(encounter.targets).toHaveLength(1);
		expect(items()).toHaveLength(1);
		expect(trackEvent).toHaveBeenCalledWith(expect.objectContaining({ label: 'remove-target' }));
	});

	it('copies a target to a clone beside it and adds a default target at the end', () => {
		const encounter = new FakeEncounter([TargetProto.create({ level: 60 })]);
		mount(encounter);

		const copy = openMenu(0).querySelector('[data-testid="targets-picker-copy"]')!;
		act(() => void fireEvent.click(copy));
		expect(encounter.targets.map(target => target.level)).toEqual([60, 60]);
		expect(encounter.targets[1]).not.toBe(encounter.targets[0]);

		act(() => void fireEvent.click(document.querySelector('[data-testid="tree-add"]')!));
		expect(encounter.targets).toHaveLength(3);
		expect(items()).toHaveLength(3);
		expect(trackEvent).toHaveBeenLastCalledWith(expect.objectContaining({ label: 'add-target' }));
	});

	it('moves a target by dragging its handle, in one write', () => {
		const encounter = new FakeEncounter([TargetProto.create({ level: 60 }), TargetProto.create({ level: 61 }), TargetProto.create({ level: 62 })]);
		mount(encounter);
		const setTargets = vi.spyOn(encounter, 'setTargets');
		items().forEach(place);

		grab(2);
		dragEvent('drop', items()[0], 0);

		expect(setTargets).toHaveBeenCalledTimes(1);
		expect(encounter.targets.map(target => target.level)).toEqual([62, 60, 61]);
	});

	it('drops a target released over a target input beside the target row holding it, never into the input list', () => {
		const encounter = new FakeEncounter([
			TargetProto.create({ level: 60, targetInputs: [TargetInput.create({ label: 'Stacks', inputType: InputType.Number })] }),
			TargetProto.create({ level: 61 }),
		]);
		mount(encounter);
		const setTargets = vi.spyOn(encounter, 'setTargets');
		items().forEach(place);
		const input = document.querySelector<HTMLElement>('[data-testid="target-inputs-picker-item"]')!;
		place(input, 0);

		grab(1);
		dragEvent('drop', input, 0);

		expect(setTargets).toHaveBeenCalledTimes(1);
		expect(encounter.targets.map(target => target.level)).toEqual([61, 60]);
		expect(encounter.targets.map(target => target.targetInputs.length)).toEqual([0, 1]);
	});

	it('reads each target row from its own index rather than from the first, and writes a level change back to it', async () => {
		const encounter = new FakeEncounter([TargetProto.create({ level: 63 }), TargetProto.create({ level: 61 })]);
		mount(encounter);

		const levelOf = (index: number) => targetRoots()[index].querySelector<HTMLElement>('#target-' + index + '-picker-level')!;
		expect(levelOf(0).textContent).toBe('63');
		expect(levelOf(1).textContent).toBe('61');

		await userEvent.click(levelOf(1));
		await userEvent.click(screen.getByRole('option', { name: '62' }));

		expect(encounter.targets.map(target => target.level)).toEqual([63, 62]);
	});

	it("sets the armor to the game files' creature armor of the new level, and the armor field shows it", async () => {
		const target = TargetProto.create({ level: 63 });
		target.stats[Stat.StatArmor] = 1234;
		const encounter = new FakeEncounter([target]);
		mount(encounter);

		await userEvent.click(document.querySelector<HTMLElement>('#target-0-picker-level')!);
		await userEvent.click(screen.getByRole('option', { name: '60' }));

		expect(CREATURE_ARMOR_BY_LEVEL[60]).toBeGreaterThan(0);
		expect(encounter.targets[0].stats[Stat.StatArmor]).toBe(CREATURE_ARMOR_BY_LEVEL[60]);
		expect(document.querySelector<HTMLInputElement>(`#target-0-picker-stats-${Stat.StatArmor}`)!.value).toBe(String(CREATURE_ARMOR_BY_LEVEL[60]));
	});

	it('disables the dual-wield miss penalty until dual wield is on', () => {
		const encounter = new FakeEncounter([TargetProto.create({ dualWield: false })]);
		mount(encounter);
		const penalty = () => document.querySelector('#target-0-picker-dw-miss-penalty')!.closest('[data-testid="boolean-picker-root"]')!;
		expect(penalty().hasAttribute('data-disabled')).toBe(true);

		act(() => encounter.modifyTarget(0, target => (target.dualWield = true)));
		expect(penalty().hasAttribute('data-disabled')).toBe(false);
	});

	it('offers only the creature families the mob type holds, and drops one a new mob type does not', async () => {
		const encounter = new FakeEncounter([TargetProto.create({ mobType: MobType.MobTypeMechanical })]);
		mount(encounter);
		const family = () => document.getElementById('target-0-picker-creature-family');
		expect(family()).toBeNull();

		act(() => encounter.modifyTarget(0, target => (target.mobType = MobType.MobTypeHumanoid)));
		await userEvent.click(family()!);
		expect(screen.getAllByRole('option').map(option => option.textContent)).toEqual([
			`family-${CreatureFamily.CreatureFamilyUnknown}`,
			`family-${CreatureFamily.CreatureFamilyNaga}`,
			`family-${CreatureFamily.CreatureFamilyAquatic}`,
		]);
		await userEvent.click(screen.getByRole('option', { name: `family-${CreatureFamily.CreatureFamilyNaga}` }));
		expect(encounter.targets[0].creatureFamily).toBe(CreatureFamily.CreatureFamilyNaga);

		await userEvent.click(document.getElementById('target-0-picker-mob-type')!);
		await userEvent.click(screen.getByRole('option', { name: `mob-${MobType.MobTypeDemon}` }));
		expect(encounter.targets[0].mobType).toBe(MobType.MobTypeDemon);
		expect(encounter.targets[0].creatureFamily).toBe(CreatureFamily.CreatureFamilyUnknown);
	});

	// Encounters register a preset per raid size and difficulty, all sharing one npc id — which is
	// this picker's value, and all the sim resolves an AI from.
	describe('AI options', () => {
		const preset = (path: string, id: number) => PresetTarget.create({ path, target: TargetProto.create({ id }) });
		const aiOptions = async () => {
			await userEvent.click(document.getElementById('target-0-picker-ai')!);
			return screen.getAllByRole('option').map(option => option.textContent);
		};

		it('offers one option per npc id, however many presets share it', async () => {
			mount(
				new FakeEncounter(
					[defaultTarget()],
					[preset('Molten Core/Lucifron', 12118), preset('Molten Core/Lucifron (alternate)', 12118), preset('x/Ragnaros', 11502)],
				),
			);

			expect(await aiOptions()).toEqual(['common.none', 'Molten Core/Lucifron', 'x/Ragnaros']);
		});

		it('keeps the preset the sim resolves the id to, which is the first one registered', async () => {
			mount(new FakeEncounter([defaultTarget()], [preset('first', 71466), preset('second', 71466)]));

			expect(await aiOptions()).toEqual(['common.none', 'first']);
		});
	});

	describe('target inputs', () => {
		const withInputs = () =>
			new FakeEncounter([
				TargetProto.create({
					targetInputs: [
						TargetInput.create({ label: 'Stacks', inputType: InputType.Number, numberValue: 3 }),
						TargetInput.create({ label: 'Enraged', inputType: InputType.Bool, boolValue: true }),
						TargetInput.create({ label: 'Phase', inputType: InputType.Enum, enumValue: 1, enumOptions: ['One', 'Two'] }),
					],
				}),
			]);

		it('renders one picker per input dispatched on its declared type, shows each its own value, offers no list actions, and writes back to its own index', () => {
			const encounter = withInputs();
			mount(encounter);
			const inputRoots = [...document.querySelectorAll('[data-testid="target-input-picker-root"]')];

			expect(inputRoots).toHaveLength(3);
			const matchesRoot = (node: Element, index: number) =>
				index === 2
					? within(node.firstElementChild as HTMLElement).queryByRole('combobox') !== null
					: node.firstElementChild!.getAttribute('data-testid') === ['number-picker-root', 'boolean-picker-root'][index];
			expect(inputRoots.map((node, index) => matchesRoot(node, index))).toEqual([true, true, true]);

			// `float: true`, so the number box shows two decimals.
			expect(inputRoots[0].querySelector('input')!.value).toBe('3.00');
			expect(inputRoots[1].querySelector('input')!.checked).toBe(true);
			expect(within(inputRoots[2] as HTMLElement).getByRole('combobox').textContent).toBe('Two');

			const list = document.querySelector('[data-testid="target-inputs-picker"]')!;
			expect(list.getAttribute('data-density')).toBe('compact');
			expect(list.querySelector('[data-testid="tree-add"]')).toBeNull();
			expect(list.querySelector('[data-testid="tree-item-header"]')).toBeNull();
			expect(list.querySelectorAll('[data-testid="target-inputs-picker-item"]')).toHaveLength(3);

			const numberInput = inputRoots[0].querySelector('input')!;
			act(() => void fireEvent.change(numberInput, { target: { value: '7' } }));

			expect(encounter.targets[0].targetInputs.map(input => input.numberValue)).toEqual([7, 0, 0]);
		});

		it('renders no list for a target that declares no input', () => {
			mount(new FakeEncounter());

			expect(document.querySelector('[data-testid="target-inputs-picker"]')).toBeNull();
		});
	});
});
