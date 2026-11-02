import { ItemQuality, Profession } from '@generated/proto/common';
import { DungeonDifficulty, RepLevel, UIEnchant as Enchant, UIItem as Item, UIItem_FactionRestriction } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { professionNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { Sim } from '@sim/sim';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemListRowView } from './ItemListRow';
import { ItemSource } from './ItemSource';

interface DbStub {
	zones?: Record<number, { id: number; name: string }>;
	npcs?: Record<number, { id: number; name: string }>;
	factions?: Record<number, { id: number; name: string }>;
}

const sim = ({ zones = {}, npcs = {}, factions = {} }: DbStub = {}) =>
	({
		db: {
			getZone: (id: number) => zones[id] ?? null,
			getNpc: (id: number) => npcs[id] ?? null,
			getRepFaction: (id: number) => factions[id] ?? null,
		},
	}) as unknown as Sim;

const mount = (item: Item, dbStub: DbStub = {}) => render(<ItemSource item={item} sim={sim(dbStub)} />);

describe('ItemSource', () => {
	it('counts a repeated source once', () => {
		const crafted = { source: { oneofKind: 'crafted' as const, crafted: { profession: Profession.Blacksmithing, spellId: 3321 } } };
		const { getByRole, queryByTestId } = mount(Item.create({ id: 3471, sources: [crafted, crafted] }));
		expect(queryByTestId('item-source-more')).toBeNull();
		expect(getByRole('link').textContent).toBe(professionNames.get(Profession.Blacksmithing));
	});

	it('links to the item page for an item with a pvp source', () => {
		const item = Item.create({ id: 200, sources: [{ source: { oneofKind: 'pvp', pvp: {} } }] });
		const { container, getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeItemUrl(200));
		expect(container.textContent).toBe('common.sources.pvp');
	});

	it('shows the rep source before the pvp source, and the pvp source before the vendor', () => {
		const rep = { source: { oneofKind: 'rep', rep: { repFactionId: 730, repLevel: RepLevel.RepLevelRevered, factionId: 0 } } } as const;
		const vendor = { source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'Vendor', zoneId: 0 } } } as const;
		const pvp = { source: { oneofKind: 'pvp', pvp: {} } } as const;

		const withRep = mount(Item.create({ id: 210, sources: [pvp, rep] }), { factions: { 730: { id: 730, name: 'Stormpike Guard' } } });
		expect(withRep.container.textContent).toContain('Stormpike Guard');
		withRep.unmount();

		const withVendor = mount(Item.create({ id: 211, sources: [vendor, pvp] }));
		expect(withVendor.getAllByRole('link')[0].textContent).toBe('common.sources.pvp');
	});

	it('shows the quest of an item whose sources list the vendor first', () => {
		const item = Item.create({
			id: 212,
			sources: [
				{ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'Vendor', zoneId: 0 } } },
				{ source: { oneofKind: 'quest', quest: { id: 7, name: 'Bounty' } } },
			],
		});
		const { getByRole } = mount(item);

		expect(getByRole('link').textContent).toBe('common.sources.questBounty');
	});

	it('renders nothing for an item with no sources', () => {
		const item = Item.create({ id: 300, name: 'Plain Sword', sources: [], randomSuffixOptions: [1] });
		const { container } = mount(item);

		expect(container.innerHTML).toBe('');
	});

	it('links to the crafting spell for a crafted source with a spell id', () => {
		const item = Item.create({
			id: 400,
			sources: [{ source: { oneofKind: 'crafted', crafted: { profession: Profession.Blacksmithing, spellId: 5000 } } }],
		});
		const { container, getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeSpellUrl(5000));
		expect(container.textContent).toBe(professionNames.get(Profession.Blacksmithing));
	});

	it('links to the item page for a crafted source without a spell id', () => {
		const item = Item.create({
			id: 401,
			sources: [{ source: { oneofKind: 'crafted', crafted: { profession: Profession.Tailoring, spellId: 0 } } }],
		});
		const { getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeItemUrl(401));
	});

	it('renders nothing and logs an error when the drop source zone is unknown', () => {
		const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
		const item = Item.create({
			id: 500,
			sources: [{ source: { oneofKind: 'drop', drop: { zoneId: 1, npcId: 1, difficulty: 0, otherName: '', category: '' } } }],
		});
		const { container } = mount(item, { zones: {}, npcs: {} });

		expect(container.innerHTML).toBe('');
		expect(errorSpy).toHaveBeenCalled();
		errorSpy.mockRestore();
	});

	it('links to the npc for a drop source with a known npc, showing zone, difficulty, and npc name', () => {
		const item = Item.create({
			id: 600,
			sources: [
				{
					source: {
						oneofKind: 'drop',
						drop: { zoneId: 10, npcId: 20, difficulty: DungeonDifficulty.DifficultyRaid, otherName: '', category: 'Hard Mode' },
					},
				},
			],
		});
		const { container, getByRole } = mount(item, { zones: { 10: { id: 10, name: 'Icecrown Citadel' } }, npcs: { 20: { id: 20, name: 'The Lich King' } } });

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(20));
		expect(container.textContent).toBe('Icecrown Citadel (common.dungeon_difficulties.raid)The Lich King - Hard Mode');
	});

	it('shows an open world drop with no difficulty', () => {
		const item = Item.create({
			id: 602,
			sources: [
				{
					source: {
						oneofKind: 'drop',
						drop: { zoneId: 28, npcId: 4494, difficulty: DungeonDifficulty.DifficultyUnknown, otherName: '', category: '' },
					},
				},
			],
		});
		const { container, getByRole } = mount(item, {
			zones: { 28: { id: 28, name: 'Western Plaguelands' } },
			npcs: { 4494: { id: 4494, name: 'Scarlet Spellbinder' } },
		});

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(4494));
		expect(container.textContent).toBe('Western PlaguelandsScarlet Spellbinder');
	});

	it('names a drop with no zone and links to the item, logging nothing', () => {
		const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
		const item = Item.create({
			id: 603,
			sources: [{ source: { oneofKind: 'drop', drop: { zoneId: 0, npcId: 0, difficulty: 0, otherName: 'World Drop', category: '' } } }],
		});
		const { container, getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeItemUrl(603));
		expect(container.textContent).toBe('World Drop');
		expect(errorSpy).not.toHaveBeenCalled();
		errorSpy.mockRestore();
	});

	it('shows an enchant, which passes its own item id, its sources and no faction restriction', () => {
		const enchant = Enchant.create({
			itemId: 2304,
			sources: [{ source: { oneofKind: 'crafted', crafted: { profession: Profession.Leatherworking, spellId: 2152 } } }],
		});
		const { container, getByRole } = render(<ItemSource item={{ id: enchant.itemId, sources: enchant.sources }} sim={sim()} />);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeSpellUrl(2152));
		expect(container.textContent).toBe(professionNames.get(Profession.Leatherworking));
	});

	it.each([
		['with no npc but another name', 'Trash', 'Deadwind PassTrash'],
		['with neither npc nor other name', '', 'Deadwind Pass'],
	])('links to the zone for a drop source %s', (_case, otherName, text) => {
		const item = Item.create({
			id: 601,
			sources: [{ source: { oneofKind: 'drop', drop: { zoneId: 11, npcId: 0, difficulty: 0, otherName, category: '' } } }],
		});
		const { container, getByRole } = mount(item, { zones: { 11: { id: 11, name: 'Deadwind Pass' } } });

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeZoneUrl(11));
		expect(container.textContent).toBe(text);
	});

	it.each([
		['an alliance', UIItem_FactionRestriction.ALLIANCE_ONLY, 'alliance'],
		['a horde', UIItem_FactionRestriction.HORDE_ONLY, 'horde'],
	])('shows %s icon for a quest available only to that faction', (_case, factionRestriction, icon) => {
		const item = Item.create({
			id: 700,
			factionRestriction,
			sources: [{ source: { oneofKind: 'quest', quest: { id: 1, name: 'A Call to Arms' } } }],
		});
		const { getByRole, getByAltText } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeQuestUrl(1));
		expect(getByAltText('').getAttribute('src')).toBe(`/forever/assets/img/${icon}.png`);
	});

	it('shows no faction icon for a quest with no faction restriction', () => {
		const item = Item.create({
			id: 702,
			factionRestriction: UIItem_FactionRestriction.UNSPECIFIED,
			sources: [{ source: { oneofKind: 'quest', quest: { id: 3, name: 'Neutral Ground' } } }],
		});
		const { container, queryByRole } = mount(item);

		expect(queryByRole('img')).toBeNull();
		expect(container.textContent).toBe('common.sources.questNeutral Ground');
	});

	it('finds and renders a rep source even when it is not the first entry', () => {
		const item = Item.create({
			id: 800,
			sources: [
				{ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'Ignored Vendor', zoneId: 0 } } },
				{ source: { oneofKind: 'rep', rep: { repFactionId: 270, repLevel: RepLevel.RepLevelRevered, factionId: 0 } } },
			],
		});
		const { getByRole } = mount(item, { factions: { 270: { id: 270, name: 'Zandalar Tribe' } } });

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeFactionUrl(270));
		expect(getByRole('link').textContent).toBe(`Zandalar Tribe - ${REP_LEVEL_NAMES[RepLevel.RepLevelRevered]}`);
	});

	describe('the other sources', () => {
		const vendor = (npcId: number, npcName: string) => ({ source: { oneofKind: 'soldBy', soldBy: { npcId, npcName, zoneId: 0 } } }) as const;
		const drop = (npcId: number) => ({ source: { oneofKind: 'drop', drop: { zoneId: 28, npcId, difficulty: 0, otherName: '', category: '' } } }) as const;
		const quest = { source: { oneofKind: 'quest', quest: { id: 7, name: 'Bounty' } } } as const;
		const db: DbStub = { zones: { 28: { id: 28, name: 'Western Plaguelands' } }, npcs: { 4494: { id: 4494, name: 'Scarlet Spellbinder' } } };
		const threeSources = () => Item.create({ id: 900, sources: [vendor(5, 'Quartermaster'), quest, drop(4494)] });

		it('shows the first source by kind and a count of the rest, named from the locale', () => {
			mount(threeSources(), db);
			const more = screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' });

			expect(more.textContent).toBe('+2');
			expect(screen.getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(4494));
			expect(screen.queryByRole('dialog')).toBeNull();
		});

		it('shows no count for one source', () => {
			mount(Item.create({ id: 901, sources: [vendor(5, 'Quartermaster')] }));

			expect(screen.queryByRole('button')).toBeNull();
		});

		it('counts only the sources it can name', () => {
			const unknownZone = { source: { oneofKind: 'drop', drop: { zoneId: 1, npcId: 1, difficulty: 0, otherName: '', category: '' } } } as const;
			const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
			mount(Item.create({ id: 902, sources: [unknownZone, vendor(5, 'Quartermaster')] }));

			expect(screen.getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(5));
			expect(screen.queryByRole('button')).toBeNull();
			expect(errorSpy).toHaveBeenCalled();
			errorSpy.mockRestore();
		});

		it('lists the others in the same kind order, each with its own link and the text the cell gives it', () => {
			mount(threeSources(), db);
			fireEvent.click(screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' }));
			const links = within(screen.getByRole('dialog')).getAllByRole('link');

			expect(links.map(link => link.getAttribute('href'))).toEqual([ActionId.makeQuestUrl(7), ActionId.makeNpcUrl(5)]);
			expect(links.map(link => link.textContent)).toEqual(['common.sources.questBounty', 'common.sources.sold_byQuartermaster']);
		});

		it('keeps a name that holds markup as plain text', () => {
			mount(Item.create({ id: 903, sources: [quest, vendor(5, '<b>Quartermaster</b>')] }), db);
			fireEvent.click(screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' }));
			const popup = screen.getByRole('dialog');

			expect(popup.querySelector('b')).toBeNull();
			expect(popup.textContent).toContain('<b>Quartermaster</b>');
		});

		it('opens from the keyboard, as a button, and closes on Escape', async () => {
			mount(threeSources(), db);
			const more = screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' });
			expect(more.tagName).toBe('BUTTON');

			more.focus();
			fireEvent.click(more);
			expect(screen.getByRole('dialog')).toBeTruthy();

			fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
			await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
		});

		it('equips nothing on a press of the count or a link in the popup, and opens no row link', () => {
			const onEquip = vi.fn();
			render(
				<div data-testid="row">
					<ItemListRowView
						name="Crusader"
						quality={ItemQuality.ItemQualityUncommon}
						nameDescription=""
						iconUrl="icon.jpg"
						href="https://example.com/item"
						active={false}
						onEquip={onEquip}
						source={<ItemSource item={threeSources()} sim={sim(db)} />}
						favourited={false}
						favouriteTooltipId="favourite"
						onToggleFavourite={vi.fn()}
					/>
				</div>,
			);

			const more = screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' });
			fireEvent.click(more);
			const popup = screen.getByRole('dialog');
			expect(screen.getByTestId('row').contains(popup)).toBe(false);
			fireEvent.click(within(popup).getAllByRole('link')[0]);

			expect(onEquip).not.toHaveBeenCalled();

			fireEvent.click(screen.getByTestId('selector-modal-list-item-row-link'));
			expect(onEquip).toHaveBeenCalledOnce();
		});
	});

	it('links to the vendor for a soldBy source', () => {
		const item = Item.create({
			id: 900,
			sources: [{ source: { oneofKind: 'soldBy', soldBy: { npcId: 5, npcName: "Sha'tari Quartermaster", zoneId: 0 } } }],
		});
		const { container, getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(5));
		expect(container.textContent).toBe(`common.sources.sold_by${"Sha'tari Quartermaster"}`);
	});

	it('opens external links in a new tab, with noopener and noreferrer', () => {
		const item = Item.create({ id: 1000, sources: [{ source: { oneofKind: 'pvp', pvp: {} } }] });
		const { getByRole } = mount(item);

		expect(getByRole('link').getAttribute('target')).toBe('_blank');
		expect(getByRole('link').getAttribute('rel')!.split(/\s+/)).toEqual(expect.arrayContaining(['noopener', 'noreferrer']));
	});

	it('opts every source anchor out of the wowhead link icon', () => {
		const item = Item.create({ id: 100, sources: [{ source: { oneofKind: 'pvp', pvp: {} } }] });
		const { getByRole } = mount(item);

		expect(getByRole('link').getAttribute('data-whtticon')).toBe('false');
	});
});
