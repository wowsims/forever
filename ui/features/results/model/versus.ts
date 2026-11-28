import i18n from '@i18n/config';
import { translateCreatureFamily, translateMobType } from '@i18n/localization';
import type { VersusLog } from '@sim/proto/combat_log';

/** The target kind a hit took a versus bonus against: "Humanoid", or "Humanoid (Orc)" with a creature family. */
export const versusTargetName = (versus: VersusLog): string =>
	versus.creatureFamily === undefined
		? translateMobType(versus.mobType)
		: i18n.t('results_tab.details.timeline.versus.tag_family', {
				target: translateMobType(versus.mobType),
				family: translateCreatureFamily(versus.creatureFamily),
			});
