import { Class } from '@generated/proto/common';

import type { HunterSpecs } from '../../proto/spec_types';
import { IconSize, PlayerClass } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { Hunter as HunterSpec } from '../specs/hunter';
import { getClassArmorTypes, getClassRaces, getClassRangedWeaponTypes, getClassWeaponTypes } from './capabilities';

export class Hunter extends PlayerClass<Class.ClassHunter> {
	static classID = Class.ClassHunter as Class.ClassHunter;
	static friendlyName = 'Hunter';
	static hexColor = '#abd473';
	static specs: Record<string, PlayerSpec<HunterSpecs>> = {
		[HunterSpec.friendlyName]: HunterSpec,
	};
	static races = getClassRaces(Hunter.classID);
	static armorTypes = getClassArmorTypes(Hunter.classID);
	static weaponTypes = getClassWeaponTypes(Hunter.classID);
	static rangedWeaponTypes = getClassRangedWeaponTypes(Hunter.classID);

	readonly classID = Hunter.classID;
	readonly friendlyName = Hunter.friendlyName;
	readonly hexColor = Hunter.hexColor;
	readonly specs = Hunter.specs;
	readonly races = Hunter.races;
	readonly armorTypes = Hunter.armorTypes;

	readonly weaponTypes = Hunter.weaponTypes;
	readonly rangedWeaponTypes = Hunter.rangedWeaponTypes;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_hunter.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Hunter.getIcon(size);
	};
}
