import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { PresetBuild } from './preset_utils';
import { PresetBuildChip } from './PresetBuildChip';

const build = { name: 'Phase 1' } as PresetBuild;

describe('PresetBuildChip', () => {
	it('applies its build from the name and anchors the tooltip by that name', () => {
		const onApply = vi.fn();
		render(<PresetBuildChip build={build} active tooltipId="builds" onApply={onApply} />);

		const chip = screen.getByTestId('saved-data-set-chip');
		expect(chip.hasAttribute('data-active')).toBe(true);
		expect(chip.getAttribute('data-tooltip-id')).toBe('builds');
		expect(chip.getAttribute('data-tooltip-content')).toBe('Phase 1');

		fireEvent.click(screen.getByTestId('saved-data-set-name'));
		expect(onApply).toHaveBeenCalledWith(build);
	});
});
