// The loading state of the sheet. The crit-cap row is the last row of the Physical group for this
// display-stat list, so it is the last row either way; what is pinned here is that it skeletons
// with its siblings instead of reading as settled while the stats are still in flight.
import { Class, PseudoStat, Race, Stat, WeaponSkill } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateWeaponSkill } from '@i18n/localization';
import { getClassWeaponSkills } from '@sim/player/classes/capabilities';
import { WeaponSkillGroup, weaponSkillsInGroups } from '@sim/player/weapon_skill_groups';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { CharacterStats, CharacterStatsHeaderView, type CharacterStatsMode, CharacterStatsPanelView } from './CharacterStats';
import { buildRows } from './utils/rows';
import { statDisplayString } from './utils/stat_display';

const KEY = 7;

const DISPLAY_STATS = [
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
];

const CRIT_CAP_DELTA = -13;

const CRIT_CAP_INFO = {
	glancing: 24,
	suppression: 4.8,
	remainingMeleeHitCap: 0,
	remainingExpertiseCap: 0,
	debuffCrit: 0,
	specSpecificOffset: 0,
	baseCritCap: 43,
	playerCritCapDelta: CRIT_CAP_DELTA,
};

interface SheetSetup {
	displayStats: Array<UnitStat>;
	displayWeaponSkillGroups?: Array<WeaponSkillGroup>;
	currentStats: object;
	inFrontOfTarget?: boolean;
	player: Record<string, unknown>;
}

const renderSheet = ({ displayStats, displayWeaponSkillGroups, currentStats, inFrontOfTarget = false, player }: SheetSetup) => {
	const store = createSimStore();
	seedKeyed(store, 'players', KEY, {
		name: 'P',
		race: Race.RaceOrc,
		gear: { id: 'gear' },
		bonusStats: new Stats(),
		inFrontOfTarget,
		weaponSkills: [],
		currentStats,
		v: zeroVersions(PLAYER_FIELDS),
	} as unknown as PlayerSlice);

	const view = renderWithHost(<CharacterStats />, {
		player: {
			sim: { store },
			storeKey: KEY,
			getClass: () => Class.ClassWarrior,
			getRace: () => Race.RaceOrc,
			getCurrentStats: () => currentStats,
			getDebuffStats: () => new Stats(),
			getEquippedItem: () => null,
			isDualWielding: () => false,
			...player,
		},
		sim: { store },
		individualConfig: { displayStats, displayWeaponSkillGroups },
		getStorageKey: (keyPart: string) => `test${keyPart}`,
	});
	const rows = Array.from(view.container.querySelectorAll('[data-testid="character-stats-table-row"]'));
	return { view, rows };
};

const statsWhen = (settled: boolean) => (settled ? { finalStats: new Stats().toProto() } : {});

const renderStats = (settled: boolean) =>
	renderSheet({
		displayStats: DISPLAY_STATS,
		currentStats: statsWhen(settled),
		player: {
			// Melee DPS but not a tank: the crit cap row is shown, the miss/avoidance/crit-immunity trio is not.
			getPlayerSpec: () => ({ isTankSpec: false, isMeleeDpsSpec: true }),
			getMeleeCritCap: () => CRIT_CAP_DELTA,
			getMeleeCritCapInfo: () => CRIT_CAP_INFO,
			getBaseDefense: () => 300,
		},
	});

// The tank trio is appended to the Defense group, so that group has to be populated for the
// rows to exist at all.
const TANK_DISPLAY_STATS = [...DISPLAY_STATS, UnitStat.fromStat(Stat.StatDefenseRating), UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent)];

const MISS_INFO = { base: 5, defense: 2, debuffs: 0, total: 7 };
const AVOIDANCE_INFO = { miss: 7, dodge: 18, parry: 14, block: 25, total: 64, shear: 57 };
const CRIT_IMMUNITY_INFO = { total: 5.6, delta: 0, defense: 4.6, talents: 1 };

// The tank sheet adds three rows the DPS sheet has no equivalent for -- miss, avoidance and
// crit immunity -- appended to the Defense group. What is pinned here is that they skeleton with
// every other row rather than reading as settled while the stats are still in flight.
const renderTankStats = (settled: boolean) =>
	renderSheet({
		displayStats: TANK_DISPLAY_STATS,
		currentStats: statsWhen(settled),
		inFrontOfTarget: true,
		player: {
			getPlayerSpec: () => ({ isTankSpec: true, isMeleeDpsSpec: false }),
			isSpec: () => false,
			getMissChanceInfo: () => MISS_INFO,
			getAvoidanceInfo: () => AVOIDANCE_INFO,
			getCritImmunityInfo: () => CRIT_IMMUNITY_INFO,
			getCritImmunity: () => CRIT_IMMUNITY_INFO.delta,
			getBaseDefense: () => 350,
		},
	});

const critCapRow = (rows: Array<Element>) => {
	const row = rows.at(-1)!;
	expect(row.querySelector('.ui-character-stats-label')!.textContent).toBe(i18n.t('sidebar.character_stats.melee_crit_cap'));
	return row;
};

describe('CharacterStats sheet', () => {
	it('lays the stats out on a plain table, with no row or column rules', () => {
		const { view } = renderStats(true);

		expect(view.container.querySelector('table')!.getAttribute('data-variant')).toBe('plain');
	});
});

describe('CharacterStats Key/All toggle', () => {
	const rowLabels = (view: ReturnType<typeof renderWithHost>) =>
		Array.from(view.container.querySelectorAll('[data-testid="character-stats-table-row"]')).map(row => row.firstElementChild!.textContent);
	const modeButton = (mode: 'key' | 'all') => screen.getByRole('button', { name: i18n.t(`sidebar.character_stats.mode.${mode}`) });

	it('opens on Key, showing the spec display stats and the crit cap', () => {
		const { view } = renderStats(true);

		expect(modeButton('key').getAttribute('aria-pressed')).toBe('true');
		expect(rowLabels(view)).toEqual([
			...DISPLAY_STATS.map(stat => stat.getShortName(Class.ClassWarrior)),
			i18n.t('sidebar.character_stats.melee_crit_cap'),
		]);
	});

	it('shows every group on All, with the rows of the open ones, and the Key rows again on Key', () => {
		const { view } = renderStats(true);
		const keyLabels = rowLabels(view);

		fireEvent.click(modeButton('all'));
		const allGroups = buildRows({ getPlayerSpec: () => ({ isTankSpec: false, isMeleeDpsSpec: true }) } as never, UnitStat.getAll()).groups;
		const shownRows = allGroups.filter(group => ['primary', 'attributes', 'physical'].includes(group.key)).flatMap(group => group.rows);
		expect(modeButton('all').getAttribute('aria-pressed')).toBe('true');
		expect(groupTriggers().map(trigger => trigger.getAttribute('data-testid'))).toEqual(
			['attributes', 'physical', 'spell', 'defense', 'resistance'].map(key => `character-stats-group-${key}-trigger`),
		);
		expect(rowLabels(view).length).toBe(shownRows.length);
		expect(rowLabels(view)).toEqual(expect.arrayContaining(keyLabels));

		fireEvent.click(modeButton('key'));
		expect(rowLabels(view)).toEqual(keyLabels);
	});
});

const groupTriggers = () => screen.queryAllByTestId(/^character-stats-group-[a-z]+-trigger$/);
const groupTrigger = (key: string) => screen.getByTestId(`character-stats-group-${key}-trigger`);
const expandedGroups = () => groupTriggers().filter(trigger => trigger.getAttribute('aria-expanded') === 'true');
const showAll = () => fireEvent.click(screen.getByRole('button', { name: i18n.t('sidebar.character_stats.mode.all') }));

describe('CharacterStats groups', () => {
	afterEach(() => {
		window.localStorage.clear();
	});

	it('keeps health and mana above the groups, under no header', () => {
		renderStats(true);
		showAll();

		const healthLabel = UnitStat.fromStat(Stat.StatHealth).getShortName(Class.ClassWarrior);
		const health = screen.getAllByTestId('character-stats-table-row').find(row => row.firstElementChild!.textContent === healthLabel)!;
		expect(screen.getByTestId('character-stats-groups').contains(health)).toBe(false);
	});

	it.each([
		['a melee spec', renderStats, 'physical'],
		['a tank', renderTankStats, 'defense'],
	])('opens Attributes and the main group for %s, and leaves the rest closed', (_spec, render, main) => {
		render(true);
		showAll();

		expect(expandedGroups()).toEqual([groupTrigger('attributes'), groupTrigger(main)]);
	});

	it('keeps the open groups under the spec storage key, and a remount reads them back', () => {
		const { view } = renderStats(true);
		showAll();

		fireEvent.click(groupTrigger('spell'));
		fireEvent.click(groupTrigger('attributes'));
		expect(JSON.parse(window.localStorage.getItem('test__characterStatsGroups')!)).toEqual(['physical', 'spell']);

		view.unmount();
		renderStats(true);
		showAll();
		expect(expandedGroups()).toEqual([groupTrigger('physical'), groupTrigger('spell')]);
	});
});

describe('CharacterStats weapon skills', () => {
	afterEach(() => {
		window.localStorage.clear();
	});

	const renderWeaponSkills = (playerClass: Class, displayWeaponSkillGroups?: Array<WeaponSkillGroup>) =>
		renderSheet({
			displayStats: DISPLAY_STATS,
			displayWeaponSkillGroups,
			currentStats: statsWhen(true),
			player: {
				getClass: () => playerClass,
				getPlayerSpec: () => ({ isTankSpec: false, isMeleeDpsSpec: true }),
				getMeleeCritCap: () => CRIT_CAP_DELTA,
				getMeleeCritCapInfo: () => CRIT_CAP_INFO,
				getBaseDefense: () => 300,
			},
		});
	const shownSkills = (playerClass: Class) => getClassWeaponSkills(playerClass).filter(skill => screen.queryByText(translateWeaponSkill(skill)));

	it.each([
		['a hunter', Class.ClassHunter, [WeaponSkillGroup.Melee, WeaponSkillGroup.Ranged], WeaponSkill.WeaponSkillThrown],
		['a warrior', Class.ClassWarrior, [WeaponSkillGroup.Melee], WeaponSkill.WeaponSkillBows],
	])('shows %s the spec weapon skill groups on Key and every class weapon skill on All', (_spec, playerClass, groups, keyHidden) => {
		renderWeaponSkills(playerClass, groups);
		fireEvent.click(groupTrigger('weapon-skill'));

		const keySkills = weaponSkillsInGroups(groups);
		expect(shownSkills(playerClass)).toEqual(getClassWeaponSkills(playerClass).filter(skill => keySkills.includes(skill)));
		expect(shownSkills(playerClass)).not.toContain(keyHidden);

		showAll();
		expect(shownSkills(playerClass)).toEqual(getClassWeaponSkills(playerClass));
	});

	it('leaves the weapon skill item out on Key for a spec that names no weapon skills', () => {
		renderWeaponSkills(Class.ClassWarrior);
		expect(screen.queryByTestId('character-stats-group-weapon-skill-trigger')).toBeNull();

		showAll();
		expect(screen.queryByTestId('character-stats-group-weapon-skill-trigger')).not.toBeNull();
	});
});

describe('CharacterStats loading state', () => {
	it('shows a skeleton in every row, crit cap included, while the stats are still loading', () => {
		const { view, rows } = renderStats(false);

		expect(rows.length).toBe(DISPLAY_STATS.length + 1);
		expect(view.container.querySelector('table')!.getAttribute('aria-busy')).toBe('true');
		for (const row of rows) expect(row.querySelector('.ui-character-stats-value > .ui-skeleton')).not.toBeNull();
		expect(view.container.querySelectorAll('[data-testid="stat-value-link"]').length).toBe(0);
		expect(critCapRow(rows).textContent).toBe(i18n.t('sidebar.character_stats.melee_crit_cap'));
	});

	it('replaces every skeleton with its value once the stats arrive', () => {
		const { view, rows } = renderStats(true);

		expect(rows.length).toBe(DISPLAY_STATS.length + 1);
		expect(view.container.querySelector('table')!.hasAttribute('aria-busy')).toBe(false);
		expect(view.container.querySelectorAll('.ui-skeleton').length).toBe(0);
		expect(view.container.querySelectorAll('[data-testid="stat-value-link"]').length).toBe(rows.length);
		expect(critCapRow(rows).querySelector('[data-testid="stat-value-link"]')!.textContent).toContain('13.00%');
	});
});

describe('CharacterStats tank rows', () => {
	const labels = [
		i18n.t('sidebar.character_stats.tank_caps.miss_label'),
		i18n.t('sidebar.character_stats.tank_caps.avoidance_label'),
		i18n.t('sidebar.character_stats.tank_caps.crit_immunity_label'),
	];

	const tankRows = (rows: Array<Element>) => rows.filter(row => labels.includes(row.querySelector('.ui-character-stats-label')!.textContent ?? ''));

	it('skeletons miss, avoidance and crit immunity while the stats are still loading', () => {
		const { rows } = renderTankStats(false);
		const trio = tankRows(rows);

		expect(trio.length).toBe(3);
		for (const row of trio) {
			expect(row.querySelector('.ui-character-stats-value > .ui-skeleton')).not.toBeNull();
			expect(row.querySelector('[data-testid="stat-value-link"]')).toBeNull();
		}
	});

	it('replaces those three skeletons with their values once the stats arrive', () => {
		const { rows } = renderTankStats(true);
		const trio = tankRows(rows);

		expect(trio.length).toBe(3);
		for (const row of trio) {
			expect(row.querySelector('.ui-skeleton')).toBeNull();
			expect(row.querySelector('[data-testid="stat-value-link"]')).not.toBeNull();
		}
		expect(trio[0].querySelector('[data-testid="stat-value-link"]')!.textContent).toContain('7.00%');
		expect(trio[1].querySelector('[data-testid="stat-value-link"]')!.textContent).toContain('64.00%');
	});
});

const MELEE_HIT_DISPLAY_STATS = [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent), UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent)];

const MELEE_HIT_STATS = new Stats()
	.withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 3)
	.withPseudoStat(PseudoStat.PseudoStatMainHandHitRating, 30)
	.withPseudoStat(PseudoStat.PseudoStatOffHandHitPercent, 2)
	.withPseudoStat(PseudoStat.PseudoStatOffHandHitRating, 20);

// i18next is stubbed empty under vitest, so `statDisplayString` is used to compute these rather than hardcoding its output.
const MAIN_HAND_VALUE = statDisplayString({} as never, MELEE_HIT_STATS, UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent), true);
const OFF_HAND_VALUE = statDisplayString({} as never, MELEE_HIT_STATS, UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent), true);

const EVEN_HANDS_STATS = MELEE_HIT_STATS.withPseudoStat(PseudoStat.PseudoStatOffHandHitPercent, 3).withPseudoStat(PseudoStat.PseudoStatOffHandHitRating, 30);

const renderMeleeHitRow = (dualWielding: boolean, stats = MELEE_HIT_STATS) => {
	const { rows } = renderSheet({
		displayStats: MELEE_HIT_DISPLAY_STATS,
		currentStats: { finalStats: stats.toProto() },
		player: {
			getPlayerSpec: () => ({ isTankSpec: false, isMeleeDpsSpec: false }),
			isDualWielding: () => dualWielding,
		},
	});
	return { meleeHitRow: rows[0] };
};

const MAIN_HAND_LABEL = i18n.t('slots.main_hand', { ns: 'character' });
const OFF_HAND_LABEL = i18n.t('slots.off_hand', { ns: 'character' });

describe('CharacterStats combined melee hand rows', () => {
	it.each([
		['single wielding', false, MELEE_HIT_STATS, [MAIN_HAND_VALUE]],
		['dual wielding with even hands', true, EVEN_HANDS_STATS, [MAIN_HAND_VALUE]],
		[
			'dual wielding with uneven hands',
			true,
			MELEE_HIT_STATS,
			[`${i18n.t('sidebar.character_stats.hands.main')} ${MAIN_HAND_VALUE}`, `${i18n.t('sidebar.character_stats.hands.off')} ${OFF_HAND_VALUE}`],
		],
	])('shows one value while %s, and each hand prefixed MH and OH only when they differ', (_case, dualWielding, stats, lines) => {
		const { meleeHitRow } = renderMeleeHitRow(dualWielding, stats);
		const links = meleeHitRow.querySelectorAll('[data-testid="stat-value-link"]');

		expect([...links].map(link => link.parentElement!.textContent)).toEqual(lines);
	});

	it.each([
		['main', 0, MAIN_HAND_LABEL, MAIN_HAND_VALUE],
		['off', 1, OFF_HAND_LABEL, OFF_HAND_VALUE],
	])("opens the %s hand's own breakdown, headed by its hand, when its half is hovered", async (_hand, index, label, value) => {
		const { meleeHitRow } = renderMeleeHitRow(true);
		const links = meleeHitRow.querySelectorAll('[data-testid="stat-value-link"]');

		fireEvent.mouseEnter(links[index]);
		expect((await screen.findByTestId('character-stats-tooltip-hand')).textContent).toBe(label);
		expect((await screen.findAllByText(value)).length).toBeGreaterThan(1);
	});
});

const LABELS = { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' };

const mount = (mode: CharacterStatsMode) => {
	const onModeChange = vi.fn();
	return { onModeChange, ...render(<CharacterStatsHeaderView mode={mode} onModeChange={onModeChange} labels={LABELS} />) };
};

const pressed = () => screen.getAllByRole('button').map(button => [button.textContent, button.getAttribute('aria-pressed')]);

describe('CharacterStatsHeaderView', () => {
	it('sets the heading beside a labelled Key/All group', () => {
		mount('key');

		expect(screen.getByTestId('character-stats-label').textContent).toBe('Stats');
		expect(screen.getByRole('group', { name: 'Stats shown' })).toBeTruthy();
	});

	it.each<[CharacterStatsMode, string, string]>([
		['key', 'true', 'false'],
		['all', 'false', 'true'],
	])('presses the mode it is given (%s)', (mode, key, all) => {
		mount(mode);
		expect(pressed()).toEqual([
			['Key', key],
			['All', all],
		]);
	});

	it('reports the other mode on a click, and nothing for the one already shown', () => {
		const { onModeChange } = mount('key');

		fireEvent.click(screen.getByRole('button', { name: 'Key' }));
		expect(onModeChange).not.toHaveBeenCalled();
		fireEvent.click(screen.getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
	});
});

describe('CharacterStatsPanelView', () => {
	it('shows the title above its content, and reports a mode change from the Key/All group', () => {
		const onModeChange = vi.fn();
		render(
			<CharacterStatsPanelView mode="key" onModeChange={onModeChange} labels={LABELS}>
				<table data-testid="stats-table" />
			</CharacterStatsPanelView>,
		);

		const title = screen.getByTestId('character-stats-label');
		expect(title.textContent).toBe('Stats');
		expect(title.compareDocumentPosition(screen.getByTestId('stats-table')) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
		fireEvent.click(screen.getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
	});
});
