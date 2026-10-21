import { Class, Spec } from '@generated/proto/common';

import { CURRENT_PHASE, getSpecSitePath, LaunchStatus } from '../../constants/other';
import { isSpecDualWieldCapable } from '../classes/capabilities';
import { IconSize } from '../player_class';
import { PlayerSpec, SimStatus } from '../player_spec';

export class DpsPriest extends PlayerSpec<Spec.SpecDpsPriest> {
	static specIndex = 0;
	static specID = Spec.SpecDpsPriest as Spec.SpecDpsPriest;
	static classID = Class.ClassPriest as Class.ClassPriest;
	static friendlyName = 'DPS';
	static simLink = getSpecSitePath('priest', 'dps');

	static isTankSpec = false;
	static isHealingSpec = false;
	static isRangedDpsSpec = true;
	static isMeleeDpsSpec = false;
	static canDualWield = isSpecDualWieldCapable(DpsPriest.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.Alpha,
	};

	readonly specIndex = DpsPriest.specIndex;
	readonly specID = DpsPriest.specID;
	readonly classID = DpsPriest.classID;
	readonly friendlyName = DpsPriest.friendlyName;
	readonly simLink = DpsPriest.simLink;

	readonly isTankSpec = DpsPriest.isTankSpec;
	readonly isHealingSpec = DpsPriest.isHealingSpec;
	readonly isRangedDpsSpec = DpsPriest.isRangedDpsSpec;
	readonly isMeleeDpsSpec = DpsPriest.isMeleeDpsSpec;

	readonly canDualWield = DpsPriest.canDualWield;

	readonly launch = DpsPriest.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_shadow_shadowwordpain.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return DpsPriest.getIcon(size);
	};
}

export class HealerPriest extends PlayerSpec<Spec.SpecHealerPriest> {
	static specIndex = 1;
	static specID = Spec.SpecHealerPriest as Spec.SpecHealerPriest;
	static classID = Class.ClassPriest as Class.ClassPriest;
	static friendlyName = 'Healer';
	static simLink = getSpecSitePath('priest', 'healer');

	static isTankSpec = false;
	static isHealingSpec = true;
	static isRangedDpsSpec = false;
	static isMeleeDpsSpec = false;
	static canDualWield = isSpecDualWieldCapable(HealerPriest.specID);

	static launch: SimStatus = {
		phase: CURRENT_PHASE,
		status: LaunchStatus.GearPlanner,
	};

	readonly specIndex = HealerPriest.specIndex;
	readonly specID = HealerPriest.specID;
	readonly classID = HealerPriest.classID;
	readonly friendlyName = HealerPriest.friendlyName;
	readonly simLink = HealerPriest.simLink;

	readonly isTankSpec = HealerPriest.isTankSpec;
	readonly isHealingSpec = HealerPriest.isHealingSpec;
	readonly isRangedDpsSpec = HealerPriest.isRangedDpsSpec;
	readonly isMeleeDpsSpec = HealerPriest.isMeleeDpsSpec;

	readonly canDualWield = HealerPriest.canDualWield;

	readonly launch = HealerPriest.launch;

	static getIcon = (size: IconSize): string => {
		return `https://wow.zamimg.com/images/wow/icons/${size}/spell_holy_guardianspirit.jpg`;
	};

	getIcon = (size: IconSize): string => {
		return HealerPriest.getIcon(size);
	};
}
