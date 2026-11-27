import { UIItem } from '@generated/proto/ui';
import { EquippedItem } from '@sim/proto/equipped_item';
import { render, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { useEquippedItemWowheadDataset } from './useEquippedItemWowheadDataset';

// Every item has the site's own tooltip here, so no test fetches the tooltip file.
vi.mock('@sim/proto/tooltips', async importOriginal => ({ ...(await importOriginal<object>()), hasItemTooltip: async () => true }));

const Probe = ({ item }: { item: EquippedItem }) => {
	const wowhead = useEquippedItemWowheadDataset(item);
	return <a {...wowhead} />;
};

describe('useEquippedItemWowheadDataset', () => {
	it('opens the item’s own tooltip, with the suffix it carries', async () => {
		const belt = new EquippedItem({ item: UIItem.create({ id: 16723, scalingOptions: { 0: { ilvl: 58 } } }) });
		const { container } = render(<Probe item={belt} />);
		const anchor = container.querySelector('a')!;

		await waitFor(() => expect(anchor.dataset.itemTooltip).toBe('16723'));
		expect(anchor.hasAttribute('data-item-suffix')).toBe(false);
	});
});
