import { Class } from '@generated/proto/common';

import type { PriestSpecs } from '../../proto/spec_types';
import { IconSize, PlayerClass } from '../player_class';
import { PlayerSpec } from '../player_spec';
import { DpsPriest, HealerPriest } from '../specs/priest';
import { getClassArmorTypes, getClassRaces, getClassRangedWeaponTypes, getClassWeaponTypes } from './capabilities';

export class Priest extends PlayerClass<Class.ClassPriest> {
	static classID = Class.ClassPriest as Class.ClassPriest;
	static friendlyName = 'Priest';
	static hexColor = '#fff';
	static specs: Record<string, PlayerSpec<PriestSpecs>> = {
		[DpsPriest.friendlyName]: DpsPriest,
		[HealerPriest.friendlyName]: HealerPriest,
	};
	static races = getClassRaces(Priest.classID);
	static armorTypes = getClassArmorTypes(Priest.classID);
	static weaponTypes = getClassWeaponTypes(Priest.classID);
	static rangedWeaponTypes = getClassRangedWeaponTypes(Priest.classID);

	readonly classID = Priest.classID;
	readonly friendlyName = Priest.friendlyName;
	readonly hexColor = Priest.hexColor;
	readonly specs = Priest.specs;
	readonly races = Priest.races;
	readonly armorTypes = Priest.armorTypes;
	readonly weaponTypes = Priest.weaponTypes;
	readonly rangedWeaponTypes = Priest.rangedWeaponTypes;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_shadow_shadowwordpain.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Priest.getIcon(size);
	};
}
