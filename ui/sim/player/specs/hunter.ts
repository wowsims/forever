import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class Hunter extends PlayerSpec<Spec.SpecHunter> {
	static specIndex = 0;
	static specID = Spec.SpecHunter as Spec.SpecHunter;
	static classID = Class.ClassHunter as Class.ClassHunter;
	static friendlyName = 'Hunter';
	static simLink = getSpecSitePath('hunter', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;

	static canDualWield = isSpecDualWieldCapable(Hunter.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = Hunter.specIndex;
	readonly specID = Hunter.specID;
	readonly classID = Hunter.classID;
	readonly friendlyName = Hunter.friendlyName;
	readonly simLink = Hunter.simLink;

	readonly isTankSpec = Hunter.isTankSpec;
	readonly isHealingSpec = Hunter.isHealingSpec;
	readonly isRangedDpsSpec = Hunter.isRangedDpsSpec;
	readonly isMeleeDpsSpec = Hunter.isMeleeDpsSpec;
	readonly canDualWield = Hunter.canDualWield;

	readonly launch = Hunter.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/class_hunter.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return Hunter.getIcon(size);
	};
}
