import { Faction, ItemQuality, ItemSlot, Profession, Race } from '@generated/proto/common';
import {
	DatabaseFilters,
	DungeonDifficulty,
	RepLevel,
	SourceFilterOption,
	UIEnchant as Enchant,
	UIItem as Item,
	UIItem_FactionRestriction,
} from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { ActionId } from '@sim/proto/action_id';
import { professionNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { Sim } from '@sim/sim';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Dialog } from '@ui-kit/Dialog';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import type { ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';
import { ItemList } from './ItemList';
import { ItemListRowView } from './ItemListRow';
import { ItemSource } from './ItemSource';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());
// Every row, so the test is about the faction the list hands a row rather than what it windowed.
vi.mock('@ui-kit/VirtualList', () => ({
	VirtualList: ({ count, renderRow }: any) => Array.from({ length: count }, (_unused, index) => <div key={index}>{renderRow(index)}</div>),
}));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', name: '', href: '', ready: true }) }));
vi.mock('../GearPicker/ItemNoticeIcon', () => ({ ItemNoticeIcon: () => null }));
vi.mock('@features/bulk/model/gear/items', () => ({ hasBulkItem: () => false, addBulkItem: vi.fn(), removeBulkItem: vi.fn() }));

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

const vendor = (npcId: number, npcName: string) => ({ source: { oneofKind: 'soldBy', soldBy: { npcId, npcName, zoneId: 0, faction: 0 } } }) as const;

describe('ItemSource', () => {
	it('links to the item page for an item with a pvp source', () => {
		const item = Item.create({ id: 200, sources: [{ source: { oneofKind: 'pvp', pvp: {} } }] });
		const { container, getByRole } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeItemUrl(200));
		expect(container.textContent).toBe('common.sources.pvp');
	});

	it('shows no pvp source beside any other, and does not count it', () => {
		const rep = { source: { oneofKind: 'rep', rep: { repFactionId: 730, repLevel: RepLevel.RepLevelRevered, factionId: 0 } } } as const;
		const pvp = { source: { oneofKind: 'pvp', pvp: {} } } as const;

		const withRep = mount(Item.create({ id: 210, sources: [pvp, rep] }), { factions: { 730: { id: 730, name: 'Stormpike Guard' } } });
		expect(withRep.container.textContent).toContain('Stormpike Guard');
		expect(withRep.container.textContent).not.toContain('common.sources.pvp');
		expect(withRep.queryByRole('button')).toBeNull();
		withRep.unmount();

		const frostwolfCloak = mount(Item.create({ id: 19083, sources: [vendor(13218, 'Grunnda Wolfheart'), vendor(13219, 'Jekyll Flandring'), pvp] }));
		expect(frostwolfCloak.getByRole('link').textContent).toBe('common.sources.sold_byGrunnda Wolfheart');
		expect(frostwolfCloak.getByRole('button').textContent).toBe('+1');
	});

	describe('the character faction', () => {
		const questFor = (id: number, faction: Faction) => ({ source: { oneofKind: 'quest', quest: { id, name: `Quest ${id}`, faction } } }) as const;
		const vendorFor = (npcId: number, faction: Faction) =>
			({ source: { oneofKind: 'soldBy', soldBy: { npcId, npcName: `Vendor ${npcId}`, zoneId: 0, faction } } }) as const;
		const craftedSource = { source: { oneofKind: 'crafted', crafted: { profession: Profession.Tailoring, spellId: 1 } } } as const;
		const show = (sources: Item['sources'], faction?: Faction) =>
			render(<ItemSource item={Item.create({ id: 22005, sources })} sim={sim()} faction={faction} />);
		const darkmantleCap = [questFor(9004, Faction.Alliance), questFor(9010, Faction.Horde)];

		it.each([
			[Faction.Alliance, 9004],
			[Faction.Horde, 9010],
		])('shows side %i its own quest of two, uncounted', (faction, quest) => {
			const { getByRole, queryByRole } = show(darkmantleCap, faction);
			expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeQuestUrl(quest));
			expect(queryByRole('button')).toBeNull();
		});

		it("leaves the other side's vendor out of the cell and the count, keeping a vendor for both", () => {
			const { getByRole, queryByRole } = show(
				[vendorFor(1, Faction.Horde), vendorFor(2, Faction.Unknown), vendorFor(3, Faction.Alliance)],
				Faction.Alliance,
			);
			expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(2));
			expect(getByRole('button').textContent).toBe('+1');
			expect(queryByRole('link', { name: /Vendor 1/ })).toBeNull();
		});

		it("shows the other side's sources when they are all the item has", () => {
			const { getByRole } = show([vendorFor(1, Faction.Horde), vendorFor(2, Faction.Horde)], Faction.Alliance);
			expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(1));
			expect(getByRole('button').textContent).toBe('+1');
		});

		it('leaves the other kinds alone', () => {
			const { getByRole } = show([craftedSource, vendorFor(1, Faction.Horde), vendorFor(2, Faction.Alliance)], Faction.Alliance);
			expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeSpellUrl(1));
			expect(getByRole('button').textContent).toBe('+1');
		});

		it('shows both sides with no faction given', () => {
			const { getByRole } = show(darkmantleCap);
			expect(getByRole('button').textContent).toBe('+1');
		});

		it('follows a race change on a real player, in a row of the item list', async () => {
			await serveDatabase();
			const { player, host } = await makeIndividualSim();
			player.setRace(Race.RaceHuman);
			player.sim.setFilters(DatabaseFilters.create({ sources: [SourceFilterOption.SourceQuest] }));
			const itemData = {
				id: 22005,
				item: Item.create({ id: 22005, sources: darkmantleCap }),
				actionId: {},
				name: 'Darkmantle Cap',
				searchText: 'darkmantle cap',
				quality: ItemQuality.ItemQualityRare,
				phase: 1,
				nameDescription: '',
				onEquip: vi.fn(),
			} as unknown as ItemData<ItemListType, ReactNode>;
			const tab = { label: SelectorModalTabs.Items, itemData: [itemData], computeEP: () => 0, equippedToItem: () => null };
			render(
				<SimHostProvider host={host}>
					<ItemList tab={tab} slot={ItemSlot.ItemSlotNeck} equippedItem={null} />
				</SimHostProvider>,
			);
			const questHref = () => within(screen.getByTestId('selector-modal-list-item-source-container')).getByRole('link').getAttribute('href');
			expect(questHref()).toBe(ActionId.makeQuestUrl(9004));

			act(() => player.setRace(Race.RaceOrc));
			expect(questHref()).toBe(ActionId.makeQuestUrl(9010));
		});
	});

	it('shows the quest of an item whose sources list the vendor first', () => {
		const item = Item.create({
			id: 212,
			sources: [
				{ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'Vendor', zoneId: 0, faction: 0 } } },
				{ source: { oneofKind: 'quest', quest: { id: 7, name: 'Bounty', faction: 0 } } },
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

	it('shows a zone-level drop with its zone and difficulty, linked to the zone', () => {
		const item = Item.create({
			id: 10298,
			sources: [
				{
					source: {
						oneofKind: 'drop',
						drop: { zoneId: 721, npcId: 0, difficulty: DungeonDifficulty.DifficultyDungeon, otherName: 'Multiple NPCs', category: '' },
					},
				},
			],
		});
		const { container, getByRole } = mount(item, { zones: { 721: { id: 721, name: 'Gnomeregan' } } });

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeZoneUrl(721));
		expect(container.textContent).toBe('Gnomeregan (common.dungeon_difficulties.dungeon)Multiple NPCs');
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
			sources: [{ source: { oneofKind: 'quest', quest: { id: 1, name: 'A Call to Arms', faction: 0 } } }],
		});
		const { getByRole, getByAltText } = mount(item);

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeQuestUrl(1));
		expect(getByAltText('').getAttribute('src')).toBe(`/forever/assets/img/${icon}.png`);
	});

	it('shows no faction icon for a quest with no faction restriction', () => {
		const item = Item.create({
			id: 702,
			factionRestriction: UIItem_FactionRestriction.UNSPECIFIED,
			sources: [{ source: { oneofKind: 'quest', quest: { id: 3, name: 'Neutral Ground', faction: 0 } } }],
		});
		const { container, queryByRole } = mount(item);

		expect(queryByRole('img')).toBeNull();
		expect(container.textContent).toBe('common.sources.questNeutral Ground');
	});

	it('finds and renders a rep source even when it is not the first entry', () => {
		const item = Item.create({
			id: 800,
			sources: [
				{ source: { oneofKind: 'soldBy', soldBy: { npcId: 1, npcName: 'Ignored Vendor', zoneId: 0, faction: 0 } } },
				{ source: { oneofKind: 'rep', rep: { repFactionId: 270, repLevel: RepLevel.RepLevelRevered, factionId: 0 } } },
			],
		});
		const { getByRole } = mount(item, { factions: { 270: { id: 270, name: 'Zandalar Tribe' } } });

		expect(getByRole('link').getAttribute('href')).toBe(ActionId.makeFactionUrl(270));
		expect(getByRole('link').textContent).toBe(`Zandalar Tribe - ${REP_LEVEL_NAMES[RepLevel.RepLevelRevered]}`);
	});

	describe('the other sources', () => {
		const drop = (npcId: number) => ({ source: { oneofKind: 'drop', drop: { zoneId: 28, npcId, difficulty: 0, otherName: '', category: '' } } }) as const;
		const quest = { source: { oneofKind: 'quest', quest: { id: 7, name: 'Bounty', faction: 0 } } } as const;
		const db: DbStub = { zones: { 28: { id: 28, name: 'Western Plaguelands' } }, npcs: { 4494: { id: 4494, name: 'Scarlet Spellbinder' } } };
		const threeSources = () => Item.create({ id: 900, sources: [vendor(5, 'Quartermaster'), quest, drop(4494)] });
		const moreButton = () => screen.getByRole('button', { name: 'gear_tab.gear_picker.more_sources' });
		const openList = () => {
			fireEvent.click(moreButton());
			return screen.findByTestId('sim-popover-popup');
		};
		const listGone = () => waitFor(() => expect(screen.queryByTestId('sim-popover-popup')).toBeNull());

		it('shows the first source by kind and a count of the rest, named from the locale', () => {
			mount(threeSources(), db);

			expect(moreButton().textContent).toBe('+2');
			expect(screen.getByRole('link').getAttribute('href')).toBe(ActionId.makeNpcUrl(4494));
			expect(screen.queryByTestId('sim-popover-popup')).toBeNull();
		});

		it('builds the list of others only when it is shown', async () => {
			const getNpc = vi.fn((id: number) => ({ id, name: `Npc ${id}` }));
			const stub = { db: { getZone: () => ({ id: 28, name: 'Western Plaguelands' }), getNpc } } as unknown as Sim;
			render(<ItemSource item={Item.create({ id: 904, sources: [drop(1), drop(2)] })} sim={stub} />);
			expect(getNpc.mock.calls.map(([id]) => id)).not.toContain(2);

			await openList();
			expect(getNpc.mock.calls.map(([id]) => id)).toContain(2);
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

		it('lists the others in the same kind order, each with its own link and the text the cell gives it', async () => {
			mount(threeSources(), db);
			const links = within(await openList()).getAllByRole('link');

			expect(links.map(link => link.getAttribute('href'))).toEqual([ActionId.makeQuestUrl(7), ActionId.makeNpcUrl(5)]);
			expect(links.map(link => link.textContent)).toEqual(['common.sources.questBounty', 'common.sources.sold_byQuartermaster']);
		});

		it('keeps a name that holds markup as plain text', async () => {
			mount(Item.create({ id: 903, sources: [quest, vendor(5, '<b>Quartermaster</b>')] }), db);
			const list = await openList();

			expect(list.querySelector('b')).toBeNull();
			expect(list.textContent).toContain('<b>Quartermaster</b>');
		});

		it('shows the list when Tab reaches the count, on a real button', async () => {
			mount(threeSources(), db);
			expect(moreButton().tagName).toBe('BUTTON');

			await userEvent.tab();
			expect(document.activeElement).toBe(screen.getByRole('link'));
			expect(screen.queryByTestId('sim-popover-popup')).toBeNull();

			await userEvent.tab();
			expect(document.activeElement).toBe(moreButton());
			expect(await screen.findByTestId('sim-popover-popup')).toBeTruthy();
		});

		it('opens the list on a press that focuses the count first, as a tap does, and closes it on a second press', async () => {
			mount(threeSources(), db);

			moreButton().focus();
			await openList();

			fireEvent.click(moreButton());
			await listGone();
		});

		it('leaves an open list alone on a mouse click', async () => {
			mount(threeSources(), db);
			await openList();

			fireEvent(moreButton(), new PointerEvent('click', { bubbles: true, cancelable: true, pointerType: 'mouse' }));

			expect(screen.getByTestId('sim-popover-popup').hasAttribute('data-open')).toBe(true);
		});

		it.each([
			['the count', () => moreButton()],
			['elsewhere in the dialog', () => screen.getByRole('textbox')],
		])('closes only the list on Escape, with focus on %s', async (_, focused) => {
			const onOpenChange = vi.fn();
			render(
				<Dialog open onOpenChange={onOpenChange}>
					<input aria-label="Search" />
					<ItemSource item={threeSources()} sim={sim(db)} />
				</Dialog>,
			);
			await openList();

			fireEvent.keyDown(focused(), { key: 'Escape' });
			await listGone();
			expect(onOpenChange).not.toHaveBeenCalled();

			fireEvent.keyDown(focused(), { key: 'Escape' });
			expect(onOpenChange).toHaveBeenCalledOnce();
		});

		it('equips nothing on a press of the count or a link in the list, and opens no row link', async () => {
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

			const list = await openList();
			fireEvent.click(within(list).getAllByRole('link')[0]);
			fireEvent.click(moreButton());

			expect(onEquip).not.toHaveBeenCalled();

			fireEvent.click(screen.getByTestId('selector-modal-list-item-row-link'));
			expect(onEquip).toHaveBeenCalledOnce();
		});
	});

	it('links to the vendor for a soldBy source', () => {
		const item = Item.create({
			id: 900,
			sources: [{ source: { oneofKind: 'soldBy', soldBy: { npcId: 5, npcName: "Sha'tari Quartermaster", zoneId: 0, faction: 0 } } }],
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
