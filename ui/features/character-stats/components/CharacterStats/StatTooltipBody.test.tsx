import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { makeStatAttribution, showStat } from '@testing/storybook/fixtures/character-stats';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { StatTooltipBody, type StatTooltipBodyProps } from './StatTooltipBody';

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
});
