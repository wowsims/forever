import type { Profession } from '@generated/proto/common';
import type { UIItem, UIItemSource } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { Database } from '@sim/proto/database';
import { professionNames } from '@sim/proto/names';
import { pushTo } from '@sim/utils/collections';

export type DropsCategory = 'drop' | 'quest' | 'crafted' | 'rep' | 'soldBy';

export interface DropsNames {
	zone: (id: number) => string | null;
	npc: (id: number) => string | null;
	faction: (id: number) => string | null;
}

export const dropsNamesOf = (db: Database): DropsNames => ({
	zone: id => db.getZone(id)?.name ?? null,
	npc: id => db.getNpc(id)?.name ?? null,
	faction: id => db.getRepFaction(id)?.name ?? null,
});

const CATEGORIES: readonly DropsCategory[] = ['drop', 'quest', 'crafted', 'rep', 'soldBy'];
const WORLD_DROP_KEY = 'drop:0:other:World Drop';
const OTHER_QUEST_KEY = 'quest:other';
const OTHER_VENDOR_KEY = 'soldBy:0';

const t = (key: string) => i18n.t(`bulk_tab.drops.sources.${key}`);

/** The leaf key a source counts under; null for PvP and for a source with nothing to key on. */
const dropsSourceKey = (source: UIItemSource, names: DropsNames): string | null => {
	switch (source.source.oneofKind) {
		case 'drop': {
			const { zoneId, npcId, otherName } = source.source.drop;
			if (npcId > 0) return `drop:${zoneId}:${npcId}`;
			return otherName ? `drop:${zoneId}:other:${otherName}` : null;
		}
		case 'quest': {
			const { zoneId } = source.source.quest;
			return zoneId > 0 && names.zone(zoneId) !== null ? `quest:${zoneId}` : OTHER_QUEST_KEY;
		}
		case 'crafted':
			return `crafted:${source.source.crafted.profession}`;
		case 'rep':
			return `rep:${source.source.rep.repFactionId}`;
		case 'soldBy': {
			const { zoneId } = source.source.soldBy;
			return zoneId > 0 && names.zone(zoneId) !== null ? `soldBy:${zoneId}` : OTHER_VENDOR_KEY;
		}
		default:
			return null;
	}
};

/** Each key of the item once, in item order; the tests read the keys here. */
export const dropsItemSourceKeys = (item: UIItem, names: DropsNames): string[] => {
	const keys = new Set<string>();
	for (const source of item.sources) {
		const key = dropsSourceKey(source, names);
		if (key !== null) keys.add(key);
	}
	return [...keys];
};

export interface DropsSourceLabel {
	label: string;
	/** The zone or the category, under the label. */
	detail: string;
}

export const dropsSourceLabel = (key: string, names: DropsNames): DropsSourceLabel => {
	const [category, first, second] = key.split(':');
	const categoryLabel = t(`categories.${category}`);
	const id = Number(first);
	switch (category) {
		case 'drop': {
			if (second === 'other') {
				const name = key.slice(`drop:${first}:other:`.length);
				return { label: key === WORLD_DROP_KEY ? t('world_drop') : name, detail: id > 0 ? (names.zone(id) ?? String(id)) : categoryLabel };
			}
			return { label: names.npc(Number(second)) ?? `#${second}`, detail: names.zone(id) ?? String(id) };
		}
		case 'quest':
			return { label: key === OTHER_QUEST_KEY ? t('other_quests') : (names.zone(id) ?? String(id)), detail: categoryLabel };
		case 'crafted':
			return { label: professionNames.get(id as Profession) ?? String(id), detail: categoryLabel };
		case 'rep':
			return { label: names.faction(id) ?? String(id), detail: categoryLabel };
		default:
			return { label: key === OTHER_VENDOR_KEY ? t('other_vendors') : (names.zone(id) ?? String(id)), detail: categoryLabel };
	}
};

export interface DropsSourceNode {
	/** A category ('drop'), a drop zone ('drop:<zoneId>') or a leaf key. */
	key: string;
	label: string;
	/** Items under the node, each once. */
	itemIds: readonly number[];
	/** Leaf keys under the node; a leaf lists itself. */
	leafKeys: readonly string[];
	children: readonly DropsSourceNode[];
}

// Ties go by key, so siblings that share a label keep one order whatever items the tree is built from.
const byLabel = (a: DropsSourceNode, b: DropsSourceNode) => a.label.localeCompare(b.label) || a.key.localeCompare(b.key);

const parent = (key: string, label: string, children: DropsSourceNode[]): DropsSourceNode => ({
	key,
	label,
	itemIds: [...new Set(children.flatMap(child => child.itemIds))],
	leafKeys: children.flatMap(child => child.leafKeys),
	children,
});

/** Built from the items the phase lets through; a category with no item is left out. */
export const dropsSourceTree = (items: readonly UIItem[], names: DropsNames): DropsSourceNode[] => {
	const itemsByKey = new Map<string, number[]>();
	for (const item of items) {
		for (const key of dropsItemSourceKeys(item, names)) pushTo(itemsByKey, key, item.id);
	}

	const leaf = (key: string): DropsSourceNode => ({
		key,
		label: dropsSourceLabel(key, names).label,
		itemIds: itemsByKey.get(key)!,
		leafKeys: [key],
		children: [],
	});
	const leavesOf = (category: DropsCategory) => [...itemsByKey.keys()].filter(key => key.startsWith(`${category}:`)).map(leaf);
	// A catch-all leaf sorts after the named ones.
	const withLastKey = (leaves: DropsSourceNode[], lastKey: string) => [
		...leaves.filter(l => l.key !== lastKey).sort(byLabel),
		...leaves.filter(l => l.key === lastKey),
	];

	const tree: DropsSourceNode[] = [];
	for (const category of CATEGORIES) {
		const leaves = leavesOf(category);
		if (leaves.length === 0) continue;
		let children: DropsSourceNode[];
		if (category === 'drop') {
			const zoned = new Map<number, DropsSourceNode[]>();
			const unzoned: DropsSourceNode[] = [];
			for (const node of leaves) {
				const zoneId = Number(node.key.split(':')[1]);
				if (zoneId > 0) pushTo(zoned, zoneId, node);
				else unzoned.push(node);
			}
			const zones = [...zoned].map(([zoneId, nodes]) => {
				const bosses = nodes.filter(n => !n.key.includes(':other:')).sort(byLabel);
				const others = nodes.filter(n => n.key.includes(':other:')).sort(byLabel);
				return parent(`drop:${zoneId}`, names.zone(zoneId) ?? String(zoneId), [...bosses, ...others]);
			});
			children = [...zones.sort(byLabel), ...unzoned.sort(byLabel)];
		} else if (category === 'quest') {
			children = withLastKey(leaves, OTHER_QUEST_KEY);
		} else if (category === 'soldBy') {
			children = withLastKey(leaves, OTHER_VENDOR_KEY);
		} else {
			children = leaves.sort(byLabel);
		}
		tree.push(parent(category, t(`categories.${category}`), children));
	}
	return tree;
};

export type DropsNodeState = 'checked' | 'unchecked' | 'mixed';

export const dropsNodeState = (node: DropsSourceNode, ticked: ReadonlySet<string>): DropsNodeState => {
	const count = node.leafKeys.filter(key => ticked.has(key)).length;
	if (count === 0) return 'unchecked';
	return count === node.leafKeys.length ? 'checked' : 'mixed';
};

/** The stored ticks after a node is ticked or unticked: sorted, each once, keys not in the tree kept. */
export const dropsTicksAfter = (ticked: readonly string[], node: DropsSourceNode, checked: boolean): string[] => {
	const next = new Set(ticked);
	for (const key of node.leafKeys) {
		if (checked) next.add(key);
		else next.delete(key);
	}
	return [...next].sort();
};

export const DROPS_L60_DUNGEON_ZONE_IDS: readonly number[] = [1584, 1583, 2557, 2057, 2017];

export const dropsQuickPickTicks = (ticked: readonly string[], tree: readonly DropsSourceNode[], zoneIds: readonly number[]): string[] => {
	const zones = tree.find(node => node.key === 'drop')?.children ?? [];
	const next = new Set(ticked);
	for (const zone of zones) {
		if (zoneIds.some(id => zone.key === `drop:${id}`)) for (const key of zone.leafKeys) next.add(key);
	}
	return [...next].sort();
};
