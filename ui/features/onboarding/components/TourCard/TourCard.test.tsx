import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { TourCard, type TourCardProps } from './TourCard';

const card = (props: Partial<TourCardProps> = {}) => {
	const handlers = { onNext: vi.fn(), onPrevious: vi.fn(), onSkip: vi.fn() };
	render(<TourCard kind="step" title="Gear" body="Click a slot." position={2} total={5} anchor={null} {...handlers} {...props} />);
	return handlers;
};

describe('TourCard', () => {
	it('shows the step, how far along the tour is, and moves on, back or out', () => {
		const { onNext, onPrevious, onSkip } = card();
		expect(screen.getByRole('dialog').getAttribute('aria-labelledby')).toBe(screen.getByTestId('tour-card-title').id);
		expect(screen.getByTestId('tour-card-title').textContent).toBe('Gear');
		expect(screen.getByTestId('tour-card-progress').textContent).toContain(i18n.t('onboarding.progress', { position: 2, total: 5 }));
		fireEvent.click(screen.getByTestId('tour-card-next'));
		fireEvent.click(screen.getByTestId('tour-card-previous'));
		fireEvent.click(screen.getByTestId('tour-card-skip'));
		expect(onNext).toHaveBeenCalledOnce();
		expect(onPrevious).toHaveBeenCalledOnce();
		expect(onSkip).toHaveBeenCalledOnce();
	});

	it('leaves the tour on Escape', () => {
		const { onSkip } = card();
		fireEvent.keyDown(screen.getByTestId('tour-card-next'), { key: 'Escape' });
		expect(onSkip).toHaveBeenCalledOnce();
	});

	it('opens without a count or a way back, and closes with Done instead of Skip', () => {
		card({ kind: 'welcome', onPrevious: undefined });
		expect(screen.queryByTestId('tour-card-progress')).toBeNull();
		expect(screen.queryByTestId('tour-card-previous')).toBeNull();
		expect(screen.getByTestId('tour-card-next').textContent).toBe(i18n.t('onboarding.start'));
	});

	it('ends the tour from the last card', () => {
		const { onSkip, onNext } = card({ kind: 'finish' });
		expect(screen.queryByTestId('tour-card-skip')).toBeNull();
		expect(screen.getByTestId('tour-card-next').textContent).toBe(i18n.t('onboarding.finish'));
		fireEvent.click(screen.getByTestId('tour-card-next'));
		expect(onSkip).toHaveBeenCalledOnce();
		expect(onNext).not.toHaveBeenCalled();
	});

	it('offers to pass over a step that waits for the player', () => {
		card({ waiting: true });
		expect(screen.getByTestId('tour-card-next').textContent).toBe(i18n.t('onboarding.skip_step'));
	});
});
