// The per-spec function table: how to create, compare, copy and serialize each
// spec's rotation, talents and options protos, plus the narrowing helper that
// reaches it. The types it is indexed by live in ./spec_types.
import { Player } from '@generated/proto/api';
import { Spec } from '@generated/proto/common';
import {
	BalanceDruid,
	BalanceDruid_Options,
	BalanceDruid_Rotation,
	DruidTalents,
	FeralBearDruid,
	FeralBearDruid_Options,
	FeralBearDruid_Rotation,
	FeralCatDruid,
	FeralCatDruid_Options,
	FeralCatDruid_Rotation,
	RestorationDruid,
	RestorationDruid_Options,
	RestorationDruid_Rotation,
} from '@generated/proto/druid';
import { Hunter, Hunter_Options, Hunter_Rotation, HunterTalents } from '@generated/proto/hunter';
import { Mage, Mage_Options, Mage_Rotation, MageTalents } from '@generated/proto/mage';
import {
	HolyPaladin,
	HolyPaladin_Options,
	HolyPaladin_Rotation,
	PaladinTalents,
	ProtectionPaladin,
	ProtectionPaladin_Options,
	ProtectionPaladin_Rotation,
	RetributionPaladin,
	RetributionPaladin_Options,
	RetributionPaladin_Rotation,
} from '@generated/proto/paladin';
import {
	DpsPriest,
	DpsPriest_Options,
	DpsPriest_Rotation,
	HealerPriest,
	HealerPriest_Options,
	HealerPriest_Rotation,
	PriestTalents,
} from '@generated/proto/priest';
import { Rogue, Rogue_Options, Rogue_Rotation, RogueTalents } from '@generated/proto/rogue';
import {
	ElementalShaman,
	ElementalShaman_Options,
	ElementalShaman_Rotation,
	EnhancementShaman,
	EnhancementShaman_Options,
	EnhancementShaman_Rotation,
	RestorationShaman,
	RestorationShaman_Options,
	RestorationShaman_Rotation,
	ShamanTalents,
} from '@generated/proto/shaman';
import type { SavedSettings } from '@generated/proto/ui';
import { Warlock, Warlock_Options, Warlock_Rotation, WarlockTalents } from '@generated/proto/warlock';
import {
	DpsWarrior,
	DpsWarrior_Options,
	DpsWarrior_Rotation,
	ProtectionWarrior,
	ProtectionWarrior_Options,
	ProtectionWarrior_Rotation,
	WarriorTalents,
} from '@generated/proto/warrior';

import type { SpecOptions, SpecRotation, SpecTypeFunctions } from './spec_types';
import { UnknownRotation, UnknownSpecOptions, UnknownTalents } from './spec_types';

export const specTypeFunctions: Record<Spec, SpecTypeFunctions<any>> = {
	[Spec.SpecUnknown]: {
		rotationCreate: () => new UnknownRotation(),
		rotationEquals: (_a, _b) => true,
		rotationCopy: _a => new UnknownRotation(),
		rotationToJson: _a => undefined,
		rotationFromJson: _obj => new UnknownRotation(),

		talentsCreate: () => new UnknownTalents(),
		talentsEquals: (_a, _b) => true,
		talentsCopy: _a => new UnknownTalents(),
		talentsToJson: _a => undefined,
		talentsFromJson: _obj => new UnknownTalents(),

		optionsCreate: () => new UnknownSpecOptions(),
		optionsEquals: (_a, _b) => true,
		optionsCopy: _a => new UnknownSpecOptions(),
		optionsToJson: _a => undefined,
		optionsFromJson: _obj => new UnknownSpecOptions(),
		optionsFromPlayer: _player => new UnknownSpecOptions(),
	},

	// Druid
	[Spec.SpecBalanceDruid]: {
		rotationCreate: () => BalanceDruid_Rotation.create(),
		rotationEquals: (a, b) => BalanceDruid_Rotation.equals(a as BalanceDruid_Rotation, b as BalanceDruid_Rotation),
		rotationCopy: a => BalanceDruid_Rotation.clone(a as BalanceDruid_Rotation),
		rotationToJson: a => BalanceDruid_Rotation.toJson(a as BalanceDruid_Rotation),
		rotationFromJson: obj => BalanceDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => BalanceDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => BalanceDruid_Options.equals(a as BalanceDruid_Options, b as BalanceDruid_Options),
		optionsCopy: a => BalanceDruid_Options.clone(a as BalanceDruid_Options),
		optionsToJson: a => BalanceDruid_Options.toJson(a as BalanceDruid_Options),
		optionsFromJson: obj => BalanceDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'balanceDruid'
				? player.spec.balanceDruid.options || BalanceDruid_Options.create()
				: BalanceDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecFeralCatDruid]: {
		rotationCreate: () => FeralCatDruid_Rotation.create(),
		rotationEquals: (a, b) => FeralCatDruid_Rotation.equals(a as FeralCatDruid_Rotation, b as FeralCatDruid_Rotation),
		rotationCopy: a => FeralCatDruid_Rotation.clone(a as FeralCatDruid_Rotation),
		rotationToJson: a => FeralCatDruid_Rotation.toJson(a as FeralCatDruid_Rotation),
		rotationFromJson: obj => FeralCatDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => FeralCatDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => FeralCatDruid_Options.equals(a as FeralCatDruid_Options, b as FeralCatDruid_Options),
		optionsCopy: a => FeralCatDruid_Options.clone(a as FeralCatDruid_Options),
		optionsToJson: a => FeralCatDruid_Options.toJson(a as FeralCatDruid_Options),
		optionsFromJson: obj => FeralCatDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'feralCatDruid'
				? player.spec.feralCatDruid.options || FeralCatDruid_Options.create()
				: FeralCatDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecFeralBearDruid]: {
		rotationCreate: () => FeralBearDruid_Rotation.create(),
		rotationEquals: (a, b) => FeralBearDruid_Rotation.equals(a as FeralBearDruid_Rotation, b as FeralBearDruid_Rotation),
		rotationCopy: a => FeralBearDruid_Rotation.clone(a as FeralBearDruid_Rotation),
		rotationToJson: a => FeralBearDruid_Rotation.toJson(a as FeralBearDruid_Rotation),
		rotationFromJson: obj => FeralBearDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => FeralBearDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => FeralBearDruid_Options.equals(a as FeralBearDruid_Options, b as FeralBearDruid_Options),
		optionsCopy: a => FeralBearDruid_Options.clone(a as FeralBearDruid_Options),
		optionsToJson: a => FeralBearDruid_Options.toJson(a as FeralBearDruid_Options),
		optionsFromJson: obj => FeralBearDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'feralBearDruid'
				? player.spec.feralBearDruid.options || FeralBearDruid_Options.create()
				: FeralBearDruid_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRestorationDruid]: {
		rotationCreate: () => RestorationDruid_Rotation.create(),
		rotationEquals: (a, b) => RestorationDruid_Rotation.equals(a as RestorationDruid_Rotation, b as RestorationDruid_Rotation),
		rotationCopy: a => RestorationDruid_Rotation.clone(a as RestorationDruid_Rotation),
		rotationToJson: a => RestorationDruid_Rotation.toJson(a as RestorationDruid_Rotation),
		rotationFromJson: obj => RestorationDruid_Rotation.fromJson(obj),

		talentsCreate: () => DruidTalents.create(),
		talentsEquals: (a, b) => DruidTalents.equals(a as DruidTalents, b as DruidTalents),
		talentsCopy: a => DruidTalents.clone(a as DruidTalents),
		talentsToJson: a => DruidTalents.toJson(a as DruidTalents),
		talentsFromJson: obj => DruidTalents.fromJson(obj),

		optionsCreate: () => RestorationDruid_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RestorationDruid_Options.equals(a as RestorationDruid_Options, b as RestorationDruid_Options),
		optionsCopy: a => RestorationDruid_Options.clone(a as RestorationDruid_Options),
		optionsToJson: a => RestorationDruid_Options.toJson(a as RestorationDruid_Options),
		optionsFromJson: obj => RestorationDruid_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'restorationDruid'
				? player.spec.restorationDruid.options || RestorationDruid_Options.create()
				: RestorationDruid_Options.create({ classOptions: {} }),
	},
	// Hunter
	[Spec.SpecHunter]: {
		rotationCreate: () => Hunter_Rotation.create(),
		rotationEquals: (a, b) => Hunter_Rotation.equals(a as Hunter_Rotation, b as Hunter_Rotation),
		rotationCopy: a => Hunter_Rotation.clone(a as Hunter_Rotation),
		rotationToJson: a => Hunter_Rotation.toJson(a as Hunter_Rotation),
		rotationFromJson: obj => Hunter_Rotation.fromJson(obj),

		talentsCreate: () => HunterTalents.create(),
		talentsEquals: (a, b) => HunterTalents.equals(a as HunterTalents, b as HunterTalents),
		talentsCopy: a => HunterTalents.clone(a as HunterTalents),
		talentsToJson: a => HunterTalents.toJson(a as HunterTalents),
		talentsFromJson: obj => HunterTalents.fromJson(obj),

		optionsCreate: () => Hunter_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Hunter_Options.equals(a as Hunter_Options, b as Hunter_Options),
		optionsCopy: a => Hunter_Options.clone(a as Hunter_Options),
		optionsToJson: a => Hunter_Options.toJson(a as Hunter_Options),
		optionsFromJson: obj => Hunter_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'hunter' ? player.spec.hunter.options || Hunter_Options.create() : Hunter_Options.create({ classOptions: {} }),
	},
	// Mage
	[Spec.SpecMage]: {
		rotationCreate: () => Mage_Rotation.create(),
		rotationEquals: (a, b) => Mage_Rotation.equals(a as Mage_Rotation, b as Mage_Rotation),
		rotationCopy: a => Mage_Rotation.clone(a as Mage_Rotation),
		rotationToJson: a => Mage_Rotation.toJson(a as Mage_Rotation),
		rotationFromJson: obj => Mage_Rotation.fromJson(obj),

		talentsCreate: () => MageTalents.create(),
		talentsEquals: (a, b) => MageTalents.equals(a as MageTalents, b as MageTalents),
		talentsCopy: a => MageTalents.clone(a as MageTalents),
		talentsToJson: a => MageTalents.toJson(a as MageTalents),
		talentsFromJson: obj => MageTalents.fromJson(obj),

		optionsCreate: () => Mage_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Mage_Options.equals(a as Mage_Options, b as Mage_Options),
		optionsCopy: a => Mage_Options.clone(a as Mage_Options),
		optionsToJson: a => Mage_Options.toJson(a as Mage_Options),
		optionsFromJson: obj => Mage_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'mage' ? player.spec.mage.options || Mage_Options.create() : Mage_Options.create({ classOptions: {} }),
	},
	// Paladin
	[Spec.SpecHolyPaladin]: {
		rotationCreate: () => HolyPaladin_Rotation.create(),
		rotationEquals: (a, b) => HolyPaladin_Rotation.equals(a as HolyPaladin_Rotation, b as HolyPaladin_Rotation),
		rotationCopy: a => HolyPaladin_Rotation.clone(a as HolyPaladin_Rotation),
		rotationToJson: a => HolyPaladin_Rotation.toJson(a as HolyPaladin_Rotation),
		rotationFromJson: obj => HolyPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => HolyPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => HolyPaladin_Options.equals(a as HolyPaladin_Options, b as HolyPaladin_Options),
		optionsCopy: a => HolyPaladin_Options.clone(a as HolyPaladin_Options),
		optionsToJson: a => HolyPaladin_Options.toJson(a as HolyPaladin_Options),
		optionsFromJson: obj => HolyPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'holyPaladin'
				? player.spec.holyPaladin.options || HolyPaladin_Options.create()
				: HolyPaladin_Options.create({ classOptions: {} }),
	},
	[Spec.SpecProtectionPaladin]: {
		rotationCreate: () => ProtectionPaladin_Rotation.create(),
		rotationEquals: (a, b) => ProtectionPaladin_Rotation.equals(a as ProtectionPaladin_Rotation, b as ProtectionPaladin_Rotation),
		rotationCopy: a => ProtectionPaladin_Rotation.clone(a as ProtectionPaladin_Rotation),
		rotationToJson: a => ProtectionPaladin_Rotation.toJson(a as ProtectionPaladin_Rotation),
		rotationFromJson: obj => ProtectionPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => ProtectionPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ProtectionPaladin_Options.equals(a as ProtectionPaladin_Options, b as ProtectionPaladin_Options),
		optionsCopy: a => ProtectionPaladin_Options.clone(a as ProtectionPaladin_Options),
		optionsToJson: a => ProtectionPaladin_Options.toJson(a as ProtectionPaladin_Options),
		optionsFromJson: obj => ProtectionPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'protectionPaladin'
				? player.spec.protectionPaladin.options || ProtectionPaladin_Options.create()
				: ProtectionPaladin_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRetributionPaladin]: {
		rotationCreate: () => RetributionPaladin_Rotation.create(),
		rotationEquals: (a, b) => RetributionPaladin_Rotation.equals(a as RetributionPaladin_Rotation, b as RetributionPaladin_Rotation),
		rotationCopy: a => RetributionPaladin_Rotation.clone(a as RetributionPaladin_Rotation),
		rotationToJson: a => RetributionPaladin_Rotation.toJson(a as RetributionPaladin_Rotation),
		rotationFromJson: obj => RetributionPaladin_Rotation.fromJson(obj),

		talentsCreate: () => PaladinTalents.create(),
		talentsEquals: (a, b) => PaladinTalents.equals(a as PaladinTalents, b as PaladinTalents),
		talentsCopy: a => PaladinTalents.clone(a as PaladinTalents),
		talentsToJson: a => PaladinTalents.toJson(a as PaladinTalents),
		talentsFromJson: obj => PaladinTalents.fromJson(obj),

		optionsCreate: () => RetributionPaladin_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RetributionPaladin_Options.equals(a as RetributionPaladin_Options, b as RetributionPaladin_Options),
		optionsCopy: a => RetributionPaladin_Options.clone(a as RetributionPaladin_Options),
		optionsToJson: a => RetributionPaladin_Options.toJson(a as RetributionPaladin_Options),
		optionsFromJson: obj => RetributionPaladin_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'retributionPaladin'
				? player.spec.retributionPaladin.options || RetributionPaladin_Options.create()
				: RetributionPaladin_Options.create({ classOptions: {} }),
	},
	// Priest
	[Spec.SpecDpsPriest]: {
		rotationCreate: () => DpsPriest_Rotation.create(),
		rotationEquals: (a, b) => DpsPriest_Rotation.equals(a as DpsPriest_Rotation, b as DpsPriest_Rotation),
		rotationCopy: a => DpsPriest_Rotation.clone(a as DpsPriest_Rotation),
		rotationToJson: a => DpsPriest_Rotation.toJson(a as DpsPriest_Rotation),
		rotationFromJson: obj => DpsPriest_Rotation.fromJson(obj),

		talentsCreate: () => PriestTalents.create(),
		talentsEquals: (a, b) => PriestTalents.equals(a as PriestTalents, b as PriestTalents),
		talentsCopy: a => PriestTalents.clone(a as PriestTalents),
		talentsToJson: a => PriestTalents.toJson(a as PriestTalents),
		talentsFromJson: obj => PriestTalents.fromJson(obj),

		optionsCreate: () => DpsPriest_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => DpsPriest_Options.equals(a as DpsPriest_Options, b as DpsPriest_Options),
		optionsCopy: a => DpsPriest_Options.clone(a as DpsPriest_Options),
		optionsToJson: a => DpsPriest_Options.toJson(a as DpsPriest_Options),
		optionsFromJson: obj => DpsPriest_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'dpsPriest' ? player.spec.dpsPriest.options || DpsPriest_Options.create() : DpsPriest_Options.create({ classOptions: {} }),
	},
	[Spec.SpecHealerPriest]: {
		rotationCreate: () => HealerPriest_Rotation.create(),
		rotationEquals: (a, b) => HealerPriest_Rotation.equals(a as HealerPriest_Rotation, b as HealerPriest_Rotation),
		rotationCopy: a => HealerPriest_Rotation.clone(a as HealerPriest_Rotation),
		rotationToJson: a => HealerPriest_Rotation.toJson(a as HealerPriest_Rotation),
		rotationFromJson: obj => HealerPriest_Rotation.fromJson(obj),

		talentsCreate: () => PriestTalents.create(),
		talentsEquals: (a, b) => PriestTalents.equals(a as PriestTalents, b as PriestTalents),
		talentsCopy: a => PriestTalents.clone(a as PriestTalents),
		talentsToJson: a => PriestTalents.toJson(a as PriestTalents),
		talentsFromJson: obj => PriestTalents.fromJson(obj),

		optionsCreate: () => HealerPriest_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => HealerPriest_Options.equals(a as HealerPriest_Options, b as HealerPriest_Options),
		optionsCopy: a => HealerPriest_Options.clone(a as HealerPriest_Options),
		optionsToJson: a => HealerPriest_Options.toJson(a as HealerPriest_Options),
		optionsFromJson: obj => HealerPriest_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'healerPriest'
				? player.spec.healerPriest.options || HealerPriest_Options.create()
				: HealerPriest_Options.create({ classOptions: {} }),
	},
	// Rogue
	[Spec.SpecRogue]: {
		rotationCreate: () => Rogue_Rotation.create(),
		rotationEquals: (a, b) => Rogue_Rotation.equals(a as Rogue_Rotation, b as Rogue_Rotation),
		rotationCopy: a => Rogue_Rotation.clone(a as Rogue_Rotation),
		rotationToJson: a => Rogue_Rotation.toJson(a as Rogue_Rotation),
		rotationFromJson: obj => Rogue_Rotation.fromJson(obj),

		talentsCreate: () => RogueTalents.create(),
		talentsEquals: (a, b) => RogueTalents.equals(a as RogueTalents, b as RogueTalents),
		talentsCopy: a => RogueTalents.clone(a as RogueTalents),
		talentsToJson: a => RogueTalents.toJson(a as RogueTalents),
		talentsFromJson: obj => RogueTalents.fromJson(obj),

		optionsCreate: () => Rogue_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Rogue_Options.equals(a as Rogue_Options, b as Rogue_Options),
		optionsCopy: a => Rogue_Options.clone(a as Rogue_Options),
		optionsToJson: a => Rogue_Options.toJson(a as Rogue_Options),
		optionsFromJson: obj => Rogue_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'rogue' ? player.spec.rogue.options || Rogue_Options.create() : Rogue_Options.create({ classOptions: {} }),
	},
	// Shaman
	[Spec.SpecElementalShaman]: {
		rotationCreate: () => ElementalShaman_Rotation.create(),
		rotationEquals: (a, b) => ElementalShaman_Rotation.equals(a as ElementalShaman_Rotation, b as ElementalShaman_Rotation),
		rotationCopy: a => ElementalShaman_Rotation.clone(a as ElementalShaman_Rotation),
		rotationToJson: a => ElementalShaman_Rotation.toJson(a as ElementalShaman_Rotation),
		rotationFromJson: obj => ElementalShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => ElementalShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ElementalShaman_Options.equals(a as ElementalShaman_Options, b as ElementalShaman_Options),
		optionsCopy: a => ElementalShaman_Options.clone(a as ElementalShaman_Options),
		optionsToJson: a => ElementalShaman_Options.toJson(a as ElementalShaman_Options),
		optionsFromJson: obj => ElementalShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'elementalShaman'
				? player.spec.elementalShaman.options || ElementalShaman_Options.create()
				: ElementalShaman_Options.create({ classOptions: {} }),
	},
	[Spec.SpecEnhancementShaman]: {
		rotationCreate: () => EnhancementShaman_Rotation.create(),
		rotationEquals: (a, b) => EnhancementShaman_Rotation.equals(a as EnhancementShaman_Rotation, b as EnhancementShaman_Rotation),
		rotationCopy: a => EnhancementShaman_Rotation.clone(a as EnhancementShaman_Rotation),
		rotationToJson: a => EnhancementShaman_Rotation.toJson(a as EnhancementShaman_Rotation),
		rotationFromJson: obj => EnhancementShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => EnhancementShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => EnhancementShaman_Options.equals(a as EnhancementShaman_Options, b as EnhancementShaman_Options),
		optionsCopy: a => EnhancementShaman_Options.clone(a as EnhancementShaman_Options),
		optionsToJson: a => EnhancementShaman_Options.toJson(a as EnhancementShaman_Options),
		optionsFromJson: obj => EnhancementShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'enhancementShaman'
				? player.spec.enhancementShaman.options || EnhancementShaman_Options.create()
				: EnhancementShaman_Options.create({ classOptions: {} }),
	},
	[Spec.SpecRestorationShaman]: {
		rotationCreate: () => RestorationShaman_Rotation.create(),
		rotationEquals: (a, b) => RestorationShaman_Rotation.equals(a as RestorationShaman_Rotation, b as RestorationShaman_Rotation),
		rotationCopy: a => RestorationShaman_Rotation.clone(a as RestorationShaman_Rotation),
		rotationToJson: a => RestorationShaman_Rotation.toJson(a as RestorationShaman_Rotation),
		rotationFromJson: obj => RestorationShaman_Rotation.fromJson(obj),

		talentsCreate: () => ShamanTalents.create(),
		talentsEquals: (a, b) => ShamanTalents.equals(a as ShamanTalents, b as ShamanTalents),
		talentsCopy: a => ShamanTalents.clone(a as ShamanTalents),
		talentsToJson: a => ShamanTalents.toJson(a as ShamanTalents),
		talentsFromJson: obj => ShamanTalents.fromJson(obj),

		optionsCreate: () => RestorationShaman_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => RestorationShaman_Options.equals(a as RestorationShaman_Options, b as RestorationShaman_Options),
		optionsCopy: a => RestorationShaman_Options.clone(a as RestorationShaman_Options),
		optionsToJson: a => RestorationShaman_Options.toJson(a as RestorationShaman_Options),
		optionsFromJson: obj => RestorationShaman_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'restorationShaman'
				? player.spec.restorationShaman.options || RestorationShaman_Options.create()
				: RestorationShaman_Options.create({ classOptions: {} }),
	},
	// Warlock
	[Spec.SpecWarlock]: {
		rotationCreate: () => Warlock_Rotation.create(),
		rotationEquals: (a, b) => Warlock_Rotation.equals(a as Warlock_Rotation, b as Warlock_Rotation),
		rotationCopy: a => Warlock_Rotation.clone(a as Warlock_Rotation),
		rotationToJson: a => Warlock_Rotation.toJson(a as Warlock_Rotation),
		rotationFromJson: obj => Warlock_Rotation.fromJson(obj),

		talentsCreate: () => WarlockTalents.create(),
		talentsEquals: (a, b) => WarlockTalents.equals(a as WarlockTalents, b as WarlockTalents),
		talentsCopy: a => WarlockTalents.clone(a as WarlockTalents),
		talentsToJson: a => WarlockTalents.toJson(a as WarlockTalents),
		talentsFromJson: obj => WarlockTalents.fromJson(obj),

		optionsCreate: () => Warlock_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => Warlock_Options.equals(a as Warlock_Options, b as Warlock_Options),
		optionsCopy: a => Warlock_Options.clone(a as Warlock_Options),
		optionsToJson: a => Warlock_Options.toJson(a as Warlock_Options),
		optionsFromJson: obj => Warlock_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'warlock' ? player.spec.warlock.options || Warlock_Options.create() : Warlock_Options.create({ classOptions: {} }),
	},
	// Warrior
	[Spec.SpecDpsWarrior]: {
		rotationCreate: () => DpsWarrior_Rotation.create(),
		rotationEquals: (a, b) => DpsWarrior_Rotation.equals(a as DpsWarrior_Rotation, b as DpsWarrior_Rotation),
		rotationCopy: a => DpsWarrior_Rotation.clone(a as DpsWarrior_Rotation),
		rotationToJson: a => DpsWarrior_Rotation.toJson(a as DpsWarrior_Rotation),
		rotationFromJson: obj => DpsWarrior_Rotation.fromJson(obj),

		talentsCreate: () => WarriorTalents.create(),
		talentsEquals: (a, b) => WarriorTalents.equals(a as WarriorTalents, b as WarriorTalents),
		talentsCopy: a => WarriorTalents.clone(a as WarriorTalents),
		talentsToJson: a => WarriorTalents.toJson(a as WarriorTalents),
		talentsFromJson: obj => WarriorTalents.fromJson(obj),

		optionsCreate: () => DpsWarrior_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => DpsWarrior_Options.equals(a as DpsWarrior_Options, b as DpsWarrior_Options),
		optionsCopy: a => DpsWarrior_Options.clone(a as DpsWarrior_Options),
		optionsToJson: a => DpsWarrior_Options.toJson(a as DpsWarrior_Options),
		optionsFromJson: obj => DpsWarrior_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'dpsWarrior'
				? player.spec.dpsWarrior.options || DpsWarrior_Options.create()
				: DpsWarrior_Options.create({ classOptions: {} }),
	},
	[Spec.SpecProtectionWarrior]: {
		rotationCreate: () => ProtectionWarrior_Rotation.create(),
		rotationEquals: (a, b) => ProtectionWarrior_Rotation.equals(a as ProtectionWarrior_Rotation, b as ProtectionWarrior_Rotation),
		rotationCopy: a => ProtectionWarrior_Rotation.clone(a as ProtectionWarrior_Rotation),
		rotationToJson: a => ProtectionWarrior_Rotation.toJson(a as ProtectionWarrior_Rotation),
		rotationFromJson: obj => ProtectionWarrior_Rotation.fromJson(obj),

		talentsCreate: () => WarriorTalents.create(),
		talentsEquals: (a, b) => WarriorTalents.equals(a as WarriorTalents, b as WarriorTalents),
		talentsCopy: a => WarriorTalents.clone(a as WarriorTalents),
		talentsToJson: a => WarriorTalents.toJson(a as WarriorTalents),
		talentsFromJson: obj => WarriorTalents.fromJson(obj),

		optionsCreate: () => ProtectionWarrior_Options.create({ classOptions: {} }),
		optionsEquals: (a, b) => ProtectionWarrior_Options.equals(a as ProtectionWarrior_Options, b as ProtectionWarrior_Options),
		optionsCopy: a => ProtectionWarrior_Options.clone(a as ProtectionWarrior_Options),
		optionsToJson: a => ProtectionWarrior_Options.toJson(a as ProtectionWarrior_Options),
		optionsFromJson: obj => ProtectionWarrior_Options.fromJson(obj),
		optionsFromPlayer: player =>
			player.spec.oneofKind == 'protectionWarrior'
				? player.spec.protectionWarrior.options || ProtectionWarrior_Options.create()
				: ProtectionWarrior_Options.create(),
	},
};

// True when `entry` holds spec options for the spec `live` is; none, or another spec's, holds nothing to apply.
export const isSameSpecArm = (entry: Player['spec'], live: Player['spec']): boolean => entry.oneofKind !== undefined && entry.oneofKind === live.oneofKind;

// An entry without spec options for the live spec takes the live ones, which loading it keeps.
export const withLiveSpecOptions = (settings: SavedSettings, live: SavedSettings | null): SavedSettings =>
	!live || isSameSpecArm(settings.spec, live.spec) ? settings : { ...settings, spec: live.spec };

// Returns a copy of playerOptions, with the class field set.
export function withSpec<SpecType extends Spec>(spec: Spec, player: Player, specOptions: SpecOptions<SpecType>): Player {
	const copy = Player.clone(player);

	switch (spec) {
		// Druid
		case Spec.SpecBalanceDruid:
			copy.spec = {
				oneofKind: 'balanceDruid',
				balanceDruid: BalanceDruid.create({
					options: specOptions as BalanceDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecFeralCatDruid:
			copy.spec = {
				oneofKind: 'feralCatDruid',
				feralCatDruid: FeralCatDruid.create({
					options: specOptions as FeralCatDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecFeralBearDruid:
			copy.spec = {
				oneofKind: 'feralBearDruid',
				feralBearDruid: FeralBearDruid.create({
					options: specOptions as FeralBearDruid_Options,
				}),
			};
			return copy;
		case Spec.SpecRestorationDruid:
			copy.spec = {
				oneofKind: 'restorationDruid',
				restorationDruid: RestorationDruid.create({
					options: specOptions as RestorationDruid_Options,
				}),
			};
			return copy;
		// Hunter
		case Spec.SpecHunter:
			copy.spec = {
				oneofKind: 'hunter',
				hunter: Hunter.create({
					options: specOptions as Hunter_Options,
				}),
			};
			return copy;
		// Mage
		case Spec.SpecMage:
			copy.spec = {
				oneofKind: 'mage',
				mage: Mage.create({
					options: specOptions as Mage_Options,
				}),
			};
			return copy;
		// Paladin
		case Spec.SpecHolyPaladin:
			copy.spec = {
				oneofKind: 'holyPaladin',
				holyPaladin: HolyPaladin.create({
					options: specOptions as HolyPaladin_Options,
				}),
			};
			return copy;
		case Spec.SpecProtectionPaladin:
			copy.spec = {
				oneofKind: 'protectionPaladin',
				protectionPaladin: ProtectionPaladin.create({
					options: specOptions as ProtectionPaladin_Options,
				}),
			};
			return copy;
		case Spec.SpecRetributionPaladin:
			copy.spec = {
				oneofKind: 'retributionPaladin',
				retributionPaladin: RetributionPaladin.create({
					options: specOptions as RetributionPaladin_Options,
				}),
			};
			return copy;
		// Priest
		case Spec.SpecDpsPriest:
			copy.spec = {
				oneofKind: 'dpsPriest',
				dpsPriest: DpsPriest.create({
					options: specOptions as DpsPriest_Options,
				}),
			};
			return copy;
		case Spec.SpecHealerPriest:
			copy.spec = {
				oneofKind: 'healerPriest',
				healerPriest: HealerPriest.create({
					options: specOptions as HealerPriest_Options,
				}),
			};
			return copy;
		// Rogue
		case Spec.SpecRogue:
			copy.spec = {
				oneofKind: 'rogue',
				rogue: Rogue.create({
					options: specOptions as Rogue_Options,
				}),
			};
			return copy;
		// Shaman
		case Spec.SpecElementalShaman:
			copy.spec = {
				oneofKind: 'elementalShaman',
				elementalShaman: ElementalShaman.create({
					options: specOptions as ElementalShaman_Options,
				}),
			};
			return copy;
		case Spec.SpecEnhancementShaman:
			copy.spec = {
				oneofKind: 'enhancementShaman',
				enhancementShaman: EnhancementShaman.create({
					options: specOptions as EnhancementShaman_Options,
				}),
			};
			return copy;
		case Spec.SpecRestorationShaman:
			copy.spec = {
				oneofKind: 'restorationShaman',
				restorationShaman: RestorationShaman.create({
					options: specOptions as RestorationShaman_Options,
				}),
			};
			return copy;
		// Warlock
		case Spec.SpecWarlock:
			copy.spec = {
				oneofKind: 'warlock',
				warlock: Warlock.create({
					options: specOptions as Warlock_Options,
				}),
			};
			return copy;
		// Warrior
		case Spec.SpecDpsWarrior:
			copy.spec = {
				oneofKind: 'dpsWarrior',
				dpsWarrior: DpsWarrior.create({
					options: specOptions as DpsWarrior_Options,
				}),
			};
			return copy;
		case Spec.SpecProtectionWarrior:
			copy.spec = {
				oneofKind: 'protectionWarrior',
				protectionWarrior: ProtectionWarrior.create({
					options: specOptions as ProtectionWarrior_Options,
				}),
			};
			return copy;
		default:
			return copy;
	}
}

/**
 * The spec rotation a Simple rotation's json stores. An empty json is the spec's default rotation: a Simple
 * rotation gets its json only once a rotation option is set, so picking a cooldown first leaves it empty.
 */
export function simpleSpecRotationFromJson<SpecType extends Spec>(functions: SpecTypeFunctions<SpecType>, jsonStr: string): SpecRotation<SpecType> {
	if (!jsonStr) {
		return functions.rotationCreate();
	}

	try {
		const json = JSON.parse(jsonStr);
		return functions.rotationFromJson(json);
	} catch (e) {
		console.warn(`Error parsing rotation spec options: ${e}\n\nSpec options: '${jsonStr}'`);
		return functions.rotationCreate();
	}
}
