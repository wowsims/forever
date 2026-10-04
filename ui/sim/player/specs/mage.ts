import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class Mage extends PlayerSpec<Spec.SpecMage> {
	static specIndex = 0;
	static specID = Spec.SpecMage as Spec.SpecMage;
	static classID = Class.ClassMage as Class.ClassMage;
	static friendlyName = 'Mage';
	static simLink = getSpecSitePath('mage', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = isSpecDualWieldCapable(Mage.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = Mage.specIndex;
	readonly specID = Mage.specID;
	readonly classID = Mage.classID;
	readonly friendlyName = Mage.friendlyName;
	readonly simLink = Mage.simLink;

	readonly isTankSpec = Mage.isTankSpec;
	readonly isHealingSpec = Mage.isHealingSpec;
	readonly isRangedDpsSpec = Mage.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Mage.isMeleeDpsSpec;

	readonly canDualWield = Mage.canDualWield;

	readonly launch = Mage.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_mage.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Mage.getIcon(size);
	};
}
