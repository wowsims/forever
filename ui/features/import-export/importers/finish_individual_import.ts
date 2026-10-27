import type { Class, EquipmentSpec, Profession, Race } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Database } from '@sim/proto/database';
import { classNames } from '@sim/proto/names';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { toastManager } from '@ui-kit/Toast';

export interface IndividualImport {
	charClass: Class;
	race: Race;
	equipmentSpec: EquipmentSpec;
	talentsStr: string;
	// The talent string was built against this sim's trees (the Armory import), so it is applied.
	talentsForTheseTrees?: boolean;
	professions: Profession[];
	missingEnchants?: number[];
	missingItems?: number[];
}

export const finishIndividualImport = async (
	host: IndividualSimHost<any>,
	{ charClass, race, equipmentSpec, talentsStr, talentsForTheseTrees = false, professions, missingEnchants = [], missingItems = [] }: IndividualImport,
): Promise<void> => {
	if (charClass != host.player.getClass()) {
		throw new Error(i18n.t('import.errors.wrong_class', { expected: host.player.getPlayerClass().friendlyName, found: classNames.get(charClass) }));
	}

	await Database.loadLeftoversIfNecessary(equipmentSpec);

	const gear = host.sim.db.lookupEquipmentSpec(equipmentSpec);

	// A talent string is positional: digit i of tree t is talents[i] of whatever tree the exporter
	// was built against. Forever's trees are a different set, so applying another exporter's string
	// would spend the points on other talents rather than fail. Only a string built against these
	// trees is applied.
	const importedTalents = !!talentsStr && talentsStr != '--';

	batch(() => {
		host.player.setRace(race);
		host.player.setGear(gear);
		if (talentsForTheseTrees) {
			host.player.setTalentsString(talentsStr);
		}
		if (professions.length > 0) {
			host.player.setProfessions(professions);
		}
	});

	if (importedTalents && !talentsForTheseTrees) {
		toastManager.add({
			variant: 'warning',
			body: i18n.t('import.talents_not_imported'),
			delay: 8000,
		});
	}

	if (missingItems.length == 0 && missingEnchants.length == 0) {
		toastManager.add({ variant: 'success', body: i18n.t('import.success') });
	} else {
		toastManager.add({
			variant: 'info',
			body:
				i18n.t('import.missing_ids.title') +
				(missingItems.length == 0 ? '' : `\n\n${i18n.t('import.missing_ids.items')} ` + missingItems.join(', ')) +
				(missingEnchants.length == 0 ? '' : `\n\n${i18n.t('import.missing_ids.enchants')} ` + missingEnchants.join(', ')),
		});
	}
};
