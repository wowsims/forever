import { OtherAction } from '@generated/proto/common';
import type { ActionId } from '@sim/proto/action_id';

export const MELEE_ACTION_CATEGORY = 1;
export const SPELL_ACTION_CATEGORY = 2;
export const DEFAULT_ACTION_CATEGORY = 3;

// A fixed sort category for casts whose default MELEE/SPELL/DEFAULT bucket would order them
// badly. An entry belongs here only if its id is reachable - a spellId in a committed APL, a
// `SpellID:` literal in sim/**/*.go, an id the game files' spellIcons table knows, a reference in
// ui/*.ts(x), or one in a committed ui/ JSON. Spells register under a rank id (Arcane Shot is
// 27019 here, not 3044). Flat Record<number, number> keyed by anyId(), so an OtherAction id and a
// spellId of the same numeric value are indistinguishable; safe only because OtherAction values
// are small and spell ids are large.
export const rotationCategoryOverrides: Record<number, number> = {
	[OtherAction.OtherActionMove]: 0,
	[OtherAction.OtherActionAttack]: 0.01,
	[OtherAction.OtherActionShoot]: 0.5,

	// Wand: the row opens with the channel, which deals nothing itself, so it sorts as a spell
	[5019]: SPELL_ACTION_CATEGORY, // Shoot

	// Hunter
	[34490]: MELEE_ACTION_CATEGORY + 0.29, // Silencing Shot

	// Paladin
	[20424]: MELEE_ACTION_CATEGORY + 0.0001, // Seal of Command
	[25742]: MELEE_ACTION_CATEGORY + 0.0011, // Seal of Righteousness (rank 1)
	[25740]: MELEE_ACTION_CATEGORY + 0.0012, // Seal of Righteousness (rank 2)
	[25739]: MELEE_ACTION_CATEGORY + 0.0013, // Seal of Righteousness (rank 3)
	[25738]: MELEE_ACTION_CATEGORY + 0.0014, // Seal of Righteousness (rank 4)
	[25737]: MELEE_ACTION_CATEGORY + 0.0015, // Seal of Righteousness (rank 5)
	[25736]: MELEE_ACTION_CATEGORY + 0.0016, // Seal of Righteousness (rank 6)
	[25735]: MELEE_ACTION_CATEGORY + 0.0017, // Seal of Righteousness (rank 7)
	[25713]: MELEE_ACTION_CATEGORY + 0.0018, // Seal of Righteousness (rank 8)
	[1311647]: MELEE_ACTION_CATEGORY + 0.0021, // Seal of Fury (rank 1)
	[1311654]: MELEE_ACTION_CATEGORY + 0.0022, // Seal of Fury (rank 2)
	[20231]: MELEE_ACTION_CATEGORY + 0.0023, // Seal of Fury (rank 3)
	[20415]: MELEE_ACTION_CATEGORY + 0.0024, // Seal of Fury (rank 4)
	[20416]: MELEE_ACTION_CATEGORY + 0.0025, // Seal of Fury (rank 5)
	[20417]: MELEE_ACTION_CATEGORY + 0.0026, // Seal of Fury (rank 6)
	[20418]: MELEE_ACTION_CATEGORY + 0.0027, // Seal of Fury (rank 7)
	[20267]: MELEE_ACTION_CATEGORY + 0.0031, // Judgment of Light (rank 1) -- proc
	[20341]: MELEE_ACTION_CATEGORY + 0.0032, // Judgment of Light (rank 2) -- proc
	[20342]: MELEE_ACTION_CATEGORY + 0.0033, // Judgment of Light (rank 3) -- proc
	[20343]: MELEE_ACTION_CATEGORY + 0.0034, // Judgment of Light (rank 4) -- proc
	[20268]: MELEE_ACTION_CATEGORY + 0.0041, // Judgment of Wisdom (rank 1) -- proc
	[20352]: MELEE_ACTION_CATEGORY + 0.0042, // Judgment of Wisdom (rank 2) -- proc
	[20353]: MELEE_ACTION_CATEGORY + 0.0043, // Judgment of Wisdom (rank 3) -- proc
	[679]: MELEE_ACTION_CATEGORY + 0.011, // Holy Strike (rank 1)
	[678]: MELEE_ACTION_CATEGORY + 0.012, // Holy Strike (rank 2)
	[1866]: MELEE_ACTION_CATEGORY + 0.013, // Holy Strike (rank 3)
	[680]: MELEE_ACTION_CATEGORY + 0.014, // Holy Strike (rank 4)
	[2495]: MELEE_ACTION_CATEGORY + 0.015, // Holy Strike (rank 5)
	[5569]: MELEE_ACTION_CATEGORY + 0.016, // Holy Strike (rank 6)
	[10332]: MELEE_ACTION_CATEGORY + 0.017, // Holy Strike (rank 7)
	[10333]: MELEE_ACTION_CATEGORY + 0.018, // Holy Strike (rank 8)
	[20271]: MELEE_ACTION_CATEGORY + 0.02, // Judgment
	[20425]: MELEE_ACTION_CATEGORY + 0.0211, // Judgment of Command (rank 1)
	[20962]: MELEE_ACTION_CATEGORY + 0.0212, // Judgment of Command (rank 2)
	[20961]: MELEE_ACTION_CATEGORY + 0.0213, // Judgment of Command (rank 3)
	[20967]: MELEE_ACTION_CATEGORY + 0.0214, // Judgment of Command (rank 4)
	[20968]: MELEE_ACTION_CATEGORY + 0.0215, // Judgment of Command (rank 5)
	[20187]: MELEE_ACTION_CATEGORY + 0.0221, // Judgment of Righteousness (rank 1)
	[20280]: MELEE_ACTION_CATEGORY + 0.0222, // Judgment of Righteousness (rank 2)
	[20281]: MELEE_ACTION_CATEGORY + 0.0223, // Judgment of Righteousness (rank 3)
	[20282]: MELEE_ACTION_CATEGORY + 0.0224, // Judgment of Righteousness (rank 4)
	[20283]: MELEE_ACTION_CATEGORY + 0.0225, // Judgment of Righteousness (rank 5)
	[20284]: MELEE_ACTION_CATEGORY + 0.0226, // Judgment of Righteousness (rank 6)
	[20285]: MELEE_ACTION_CATEGORY + 0.0227, // Judgment of Righteousness (rank 7)
	[20286]: MELEE_ACTION_CATEGORY + 0.0228, // Judgment of Righteousness (rank 8)
	[1311650]: MELEE_ACTION_CATEGORY + 0.0231, // Judgment of Fury (rank 1)
	[1311655]: MELEE_ACTION_CATEGORY + 0.0232, // Judgment of Fury (rank 2)
	[20183]: MELEE_ACTION_CATEGORY + 0.0233, // Judgment of Fury (rank 3)
	[20411]: MELEE_ACTION_CATEGORY + 0.0234, // Judgment of Fury (rank 4)
	[20412]: MELEE_ACTION_CATEGORY + 0.0235, // Judgment of Fury (rank 5)
	[20413]: MELEE_ACTION_CATEGORY + 0.0236, // Judgment of Fury (rank 6)
	[20414]: MELEE_ACTION_CATEGORY + 0.0237, // Judgment of Fury (rank 7)
	[20232]: MELEE_ACTION_CATEGORY + 0.0238, // Judgment of Fury (taunt)
	[1311604]: MELEE_ACTION_CATEGORY + 0.0311, // Holy Shock (rank 1) - damage
	[25912]: MELEE_ACTION_CATEGORY + 0.0312, // Holy Shock (rank 2) - damage
	[25911]: MELEE_ACTION_CATEGORY + 0.0313, // Holy Shock (rank 3) - damage
	[25902]: MELEE_ACTION_CATEGORY + 0.0314, // Holy Shock (rank 4) - damage
	[1311605]: MELEE_ACTION_CATEGORY + 0.0321, // Holy Shock (rank 1) - heal
	[25914]: MELEE_ACTION_CATEGORY + 0.0322, // Holy Shock (rank 2) - heal
	[25913]: MELEE_ACTION_CATEGORY + 0.0323, // Holy Shock (rank 3) - heal
	[25903]: MELEE_ACTION_CATEGORY + 0.0324, // Holy Shock (rank 4) - heal
	[26573]: MELEE_ACTION_CATEGORY + 0.041, // Consecration (rank 1)
	[20116]: MELEE_ACTION_CATEGORY + 0.042, // Consecration (rank 2)
	[20922]: MELEE_ACTION_CATEGORY + 0.043, // Consecration (rank 3)
	[20923]: MELEE_ACTION_CATEGORY + 0.044, // Consecration (rank 4)
	[20924]: MELEE_ACTION_CATEGORY + 0.045, // Consecration (rank 5)
	[24275]: MELEE_ACTION_CATEGORY + 0.051, // Hammer of Wrath (rank 1)
	[24274]: MELEE_ACTION_CATEGORY + 0.052, // Hammer of Wrath (rank 2)
	[24239]: MELEE_ACTION_CATEGORY + 0.053, // Hammer of Wrath (rank 3)
	[20216]: MELEE_ACTION_CATEGORY + 0.06, // Divine Favor
	[20375]: MELEE_ACTION_CATEGORY + 0.0711, // Seal of Command (rank 1)
	[20915]: MELEE_ACTION_CATEGORY + 0.0712, // Seal of Command (rank 2)
	[20918]: MELEE_ACTION_CATEGORY + 0.0713, // Seal of Command (rank 3)
	[20919]: MELEE_ACTION_CATEGORY + 0.0714, // Seal of Command (rank 4)
	[20920]: MELEE_ACTION_CATEGORY + 0.0715, // Seal of Command (rank 5)
	[21084]: MELEE_ACTION_CATEGORY + 0.0721, // Seal of Righteousness (rank 1)
	[20287]: MELEE_ACTION_CATEGORY + 0.0722, // Seal of Righteousness (rank 2)
	[20288]: MELEE_ACTION_CATEGORY + 0.0723, // Seal of Righteousness (rank 3)
	[20289]: MELEE_ACTION_CATEGORY + 0.0724, // Seal of Righteousness (rank 4)
	[20290]: MELEE_ACTION_CATEGORY + 0.0725, // Seal of Righteousness (rank 5)
	[20291]: MELEE_ACTION_CATEGORY + 0.0726, // Seal of Righteousness (rank 6)
	[20292]: MELEE_ACTION_CATEGORY + 0.0727, // Seal of Righteousness (rank 7)
	[20293]: MELEE_ACTION_CATEGORY + 0.0728, // Seal of Righteousness (rank 8)
	[1311649]: MELEE_ACTION_CATEGORY + 0.0731, // Seal of Fury (rank 1)
	[1311656]: MELEE_ACTION_CATEGORY + 0.0732, // Seal of Fury (rank 2)
	[20163]: MELEE_ACTION_CATEGORY + 0.0733, // Seal of Fury (rank 3)
	[20419]: MELEE_ACTION_CATEGORY + 0.0734, // Seal of Fury (rank 4)
	[20421]: MELEE_ACTION_CATEGORY + 0.0735, // Seal of Fury (rank 5)
	[20422]: MELEE_ACTION_CATEGORY + 0.0736, // Seal of Fury (rank 6)
	[20423]: MELEE_ACTION_CATEGORY + 0.0737, // Seal of Fury (rank 7)
	[20165]: MELEE_ACTION_CATEGORY + 0.0741, // Seal of Light (rank 1)
	[20347]: MELEE_ACTION_CATEGORY + 0.0742, // Seal of Light (rank 2)
	[20348]: MELEE_ACTION_CATEGORY + 0.0743, // Seal of Light (rank 3)
	[20349]: MELEE_ACTION_CATEGORY + 0.0744, // Seal of Light (rank 4)
	[20166]: MELEE_ACTION_CATEGORY + 0.0751, // Seal of Wisdom (rank 1)
	[20356]: MELEE_ACTION_CATEGORY + 0.0752, // Seal of Wisdom (rank 2)
	[20357]: MELEE_ACTION_CATEGORY + 0.0753, // Seal of Wisdom (rank 3)
	[21082]: MELEE_ACTION_CATEGORY + 0.0761, // Seal of the Crusader (rank 1)
	[20162]: MELEE_ACTION_CATEGORY + 0.0762, // Seal of the Crusader (rank 2)
	[20305]: MELEE_ACTION_CATEGORY + 0.0763, // Seal of the Crusader (rank 3)
	[20306]: MELEE_ACTION_CATEGORY + 0.0764, // Seal of the Crusader (rank 4)
	[20307]: MELEE_ACTION_CATEGORY + 0.0765, // Seal of the Crusader (rank 5)
	[20308]: MELEE_ACTION_CATEGORY + 0.0766, // Seal of the Crusader (rank 6)
	[20164]: MELEE_ACTION_CATEGORY + 0.0771, // Seal of Justice (rank 1)
	[15851]: MELEE_ACTION_CATEGORY + 0.08, // Dragonbreath Chili
	[10646]: MELEE_ACTION_CATEGORY + 0.09, // Goblin Sapper Charge
	[12662]: MELEE_ACTION_CATEGORY + 0.1, // Demonic Rune
	[440667]: MELEE_ACTION_CATEGORY + 0.11, // Vindication
	[20185]: MELEE_ACTION_CATEGORY + 0.2011, // Judgment of Light (rank 1)
	[20344]: MELEE_ACTION_CATEGORY + 0.2012, // Judgment of Light (rank 2)
	[20345]: MELEE_ACTION_CATEGORY + 0.2013, // Judgment of Light (rank 3)
	[20346]: MELEE_ACTION_CATEGORY + 0.2014, // Judgment of Light (rank 4)
	[20186]: MELEE_ACTION_CATEGORY + 0.2021, // Judgment of Wisdom (rank 1)
	[20354]: MELEE_ACTION_CATEGORY + 0.2022, // Judgment of Wisdom (rank 2)
	[20355]: MELEE_ACTION_CATEGORY + 0.2023, // Judgment of Wisdom (rank 3)
	[21183]: MELEE_ACTION_CATEGORY + 0.2031, // Judgment of the Crusader (rank 1)
	[20188]: MELEE_ACTION_CATEGORY + 0.2032, // Judgment of the Crusader (rank 2)
	[20300]: MELEE_ACTION_CATEGORY + 0.2033, // Judgment of the Crusader (rank 3)
	[20301]: MELEE_ACTION_CATEGORY + 0.2034, // Judgment of the Crusader (rank 4)
	[20302]: MELEE_ACTION_CATEGORY + 0.2035, // Judgment of the Crusader (rank 5)
	[20303]: MELEE_ACTION_CATEGORY + 0.2036, // Judgment of the Crusader (rank 6)
	[20184]: MELEE_ACTION_CATEGORY + 0.2041, // Judgment of Justice (rank 1)

	// Rogue
	[6774]: MELEE_ACTION_CATEGORY + 0.1, // Slice and Dice
	[8647]: MELEE_ACTION_CATEGORY + 0.2, // Expose Armor

	// Shaman
	[8232]: 0.11, // Windfury Weapon
	[8024]: 0.12, // Flametongue Weapon
	[8033]: 0.12, // Frostbrand Weapon
	[17364]: MELEE_ACTION_CATEGORY + 0.1, // Stormstrike

	// Warlock
	[603]: SPELL_ACTION_CATEGORY + 0.01, // Curse of Doom
	[30108]: SPELL_ACTION_CATEGORY + 0.3, // Unstable Affliction
	[17962]: SPELL_ACTION_CATEGORY + 0.32, // Conflagrate

	// Mage
	[12472]: SPELL_ACTION_CATEGORY + 0.4, // Icy Veins
	[11129]: SPELL_ACTION_CATEGORY + 0.4, // Combustion
	[12042]: SPELL_ACTION_CATEGORY + 0.4, // Arcane Power
	[11958]: SPELL_ACTION_CATEGORY + 0.41, // Cold Snap
	[12043]: SPELL_ACTION_CATEGORY + 0.41, // Presence of Mind
	[31687]: SPELL_ACTION_CATEGORY + 0.41, // Water Elemental
	[12051]: SPELL_ACTION_CATEGORY + 0.52, // Evocate
	[12536]: SPELL_ACTION_CATEGORY + 0.61, // Clearcasting

	// Warrior
	[23894]: MELEE_ACTION_CATEGORY + 0.1, // Bloodthirst
	[23925]: MELEE_ACTION_CATEGORY + 0.1, // Shield Slam
	[1680]: MELEE_ACTION_CATEGORY + 0.24, // Whirlwind
	// The swing and the tag 1 queue spell share a category, so their rows sit together
	[11567]: MELEE_ACTION_CATEGORY + 0.3, // Heroic Strike (rank 8)
	[25286]: MELEE_ACTION_CATEGORY + 0.3, // Heroic Strike (rank 9)
	[412609]: SPELL_ACTION_CATEGORY + 0.51, // Deep Wounds
	[2565]: SPELL_ACTION_CATEGORY + 0.62, // Shield Block
	// The stances sit together above every other cast row, Move and the auto attack included
	[71]: -0.1, // Defensive Stance
	[2457]: -0.1, // Battle Stance
	[2458]: -0.1, // Berserker Stance
};

export function actionCategory(actionId: ActionId, meleeKeys: ReadonlySet<string>, spellKeys: ReadonlySet<string>): number {
	const fixedCategory = rotationCategoryOverrides[actionId.anyId()];
	if (fixedCategory != null) return fixedCategory;
	const key = actionId.equalityKey();
	if (meleeKeys.has(key)) return MELEE_ACTION_CATEGORY;
	if (spellKeys.has(key)) return SPELL_ACTION_CATEGORY;
	return DEFAULT_ACTION_CATEGORY;
}
