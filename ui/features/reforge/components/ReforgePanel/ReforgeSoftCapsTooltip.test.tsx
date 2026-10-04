import { StatCapType } from '@generated/proto/api';
import { Class, PseudoStat } from '@generated/proto/common';
import { StatCap } from '@sim/proto/stats';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ReforgeSoftCapsTooltip } from './ReforgeSoftCapsTooltip';

const critTaken = () =>
	StatCap.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent, {
		breakpoints: [5.6],
		capType: StatCapType.TypeSoftCap,
		postCapEPs: [0],
	});

describe('ReforgeSoftCapsTooltip', () => {
	// Prot Paladin's only soft cap is ReducedCritTakenPercent, which has no single rating
	// conversion, so the post-cap EP cell gets null here; a .toFixed() on it would take the whole
	// sidebar down when the Suggest Reforges button is hovered.
	it('renders a soft cap whose stat has no percent conversion', () => {
		const { container } = render(<ReforgeSoftCapsTooltip playerClass={Class.ClassPaladin} softCaps={[critTaken()]} additionalInformation={{}} />);

		expect(container.querySelectorAll('tr').length).toBeGreaterThan(0);
		expect(container.textContent).toContain('0.00');
	});

	it('gives each soft-capped stat its own table, with one row per breakpoint', () => {
		const hit = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
			breakpoints: [5, 9],
			capType: StatCapType.TypeThreshold,
			postCapEPs: [0],
		});

		render(<ReforgeSoftCapsTooltip playerClass={Class.ClassWarrior} softCaps={[hit, critTaken()]} additionalInformation={{}} />);
		const [hitTable, critTable] = screen.getAllByTestId('table');

		expect(screen.getAllByTestId('table')).toHaveLength(2);
		expect(hitTable.querySelectorAll('tbody tr')).toHaveLength(3);
		expect(critTable.querySelectorAll('tbody tr')).toHaveLength(2);
	});
});
