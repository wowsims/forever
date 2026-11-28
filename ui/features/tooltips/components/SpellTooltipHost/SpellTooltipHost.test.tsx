import { Class } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { Database } from '@sim/proto/database';
import { Gear } from '@sim/proto/gear';
import { enchantTooltipAnchor, itemTooltipAnchor, loadTooltips, spellTooltipAnchor } from '@sim/proto/tooltips';
import { renderWithHost } from '@testing/react';
import { fireEvent, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

const store = vi.hoisted(() => ({ currentStats: {}, talentsString: '' }) as Record<string, unknown>);
vi.mock('@sim/hooks/usePlayerStore', () => ({ usePlayerStore: (field: string) => store[field] }));

const { SpellTooltipHost } = await import('./SpellTooltipHost');

const FIREBALL = 100;
const IMPROVED_REVENGE = 200;
const AGILITY_SPELL = 300;
const ELIXIR = 400;
const AGILITY_EFFECT = 50;

const TOOLTIPS = {
	spells: {
		[FIREBALL]: { n: 'Fireball', d: ['Hurls a fiery ball.'] },
		[IMPROVED_REVENGE]: { n: 'Improved Revenge', m: 3, d: ['Increases Revenge damage.'] },
		[AGILITY_SPELL]: { n: 'Enchant Gloves - Agility', d: ['Permanently enchants gloves.'] },
	},
	items: { [ELIXIR]: { n: 'Elixir of the Mongoose', q: 2 } },
	sets: {},
};

const ENCHANTS = [UIEnchant.create({ effectId: AGILITY_EFFECT, spellId: AGILITY_SPELL, name: 'Enchant Gloves - Agility' })];

const wowheadLink = () => screen.findByRole('link', { name: i18n.t('tooltips.wowhead_link') });

/** Hovers the anchor and returns the link its tooltip shows. */
const linkOf = async (anchor: HTMLElement) => {
	fireEvent.mouseOver(anchor);
	return wowheadLink();
};

const renderHost = (anchors: React.ReactNode) =>
	renderWithHost(
		<>
			{anchors}
			<SpellTooltipHost />
		</>,
		{ player: { getClass: () => Class.ClassWarrior } },
	);

describe('SpellTooltipHost', () => {
	beforeAll(async () => {
		vi.stubGlobal('fetch', async (url: string) => ({ ok: true, json: async () => (url.endsWith('/tooltips.json') ? TOOLTIPS : {}) }));
		await loadTooltips();
	});

	beforeEach(() => {
		store.gear = new Gear({});
		vi.spyOn(Database, 'getSync').mockReturnValue({
			lookupItemSpec: () => null,
			getItemById: () => undefined,
			getEnchantsByEffectId: (effectId: number) => ENCHANTS.filter(enchant => enchant.effectId === effectId),
		} as unknown as Database);
	});

	it('links a spell to the page its anchor links to, in a new tab', async () => {
		const href = ActionId.makeSpellUrl(FIREBALL);
		renderHost(<a href={href}>Fireball</a>);
		const link = await linkOf(screen.getByText('Fireball'));
		expect(link.getAttribute('href')).toBe(href);
		expect(link.getAttribute('target')).toBe('_blank');
		expect(link.getAttribute('rel')).toBe('noopener noreferrer');
	});

	it('links a talent to the page its anchor links to, at the rank taken', async () => {
		const href = ActionId.makeSpellUrl(IMPROVED_REVENGE, 2);
		renderHost(
			<a href={href} {...spellTooltipAnchor(IMPROVED_REVENGE, { talent: true })}>
				Improved Revenge
			</a>,
		);
		const link = await linkOf(screen.getByText('Improved Revenge'));
		expect(screen.getByTestId('talent-tooltip').contains(link)).toBe(true);
		expect(link.getAttribute('href')).toBe(href);
		expect(new URL(href).searchParams.get('rank')).toBe('2');
	});

	it('links a spell anchor with no link of its own to the spell’s page', async () => {
		renderHost(<span {...spellTooltipAnchor(FIREBALL)}>Fireball</span>);
		const link = await linkOf(screen.getByText('Fireball'));
		expect(link.getAttribute('href')).toContain(`/spell=${FIREBALL}`);
	});

	it('links an item to its page with its random suffix and enchant', async () => {
		renderHost(<span {...itemTooltipAnchor(ELIXIR, { randomSuffix: 7, enchant: 9 })}>Elixir</span>);
		const link = new URL((await linkOf(screen.getByText('Elixir'))).getAttribute('href')!);
		expect(link.pathname).toContain(`/item=${ELIXIR}`);
		expect(link.searchParams.get('bonus')).toBe('7');
		expect(link.searchParams.get('ench')).toBe('9');
	});

	it('links an enchant to the spell that applies it', async () => {
		renderHost(<span {...enchantTooltipAnchor({ effectId: AGILITY_EFFECT, spellId: AGILITY_SPELL, itemId: 0 })}>Agility</span>);
		const link = await linkOf(screen.getByText('Agility'));
		expect(screen.getByTestId('enchant-tooltip').contains(link)).toBe(true);
		expect(link.getAttribute('href')).toContain(`/spell=${AGILITY_SPELL}`);
	});
});
