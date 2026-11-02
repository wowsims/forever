import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { StatFilterList, StatFilterPopover } from './StatFilterPopover';

const health = UnitStat.fromStat(Stat.StatHealth);
const strength = UnitStat.fromStat(Stat.StatStrength);
const agility = UnitStat.fromStat(Stat.StatAgility);
const meleeHit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);
const hitRating = UnitStat.fromStat(Stat.StatHitRating);
const mainHandDps = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps);
const physicalDamage = UnitStat.fromStat(Stat.StatPhysicalDamage);

const universe = [health, strength, agility, meleeHit, hitRating, mainHandDps, physicalDamage];

const EVERY_GROUP = ['primary', 'attributes', 'physical', 'spell', 'defense', 'resistance', 'gear', 'other'];

const mountList = (props: { shown?: UnitStat[]; customised?: boolean; universe?: UnitStat[]; openGroups?: string[] } = {}) => {
	const onChange = vi.fn();
	const onReset = vi.fn();
	const onOpenGroupsChange = vi.fn();
	render(
		<StatFilterList
			universe={props.universe ?? universe}
			shown={props.shown ?? [strength, meleeHit]}
			customised={props.customised ?? false}
			playerClass={Class.ClassWarrior}
			onChange={onChange}
			onReset={onReset}
			openGroups={props.openGroups ?? EVERY_GROUP}
			onOpenGroupsChange={onOpenGroupsChange}
			idPrefix="test-filter"
		/>,
	);
	return { onChange, onReset, onOpenGroupsChange };
};

const checkbox = (stat: UnitStat) => screen.getByRole('checkbox', { name: stat.getShortName(Class.ClassWarrior) });
const keys = (stats: ReadonlyArray<UnitStat>) => stats.map(stat => stat.getKey());
const group = (key: string) => screen.getByTestId(`test-filter-group-${key}`);
const groupTrigger = (key: string) => screen.getByTestId(`test-filter-group-${key}-trigger`);
const groupTriggers = () => screen.queryAllByTestId(/^test-filter-group-[a-z]+-trigger$/);
const groupTitle = (label: string, shown: number, total: number) => i18n.t('stat_filters.group_count', { group: label, shown, total });
const expectGroupOrder = (key: string, stats: ReadonlyArray<UnitStat>) => {
	const boxes = within(group(key)).getAllByRole('checkbox');
	expect(boxes).toHaveLength(stats.length);
	stats.forEach((stat, index) => expect(boxes[index]).toBe(checkbox(stat)));
};

describe('StatFilterList', () => {
	it('groups the universe under the character stats headings with shown/total counts, Other last and empty groups left out', () => {
		mountList();

		expect(groupTriggers().map(trigger => trigger.textContent)).toEqual([
			groupTitle(i18n.t('stat_filters.groups.primary'), 0, 1),
			groupTitle(i18n.t('sidebar.character_stats.groups.attributes'), 1, 2),
			groupTitle(i18n.t('sidebar.character_stats.groups.physical'), 1, 1),
			groupTitle(i18n.t('stat_filters.groups.gear'), 0, 2),
			groupTitle(i18n.t('stat_filters.groups.other'), 0, 1),
		]);
	});

	it('expands the open groups only, and a collapsed group still counts what it holds', () => {
		mountList({ openGroups: ['attributes'] });

		expect(groupTriggers().filter(trigger => trigger.getAttribute('aria-expanded') === 'true')).toEqual([groupTrigger('attributes')]);
		expect(screen.queryByRole('checkbox', { name: hitRating.getShortName(Class.ClassWarrior) })).toBeNull();
		expect(groupTrigger('physical').textContent).toBe(groupTitle(i18n.t('sidebar.character_stats.groups.physical'), 1, 1));
	});

	it('hands back the new open groups when a header is clicked', () => {
		const { onOpenGroupsChange } = mountList({ openGroups: ['attributes'] });

		fireEvent.click(groupTrigger('gear'));

		expect(onOpenGroupsChange.mock.lastCall![0]).toEqual(['attributes', 'gear']);
	});

	it('puts rating stats with no percent, and weapon DPS, under Gear', () => {
		mountList();

		const gear = group('gear');
		expect(within(gear).getByRole('checkbox', { name: hitRating.getShortName(Class.ClassWarrior) })).toBeTruthy();
		expect(within(gear).getByRole('checkbox', { name: mainHandDps.getShortName(Class.ClassWarrior) })).toBeTruthy();
		expect(within(gear).getAllByRole('checkbox')).toHaveLength(2);
	});

	it('leaves Gear out when the universe has no rating stat or weapon DPS', () => {
		mountList({ universe: [health, strength, physicalDamage] });

		expect(screen.queryByTestId('test-filter-group-gear')).toBeNull();
	});

	// The test setup loads no locale, so every stat reads as its own key; the real names are what collide.
	describe('with the English names', () => {
		beforeAll(() => {
			i18n.addResourceBundle('en', 'translation', english, true, true);
		});
		afterAll(() => {
			i18n.removeResourceBundle('en', 'translation');
		});

		it('lists only the percent where a rating stat and a percent share a short name, and names every stat by its short name', () => {
			const pairs: Array<[Stat, PseudoStat, string]> = [
				[Stat.StatDodgeRating, PseudoStat.PseudoStatDodgePercent, 'Dodge'],
				[Stat.StatParryRating, PseudoStat.PseudoStatParryPercent, 'Parry'],
				[Stat.StatBlockRating, PseudoStat.PseudoStatBlockPercent, 'Block'],
				[Stat.StatExpertiseRating, PseudoStat.PseudoStatExpertisePercent, 'Expertise'],
			];
			const paired = pairs.flatMap(([rating, percent]) => [UnitStat.fromStat(rating), UnitStat.fromPseudoStat(percent)]);
			mountList({ universe: [hitRating, meleeHit, ...paired], shown: [] });

			for (const [, , name] of pairs) {
				expect(screen.getAllByRole('checkbox', { name })).toHaveLength(1);
			}
			expect(screen.queryByRole('checkbox', { name: /Rating|%/ })).toBeNull();
			expect(screen.getByRole('checkbox', { name: 'Hit' })).toBeTruthy();
			expect(screen.getByRole('checkbox', { name: 'Melee Hit' })).toBeTruthy();
			expect(within(group('gear')).getAllByRole('checkbox')).toHaveLength(1);
		});

		it('keeps the short name of a rating stat whose percent is not in the universe', () => {
			mountList({ universe: [UnitStat.fromStat(Stat.StatDodgeRating)], shown: [] });

			expect(within(group('gear')).getByRole('checkbox', { name: 'Dodge' })).toBeTruthy();
		});

		describe('where a percent row stands for the rating stat it hides', () => {
			const dodgeRating = UnitStat.fromStat(Stat.StatDodgeRating);
			const dodgePercent = UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent);
			const pair = [dodgeRating, dodgePercent];
			const dodgeRow = () => screen.getByRole('checkbox', { name: 'Dodge' });

			it('reads ticked when only the hidden rating stat is shown, and counts it', () => {
				mountList({ universe: pair, shown: [dodgeRating] });

				expect(screen.queryByTestId('test-filter-group-gear')).toBeNull();
				expect(groupTrigger('defense').textContent).toBe(groupTitle(i18n.t('sidebar.character_stats.groups.defense'), 1, 1));
				expect(dodgeRow().getAttribute('aria-checked')).toBe('true');
			});

			it('shows the rating stat along with the percent when ticked', () => {
				const { onChange } = mountList({ universe: pair, shown: [] });

				fireEvent.click(dodgeRow());

				expect(keys(onChange.mock.lastCall![0])).toEqual(keys(pair));
			});

			it('hides the rating stat along with the percent when unticked', () => {
				const { onChange } = mountList({ universe: pair, shown: pair });

				fireEvent.click(dodgeRow());

				expect(onChange.mock.lastCall![0]).toEqual([]);
			});

			it('hides a rating stat shown alone when its row is unticked', () => {
				const { onChange } = mountList({ universe: pair, shown: [dodgeRating] });

				fireEvent.click(dodgeRow());

				expect(onChange.mock.lastCall![0]).toEqual([]);
			});
		});
	});

	it("lists each group's stats in the group's own order, whatever the universe order", () => {
		const blockRating = UnitStat.fromStat(Stat.StatBlockRating);
		const dodgeRating = UnitStat.fromStat(Stat.StatDodgeRating);
		const healingPower = UnitStat.fromStat(Stat.StatHealingPower);
		const spellDamage = UnitStat.fromStat(Stat.StatSpellDamage);
		mountList({ universe: [mainHandDps, blockRating, dodgeRating, hitRating, healingPower, spellDamage] });

		expectGroupOrder('gear', [hitRating, dodgeRating, blockRating, mainHandDps]);
		expectGroupOrder('spell', [spellDamage, healingPower]);
	});

	it('puts the off-hand hit and crit percents under Physical, each after its melee stat', () => {
		const meleeCrit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent);
		const offHandHit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent);
		const offHandCrit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent);
		mountList({ universe: [offHandCrit, offHandHit, meleeCrit, meleeHit] });

		expectGroupOrder('physical', [meleeHit, offHandHit, meleeCrit, offHandCrit]);
	});

	it('checks the shown stats only', () => {
		mountList();

		expect(checkbox(strength).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(meleeHit).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(agility).getAttribute('aria-checked')).toBe('false');
	});

	it('hands back the whole new list in universe order on a toggle', () => {
		const { onChange } = mountList();

		fireEvent.click(checkbox(hitRating));
		expect(keys(onChange.mock.lastCall![0])).toEqual(keys([strength, meleeHit, hitRating]));

		fireEvent.click(checkbox(strength));
		expect(keys(onChange.mock.lastCall![0])).toEqual(keys([meleeHit]));
	});

	it('shows every stat on All', () => {
		const { onChange } = mountList();

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.all') }));

		expect(keys(onChange.mock.lastCall![0])).toEqual(keys(universe));
	});

	it('offers Spec defaults only once the panel is customised', () => {
		const untouched = mountList();
		expect(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }).hasAttribute('disabled')).toBe(true);
		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }));
		expect(untouched.onReset).not.toHaveBeenCalled();
	});

	it('clears the slot on Spec defaults once customised', () => {
		const { onReset } = mountList({ customised: true });

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }));

		expect(onReset).toHaveBeenCalledTimes(1);
	});
});

describe('StatFilterPopover', () => {
	const mountPopover = (customised = false) =>
		render(
			<StatFilterPopover
				universe={universe}
				shown={[strength]}
				customised={customised}
				playerClass={Class.ClassWarrior}
				onChange={() => {}}
				onReset={() => {}}
				groupsStorageKey="test__groups"
				ariaLabel="Filter test stats"
				testId="test-filter"
			/>,
		);

	it('anchors a tooltip on the trigger that reads the popover title', () => {
		mountPopover();

		const tooltipId = screen.getByTestId('test-filter-button').getAttribute('data-tooltip-id');

		expect(tooltipId).toBeTruthy();
		expect(screen.getByRole('button', { name: 'Filter test stats' }).getAttribute('data-tooltip-id')).toBe(tooltipId);
	});

	it('opens the list from the trigger it names, under the shared title', () => {
		mountPopover();

		fireEvent.click(screen.getByRole('button', { name: 'Filter test stats' }));

		const popup = screen.getByRole('dialog', { name: i18n.t('stat_filters.title') });
		expect(within(popup).getAllByRole('heading')[0].textContent).toBe(i18n.t('stat_filters.title'));
		expect(checkbox(strength).getAttribute('aria-checked')).toBe('true');
	});

	it('marks the trigger customised only while the slot is', () => {
		const { unmount } = mountPopover();
		expect(screen.getByTestId('test-filter-button').hasAttribute('data-customised')).toBe(false);
		unmount();

		mountPopover(true);
		expect(screen.getByTestId('test-filter-button').hasAttribute('data-customised')).toBe(true);
	});
});
