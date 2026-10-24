import { ActionCategory } from '@generated/proto/api';
import { OtherAction } from '@generated/proto/common';
import i18n from '@i18n/config';
import { AuraStats, SpellStats, UnitMetadata } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { bucket } from '@sim/utils/collections';

import { talentTrees } from './talent_trees';

export type ACTION_ID_SET =
	| 'auras'
	| 'stackable_auras'
	| 'icd_auras'
	| 'exclusive_effect_auras'
	| 'spells'
	| 'castable_spells'
	| 'channel_spells'
	| 'dot_spells'
	| 'castable_dot_spells'
	| 'shield_spells'
	| 'non_instant_spells'
	| 'friendly_spells'
	| 'expected_dot_spells'
	| 'expected_damage_spells'
	| 'spells_with_travelTime'
	| 'potions'
	| 'conjured_items'
	| 'talents'
	| 'talent_ranks';

/** The ranked entries that share `actionId`'s base name and tag: the ranks of its spell. */
const ranksOf = <T extends AuraStats | SpellStats>(actionId: ActionId, entries: Array<T>): Array<T> =>
	entries.filter(entry => actionId.baseName === entry.id.baseName && actionId.tag === entry.id.tag && entry.id.hasRank);

/** True when two or more of `spells` share `actionId`'s base name and tag — i.e. `actionId` has ranks. */
const spellHasRanks = (actionId: ActionId, spells: Array<AuraStats | SpellStats>) => ranksOf(actionId, spells).length > 1;

/** The highest rank `offered` of `actionId`'s spell, when one ranks above `actionId`. */
const higherRank = <T extends AuraStats | SpellStats>(actionId: ActionId, offered: Array<T>): T | undefined =>
	ranksOf(actionId, offered).reduce<T | undefined>((top, spell) => (spell.id.rank > (top?.id.rank ?? actionId.rank) ? spell : top), undefined);

const rankKey = (actionId: ActionId) => `${actionId.baseName}|${actionId.tag}`;

/** Each entry's highest rank among `entries`, or the entry itself when none ranks above it. */
const highestRanks = <T extends AuraStats | SpellStats>(entries: Array<T>): Map<T, T> => {
	const tops = new Map<string, T>();
	for (const entry of entries) {
		const key = rankKey(entry.id);
		if (entry.id.hasRank && entry.id.rank > (tops.get(key)?.id.rank ?? 0)) tops.set(key, entry);
	}
	return new Map(
		entries.map(entry => {
			const top = tops.get(rankKey(entry.id));
			return [entry, top && top.id.rank > entry.id.rank ? top : entry];
		}),
	);
};

/**
 * The menu path `actionId` is filed under. The ranks of a spell are filed under the highest rank
 * `offered`, which makes that rank the submenu's own row: picking the row picks the highest rank,
 * and the lower ranks sit behind it. `spells` decides whether `actionId` has ranks at all.
 */
const createSpellSubmenu = (
	actionId: ActionId,
	spells: Array<AuraStats | SpellStats>,
	baseMenuEntry: Array<string> = [],
	offered: Array<AuraStats | SpellStats> = spells,
): Array<string | ActionId> => {
	if (!spellHasRanks(actionId, spells)) return baseMenuEntry;
	const highest = higherRank(actionId, offered);
	return highest ? [...baseMenuEntry, highest.id] : baseMenuEntry;
};

/** The submenus the sim's categories are filed under, in the order they are listed. */
const CATEGORY_SUBMENUS: Array<[ActionCategory, string]> = [
	[ActionCategory.ActionCategoryCooldown, 'cooldowns'],
	[ActionCategory.ActionCategoryConsumable, 'consumables'],
	[ActionCategory.ActionCategoryItem, 'items'],
	[ActionCategory.ActionCategoryTalent, 'talents'],
	[ActionCategory.ActionCategoryRacial, 'racials'],
	[ActionCategory.ActionCategoryProfession, 'professions'],
	[ActionCategory.ActionCategoryExternalBuff, 'external_buffs'],
	[ActionCategory.ActionCategoryExternalCooldown, 'external_cooldowns'],
	[ActionCategory.ActionCategoryExternalDebuff, 'external_debuffs'],
];

/**
 * Sorts `entries` into the unit's shift spells, their ranks and their auras at the top level, the
 * manual groups the sim names, their category's submenu and their talent tree, and hands back the
 * rest for the caller to place. A category other than Talents wins over the tree; a talent entry
 * with no tree keeps its place.
 *
 * The categories and trees are emitted in a fixed order because the menu places a submenu where its
 * first option appears, and the metadata arrives sorted by name.
 */
const categorise = <T extends AuraStats | SpellStats>(
	metadata: UnitMetadata,
	entries: Array<T>,
	toOption: (entry: T, submenu: Array<string | ActionId>) => ActionIdOption,
	allEntries: Array<T> = entries,
) => {
	const talentTree = talentTrees(metadata.getTalents());
	// Every rank of a shift spell goes with it, so an entry matches by base name and tag.
	const shiftKeys = metadata.getShiftSpells().map(rankKey);
	const shiftPlace = (entry: T) => (entry.id.itemId ? -1 : shiftKeys.indexOf(rankKey(entry.id)));
	const highest = highestRanks(entries);
	const ranked = bucket(
		allEntries.filter(entry => entry.id.hasRank),
		entry => rankKey(entry.id),
	);
	// The lower ranks sit behind the highest, as in createSpellSubmenu.
	const option = (entry: T, path: Array<string>) => {
		const top = highest.get(entry)!;
		return toOption(entry, top !== entry && ranked[rankKey(entry.id)].length > 1 ? [...path, top.id] : path);
	};
	const isShift = (entry: T) => shiftPlace(entry) >= 0;
	const groupOf = (entry: T) => highest.get(entry)!.data.group;
	const isCategorised = (entry: T) => CATEGORY_SUBMENUS.some(([category]) => entry.data.category === category);
	const treeOf = (entry: T) =>
		entry.data.category === ActionCategory.ActionCategoryTalent || !isCategorised(entry) ? talentTree(entry.id.baseName) : undefined;

	const unshifted = entries.filter(entry => !isShift(entry));
	const ungrouped = unshifted.filter(entry => !groupOf(entry));
	const inTrees = ungrouped.flatMap(entry => {
		const tree = treeOf(entry);
		return tree ? [{ entry, tree }] : [];
	});
	inTrees.sort((a, b) => a.tree.index - b.tree.index);
	const rest = ungrouped.filter(entry => !treeOf(entry));

	return {
		// In the sim's form order, which is the order of the stance bar in game.
		shifted: entries
			.filter(isShift)
			.sort((a, b) => shiftPlace(a) - shiftPlace(b))
			.map(entry => option(entry, [])),
		inTrees: inTrees.map(({ entry, tree }) => option(entry, [tree.name])),
		grouped: Object.entries(bucket(unshifted.filter(groupOf), groupOf)).flatMap(([group, members]) => members.map(entry => option(entry, [group]))),
		categorised: CATEGORY_SUBMENUS.flatMap(([category, key]) => rest.filter(entry => entry.data.category === category).map(entry => option(entry, [key]))),
		uncategorised: rest.filter(entry => !isCategorised(entry)),
	};
};

/** The auras `keep` passes; `allAuras`, the unfiltered list, decides whether an aura has ranks. */
const auraOptions = (metadata: UnitMetadata, allAuras: Array<AuraStats>, keep: (aura: AuraStats) => boolean = () => true): Array<ActionIdOption> => {
	const toOption = (aura: AuraStats, submenu: Array<string | ActionId>) => ({ value: aura.id, submenu });
	const kept = allAuras.filter(keep);
	const { shifted, inTrees, grouped, categorised, uncategorised } = categorise(metadata, kept, toOption, allAuras);
	return [...shifted, ...inTrees, ...grouped, ...categorised, ...uncategorised.map(aura => toOption(aura, createSpellSubmenu(aura.id, allAuras, [], kept)))];
};

/** The one list a spell may be cast from, for the spells the sim only accepts in one of them. */
export type LIST_SCOPE = 'prepull' | 'priority';

export interface ActionIdOption {
	value: ActionId;
	/** The talent tree, manual group or category submenu the spell or aura is filed under, none for a shift spell or its aura, then the highest rank its lower ranks sit behind. */
	submenu?: Array<string | ActionId>;
	tooltip?: string;
	/**
	 * Read by the field, which drops the entries the list it is in cannot use.
	 *
	 * A CSS rule scoped to the enclosing list cannot hide them: the menu renders in a portal on
	 * `document.body`, where no such rule can reach it.
	 */
	listScope?: LIST_SCOPE;
}

const listScopeOf = (spell: SpellStats): LIST_SCOPE | undefined => (spell.data.prepullOnly ? 'prepull' : spell.data.encounterOnly ? 'priority' : undefined);

export const actionIdSets: Record<
	ACTION_ID_SET,
	{
		defaultLabel: string;
		getActionIDs: (metadata: UnitMetadata) => Promise<Array<ActionIdOption>>;
	}
> = {
	auras: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.auras'),
		getActionIDs: async metadata => auraOptions(metadata, metadata.getAuras()),
	},
	stackable_auras: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.stackable_auras'),
		getActionIDs: async metadata => auraOptions(metadata, metadata.getAuras(), aura => aura.data.maxStacks > 0),
	},
	icd_auras: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.icd_auras'),
		getActionIDs: async metadata => auraOptions(metadata, metadata.getAuras(), aura => aura.data.hasIcd),
	},
	exclusive_effect_auras: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.exclusive_effect_auras'),
		getActionIDs: async metadata => auraOptions(metadata, metadata.getAuras(), aura => aura.data.hasExclusiveEffect),
	},
	// Used for non categorized lists
	spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isCastable);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells, []),
				};
			});
		},
	},
	castable_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.castable_spells'),
		getActionIDs: async metadata => {
			const castableSpells = metadata.getSpells().filter(spell => spell.data.isCastable);
			const toOption = (spell: SpellStats, submenu: Array<string | ActionId>) => ({ value: spell.id, submenu, listScope: listScopeOf(spell) });
			const { shifted, inTrees, grouped, categorised, uncategorised } = categorise(metadata, castableSpells, toOption);

			const placeholders: Array<ActionId> = [
				OtherAction.OtherActionOffensivePotion,
				OtherAction.OtherActionDefensivePotion,
				OtherAction.OtherActionResourcePotion,
			].map(otherId => ActionId.fromOtherId(otherId));

			return [
				shifted,
				inTrees,
				grouped,
				uncategorised.map(spell => toOption(spell, createSpellSubmenu(spell.id, castableSpells, ['spell']))),
				categorised,
				placeholders.map(actionId => {
					return {
						value: actionId,
						submenu: ['placeholders'],
						tooltip: i18n.t('rotation_tab.apl.helpers.placeholder_tooltip'),
					};
				}),
			].flat();
		},
	},
	non_instant_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.non_instant_spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isCastable && spell.data.hasCastTime);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	friendly_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.friendly_spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isCastable && spell.data.isFriendly);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	channel_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.channel_spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isCastable && spell.data.isChanneled);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	potions: {
		defaultLabel: i18n.t('rotation_tab.apl.submenus.potions'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isPotion);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	conjured_items: {
		defaultLabel: i18n.t('rotation_tab.apl.submenus.conjured_items'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isConjured);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	dot_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.dot_spells'),
		getActionIDs: async metadata => {
			const spells = metadata
				.getSpells()
				.filter(spell => spell.data.hasDot)
				// filter duplicate dot entries from RelatedDotSpell
				.filter((value, index, self) => self.findIndex(v => v.id.anyId() === value.id.anyId()) === index);

			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	castable_dot_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.castable_dot_spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.isCastable && spell.data.hasDot);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	expected_dot_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.expected_dot_spells'),
		getActionIDs: async metadata => {
			const spells = metadata
				.getSpells()
				.filter(spell => spell.data.hasExpectedTick)
				// filter duplicate dot entries from RelatedDotSpell
				.filter((value, index, self) => self.findIndex(v => v.id.anyId() === value.id.anyId()) === index);

			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	expected_damage_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.expected_damage_spells'),
		getActionIDs: async metadata => {
			const spells = metadata
				.getSpells()
				.filter(spell => spell.data.isCastable && (spell.data.hasExpectedTick || spell.data.hasExpectedInitial))
				// filter duplicate dot entries from RelatedDotSpell
				.filter((value, index, self) => self.findIndex(v => v.id.anyId() === value.id.anyId()) === index);

			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	shield_spells: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.shield_spells'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.hasShield);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
	// Each talent at the rank the player holds, with its lower ranks behind it. Every rank of a talent
	// shares one spell id and the metadata carries only the held rank, so the lower ranks are built here.
	talents: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.talents'),
		getActionIDs: async metadata => {
			const perTalent = await Promise.all(
				metadata.getTalents().map(async held => {
					const lowerRanks = await Promise.all(Array.from({ length: held.rank - 1 }, (_, i) => ActionId.fromSpellId(held.spellId, i + 1).fill()));
					return [{ value: held }, ...lowerRanks.map(value => ({ value, submenu: [held] }))];
				}),
			);
			return perTalent.flat();
		},
	},
	// Each talent the player holds, once and without a rank, for the values that read the rank.
	talent_ranks: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.talent_ranks'),
		getActionIDs: async metadata => metadata.getTalents().map(held => ({ value: held.withoutRank() })),
	},
	spells_with_travelTime: {
		defaultLabel: i18n.t('rotation_tab.apl.helpers.action_id_sets.spells_with_travelTime'),
		getActionIDs: async metadata => {
			const spells = metadata.getSpells().filter(spell => spell.data.hasMissileSpeed);
			return spells.map(actionId => {
				return {
					value: actionId.id,
					submenu: createSpellSubmenu(actionId.id, spells),
				};
			});
		},
	},
};
