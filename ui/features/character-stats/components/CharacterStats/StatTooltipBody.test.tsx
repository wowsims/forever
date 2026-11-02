import { PseudoStat, Stat } from '@generated/proto/common';
import { computeStatAttribution, Stats, UnitStat } from '@sim/proto/stats';
import { makeStatAttribution, showStat } from '@testing/storybook/fixtures/character-stats';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// The loader that feeds i18next is stubbed empty under vitest, so `t` would otherwise echo the key; only the percent sign is given a value.
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => (key === 'sidebar.character_stats.percent_suffix' ? '%' : key) } }));

import { StatTooltipBody, type StatTooltipBodyProps } from './StatTooltipBody';
const { statDisplayString } = await import('./utils/stat_display');

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

const renderBody = (props: Partial<StatTooltipBodyProps> = {}) =>
	render(
		<StatTooltipBody
			hand={STRENGTH}
			attribution={makeStatAttribution()}
			show={showStat}
			bonusStats={new Stats()}
			bonusStatValue={0}
			total="370"
			{...props}
		/>,
	);

const rows = (container: HTMLElement) =>
	[...container.querySelectorAll('[data-testid="character-stats-tooltip-row"]')].map(row => [...row.children].map(cell => cell.textContent));

describe('StatTooltipBody', () => {
	it('lists each stage, then the total', () => {
		const { container } = renderBody();
		expect(rows(container)).toEqual([
			['sidebar.character_stats.tooltip.base', '120'],
			['sidebar.character_stats.tooltip.gear', '180'],
			['sidebar.character_stats.tooltip.talents', '0'],
			['sidebar.character_stats.tooltip.buffs', '45'],
			['sidebar.character_stats.tooltip.consumes', '25'],
			['sidebar.character_stats.tooltip.debuffs', '0'],
			['sidebar.character_stats.tooltip.total', '370'],
		]);
		expect(container.querySelector('[data-testid="character-stats-tooltip-hand"]')).toBeNull();
	});

	it('adds the bonus row only for a bonus, and heads the list with the hand when given one', () => {
		const { container } = renderBody({ bonusStats: new Stats().withStat(Stat.StatStrength, 15), bonusStatValue: 15, handLabel: 'Main Hand' });
		expect(rows(container)).toContainEqual(['sidebar.character_stats.tooltip.bonus', '15']);
		expect(container.querySelector('[data-testid="character-stats-tooltip-hand"]')!.textContent).toBe('Main Hand');
	});

	it('divides the ranged weapon speed by the final ranged haste', () => {
		const { container } = renderBody({ rangedWeaponSpeed: 3 });
		expect(rows(container).at(-1)).toEqual(['sidebar.character_stats.tooltip.eWS', `${(3 / 1.15).toFixed(2)}s`]);
	});

	describe('on the sheet path, with the real formatter', () => {
		const MELEE_HIT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);
		const BLOCK = UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent);

		// The back end's cumulative gear stage holds the base, the real gear and the bonus rating's percent in one percent pseudo stat; a gated row holds 0.
		// The bonus stats carry the rating alone.
		const tooltipRows = (
			hand: UnitStat,
			{ base = 0, realGear = 0, bonusPercent, gated = false }: { base?: number; realGear?: number; bonusPercent: number; gated?: boolean },
		) => {
			const bonusStats = new Stats().withStat(hand.getRootStat(), hand.convertPercentToRating(bonusPercent)!);
			const baseStage = new Stats().withPseudoStat(hand.getPseudoStat(), base);
			const gearStage = bonusStats.withPseudoStat(hand.getPseudoStat(), gated ? 0 : base + realGear + bonusPercent);
			const attribution = computeStatAttribution(
				{ baseStats: baseStage.toProto(), gearStats: gearStage.toProto(), finalStats: gearStage.toProto() },
				bonusStats,
				new Stats(),
				{},
			);
			const show = (deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean) => statDisplayString({} as never, deltaStats, unitStat, includeBase);
			const { container } = renderBody({
				hand,
				attribution,
				show,
				bonusStats,
				bonusStatValue: bonusStats.getStat(hand.getRootStat()),
				total: show(attribution.final, hand, true),
			});
			return Object.fromEntries(rows(container).map(([label, value]) => [label!.split('.').at(-1), value]));
		};

		it('shows a percent bonus on the Bonus line and takes it out of Gear, leaving the total', () => {
			expect(tooltipRows(MELEE_HIT, { realGear: 3, bonusPercent: 1 })).toMatchObject({ gear: '3.00%', bonus: '1.00%', total: '4.00%' });
		});

		it('keeps real gear that a negative bonus cancels out', () => {
			expect(tooltipRows(MELEE_HIT, { realGear: 3, bonusPercent: -3 })).toMatchObject({ gear: '3.00%', bonus: '-3.00%', total: '0.00%' });
		});

		it('keeps real gear that a larger negative bonus overshoots', () => {
			expect(tooltipRows(MELEE_HIT, { realGear: 1.5, bonusPercent: -2 })).toMatchObject({ gear: '1.50%', bonus: '-2.00%', total: '-0.50%' });
		});

		it('lists base, gear and bonus so they sum to the total', () => {
			expect(tooltipRows(MELEE_HIT, { base: 5, realGear: 1, bonusPercent: -3 })).toMatchObject({
				base: '5.00%',
				gear: '1.00%',
				bonus: '-3.00%',
				total: '3.00%',
			});
		});

		it('never prints a negative zero for Gear when the bonus alone moved the percent', () => {
			const lines = tooltipRows(MELEE_HIT, { base: 6.3, bonusPercent: -0.33 });
			expect(lines.gear).toBe('0.00%');
			expect(Object.values(lines)).not.toContain('-0.00%');
		});

		it('keeps Gear at zero on a block row the back end gates to zero without a shield', () => {
			expect(tooltipRows(BLOCK, { bonusPercent: 1, gated: true })).toMatchObject({ gear: '0.00%', bonus: '1.00%', total: '0.00%' });
		});
	});
});
