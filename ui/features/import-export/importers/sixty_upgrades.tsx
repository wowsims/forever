import { Class, EquipmentSpec, ItemSpec, Race } from '@generated/proto/common';
import { getEligibleItemSlots } from '@sim/proto/items';
import { nameToClass, nameToRace } from '@sim/proto/names';
import { toastManager } from '@ui-kit/Toast';

import { finishIndividualImport } from './finish_individual_import';
import type { ImporterDefinition } from './types';

const removedSuffixesBody = (itemNames: string[]) => (
	<>
		<p>Sixty Upgrades currently exports the wrong Random Suffixes. We have removed the random suffix on the following item(s):</p>
		<ul>
			{itemNames.map((itemName, index) => (
				<li key={index}>
					<strong>{itemName}</strong>
				</li>
			))}
		</ul>
	</>
);

export const SIXTY_UPGRADES_IMPORTER: ImporterDefinition = {
	title: 'Sixty Upgrades Import',
	allowFileUpload: true,
	onImport: async (host, data) => {
		let importJson: any | null;
		try {
			importJson = JSON.parse(data);
		} catch {
			throw new Error('Please use a valid Sixty Upgrades export.');
		}

		const missingItems: number[] = [];
		const missingEnchants: number[] = [];

		const charClass = nameToClass((importJson?.character?.gameClass as string) || '');
		if (charClass == Class.ClassUnknown) {
			throw new Error('Could not parse Class!');
		}

		const race = nameToRace((importJson?.character?.race as string) || '');
		if (race == Race.RaceUnknown) {
			throw new Error('Could not parse Race!');
		}

		// Sixty Upgrades names a talent by the spell id of the rank taken. A Forever talent reports
		// one id for all of its ranks, so the export does not say how many points went in and every
		// talent would import as 1/N.
		const droppedTalents = (importJson?.talents as any[] | undefined)?.length ?? 0;

		let hasRemovedRandomSuffix = false;
		const modifiedItemNames: string[] = [];
		const equipmentSpec = EquipmentSpec.create();
		(importJson.items as any[]).forEach(itemJson => {
			const itemSpec = ItemSpec.create();
			itemSpec.id = itemJson.id;
			const dbItem = host.sim.db.getItemById(itemSpec.id);

			if (!dbItem) {
				missingItems.push(itemSpec.id);
				return;
			}

			if (itemJson.enchant?.id) {
				itemSpec.enchant = itemJson.enchant.id;
				const slots = getEligibleItemSlots(dbItem);
				const enchant = slots.flatMap(slot => host.sim.db.getEnchants(slot)).some(enchant => enchant.effectId == itemSpec.enchant);
				if (!enchant) {
					missingEnchants.push(itemSpec.enchant);
					return;
				}
			}

			// As long as 60U exports the wrong suffixes we should inform the user that they need to manually add them.
			if (itemJson.suffixId) {
				hasRemovedRandomSuffix = true;
				modifiedItemNames.push(itemJson.name);
			}
			equipmentSpec.items.push(itemSpec);
		});

		await finishIndividualImport(host, {
			charClass,
			race,
			equipmentSpec,
			talentsStr: '',
			professions: [],
			missingEnchants,
			missingItems,
		});

		if (hasRemovedRandomSuffix && modifiedItemNames.length) {
			toastManager.add({ variant: 'warning', body: removedSuffixesBody(modifiedItemNames), delay: 8000 });
		}

		if (droppedTalents > 0) {
			toastManager.add({
				variant: 'warning',
				body: 'Sixty Upgrades exports do not carry talent ranks, so talents were not imported. Set them on the Talents tab.',
				delay: 8000,
			});
		}
	},
};
