import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { makeRangedWeapon, makeStatAttribution, showStat } from '@testing/storybook/fixtures/character-stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { StatRowView, type StatRowViewProps } from './StatRow';

const MELEE_HIT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);
const OFF_HAND_HIT = UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent);
const RANGED_HASTE = UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent);
const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

const renderRow = (props: Partial<StatRowViewProps>) =>
	render(
		<Table>
			<tbody>
				<StatRowView
					displayStat={{ stat: STRENGTH }}
					playerClass={Class.ClassWarrior}
					bonusStats={new Stats()}
					attribution={makeStatAttribution()}
					show={showStat}
					rangedWeapon={null}
					dualWielding={false}
					{...props}
				/>
			</tbody>
		</Table>,
	);

const links = (container: HTMLElement) => [...container.querySelectorAll('[data-testid="stat-value-link"]')].map(link => link.textContent);

describe('StatRowView', () => {
	it('shows the final value, and the bonus-stat slot beside it', () => {
		const { container } = renderRow({ bonusStatsLink: <span data-testid="bonus-slot" /> });
		expect(links(container)).toEqual(['370']);
		expect(container.querySelector('[data-testid="bonus-slot"]')).not.toBeNull();
	});

	it('colours the value by the bonus on its stat', () => {
		const { container } = renderRow({ bonusStats: new Stats().withStat(Stat.StatStrength, 10) });
		expect(container.querySelector('[data-tone]')!.getAttribute('data-tone')).toBe('positive');
	});

	it('shows a skeleton and no value while the stats are pending', () => {
		const { container } = renderRow({ pending: true });
		expect(links(container)).toEqual([]);
	});

	it('shows the off-hand half only while dual wielding', () => {
		const displayStat = { stat: MELEE_HIT, offHandStat: OFF_HAND_HIT };
		const { container, rerender } = renderRow({ displayStat });
		expect(links(container)).toEqual(['8.00%']);

		rerender(
			<Table>
				<tbody>
					<StatRowView
						displayStat={displayStat}
						playerClass={Class.ClassWarrior}
						bonusStats={new Stats()}
						attribution={makeStatAttribution()}
						show={showStat}
						rangedWeapon={null}
						dualWielding
					/>
				</tbody>
			</Table>,
		);
		expect(links(container)).toEqual(['8.00%', '7.00%']);
	});

	it("adds the effective weapon speed to a hunter's ranged haste breakdown", async () => {
		const { container } = renderRow({ displayStat: { stat: RANGED_HASTE }, playerClass: Class.ClassHunter, rangedWeapon: makeRangedWeapon() });
		fireEvent.mouseEnter(container.querySelector('[data-testid="stat-value-link"]')!);
		expect(await screen.findByText(`${(2.9 / 1.15).toFixed(2)}s`)).toBeTruthy();
	});
});
