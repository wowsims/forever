import type { APLRotation } from '@generated/proto/apl';
import type { Class, EquipmentSpec, Profession, Race } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { Database } from '@sim/proto/database';
import { classNames } from '@sim/proto/names';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { toastManager } from '@ui-kit/Toast';

// Only the class is required. Anything left out stays as it is in the sim.
export interface IndividualImport {
	charClass: Class;
	race?: Race;
	equipmentSpec?: EquipmentSpec;
	talentsStr?: string;
	// The exporter built talentsStr from this sim's trees, so it is safe to apply.
	talentsMatchTrees?: boolean;
	professions?: Profession[];
	rotation?: APLRotation;
	missingEnchants?: number[];
	missingItems?: number[];
}

export const finishIndividualImport = async (
	host: IndividualSimHost<any>,
	{
		charClass,
		race,
		equipmentSpec,
		talentsStr = '',
		talentsMatchTrees = false,
		professions = [],
		rotation,
		missingEnchants = [],
		missingItems = [],
	}: IndividualImport,
): Promise<void> => {
	if (charClass != host.player.getClass()) {
		throw new Error(i18n.t('import.errors.wrong_class', { expected: host.player.getPlayerClass().friendlyName, found: classNames.get(charClass) }));
	}

	if (equipmentSpec) {
		await Database.loadLeftoversIfNecessary(equipmentSpec);
	}
	const gear = equipmentSpec && host.sim.db.lookupEquipmentSpec(equipmentSpec);

	// A talent string is positional: digit i of tree t is talents[i] of whatever tree the exporter
	// was built against. Forever's trees are a different set, so applying a string built for other
	// trees would spend the points on other talents rather than fail. Only strings the exporter
	// marks as built from these trees are applied.
	const importedTalents = !!talentsStr && talentsStr != '--';
	const applyTalents = importedTalents && talentsMatchTrees;

	batch(() => {
		if (race !== undefined) {
			host.player.setRace(race);
		}
		if (gear) {
			host.player.setGear(gear);
		}
		if (applyTalents) {
			host.player.setTalentsString(talentsStr);
		}
		if (professions.length > 0) {
			host.player.setProfessions(professions);
		}
	});

	if (rotation) {
		host.fromProto(IndividualSimSettings.create({ player: { rotation } }), [SimSettingCategories.Rotation]);
	}

	if (importedTalents && !applyTalents) {
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
