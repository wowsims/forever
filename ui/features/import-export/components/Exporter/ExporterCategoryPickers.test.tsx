import { SimSettingCategories } from '@sim/constants/sim_settings';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// The exporters reach the addon importer, which asks GitHub for its version when its module loads.
// Answered offline here, in `vi.hoisted` so it runs before the imports, no request is left open
// when the window closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

import { defaultExportCategories, EXPORT_CATEGORY_OPTIONS } from '../../exporters';
import { ExporterCategoryPickers } from './ExporterCategoryPickers';

describe('ExporterCategoryPickers', () => {
	it('shows one checkbox per category, checked as the record says', () => {
		const categories = defaultExportCategories();
		render(<ExporterCategoryPickers categories={categories} onChange={() => {}} />);

		const boxes = screen.getAllByRole('checkbox');
		expect(boxes).toHaveLength(EXPORT_CATEGORY_OPTIONS.length);
		EXPORT_CATEGORY_OPTIONS.forEach((option, idx) => {
			expect(boxes[idx].getAttribute('aria-checked')).toBe(String(categories[option.category]));
		});
	});

	it('writes the record in place, then rings onChange', () => {
		const categories = defaultExportCategories();
		const before = categories[SimSettingCategories.Gear];
		const onChange = vi.fn(() => expect(categories[SimSettingCategories.Gear]).toBe(!before));
		render(<ExporterCategoryPickers categories={categories} onChange={onChange} />);

		const gearIdx = EXPORT_CATEGORY_OPTIONS.findIndex(option => option.category === SimSettingCategories.Gear);
		fireEvent.click(screen.getAllByRole('checkbox')[gearIdx]);

		expect(onChange).toHaveBeenCalledOnce();
		expect(categories[SimSettingCategories.Gear]).toBe(!before);
	});
});
