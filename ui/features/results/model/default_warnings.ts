import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import { professionNames } from '@sim/proto/names';
import { getTalentPoints } from '@sim/proto/utils';
import type { SimWarning } from '@sim/sim_host';
import { subscribeAll, subscribePlayerField, subscribeRaidField } from '@sim/state/subscriptions';

// Talents are three free-form trees, so there is no per-row requirement to check: the only
// thing the string can be wrong about is the point total.
const MAX_TALENT_POINTS = CHARACTER_LEVEL - 9;

export function defaultSimWarnings<SpecType extends number>(player: Player<SpecType>): Array<SimWarning> {
	return [
		{
			updateOn: subscribeAll([
				subscribePlayerField(player, 'gear'),
				subscribePlayerField(player, 'profession1'),
				subscribePlayerField(player, 'profession2'),
			]),
			getContent: () => {
				const failedProfReqs = player.getGear().getFailedProfessionRequirements(player.getProfessions());
				if (failedProfReqs.length == 0) {
					return '';
				}

				return failedProfReqs.map(fpr =>
					i18n.t('sidebar.warnings.profession_requirement', {
						itemName: fpr.name,
						professionName: professionNames.get(fpr.requiredProfession)!,
					}),
				);
			},
		},
		{
			updateOn: subscribePlayerField(player, 'talentsString'),
			getContent: () => {
				const talentPoints = getTalentPoints(player.getTalentsString());

				// Skipped during initial load, when the string is still empty.
				if (talentPoints == 0 || talentPoints >= MAX_TALENT_POINTS) {
					return '';
				}

				return i18n.t('sidebar.warnings.unspent_talent_points', {
					count: MAX_TALENT_POINTS - talentPoints,
				});
			},
		},
		{
			updateOn: subscribeRaidField(player.sim.raid, 'partyBuffs'),
			getContent: () => {
				const buffs = player.getParty()?.getBuffs();
				if (!buffs?.windfuryTotem) {
					return '';
				}

				const warnings: Array<string> = [];
				if (buffs.graceOfAirTotem) {
					warnings.push(i18n.t('sidebar.warnings.windfury_over_grace_of_air'));
				}
				if (buffs.flametongueTotem) {
					warnings.push(i18n.t('sidebar.warnings.windfury_over_flametongue'));
				}
				return warnings;
			},
		},
	];
}
