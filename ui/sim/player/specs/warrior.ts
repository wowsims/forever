import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class DpsWarrior extends PlayerSpec<Spec.SpecDpsWarrior> {
	static specIndex = 0;
	static specID = Spec.SpecDpsWarrior as Spec.SpecDpsWarrior;
	static classID = Class.ClassWarrior as Class.ClassWarrior;
	static friendlyName = 'DPS Warrior';
	static simLink = getSpecSitePath('warrior', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = true;
	static canDualWield = isSpecDualWieldCapable(DpsWarrior.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = DpsWarrior.specIndex;
	readonly specID = DpsWarrior.specID;
	readonly classID = DpsWarrior.classID;
	readonly friendlyName = DpsWarrior.friendlyName;
	readonly simLink = DpsWarrior.simLink;

	readonly isTankSpec = DpsWarrior.isTankSpec;
	readonly isHealingSpec = DpsWarrior.isHealingSpec;
	readonly isRangedDpsSpec = DpsWarrior.isRangedDpsSpec;
	readonly isMeleeDpsSpec = DpsWarrior.isMeleeDpsSpec;

	readonly canDualWield = DpsWarrior.canDualWield;

	readonly launch = DpsWarrior.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/ability_warrior_innerrage.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return DpsWarrior.getIcon(size);
	};
}

export class ProtectionWarrior extends PlayerSpec<Spec.SpecProtectionWarrior> {
	static specIndex = 2;
	static specID = Spec.SpecProtectionWarrior as Spec.SpecProtectionWarrior;
	static classID = Class.ClassWarrior as Class.ClassWarrior;
	static friendlyName = 'Protection';
	static simLink = getSpecSitePath('warrior', 'protection');

	static isTankSpec = true;
	static isHealingSpec = false;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = false;
	static canDualWield = isSpecDualWieldCapable(ProtectionWarrior.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = ProtectionWarrior.specIndex;
	readonly specID = ProtectionWarrior.specID;
	readonly classID = ProtectionWarrior.classID;
	readonly friendlyName = ProtectionWarrior.friendlyName;
	readonly simLink = ProtectionWarrior.simLink;

	readonly isTankSpec = ProtectionWarrior.isTankSpec;
	readonly isHealingSpec = ProtectionWarrior.isHealingSpec;
	readonly isRangedDpsSpec = ProtectionWarrior.isRangedDpsSpec;
	readonly isMeleeDpsSpec = ProtectionWarrior.isMeleeDpsSpec;

	readonly canDualWield = ProtectionWarrior.canDualWield;

	readonly launch = ProtectionWarrior.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/ability_warrior_defensivestance.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return ProtectionWarrior.getIcon(size);
	};
}
