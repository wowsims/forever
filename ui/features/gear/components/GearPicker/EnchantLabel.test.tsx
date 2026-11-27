import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/proto/enchants', () => ({ getEnchantDescription: async (enchant: Enchant) => enchant.name }));

const { EnchantLabel } = await import('./EnchantLabel');

const enchant = (spellId: number) => ({ spellId, itemId: 0, effectId: spellId, name: `Enchant ${spellId}`, type: 0, quality: 4 }) as unknown as Enchant;

describe('EnchantLabel', () => {
	const anchor = (container: HTMLElement) => container.querySelector<HTMLAnchorElement>('a')!;

	it('opens the enchant’s own tooltip, following the enchant in the slot', () => {
		const { container, rerender } = render(<EnchantLabel className="item-picker-enchant" enchant={enchant(1)} />);
		expect(anchor(container).dataset.enchantTooltip).toBe('1');

		rerender(<EnchantLabel className="item-picker-enchant" enchant={enchant(2)} />);
		expect(anchor(container).dataset.enchantTooltip).toBe('2');
		expect(anchor(container).dataset.enchantSpell).toBe('2');
	});

	it('keeps the quick enchant list on a label that opens it', () => {
		const { container } = render(<EnchantLabel className="item-picker-enchant" enchant={enchant(1)} tooltipId="quick-swap" />);
		expect(anchor(container).dataset.tooltipId).toBe('quick-swap');
	});
});
