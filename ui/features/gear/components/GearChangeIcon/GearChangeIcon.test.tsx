import { ItemSlot } from '@generated/proto/common';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import type { GearChangeIconViewProps } from './GearChangeIcon';

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.png', name: 'Item', href: 'https://wowhead/item', ready: true }) }));
vi.mock('@sim/hooks/useWowheadDataset', () => ({ useWowheadDataset: () => ({}) }));
vi.mock('@sim/proto/action_id/tooltip_data', () => ({ equippedItemWowheadTooltipData: () => Promise.resolve('') }));

const { GearChangeIcon, GearChangeIconView } = await import('./GearChangeIcon');
const { enchantChanged } = await import('./utils');

const item = () =>
	({
		asActionId: () => ({}),
	}) as any;

const renderIcon = (props: Parameters<typeof GearChangeIcon>[0]) =>
	renderWithHost(<GearChangeIcon {...props} />, { player: { sim: { store: createSimStore() } } });

describe('GearChangeIcon', () => {
	it('renders the empty frame for a slot with no item', () => {
		const { container } = renderIcon({ slot: ItemSlot.ItemSlotHead });

		expect(container.querySelector('[data-testid="item-picker-icon-wrapper"]')?.getAttribute('style')).toContain(
			getEmptySlotIconUrl(ItemSlot.ItemSlotHead),
		);
		expect(container.querySelector('[data-testid="gear-change-icon-link"]')?.getAttribute('href')).toBeNull();
	});

	it('paints the item icon and its wowhead link', () => {
		const { container } = renderIcon({ slot: ItemSlot.ItemSlotHead, item: item() });

		expect(container.querySelector('[data-testid="item-picker-icon-wrapper"]')?.getAttribute('style')).toContain('icon.png');
		expect(container.querySelector('[data-testid="gear-change-icon-link"]')!.getAttribute('href')).toBe('https://wowhead/item');
	});

	it.each([
		['draws no enchant marker by default', undefined, false],
		['draws the enchant marker when asked to', true, true],
	])('%s', (_, showEnchant, marked) => {
		const enchanted = { ...item(), enchant: { effectId: 1900, spellId: 20034, name: 'Crusader' } };
		const { container } = renderIcon({ slot: ItemSlot.ItemSlotMainHand, item: enchanted, previousItem: item(), showEnchant });

		expect(!!container.querySelector('[data-testid="gear-change-icon-enchant"]')).toBe(marked);
	});
});

describe('enchantChanged', () => {
	const withEnchant = (effectId?: number) => ({ ...item(), enchant: effectId ? { effectId, name: `Enchant ${effectId}` } : null }) as any;

	it.each([
		['the same enchant', withEnchant(7), withEnchant(7), false],
		['no enchant before or after', withEnchant(), withEnchant(), false],
		['an enchant that differs from the previous item', withEnchant(8), withEnchant(7), true],
		['an enchant the previous item did not have', withEnchant(8), withEnchant(), true],
		['a removed enchant', withEnchant(), withEnchant(7), false],
		['no previous item', withEnchant(8), undefined, false],
		['no item', undefined, withEnchant(7), false],
	])('answers %s', (_, after, before, expected) => {
		expect(enchantChanged(after, before)).toBe(expected);
	});
});

const renderView = (props: Partial<GearChangeIconViewProps> = {}) =>
	render(<GearChangeIconView slot={ItemSlot.ItemSlotHead} equipped iconUrl="helm.jpg" href="https://example.com/item" {...props} />);

const frame = () => screen.getByTestId('item-picker-icon-wrapper');
const link = () => screen.getByTestId('gear-change-icon-link');

describe('GearChangeIconView', () => {
	it('paints the item icon and links it, marked as carrying its own tooltip', () => {
		renderView({ wowhead: { 'data-wowhead': 'item=12640' } });
		expect(frame().style.backgroundImage).toContain('helm.jpg');
		expect(link().getAttribute('href')).toBe('https://example.com/item');
		expect(link().dataset.whtticon).toBe('false');
		expect(link().dataset.wowhead).toBe('item=12640');
	});

	it('draws the empty frame with no link for an unfilled slot', () => {
		renderView({ equipped: false });
		expect(frame().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotHead));
		expect(link().hasAttribute('href')).toBe(false);
		expect(link().hasAttribute('data-whtticon')).toBe(false);
	});

	it.each([true, false])('marks the enchant as changed only when it differs from the previous one (changed: %s)', enchantChanged => {
		renderView({ enchant: { name: 'Crusader', iconUrl: 'crusader.jpg' }, enchantChanged });
		expect(screen.getByTestId('gear-change-icon-enchant').hasAttribute('data-changed')).toBe(enchantChanged);
	});

	it('links the changed-enchant marker to the enchant, carrying its wowhead tooltip', () => {
		renderView({
			enchant: { name: 'Crusader', iconUrl: 'crusader.jpg', href: 'https://example.com/spell=20034' },
			enchantWowhead: { 'data-wowhead': 'spell=20034' },
		});

		const marker = screen.getByTestId('gear-change-icon-enchant');
		expect(marker.getAttribute('href')).toBe('https://example.com/spell=20034');
		expect(marker.dataset.wowhead).toBe('spell=20034');
		expect(marker.getAttribute('aria-label')).toBe('Crusader');
	});
});
