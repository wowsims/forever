import { APLRotation } from '@generated/proto/apl';
import { Class, EquipmentSpec, Profession, Race } from '@generated/proto/common';
import i18n from '@i18n/config';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import { Database } from '@sim/proto/database';
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

// A section counts as sent when the export has the key at all. Whatever it leaves out stays as it is in the sim.
const has = (json: Record<string, any>, key: string) => json[key] !== undefined && json[key] !== null;

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

		if (!has(importJson, 'gear') && !has(importJson, 'race') && !has(importJson, 'professions') && !has(importJson, 'rotation')) {
			throw new Error(i18n.t('import.addon.error_nothing_to_import'));
		}

		// Hand-assembled exports, such as a rotation on its own, carry no addon version.
		const addonVersion = await WSE_VERSION;
		if (addonVersion && has(importJson, 'version') && importJson['version'] != addonVersion) {
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

		let charClass: Class | undefined;
		if (has(importJson, 'class')) {
			charClass = nameToClass((importJson['class'] as string) || '');
			if (charClass == Class.ClassUnknown) {
				throw new Error(i18n.t('import.errors.cannot_parse_class'));
			}
		}

		let race: Race | undefined;
		if (has(importJson, 'race')) {
			race = nameToRace((importJson['race'] as string) || '');
			if (race == Race.RaceUnknown) {
				throw new Error(i18n.t('import.errors.cannot_parse_race'));
			}
		}

		let professions: Profession[] | undefined;
		if (has(importJson, 'professions')) {
			professions = (importJson['professions'] as Array<{ name: string; level: number }>).map(profData => nameToProfession(profData.name));
			professions.forEach((prof, i) => {
				if (prof == Profession.ProfessionUnknown) {
					throw new Error(i18n.t('import.errors.cannot_parse_profession', { profession: importJson['professions'][i] }));
				}
			});
		}

		const talentsStr = (importJson['talents'] as string) || '';

		let equipmentSpec: EquipmentSpec | undefined;
		if (has(importJson, 'gear')) {
			await Database.get();

			const gearJson = importJson['gear'];
			gearJson.items = (gearJson.items as Array<any>).filter(item => item != null);
			delete gearJson.version;

			equipmentSpec = EquipmentSpec.fromJson(gearJson, { ignoreUnknownFields: true });
		}

		// The same APLRotation JSON a settings export holds under `player.rotation`.
		let rotation: APLRotation | undefined;
		if (has(importJson, 'rotation')) {
			try {
				rotation = APLRotation.fromJson(importJson['rotation'], { ignoreUnknownFields: true });
			} catch {
				throw new Error(i18n.t('import.addon.error_invalid_rotation'));
			}
		}

		await finishIndividualImport(host, {
			charClass,
			race,
			equipmentSpec,
			talentsStr,
			professions,
			rotation,
		});
	},
};
