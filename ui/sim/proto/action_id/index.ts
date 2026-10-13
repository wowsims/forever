import { APLActionItemSwap_SwapSet } from '@generated/proto/apl';
import { ActionID as ActionIdProto, ItemRandomSuffix, OtherAction } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { IconData, UIItem as Item } from '@generated/proto/ui';

import { CHARACTER_LEVEL } from '../../constants/mechanics';
import { Database } from '../database';
import { buildWowheadTooltipDataset, wowheadEntityUrl, wowheadIconUrl, WowheadTooltipItemParams, WowheadTooltipSpellParams } from '../wowhead';

type ActionIdOptions = {
	itemId?: number;
	spellId?: number;
	otherId?: OtherAction;
	tag?: number;
	baseName?: string;
	name?: string;
	iconUrl?: string;
	randomSuffixId?: number;
	rank?: number;
	definitionId?: number;
	isEmptyPlaceholder?: boolean;
};

// Uniquely identifies a specific item / spell / thing in WoW. This object is immutable.
export class ActionId {
	readonly itemId: number;
	readonly randomSuffixId: number;
	readonly spellId: number;
	readonly otherId: OtherAction;
	readonly tag: number;
	readonly rank: number;
	readonly definitionId: number;

	readonly baseName: string; // The name without any tag additions.
	readonly name: string;
	readonly iconUrl: string;
	readonly spellIdTooltipOverride: number | null;
	// True when this ActionId was created via ActionId.empty() to represent an intentional
	// placeholder (e.g. "No Pet") rather than a missing/unresolved id. Suppresses the
	// "Empty action id!" error logged by toStringIgnoringTag() for accidental empties.
	private readonly isEmptyPlaceholder: boolean;

	private constructor({
		itemId,
		spellId,
		otherId,
		tag,
		baseName,
		name,
		iconUrl,
		randomSuffixId,
		rank,
		definitionId,
		isEmptyPlaceholder,
	}: ActionIdOptions = {}) {
		this.itemId = itemId ?? 0;
		this.randomSuffixId = randomSuffixId ?? 0;
		this.spellId = spellId ?? 0;
		this.otherId = otherId ?? OtherAction.OtherActionNone;
		this.tag = tag ?? 0;
		this.rank = rank ?? 0;
		this.definitionId = definitionId ?? 0;
		this.isEmptyPlaceholder = isEmptyPlaceholder ?? false;

		switch (otherId) {
			case OtherAction.OtherActionNone:
				break;
			case OtherAction.OtherActionWait:
				baseName = 'Wait';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_pocketwatch_01.jpg';
				break;
			case OtherAction.OtherActionManaRegen:
				name = 'Mana Tick';
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeMana];
				if (tag == 1) {
					name += ' (Casting)';
				} else if (tag == 2) {
					name += ' (Not Casting)';
				}
				break;
			case OtherAction.OtherActionEnergyRegen:
				baseName = 'Energy Tick';
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeEnergy];
				break;
			case OtherAction.OtherActionFocusRegen:
				baseName = 'Focus Tick';
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeFocus];
				break;
			case OtherAction.OtherActionManaGain:
				baseName = 'Mana Gain';
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeMana];
				break;
			case OtherAction.OtherActionRageGain:
				baseName = 'Rage Gain';
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeRage];
				break;
			case OtherAction.OtherActionAttack:
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_sword_04.jpg';
				// A name handed in is a filled one: an extra attack's, from the spell that tagged it.
				if (name) break;
				name = 'Attack';
				if (this.tag == 1) {
					name += ' (Main Hand)';
				} else if (this.tag == 2) {
					name += ' (Off Hand)';
				} else if (this.tag == 10610) {
					// Windfury Totem's swings, tagged with 10610, the spell that grants them.
					name += ' (Windfury)';
				} else if (this.tag == 99999) {
					name += ' (Boss)';
				} else if (this.tag == 99998) {
					name += ' (Add)';
				} else if (this.tag == 100000) {
					name += ' (Custom Boss)';
				} else if (this.tag == 100000 + 1) {
					name += ' (Custom Boss) - Thrash';
				}
				break;
			case OtherAction.OtherActionShoot:
				name = 'Shoot';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/ability_marksmanship.jpg';
				break;
			case OtherAction.OtherActionPet:
				break;
			case OtherAction.OtherActionRefund:
				baseName = 'Refund';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_coin_01.jpg';
				break;
			case OtherAction.OtherActionDamageTaken:
				baseName = 'Damage Taken';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_sword_04.jpg';
				break;
			case OtherAction.OtherActionHealingModel:
				baseName = 'Incoming HPS';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/spell_holy_renew.jpg';
				break;
			case OtherAction.OtherActionResourcePotion:
				baseName = 'Resource Potion';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_potion_76.jpg';
				break;
			case OtherAction.OtherActionDefensivePotion:
				baseName = 'Defensive Potion';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_potion_69.jpg';
				break;
			case OtherAction.OtherActionOffensivePotion:
				baseName = 'Offensive Potion';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_alchemy_elixir_04.jpg';
				break;
			case OtherAction.OtherActionMove:
				baseName = 'Moving';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/inv_boots_cloth_03.jpg';
				break;
			case OtherAction.OtherActionPrepull:
				baseName = 'Prepull';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/inv_misc_pocketwatch_02.jpg';
				break;
			case OtherAction.OtherActionEncounterStart:
				baseName = 'Encounter Start';
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/achievement_faction_elders.jpg';
				break;
			case OtherAction.OtherActionItemSwap:
				baseName = 'Item Swap: ';
				if (this.tag == APLActionItemSwap_SwapSet.Main) {
					baseName += 'Main';
				} else {
					baseName += 'Swapped';
				}
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/ability_dualwield.jpg';
				break;
		}
		this.baseName = baseName ?? '';
		this.name = (name || baseName) ?? '';
		this.iconUrl = iconUrl ?? '';
		if (this.name) this.name += this.rank ? ` (Rank ${this.rank})` : '';
		this.spellIdTooltipOverride = this.spellTooltipOverride?.spellId || null;
	}

	anyId(): number {
		return this.itemId || this.spellId || this.otherId;
	}

	equals(other: ActionId): boolean {
		return this.equalsIgnoringTag(other) && this.tag == other.tag;
	}

	// Every field these read is readonly and set in the constructor, so the string cannot go stale.
	private cachedEqualityKey?: string;
	private cachedEqualityKeyIgnoringTag?: string;

	// A string form of exactly what equals() compares, for use as a Map/Set key. Keep in step
	// with equalsIgnoringTag below; toString() is not a substitute, as it reports only the
	// first non-zero of itemId/spellId/otherId and drops randomSuffixId.
	equalityKey(): string {
		return (this.cachedEqualityKey ??= `${this.equalityKeyIgnoringTag()}|${this.tag}`);
	}

	// equalityKey() without the tag, matching equalsIgnoringTag below.
	equalityKeyIgnoringTag(): string {
		return (this.cachedEqualityKeyIgnoringTag ??= `${this.itemId}|${this.randomSuffixId}|${this.spellId}|${this.otherId}`);
	}

	equalsIgnoringTag(other: ActionId): boolean {
		return this.itemId == other.itemId && this.randomSuffixId == other.randomSuffixId && this.spellId == other.spellId && this.otherId == other.otherId;
	}

	static makeItemUrl(id: number, randomSuffixId?: number): string {
		const url = new URL(wowheadEntityUrl('item', id));
		url.searchParams.set('level', String(CHARACTER_LEVEL));
		if (randomSuffixId) url.searchParams.set('bonus', String(randomSuffixId));
		return url.toString();
	}
	static makeSpellUrl(id: number, rank = 0, definitionId = 0): string {
		return wowheadEntityUrl('spell', id, rank, definitionId);
	}
	static async makeItemTooltipData(id: number, params?: Omit<WowheadTooltipItemParams, 'itemId'>) {
		return buildWowheadTooltipDataset({ itemId: id, ...params });
	}
	static async makeSpellTooltipData(id: number, params?: Omit<WowheadTooltipSpellParams, 'spellId'>) {
		return buildWowheadTooltipDataset({ spellId: id, ...params });
	}
	static makeQuestUrl(id: number): string {
		return wowheadEntityUrl('quest', id);
	}
	static makeNpcUrl(id: number): string {
		return wowheadEntityUrl('npc', id);
	}
	static makeFactionUrl(id: number): string {
		return wowheadEntityUrl('faction', id);
	}
	static makeZoneUrl(id: number): string {
		return wowheadEntityUrl('zone', id);
	}

	// An extra attack is a main-hand swing tagged with the spell that granted it, so its name comes
	// from that spell rather than from a list kept by hand. A tag no spell answers keeps the bare name,
	// and is remembered so it is not asked for again on every fill.
	private static readonly extraAttackNames = new Map<number, Promise<string>>();
	private async fillExtraAttack(): Promise<ActionId> {
		let spellName = ActionId.extraAttackNames.get(this.tag);
		if (!spellName) {
			spellName = Database.getSpellIconData(this.tag).then(
				data => data.name,
				() => '',
			);
			ActionId.extraAttackNames.set(this.tag, spellName);
		}
		const name = await spellName;
		if (!name) return this;
		return new ActionId({ otherId: this.otherId, tag: this.tag, name: `Attack (${name})`, iconUrl: this.iconUrl });
	}

	// Returns an ActionId with the name and iconUrl fields filled.
	// playerIndex is the optional index of the player to whom this ID corresponds.
	async fill(playerIndex?: number): Promise<ActionId> {
		if (this.otherId == OtherAction.OtherActionAttack && this.name == 'Attack' && this.tag > 2) {
			return this.fillExtraAttack();
		}
		if (this.name || this.iconUrl) {
			return this;
		}

		if (this.otherId) {
			return this;
		}
		const tooltipData = await ActionId.getTooltipData(this);

		const baseName = tooltipData['name'];
		let name = baseName;

		let tag = this.tag;

		switch (baseName) {
			case 'Minor Speed':
				name = 'Minor Run Speed (8%)';
				break;
			case 'Arcane Missiles':
				break;
			case 'Arcane Blast':
				if (tag == 1) {
					name += ' (No Stacks)';
				} else if (tag == 2) {
					name += ` (1 Stack)`;
				} else if (tag > 2) {
					name += ` (${tag - 1} Stacks)`;
				}
				break;
			case 'Hot Streak':
				if (tag) name += ' (Crits)';
				break;
			case 'Fireball':
			case 'Flamestrike':
				if (tag == 1) name += ' (Blast Wave)';
				break;
			case 'Pyroblast':
			case 'Combustion':
				if (tag) name += ' (DoT)';
				break;
			case 'Evocation':
				if (tag == 1) {
					name += ' (1 Tick)';
				} else if (tag == 2) {
					name += ' (2 Tick)';
				} else if (tag == 3) {
					name += ' (3 Tick)';
				} else if (tag == 4) {
					name += ' (4 Tick)';
				} else if (tag == 5) {
					name += ' (5 Tick)';
				}
				break;
			case 'Mind Flay':
				if (tag == 1) {
					name += ' (1 Tick)';
				} else if (tag == 2) {
					name += ' (2 Tick)';
				} else if (tag == 3) {
					name += ' (3 Tick)';
				}
				break;
			case 'Mind Sear':
				if (tag == 1) {
					name += ' (1 Tick)';
				} else if (tag == 2) {
					name += ' (2 Tick)';
				} else if (tag == 3) {
					name += ' (3 Tick)';
				} else if (tag == 77486) {
					name += ' (Mastery)';
				}

				break;
			case 'Shoot':
				if (tag == 1) name += ' (Shot)';
				break;
			case 'Envenom':
			case 'Eviscerate':
			case 'Rupture':
			case 'Slice and Dice':
			case 'Expose Armor':
				if (tag) name += ` (${tag} CP)`;
				break;
			case 'Crimson Tempest':
				if (tag == 7) {
					name += ' (DoT)';
				} else if (tag) {
					name += ` (${tag} CP)`;
				}
				break;
			case 'Wound Poison':
				if (tag == 1) {
					name += ' (Shiv)';
				}
				break;
			case 'Bladestorm':
				if (tag == 1) {
					name += ' (Main Hand)';
				} else if (tag == 2) {
					name += ' (Off Hand)';
				}
				break;
			case 'Mutilate':
				if (tag == 0) {
					name += ' (Cast)';
				} else if (tag == 1) {
					name += ' (Main Hand)';
				} else if (tag == 2) {
					name += ' (Off Hand)';
				}
				break;
			case 'Stormstrike':
				if (tag == 0) {
					name += ' (Cast)';
				} else if (tag == 1) {
					name += ' (Main Hand)';
				} else if (tag == 2) {
					name += ' (Off Hand)';
				}
				break;
			case 'Chain Lightning':
			case 'Lightning Bolt':
			case 'Lava Burst':
				if (tag == 6) {
					name += ' (Overload)';
				} else if (tag == 7) {
					name += ' (Echo)';
				} else if (tag == 8) {
					name += ' (Echo Overload)';
				} else if (tag == 9) {
					name += ' (Mastery)';
				} else if (tag == 10) {
					name += ' (Haste)';
				} else if (tag == 11) {
					name += ' (Crit)';
				} else if (tag == 12) {
					name += ' (Agi)';
				} else if (tag) {
					name += ` (${tag} MW)`;
				}
				break;
			case 'Flame Shock':
			case 'Earth Shock':
			case 'Frost Shock':
				if (tag == 1) {
					name += ' (DoT)';
				} else if (tag == 7) {
					name += ' (Echo)';
				}
				break;
			case 'Fire Nova':
				// The cast's hits report under its spell, tagged 1, so they share its row.
				if (tag === 1) {
					name += ' (Damage)';
				}
				break;
			case 'Fulmination':
				name += ` (${tag + 2})`;
				break;
			case 'Lightning Shield':
				if (tag == 1) {
					name += ' (Wasted)';
				}
				break;
			case 'Moonfire':
			case 'Sunfire':
				if (tag == 1) {
					name += ' (DoT)';
				}
				break;
			// For targetted buffs, tag is the source player's raid index or -1 if none.
			case 'Innervate':
			case 'Mana Tide Totem':
			case 'Unholy Frenzy':
			case 'Power Infusion':
				if (tag != -1) {
					if (tag === playerIndex || playerIndex == undefined) {
						name += ` (self)`;
					} else {
						name += ` (from #${tag + 1})`;
					}
				} else {
					name += ' (raid)';
				}
				break;
			case 'Berserking':
				if (tag > 0) {
					name += ` (${10 + (tag - 1) * 5}%)`;
				}
				break;
			case 'Whirlwind':
			case 'Windfury Weapon':
			case 'Enchant Weapon - Crusader':
				if (tag == 1) {
					name += ' (Main Hand)';
				} else if (tag == 2) {
					name += ' (Off Hand)';
				}
				break;
			case 'Heroic Strike':
			case 'Cleave':
			case 'Maul':
				if (tag == 1) {
					name += ' (Queue)';
				}
				break;
			case 'Sweeping Strikes':
				if (tag == 1) {
					name += ' (Normalized)';
				}
				break;
			case 'Sunder Armor':
				if (tag == 1) {
					name += ' (Devastate)';
				}
				break;
			case 'Improved Berserker Rage':
			case 'Improved Overpower':
				name += ` (${tag}/2)`;
				break;
			case 'Thunderfury':
				if (tag == 1) {
					name += ' (ST)';
				} else if (tag == 2) {
					name += ' (MT)';
				}
				break;
			case 'Devouring Plague':
				if (tag == 1) {
					name += ' (DoT)';
					break;
				} else if (tag == 2) {
					name += ' (Heal)';
					break;
				}
				if (tag == 77486) {
					name += ' (Mastery)';
					break;
				}
			case 'Shadow Word: Death':
				if (tag == 1) {
					name += ' (No Orb)';
				}
			case 'Shadowfiend':
				if (tag == 1) {
					name += ' (Mana Restore)';
				}
				break;
			case 'Immolate':
				if (tag == 1) {
					name += ' (DoT)';
				}
				break;
			case 'Vampiric Touch':
			case 'Shadow Word: Pain':
				if (tag == 77486) {
					name += ' (Mastery)';
				}

				break;
			case 'Unstable Affliction':
			case 'Corruption':
				if (tag == 1) {
					name += ' (Malefic)';
				}
				break;
			case 'Raptor Strike':
				if (tag == 2) {
					name += ' (Cooldown)';
				}
				break;
			case 'Retribution Aura':
			case 'Holy Shield':
			case 'Thorns':
				// A damage shield's proc is tagged two past the aura that carries it: tag 2 is the
				// player's own copy, tag 1 the copy an outside caster provides, and that copy's
				// aura is the tag -1 this case names below.
				if (tag === 1 || tag === 2) {
					name += ' (Hit)';
				} else if (tag === -1) {
					name += ' (External)';
				}
				break;
			case 'Searing Totem':
			case 'Magma Totem':
			case 'Flametongue Totem':
				// A totem's hits report under the totem's spell, tagged 1, so they share its row.
				if (tag === 1) {
					name += ' (Attack)';
				} else if (tag === -1) {
					name += ' (External)';
				}
				break;
			case 'Dummy Spell':
				if (tag === 100000) {
					['Arcane', 'Fire', 'Frost', 'Holy', 'Shadow', 'Nature'].forEach((school, index) => {
						if (tag === 100000 + index) {
							name += ` (${school})`;
						}
					});
				}
				break;
			default:
				if (tag === -1) {
					name += ' (External)';
				} else if (tag) {
					name += ' (??)';
				}
				break;
		}

		let iconUrl = ActionId.makeIconUrl(tooltipData['icon']);

		const iconOverrideId = this.spellTooltipOverride || this.spellIconOverride;
		if (iconOverrideId) {
			const overrideTooltipData = await ActionId.getTooltipData(iconOverrideId);
			iconUrl = ActionId.makeIconUrl(overrideTooltipData['icon']);
		}

		return new ActionId({
			itemId: this.itemId,
			spellId: this.spellId,
			otherId: this.otherId,
			tag: this.tag,
			baseName,
			name,
			iconUrl,
			randomSuffixId: this.randomSuffixId,
			rank: this.rank || tooltipData['rank'],
			definitionId: this.definitionId,
		});
	}

	toString(): string {
		return this.toStringIgnoringTag() + (this.tag ? '-' + this.tag : '');
	}

	toStringIgnoringTag(): string {
		if (this.itemId) {
			return 'item-' + this.itemId;
		} else if (this.spellId) {
			return 'spell-' + this.spellId;
		} else if (this.otherId) {
			return 'other-' + this.otherId;
		} else {
			if (!this.isEmptyPlaceholder) {
				console.error('Empty action id!');
			}
			return this.name;
		}
	}

	toProto(): ActionIdProto {
		const protoId = ActionIdProto.create({
			tag: this.tag,
		});

		if (this.itemId) {
			protoId.rawId = {
				oneofKind: 'itemId',
				itemId: this.itemId,
			};
		} else if (this.spellId) {
			protoId.rawId = {
				oneofKind: 'spellId',
				spellId: this.spellId,
			};
			protoId.rank = this.rank;
		} else if (this.otherId) {
			protoId.rawId = {
				oneofKind: 'otherId',
				otherId: this.otherId,
			};
		}

		return protoId;
	}

	toProtoString(): string {
		return ActionIdProto.toJsonString(this.toProto());
	}

	withoutTag(): ActionId {
		return new ActionId({
			itemId: this.itemId,
			spellId: this.spellId,
			otherId: this.otherId,
			baseName: this.baseName,
			iconUrl: this.iconUrl,
			randomSuffixId: this.randomSuffixId,
			rank: this.rank,
		});
	}

	withoutRank(): ActionId {
		return new ActionId({
			itemId: this.itemId,
			spellId: this.spellId,
			otherId: this.otherId,
			tag: this.tag,
			baseName: this.baseName,
			name: this.nameWithoutRank,
			iconUrl: this.iconUrl,
			randomSuffixId: this.randomSuffixId,
			definitionId: this.definitionId,
		});
	}

	static fromEmpty(): ActionId {
		return new ActionId();
	}

	// Use this for an INTENTIONALLY empty action id, e.g. a "None" placeholder entry in a
	// picker. Unlike fromEmpty()/an accidentally-empty ActionId, this does not log
	// 'Empty action id!' when stringified.
	static empty(name: string, iconUrl?: string): ActionId {
		return new ActionId({
			baseName: name,
			iconUrl,
			isEmptyPlaceholder: true,
		});
	}

	static fromItemId(itemId: number, tag?: number, randomSuffixId?: number): ActionId {
		return new ActionId({
			itemId,
			tag,
			randomSuffixId,
		});
	}

	static fromSpellId(spellId: number, rank = 0, tag?: number): ActionId {
		return new ActionId({ spellId, rank, tag });
	}

	static fromTalent(spellId: number, rank: number, definitionId: number): ActionId {
		return new ActionId({ spellId, rank, definitionId });
	}

	static fromOtherId(otherId: OtherAction, tag?: number): ActionId {
		return new ActionId({ otherId, tag });
	}

	static fromPetName(petName: string): ActionId {
		return (
			petNameToActionId[petName] ||
			new ActionId({
				baseName: petName,
				iconUrl: petNameToIcon[petName],
			})
		);
	}

	static fromItem(item: Item): ActionId {
		return ActionId.fromItemId(item.id);
	}

	static fromRandomSuffix(item: Item, randomSuffix: ItemRandomSuffix): ActionId {
		return ActionId.fromItemId(item.id, 0, randomSuffix.id);
	}

	static fromProto(protoId: ActionIdProto): ActionId {
		if (protoId.rawId.oneofKind == 'spellId') {
			return ActionId.fromSpellId(protoId.rawId.spellId, protoId.rank, protoId.tag);
		} else if (protoId.rawId.oneofKind == 'itemId') {
			return ActionId.fromItemId(protoId.rawId.itemId, protoId.tag);
		} else if (protoId.rawId.oneofKind == 'otherId') {
			return ActionId.fromOtherId(protoId.rawId.otherId, protoId.tag);
		} else {
			return ActionId.fromEmpty();
		}
	}

	private static readonly logRegex = /{((SpellID)|(ItemID)|(OtherID)): (\d+)(, Tag: (-?\d+))?}/;
	private static readonly logRegexGlobal = new RegExp(ActionId.logRegex, 'g');
	private static fromMatch(match: RegExpMatchArray): ActionId {
		const idType = match[1];
		const id = parseInt(match[5]);
		return new ActionId({
			itemId: idType == 'ItemID' ? id : undefined,
			spellId: idType == 'SpellID' ? id : undefined,
			otherId: idType == 'OtherID' ? id : undefined,
			tag: match[7] ? parseInt(match[7]) : undefined,
		});
	}
	static fromLogString(str: string): ActionId {
		const match = str.match(ActionId.logRegex);
		if (match) {
			return ActionId.fromMatch(match);
		} else {
			console.warn('Failed to parse action id from log: ' + str);
			return ActionId.fromEmpty();
		}
	}

	static async replaceAllInString(str: string): Promise<string> {
		const matches = [...str.matchAll(ActionId.logRegexGlobal)];

		const replaceData = await Promise.all(
			matches.map(async match => {
				const actionId = ActionId.fromMatch(match);
				const filledId = await actionId.fill();
				return {
					firstIndex: match.index || 0,
					len: match[0].length,
					actionId: filledId,
				};
			}),
		);

		// Loop in reverse order so we can greedily apply the string replacements.
		for (let i = replaceData.length - 1; i >= 0; i--) {
			const data = replaceData[i];
			str = str.substring(0, data.firstIndex) + data.actionId.name + str.substring(data.firstIndex + data.len);
		}

		return str;
	}

	private static makeIconUrl(iconLabel: string): string {
		return wowheadIconUrl(iconLabel);
	}

	static async getTooltipData(actionId: ActionId): Promise<IconData> {
		if (actionId.itemId) {
			return Database.getItemIconData(actionId.itemId);
		}
		return Database.getSpellIconData(actionId.spellId);
	}

	get spellIconOverride(): ActionId | null {
		const override = spellIdIconOverrides.get(JSON.stringify({ spellId: this.spellId }));
		if (!override) return null;
		return override.itemId ? ActionId.fromItemId(override.itemId) : ActionId.fromSpellId(override.spellId!);
	}

	get spellTooltipOverride(): ActionId | null {
		const override = spellIdTooltipOverrides.get(JSON.stringify({ spellId: this.spellId, tag: this.tag }));
		if (!override) return null;
		return override.itemId ? ActionId.fromItemId(override.itemId) : ActionId.fromSpellId(override.spellId!);
	}

	get nameWithoutRank(): string {
		return this.name.replace(/ \(Rank \d+\)/g, '');
	}

	get hasRank(): boolean {
		return this.rank > 0;
	}
}

type ActionIdOverride = { itemId?: number; spellId?: number };

// Some items/spells have weird icons, so use this to show a different icon instead.
const spellIdIconOverrides: Map<string, ActionIdOverride> = new Map();

const spellIdTooltipOverrides: Map<string, ActionIdOverride> = new Map([
	// Warrior
	[JSON.stringify({ spellId: 12723, tag: 1 }), { spellId: 26654 }], // Sweeping Strikes (Normalized)
]);

export const defaultTargetIcon = 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_metamorphosis.jpg';

const petNameToActionId: Record<string, ActionId> = {
	Shadowfiend: ActionId.fromSpellId(34433),
	Felhunter: ActionId.fromSpellId(691),
	Imp: ActionId.fromSpellId(688),
	Succubus: ActionId.fromSpellId(712),
	Voidwalker: ActionId.fromSpellId(697),
	Doomguard: ActionId.fromSpellId(18540),
	Infernal: ActionId.fromSpellId(1122),
};

// The owner's spell a pet is summoned by, for a pet the results list under that spell.
export const petNameToSummonSpell: Record<string, string> = {
	'Hawk 1': 'Summon Hawk',
	'Hawk 2': 'Summon Hawk',
};

// https://wowhead.com/forever/hunter-pets
const petNameToIcon: Record<string, string> = {
	'Hawk 1': 'https://wow.zamimg.com/images/wow/icons/medium/spell_nature_ravenform.jpg',
	'Hawk 2': 'https://wow.zamimg.com/images/wow/icons/medium/spell_nature_ravenform.jpg',
	Bat: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_bat.jpg',
	Bear: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_bear.jpg',
	'Bird of Prey': 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_owl.jpg',
	Boar: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_boar.jpg',
	'Carrion Bird': 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_vulture.jpg',
	Cat: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_cat.jpg',
	'Core Hound': 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_corehound.jpg',
	Crab: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_crab.jpg',
	Crocilisk: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_crocolisk.jpg',
	Felguard: 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_summonfelguard.jpg',
	Felhunter: 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_summonfelhunter.jpg',
	Infernal: 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_summoninfernal.jpg',
	Doomguard: 'https://wow.zamimg.com/images/wow/icons/large/warlock_summon_doomguard.jpg',
	Gorilla: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_gorilla.jpg',
	Hyena: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_hyena.jpg',
	Imp: 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_summonimp.jpg',
	Raptor: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_raptor.jpg',
	Scorpid: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_scorpid.jpg',
	Spider: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_spider.jpg',
	Succubus: 'https://wow.zamimg.com/images/wow/icons/large/spell_shadow_summonsuccubus.jpg',
	Tallstrider: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_tallstrider.jpg',
	Turtle: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_turtle.jpg',
	'Wind Serpent': 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_windserpent.jpg',
	Wolf: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_wolf.jpg',
	Fox: 'https://wow.zamimg.com/images/wow/icons/medium/inv_misc_monstertail_07.jpg',
};

export function getPetIconFromName(name: string): string | ActionId | undefined {
	return petNameToActionId[name] || petNameToIcon[name];
}

export const resourceTypeToIcon: Record<ResourceType, string> = {
	[ResourceType.ResourceTypeNone]: '',
	[ResourceType.ResourceTypeHealth]: 'https://wow.zamimg.com/images/wow/icons/medium/inv_elemental_mote_life01.jpg',
	[ResourceType.ResourceTypeMana]: 'https://wow.zamimg.com/images/wow/icons/medium/inv_elemental_mote_mana.jpg',
	[ResourceType.ResourceTypeEnergy]: 'https://wow.zamimg.com/images/wow/icons/medium/spell_shadow_shadowworddominate.jpg',
	[ResourceType.ResourceTypeRage]: 'https://wow.zamimg.com/images/wow/icons/medium/spell_misc_emotionangry.jpg',
	[ResourceType.ResourceTypeComboPoints]: 'https://wow.zamimg.com/images/wow/icons/medium/inv_mace_2h_pvp410_c_01.jpg',
	[ResourceType.ResourceTypeFocus]: 'https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_focusfire.jpg',
	[ResourceType.ResourceTypeGenericResource]: 'https://wow.zamimg.com/images/wow/icons/medium/spell_holy_holybolt.jpg',
};

// Use this to connect a buff row to a cast row in the timeline view
export const buffAuraToSpellIdMap: Record<number, ActionId> = {
	15036: ActionId.fromSpellId(16387), // The party's Flametongue Totem, on the row of the totem's strikes
};
