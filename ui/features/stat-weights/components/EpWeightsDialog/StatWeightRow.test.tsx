import type { StatWeightsResult } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import * as Mechanics from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { createSimStore } from '@sim/state/sim_store';
import { stDevToConf90 } from '@sim/utils/math';
import { act, fireEvent, render } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { emptyStatWeightsResult } from '../../model/ep_math';
import { StatWeightRow, StatWeightRowView, type StatWeightRowViewProps } from './StatWeightRow';

const AGILITY = UnitStat.fromStat(Stat.StatAgility);

// The settings are the real class on a real store, so the toggle goes through the same store write
// and subscription the dialog uses. The player only answers what the row reads.
const setup = (editResult: (result: StatWeightsResult) => void = () => {}, stat: UnitStat = AGILITY) => {
	const store = createSimStore();
	const storage = new Map<string, string>();
	const player = {
		storeKey: 0,
		sim: {
			store,
			env: { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } },
		},
		playerClass: { classID: Class.ClassWarrior },
		getClass: () => Class.ClassWarrior,
		getEpWeights: () => new Stats(),
		setEpWeights: () => {},
	} as unknown as Player<any>;
	const settings = new StatWeightActionSettings(player, 'stat-weight-row-test');

	const result = emptyStatWeightsResult();
	AGILITY.setProtoValue(result.dps!.weights!, 1.9);
	AGILITY.setProtoValue(result.dps!.epValues!, 0.79);
	editResult(result);

	const view = render(
		<Table>
			<tbody>
				<StatWeightRow
					stat={stat}
					result={result}
					iterations={1000}
					epRatios={[1, 0, 0, 0, 0, 0]}
					epWeights={new Stats()}
					settings={settings}
					player={player}
					epReferenceStat={Stat.StatStrength}
					includable
					showThreatMetrics={false}
					displayMetrics={{ damage: true, healing: false, threat: false }}
				/>
			</tbody>
		</Table>,
	);
	return { ...view, settings };
};

describe('StatWeightRow', () => {
	it('switches its cells to N/A when its stat is toggled out after a result, and back when toggled in, from the same result', () => {
		const { container, settings } = setup();
		const epCell = () => container.querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!;
		const toggle = () => container.querySelector('[data-testid="swcalc-include-toggle"] [data-testid="boolean-picker-input"]')!;
		expect(epCell().textContent).toBe('0.79');

		act(() => {
			fireEvent.click(toggle());
		});
		expect(settings.isStatExcludedFromCalc(Stat.StatAgility)).toBe(true);
		expect(container.querySelectorAll('[data-notapplicable]')).toHaveLength(2);

		act(() => {
			fireEvent.click(toggle());
		});
		expect(settings.isStatExcludedFromCalc(Stat.StatAgility)).toBe(false);
		expect(container.querySelector('[data-notapplicable]')).toBeNull();
		expect(epCell().textContent).toBe('0.79');
	});

	describe('with the locale loaded', () => {
		beforeAll(() => {
			i18n.addResourceBundle('en', 'translation', english, true, true);
		});
		afterAll(() => {
			i18n.removeResourceBundle('en', 'translation');
		});

		const helpText = (stat: UnitStat, stepBelow = 0, stepAbove = 0) => {
			const view = setup(result => {
				result.stepBelow = new Stats().withUnitStat(stat, stepBelow).toProto();
				result.stepAbove = new Stats().withUnitStat(stat, stepAbove).toProto();
			}, stat);
			const text = view.container.querySelector('[data-testid="stat-weight-help"]')?.getAttribute('aria-label');
			view.unmount();
			return text;
		};

		it.each([
			['Hit Rating', UnitStat.fromStat(Stat.StatHitRating), Mechanics.HIT_RATING_PER_HIT_PERCENT],
			['Crit Rating', UnitStat.fromStat(Stat.StatCritRating), Mechanics.CRIT_RATING_PER_CRIT_PERCENT],
			['Haste Rating', UnitStat.fromStat(Stat.StatHasteRating), Mechanics.HASTE_RATING_PER_HASTE_PERCENT],
			['Expertise Rating', UnitStat.fromStat(Stat.StatExpertiseRating), Mechanics.EXPERTISE_RATING_PER_EXPERTISE_PERCENT],
			['Dodge Rating', UnitStat.fromStat(Stat.StatDodgeRating), Mechanics.DODGE_RATING_PER_DODGE_PERCENT],
			['Melee Crit Percent', UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent), Mechanics.CRIT_RATING_PER_CRIT_PERCENT],
			['Fire Hit Percent', UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire), Mechanics.HIT_RATING_PER_HIT_PERCENT],
		])('states the rating conversion of %s before a run', (_name, stat, ratingPerPercent) => {
			expect(helpText(stat)).toBe(`${ratingPerPercent} rating = 1%.`);
		});

		it('has no help icon for a stat with no conversion and no step', () => {
			expect(helpText(AGILITY)).toBeUndefined();
			expect(helpText(UnitStat.fromStat(Stat.StatDefenseRating))).toBeUndefined();
		});

		it('adds how far the sim moved the stat after a run', () => {
			const hit = Mechanics.HIT_RATING_PER_HIT_PERCENT;
			const crit = Mechanics.CRIT_RATING_PER_CRIT_PERCENT;
			expect(helpText(AGILITY, -10, 10)).toBe('Simulated step: ±10.');
			expect(helpText(UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps), -5, 5)).toBe('Simulated step: ±5.');
			expect(helpText(UnitStat.fromStat(Stat.StatHitRating), 0, 10)).toBe(`${hit} rating = 1%. Simulated step: +10 rating.`);
			expect(helpText(UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent), 0, 1)).toBe(
				`${hit} rating = 1%. Simulated step: +10 rating (+1%).`,
			);
			expect(helpText(UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent), -10 / crit, 10 / crit)).toBe(
				`${crit} rating = 1%. Simulated step: ±10 rating (±0.71%).`,
			);
		});

		it('adds no text to the first cell, which the row is found by', () => {
			const { container } = setup(() => {}, UnitStat.fromStat(Stat.StatCritRating));
			expect(container.querySelector('tr')!.firstElementChild!.textContent).toBe('Crit Rating');
		});
	});

	it("sizes the ± by the stat's own sample count, and by the run's iterations when the result has none", () => {
		const epInterval = (container: HTMLElement) => container.querySelector('[data-column-type="ep"]')!.textContent;
		const withStdev = (result: StatWeightsResult) => AGILITY.setProtoValue(result.dps!.epValuesStdev!, 6);

		const counted = setup(result => {
			withStdev(result);
			result.sampleCounts = new Stats().withStat(Stat.StatAgility, 500).toProto();
		});
		expect(epInterval(counted.container)).toContain(`(${stDevToConf90(6, 500).toFixed(2)})`);
		counted.unmount();

		const uncounted = setup(withStdev);
		expect(epInterval(uncounted.container)).toContain(`(${stDevToConf90(6, 1000).toFixed(2)})`);
	});
});

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const renderRow = (props: Partial<StatWeightRowViewProps> = {}) =>
	render(
		<Table>
			<tbody>
				<StatWeightRowView
					stat={STRENGTH}
					fullName="Spell Hit Percent"
					shortName="Spell Hit"
					metrics={[{ ratioIndex: 0, epRatio: 1 }]}
					iterations={100}
					epDelta={0}
					currentEp={1.5}
					onCurrentEpChange={() => {}}
					showThreatMetrics={false}
					{...props}
				/>
			</tbody>
		</Table>,
	);

describe('StatWeightRowView', () => {
	it('shows the help icon only with a help text', () => {
		expect(renderRow().queryByTestId('stat-weight-help')).toBeNull();
		expect(renderRow({ helpText: '14 rating = 1%.' }).getByLabelText('14 rating = 1%.')).toBeTruthy();
	});

	it('names the row, and ids its two inputs from the sanitised names', () => {
		const onChange = vi.fn();
		const { container, getByText } = renderRow({ include: { value: true, disabled: false, onChange } });
		expect(getByText('Spell Hit Percent')).toBeTruthy();
		expect(container.querySelector('[data-testid="swcalc-include-toggle"] #sw-stat-toggle-spellhitpercent')).not.toBeNull();
		expect(container.querySelector<HTMLInputElement>('[data-testid="current-ep"] input')!.id).toBe('ep-weight-stat-spellhit');
	});

	it('shows the current EP weight and reports an edit', () => {
		const onCurrentEpChange = vi.fn();
		const { container } = renderRow({ onCurrentEpChange });
		const input = container.querySelector<HTMLInputElement>('[data-testid="current-ep"] input')!;
		expect(input.value).toBe('1.50');
		input.value = '2.25';
		fireEvent.change(input);
		expect(onCurrentEpChange).toHaveBeenCalledWith(2.25);
	});

	it('leaves the Update cell empty without an include toggle, and drops it with the column', () => {
		const { container, rerender } = renderRow();
		expect(container.querySelector('[data-testid="swcalc-include-toggle"]')!.childElementCount).toBe(0);

		rerender(
			<Table>
				<tbody>
					<StatWeightRowView
						stat={STRENGTH}
						fullName="Strength"
						shortName="Str"
						metrics={[]}
						iterations={1}
						epDelta={0}
						currentEp={0}
						onCurrentEpChange={() => {}}
						showThreatMetrics={false}
						showUpdateColumn={false}
					/>
				</tbody>
			</Table>,
		);
		expect(container.querySelector('[data-testid="swcalc-include-toggle"]')).toBeNull();
	});

	it('reports an include toggle, and holds a disabled one still', () => {
		const onChange = vi.fn();
		const { container, rerender } = renderRow({ include: { value: true, disabled: false, onChange } });
		fireEvent.click(container.querySelector('#sw-stat-toggle-spellhitpercent')!);
		expect(onChange).toHaveBeenCalledWith(false);

		rerender(
			<Table>
				<tbody>
					<StatWeightRowView
						stat={STRENGTH}
						fullName="Spell Hit Percent"
						shortName="Spell Hit"
						metrics={[]}
						iterations={1}
						epDelta={0}
						currentEp={0}
						onCurrentEpChange={() => {}}
						showThreatMetrics={false}
						include={{ value: true, disabled: true, onChange }}
					/>
				</tbody>
			</Table>,
		);
		expect(container.querySelector<HTMLInputElement>('[data-testid="swcalc-include-toggle"] input')!.disabled).toBe(true);
	});

	it('renders a weight and an EP cell per metric, N/A where there is no result', () => {
		const result = emptyStatWeightsResult();
		STRENGTH.setProtoValue(result.dps!.weights!, 2);
		STRENGTH.setProtoValue(result.dps!.epValues!, 1);
		const { container } = renderRow({
			metrics: [
				{ ratioIndex: 0, epRatio: 1, statWeights: result.dps },
				{ ratioIndex: 1, epRatio: 1 },
			],
		});
		const values = [...container.querySelectorAll('[data-testid="stdev-cell"] [data-testid="results-avg"]')].map(cell => cell.textContent);
		expect(values).toEqual(['2.00', '1.00', 'sidebar.buttons.stat_weights.modal.not_applicable', 'sidebar.buttons.stat_weights.modal.not_applicable']);
	});
});
