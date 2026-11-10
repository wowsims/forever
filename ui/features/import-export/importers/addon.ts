import { APLRotation } from '@generated/proto/apl';
import { Class, EquipmentSpec, Profession, Race } from '@generated/proto/common';
import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { Database } from '@sim/proto/database';
import { fromJsonLenient } from '@sim/proto/lenient_json';
import { nameToClass, nameToProfession, nameToRace } from '@sim/proto/names';
import { toastManager } from '@ui-kit/Toast';

import { finishIndividualImport } from './finish_individual_import';
import type { ImporterDefinition } from './types';

function getWSEVersion(): Promise<string | null> {
	return fetch('https://api.github.com/repos/wowsims/exporter/releases/latest')
		.then(resp => {
			return resp.json().then(json => {
				return json.tag_name as string;
			});
		})
		.catch(_ => {
			return null;
		});
}

const WSE_VERSION = getWSEVersion();

export const ADDON_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.addon.title'),
	allowFileUpload: true,
	onImport: async (host, data) => {
		let importJson: Record<string, any>;
		try {
			importJson = JSON.parse(data);
		} catch {
			throw new Error(i18n.t('import.addon.error_invalid_export'));
		}
		if (typeof importJson != 'object' || importJson === null || Array.isArray(importJson)) {
			throw new Error(i18n.t('import.addon.error_invalid_export'));
		}

		// The class is required. Gear, race, professions, rotation and talents are optional, and anything left out stays as it is.
		const has = (key: string) => importJson[key] != null;
		const talentsStr = (importJson['talents'] as string) || '';
		// Only an export from the Forever client builds its talent string from these trees; Era/SoD
		// exports carry strings for other trees, so their talents don't count as something to import.
		const talentsMatchTrees = importJson['flavor'] === 'forever';
		const hasTalents = talentsMatchTrees && talentsStr != '' && talentsStr != '--';
		if (!['gear', 'race', 'professions', 'rotation'].some(has) && !hasTalents) {
			throw new Error(i18n.t('import.addon.error_nothing_to_import'));
		}

		const addonVersion = await WSE_VERSION;
		if (addonVersion && ((importJson['version'] as string) || '') != addonVersion) {
			toastManager.add({
				variant: 'warning',
				body: i18n.t('import.addon.outdated', { version: importJson['version'], latest: addonVersion }),
			});
		}

		// The sim only models max level, so anything below it is missing gear it cannot show.
		const level = importJson['level'];
		if (level < CHARACTER_LEVEL) {
			toastManager.add({
				variant: 'warning',
				body: i18n.t('import.addon.below_max_level', { level }),
				delay: 5000,
			});
		}

		const charClass = nameToClass((importJson['class'] as string) || '');
		if (charClass == Class.ClassUnknown) {
			throw new Error(i18n.t('import.errors.cannot_parse_class'));
		}

		let race: Race | undefined;
		if (has('race')) {
			race = nameToRace((importJson['race'] as string) || '');
			if (race == Race.RaceUnknown) {
				throw new Error(i18n.t('import.errors.cannot_parse_race'));
			}
		}

		let professions: Profession[] | undefined;
		if (has('professions')) {
			professions = (importJson['professions'] as Array<{ name: string; level: number }>).map(profData => nameToProfession(profData.name));
			professions.forEach((prof, i) => {
				if (prof == Profession.ProfessionUnknown) {
					throw new Error(i18n.t('import.errors.cannot_parse_profession', { profession: importJson['professions'][i] }));
				}
			});
		}

		let equipmentSpec: EquipmentSpec | undefined;
		if (has('gear')) {
			await Database.get();

			const gearJson = importJson['gear'];
			gearJson.items = (gearJson.items as Array<any>).filter(item => item != null);
			delete gearJson.version;

			equipmentSpec = EquipmentSpec.fromJson(gearJson, { ignoreUnknownFields: true });
		}

		// Same format as `player.rotation` in a JSON export.
		let rotation: APLRotation | undefined;
		if (has('rotation')) {
			try {
				rotation = fromJsonLenient(APLRotation, importJson['rotation']);
			} catch {
				throw new Error(i18n.t('import.addon.error_invalid_rotation'));
			}
		}

		await finishIndividualImport(host, {
			charClass,
			race,
			equipmentSpec,
			talentsStr,
			talentsMatchTrees,
			professions,
			rotation,
		});
	},
};
