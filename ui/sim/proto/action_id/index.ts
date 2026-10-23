import { APLActionItemSwap_SwapSet } from '@generated/proto/apl';
import { ActionID as ActionIdProto, ItemRandomSuffix, OtherAction } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { IconData, UIItem as Item } from '@generated/proto/ui';
import i18n from '@i18n/config';

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
				baseName = i18n.t('rotation_tab.apl.actions.wait.label');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_pocketwatch_01.jpg';
				break;
			case OtherAction.OtherActionManaRegen:
				name = i18n.t('common.action_names.mana_tick');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeMana];
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.casting')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.not_casting')})`;
				}
				break;
			case OtherAction.OtherActionEnergyRegen:
				baseName = i18n.t('common.action_names.energy_tick');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeEnergy];
				break;
			case OtherAction.OtherActionFocusRegen:
				baseName = i18n.t('common.action_names.focus_tick');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeFocus];
				break;
			case OtherAction.OtherActionManaGain:
				baseName = i18n.t('common.action_names.mana_gain');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeMana];
				break;
			case OtherAction.OtherActionRageGain:
				baseName = i18n.t('common.action_names.rage_gain');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeRage];
				break;
			case OtherAction.OtherActionAttack:
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_sword_04.jpg';
				// A name handed in is a filled one: an extra attack's, from the spell that tagged it.
				if (name) break;
				name = i18n.t('common.action_names.attack');
				if (this.tag == 1) {
					name += ` (${i18n.t('common.action_tags.main_hand')})`;
				} else if (this.tag == 2) {
					name += ` (${i18n.t('common.action_tags.off_hand')})`;
				} else if (this.tag == 10610) {
					// Windfury Totem's swings, tagged with 10610, the spell that grants them.
					name += ` (${i18n.t('common.action_tags.windfury')})`;
				} else if (this.tag == 99999) {
					name += ` (${i18n.t('common.action_tags.boss')})`;
				} else if (this.tag == 99998) {
					name += ` (${i18n.t('common.action_tags.add')})`;
				} else if (this.tag == 100000) {
					name += ` (${i18n.t('common.action_tags.custom_boss')})`;
				} else if (this.tag == 100000 + 1) {
					name += ` (${i18n.t('common.action_tags.custom_boss')}) - ${i18n.t('common.action_tags.thrash')}`;
				}
				break;
			case OtherAction.OtherActionShoot:
				name = i18n.t('common.action_names.shoot');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/ability_marksmanship.jpg';
				break;
			case OtherAction.OtherActionPet:
				break;
			case OtherAction.OtherActionRefund:
				baseName = i18n.t('common.action_names.refund');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_coin_01.jpg';
				break;
			case OtherAction.OtherActionDamageTaken:
				baseName = i18n.t('results_tab.details.columns.damage_taken');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_sword_04.jpg';
				break;
			case OtherAction.OtherActionHealingModel:
				baseName = i18n.t('settings_tab.other.incoming_hps.label');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/spell_holy_renew.jpg';
				break;
			case OtherAction.OtherActionHealthRegen:
				baseName = i18n.t('common.action_names.health_regen');
				iconUrl = resourceTypeToIcon[ResourceType.ResourceTypeHealth];
				break;
			case OtherAction.OtherActionResourcePotion:
				baseName = i18n.t('settings_tab.consumables.potions.resource');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_potion_76.jpg';
				break;
			case OtherAction.OtherActionDefensivePotion:
				baseName = i18n.t('settings_tab.consumables.potions.defensive');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_potion_69.jpg';
				break;
			case OtherAction.OtherActionOffensivePotion:
				baseName = i18n.t('settings_tab.consumables.potions.offensive');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/large/inv_alchemy_elixir_04.jpg';
				break;
			case OtherAction.OtherActionMove:
				baseName = i18n.t('common.action_names.moving');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/inv_boots_cloth_03.jpg';
				break;
			case OtherAction.OtherActionPrepull:
				baseName = i18n.t('common.action_names.prepull');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/inv_misc_pocketwatch_02.jpg';
				break;
			case OtherAction.OtherActionEncounterStart:
				baseName = i18n.t('common.action_names.encounter_start');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/achievement_faction_elders.jpg';
				break;
			case OtherAction.OtherActionItemSwap:
				baseName =
					this.tag == APLActionItemSwap_SwapSet.Main ? i18n.t('common.action_names.item_swap_main') : i18n.t('common.action_names.item_swap_swapped');
				iconUrl = 'https://wow.zamimg.com/images/wow/icons/medium/ability_dualwield.jpg';
				break;
		}
		this.baseName = baseName ?? '';
		this.name = (name || baseName) ?? '';
		this.iconUrl = iconUrl ?? '';
		if (this.name) this.name += this.rank ? ` (${i18n.t('common.action_tags.rank', { rank: this.rank })})` : '';
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
		return new ActionId({ otherId: this.otherId, tag: this.tag, name: `${i18n.t('common.action_names.attack')} (${name})`, iconUrl: this.iconUrl });
	}

	// Returns an ActionId with the name and iconUrl fields filled.
	// playerIndex is the optional index of the player to whom this ID corresponds.
	async fill(playerIndex?: number): Promise<ActionId> {
		if (this.otherId == OtherAction.OtherActionAttack && this.name == i18n.t('common.action_names.attack') && this.tag > 2) {
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
				name = i18n.t('common.action_names.minor_run_speed');
				break;
			case 'Arcane Missiles':
				break;
			case 'Arcane Blast':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.no_stacks')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.one_stack')})`;
				} else if (tag > 2) {
					name += ` (${i18n.t('common.action_tags.stacks', { stacks: tag - 1 })})`;
				}
				break;
			case 'Hot Streak':
				if (tag) name += ` (${i18n.t('common.action_tags.crits')})`;
				break;
			case 'Fireball':
			case 'Flamestrike':
			case 'Frostfire Bolt':
			case 'Pyroblast':
			case 'Combustion':
				if (tag) name += ` (${i18n.t('common.action_tags.dot')})`;
				break;
			case 'Evocation':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 1 })})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 2 })})`;
				} else if (tag == 3) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 3 })})`;
				} else if (tag == 4) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 4 })})`;
				} else if (tag == 5) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 5 })})`;
				}
				break;
			case 'Mind Flay':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 1 })})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 2 })})`;
				} else if (tag == 3) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 3 })})`;
				}
				break;
			case 'Mind Sear':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 1 })})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 2 })})`;
				} else if (tag == 3) {
					name += ` (${i18n.t('common.action_tags.ticks', { ticks: 3 })})`;
				} else if (tag == 77486) {
					name += ` (${i18n.t('common.action_tags.mastery')})`;
				}

				break;
			case 'Shoot':
				if (tag == 1) name += ` (${i18n.t('common.action_tags.shot')})`;
				break;
			case 'Envenom':
			case 'Eviscerate':
			case 'Rupture':
			case 'Slice and Dice':
			case 'Expose Armor':
				if (tag) name += ` (${i18n.t('common.action_tags.combo_points', { points: tag })})`;
				break;
			case 'Crimson Tempest':
				if (tag == 7) {
					name += ` (${i18n.t('common.action_tags.dot')})`;
				} else if (tag) {
					name += ` (${i18n.t('common.action_tags.combo_points', { points: tag })})`;
				}
				break;
			case 'Wound Poison':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.shiv')})`;
				}
				break;
			case 'Bladestorm':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.main_hand')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.off_hand')})`;
				}
				break;
			case 'Mutilate':
				if (tag == 0) {
					name += ` (${i18n.t('common.action_tags.cast')})`;
				} else if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.main_hand')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.off_hand')})`;
				}
				break;
			case 'Stormstrike':
				if (tag == 0) {
					name += ` (${i18n.t('common.action_tags.cast')})`;
				} else if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.main_hand')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.off_hand')})`;
				}
				break;
			case 'Chain Lightning':
			case 'Lightning Bolt':
			case 'Lava Burst':
				if (tag == 6) {
					name += ` (${i18n.t('common.action_tags.overload')})`;
				} else if (tag == 7) {
					name += ` (${i18n.t('common.action_tags.echo')})`;
				} else if (tag == 8) {
					name += ` (${i18n.t('common.action_tags.echo_overload')})`;
				} else if (tag == 9) {
					name += ` (${i18n.t('common.action_tags.mastery')})`;
				} else if (tag == 10) {
					name += ` (${i18n.t('common.action_tags.haste')})`;
				} else if (tag == 11) {
					name += ` (${i18n.t('common.action_tags.crit')})`;
				} else if (tag == 12) {
					name += ` (${i18n.t('common.action_tags.agi')})`;
				} else if (tag) {
					name += ` (${i18n.t('common.action_tags.maelstrom_weapon', { stacks: tag })})`;
				}
				break;
			case 'Flame Shock':
			case 'Earth Shock':
			case 'Frost Shock':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.dot')})`;
				} else if (tag == 7) {
					name += ` (${i18n.t('common.action_tags.echo')})`;
				}
				break;
			case 'Fire Nova':
				// The cast's hits report under its spell, tagged 1, so they share its row.
				if (tag === 1) {
					name += ` (${i18n.t('common.action_tags.damage')})`;
				}
				break;
			case 'Fulmination':
				name += ` (${tag + 2})`;
				break;
			case 'Lightning Shield':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.wasted')})`;
				}
				break;
			case 'Moonfire':
			case 'Sunfire':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.dot')})`;
				}
				break;
			// For targetted buffs, tag is the source player's raid index or -1 if none.
			case 'Innervate':
			case 'Mana Tide Totem':
			case 'Unholy Frenzy':
			case 'Power Infusion':
				if (tag != -1) {
					if (tag === playerIndex || playerIndex == undefined) {
						name += ` (${i18n.t('common.action_tags.self')})`;
					} else {
						name += ` (${i18n.t('common.action_tags.from_player', { number: tag + 1 })})`;
					}
				} else {
					name += ` (${i18n.t('common.action_tags.raid')})`;
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
					name += ` (${i18n.t('common.action_tags.main_hand')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.off_hand')})`;
				}
				break;
			case 'Heroic Strike':
			case 'Cleave':
			case 'Maul':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.queue')})`;
				}
				break;
			case 'Sweeping Strikes':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.normalized')})`;
				}
				break;
			case 'Sunder Armor':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.devastate')})`;
				}
				break;
			case 'Improved Berserker Rage':
			case 'Improved Overpower':
				name += ` (${tag}/2)`;
				break;
			case 'Thunderfury':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.single_target')})`;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.multi_target')})`;
				}
				break;
			case 'Devouring Plague':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.dot')})`;
					break;
				} else if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.heal')})`;
					break;
				}
				if (tag == 77486) {
					name += ` (${i18n.t('common.action_tags.mastery')})`;
					break;
				}
			case 'Shadow Word: Death':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.no_orb')})`;
				}
			case 'Shadowfiend':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.mana_restore')})`;
				}
				break;
			case 'Immolate':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.dot')})`;
				}
				break;
			case 'Vampiric Touch':
			case 'Shadow Word: Pain':
				if (tag == 77486) {
					name += ` (${i18n.t('common.action_tags.mastery')})`;
				}

				break;
			case 'Unstable Affliction':
			case 'Corruption':
				if (tag == 1) {
					name += ` (${i18n.t('common.action_tags.malefic')})`;
				}
				break;
			case 'Raptor Strike':
				if (tag == 2) {
					name += ` (${i18n.t('common.action_tags.cooldown')})`;
				}
				break;
			case 'Retribution Aura':
			case 'Holy Shield':
			case 'Thorns':
				// A damage shield's proc is tagged two past the aura that carries it: tag 2 is the
				// player's own copy, tag 1 the copy an outside caster provides, and that copy's
				// aura is the tag -1 this case names below.
				if (tag === 1 || tag === 2) {
					name += ` (${i18n.t('common.action_tags.hit')})`;
				} else if (tag === -1) {
					name += ` (${i18n.t('common.action_tags.external')})`;
				}
				break;
			case 'Searing Totem':
			case 'Magma Totem':
			case 'Flametongue Totem':
				// A totem's hits report under the totem's spell, tagged 1, so they share its row.
				if (tag === 1) {
					name += ` (${i18n.t('common.action_names.attack')})`;
				} else if (tag === -1) {
					name += ` (${i18n.t('common.action_tags.external')})`;
				}
				break;
			case 'Dummy Spell':
				if (tag === 100000) {
					[
						i18n.t('common.spell_schools.arcane'),
						i18n.t('common.spell_schools.fire'),
						i18n.t('common.spell_schools.frost'),
						i18n.t('common.spell_schools.holy'),
						i18n.t('common.spell_schools.shadow'),
						i18n.t('common.spell_schools.nature'),
					].forEach((school, index) => {
						if (tag === 100000 + index) {
							name += ` (${school})`;
						}
					});
				}
				break;
			case 'Penance':
				if (tag === 1) {
					name += ' (Heal)';
				}
				break;
			case 'Hack and Slash':
				if (tag == 1) {
					name += ' (Axe/Sword)';
				} else if (tag == 2) {
					name += ' (Fist/Dagger)';
				} else if (tag == 3) {
					name += ' (Mace)';
				}
				break;
			default:
				if (tag === -1) {
					name += ` (${i18n.t('common.action_tags.external')})`;
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
		return this.name.replace(rankSuffixPattern(), '');
	}

	get hasRank(): boolean {
		return this.rank > 0;
	}
}

// Matches the " (Rank N)" suffix the constructor appends, in the current locale's wording.
const rankSuffixPattern = (): RegExp => {
	const [before, after] = i18n.t('common.action_tags.rank', { rank: '\u0000' }).split('\u0000');
	const escape = (text: string) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	const rank = after === undefined ? escape(before) : `${escape(before)}\\d+${escape(after)}`;
	return new RegExp(` \\(${rank}\\)`, 'g');
};

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
