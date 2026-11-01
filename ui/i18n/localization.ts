import {
	AreaType,
	ArmorType,
	CreatureFamily,
	ItemSlot,
	MobType,
	Profession,
	PseudoStat,
	Race,
	RangedWeaponType,
	SpellSchool,
	Stat,
	WeaponSkill,
	WeaponType,
} from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { RaidFilterOption, SourceFilterOption } from '@generated/proto/ui';
import { BulkSimItemSlot } from '@sim/bulk/utils';
import { LaunchStatus } from '@sim/constants/other';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { PlayerClass } from '@sim/player/player_class';
import { PlayerSpec } from '@sim/player/player_spec';

import i18n from './config';
import {
	aplItemLabelI18nKeys,
	classNameToClassKey,
	getAreaTypeI18nKey,
	getArmorTypeI18nKey,
	getBulkSlotI18nKey,
	getClassI18nKey,
	getCreatureFamilyI18nKey,
	getMobTypeI18nKey,
	getPresetConfigurationCategoryI18nKey,
	getProfessionI18nKey,
	getRaceI18nKey,
	getRaidFilterI18nKey,
	getRangedWeaponTypeI18nKey,
	getSlotNameI18nKey,
	getSourceFilterI18nKey,
	getSpecI18nKey,
	getStatusI18nKey,
	getWeaponSkillI18nKey,
	getWeaponTypeI18nKey,
	protoStatNameI18nKeys,
	pseudoStatI18nKeys,
	resourceTypeI18nKeys,
	spellSchoolI18nKeys,
	statI18nKeys,
} from './entity_mapping';
import { getLang } from './locale_service';

/**
 * Entity translation functions
 */

export const translateStat = (stat: Stat): string => {
	return i18n.t(`common.stats.${statI18nKeys[stat] || Stat[stat].toLowerCase()}`, {
		defaultValue: Stat[stat],
	});
};
export const translateProtoStatName = (statName: string): string => {
	return i18n.t(`common.stats.${protoStatNameI18nKeys[statName] || statName.toLowerCase()}`, {
		defaultValue: statName,
	});
};

export const translatePseudoStat = (pseudoStat: PseudoStat): string => {
	return i18n.t(`common.stats.${pseudoStatI18nKeys[pseudoStat] || PseudoStat[pseudoStat].toLowerCase()}`, {
		defaultValue: PseudoStat[pseudoStat],
	});
};

export const translateSpellSchool = (spellSchool: SpellSchool): string => {
	return i18n.t(`common.spell_schools.${spellSchoolI18nKeys[spellSchool] || SpellSchool[spellSchool].toLowerCase()}`, {
		defaultValue: SpellSchool[spellSchool],
	});
};

export const translateMobType = (mobType: MobType): string => {
	return i18n.t(`common.mob_types.${getMobTypeI18nKey(mobType)}`, {
		defaultValue: MobType[mobType],
	});
};

export const translateCreatureFamily = (creatureFamily: CreatureFamily): string => {
	return i18n.t(`common.creature_families.${getCreatureFamilyI18nKey(creatureFamily)}`, {
		defaultValue: CreatureFamily[creatureFamily],
	});
};

export const translateAreaType = (areaType: AreaType): string => {
	return i18n.t(`common.area_types.${getAreaTypeI18nKey(areaType)}`, {
		defaultValue: AreaType[areaType],
	});
};

export const translateRace = (race: Race): string => {
	return i18n.t(`races.${getRaceI18nKey(race)}`, { ns: 'character', defaultValue: Race[race] });
};

export const translateProfession = (profession: Profession): string => {
	return i18n.t(`professions.${getProfessionI18nKey(profession)}`, { ns: 'character', defaultValue: Profession[profession] });
};

export const translateSourceFilter = (source: SourceFilterOption): string => {
	return i18n.t(`common.sources.${getSourceFilterI18nKey(source)}`, {
		defaultValue: SourceFilterOption[source],
	});
};

export const translateRaidFilter = (raid: RaidFilterOption): string => {
	return i18n.t(`common.raids.${getRaidFilterI18nKey(raid)}`, {
		defaultValue: RaidFilterOption[raid],
	});
};

export const translateArmorType = (armorType: ArmorType): string => {
	return i18n.t(`common.armor_types.${getArmorTypeI18nKey(armorType)}`, {
		defaultValue: ArmorType[armorType],
	});
};

export const translateWeaponType = (weaponType: WeaponType): string => {
	return i18n.t(`common.weapon_types.${getWeaponTypeI18nKey(weaponType)}`, {
		defaultValue: WeaponType[weaponType],
	});
};

export const translateRangedWeaponType = (rangedWeaponType: RangedWeaponType): string => {
	return i18n.t(`common.ranged_weapon_types.${getRangedWeaponTypeI18nKey(rangedWeaponType)}`, {
		defaultValue: RangedWeaponType[rangedWeaponType],
	});
};

export const translateWeaponSkill = (weaponSkill: WeaponSkill): string => {
	return i18n.t(`common.weapon_skills.${getWeaponSkillI18nKey(weaponSkill)}`, {
		defaultValue: WeaponSkill[weaponSkill],
	});
};

export const translateResourceType = (resourceType: ResourceType): string => {
	return i18n.t(`common.resource_types.${resourceTypeI18nKeys[resourceType] || ResourceType[resourceType].toLowerCase()}`, {
		defaultValue: ResourceType[resourceType],
	});
};

export const translateStatus = (status: LaunchStatus): string => {
	return i18n.t(`common.status.${getStatusI18nKey(status)}`, {
		defaultValue: LaunchStatus[status],
	});
};

export const translateClass = (className: string): string => {
	return i18n.t(`classes.${classNameToClassKey(className)}`, {
		ns: 'character',
		defaultValue: className,
	});
};

export const translateSpec = (className: string, specName: string): string => {
	const specKey = specName.toLowerCase();
	return i18n.t(`specs.${classNameToClassKey(className)}.${specKey}`, {
		ns: 'character',
		defaultValue: specName,
	});
};

export const translatePlayerClass = (playerClass: PlayerClass<any>): string => {
	return translateClass(getClassI18nKey(playerClass.classID));
};

export const translatePlayerSpec = (playerSpec: PlayerSpec<any>): string => {
	return translateSpec(getClassI18nKey(playerSpec.classID), getSpecI18nKey(playerSpec.specID));
};

/**
 * Component Translation Helpers
 */

// The spec page template is identical for every spec (see ui/index_template.html),
// so class/spec are not baked in as data-class/data-spec attributes; derive
// them from the URL the same way ui/app/spec_entry.ts derives its spec module key
// ('/forever/warrior/arms/' -> ['warrior', 'arms']). Data attributes are kept as a
// fallback for any caller that isn't served from a real spec URL.
export const extractClassAndSpecFromDataAttributes = (): { className: string; specName: string } | null => {
	const base = import.meta.env.BASE_URL || '/';
	const rel = (location.pathname.startsWith(base) ? location.pathname.slice(base.length) : location.pathname)
		.replace(/^\/+/, '')
		.replace(/index\.html$/, '')
		.replace(/\/+$/, '');
	const [className, specName] = rel.split('/');
	if (className && specName) {
		return { className, specName };
	}

	const titleElement = document.querySelector('title');
	if (titleElement) {
		const attrClassName = titleElement.getAttribute('data-class');
		const attrSpecName = titleElement.getAttribute('data-spec');
		if (attrClassName && attrSpecName) {
			return { className: attrClassName, specName: attrSpecName };
		}
	}

	const metaDescription = document.querySelector('meta[name="description"]') as HTMLMetaElement;
	if (metaDescription) {
		const attrClassName = metaDescription.getAttribute('data-class');
		const attrSpecName = metaDescription.getAttribute('data-spec');
		if (attrClassName && attrSpecName) {
			return { className: attrClassName, specName: attrSpecName };
		}
	}
	return null;
};

// A <meta>'s text is its `content` attribute; `textContent` on it is invisible to crawlers.
const setMetaDescription = (text: string): void => document.querySelector('meta[name="description"]')?.setAttribute('content', text);

export const updateLandingPageMetadata = (): void => {
	document.documentElement.lang = getLang();
	document.title = i18n.t('landing.home.title');
	setMetaDescription(i18n.t('landing.home.description'));
};

export const updateSimPageMetadata = (): void => {
	const classSpecInfo = extractClassAndSpecFromDataAttributes();
	if (!classSpecInfo) return;

	const { className, specName } = classSpecInfo;

	// Hunter, Mage, Rogue and Warlock have one spec, which has no name of its own.
	const spec = i18n.t(`specs.${classNameToClassKey(className)}.${specName.toLowerCase()}`, { ns: 'character', defaultValue: '' });
	const translationData = {
		class: translateClass(className),
		spec,
	};
	const suffix = spec ? '' : '_class';

	document.title = i18n.t(`sim.title${suffix}`, translationData);
	setMetaDescription(i18n.t(`sim.description${suffix}`, translationData));
};

export const translateItemLabel = (itemLabel: string): string => {
	try {
		const key = aplItemLabelI18nKeys[itemLabel];
		if (!key) {
			return itemLabel;
		}
		const translated = i18n.t(key);
		if (translated === key) {
			return itemLabel;
		}
		return translated;
	} catch {
		return itemLabel;
	}
};

export const translateSlotName = (slot: ItemSlot): string => {
	const key = getSlotNameI18nKey(slot);
	return i18n.t(`slots.${key}`, { ns: 'character' });
};

export const translateBulkSlotName = (slot: BulkSimItemSlot): string => {
	const key = getBulkSlotI18nKey(slot);
	return i18n.t(`slots.${key}`, { ns: 'character' });
};

export const translatePresetConfigurationCategory = (category: PresetConfigurationCategory): string => {
	return i18n.t(`common.preset.${getPresetConfigurationCategoryI18nKey(category)}`, {
		defaultValue: category,
	});
};

/**
 * Localization Initialization
 */

export const updateTranslations = (): void => {
	document.documentElement.lang = getLang();
	updateSimPageMetadata();
};

export const initLocalization = (): void => {
	const initialize = () => {
		if (!i18n.isInitialized) {
			i18n.init();
		}

		i18n.on('languageChanged', updateTranslations);

		updateTranslations();
	};

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', initialize);
	} else {
		initialize();
	}
};
