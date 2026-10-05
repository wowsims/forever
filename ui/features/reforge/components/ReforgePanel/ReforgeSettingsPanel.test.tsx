import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { ReforgeSettingsPanelView, type ReforgeSettingsPanelViewProps } from './ReforgeSettingsPanel';

const SLOTS = ['customEpToggle', 'statCaps', 'softCaps', 'freezeSlots', 'savedEpWeights'] as const;

const slotProps = Object.fromEntries(SLOTS.map(slot => [slot, <span data-slot={slot} />])) as Record<(typeof SLOTS)[number], ReactNode>;

const renderPanel = (props: Partial<ReforgeSettingsPanelViewProps> = {}) => {
	const onEditWeights = vi.fn();
	const { container } = render(<ReforgeSettingsPanelView {...slotProps} showEpHelp={false} onEditWeights={onEditWeights} {...props} />);
	return { container, onEditWeights };
};

describe('ReforgeSettingsPanelView', () => {
	it('stacks the sections in order and ends on the edit weights button', () => {
		const { container } = renderPanel();

		expect(Array.from(container.querySelectorAll('[data-slot]')).map(slot => slot.getAttribute('data-slot'))).toEqual([...SLOTS]);
		expect(container.lastElementChild).toBe(screen.getByTestId('reforge-edit-weights'));
	});

	it.each([true, false])('explains how to edit weights only while the custom EP values are off (showEpHelp %s)', showEpHelp => {
		renderPanel({ showEpHelp });

		expect(!!screen.queryByText('sidebar.buttons.suggest_reforges.enable_modification')).toBe(showEpHelp);
		expect(!!screen.queryByText('sidebar.buttons.suggest_reforges.hard_cap_info')).toBe(showEpHelp);
	});

	it('opens the weights editor from its button', () => {
		const { onEditWeights } = renderPanel();

		fireEvent.click(screen.getByTestId('reforge-edit-weights'));

		expect(onEditWeights).toHaveBeenCalledOnce();
	});
});
