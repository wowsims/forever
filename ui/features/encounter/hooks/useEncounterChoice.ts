import type { ConnectedChoiceFieldProps } from '@features/inputs/configs';
import { Encounter as EncounterProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Encounter } from '@sim/raid/encounter';
import { subscribeEncounterChange } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { encounterChoiceConfig, encounterChoices } from '../components/EncounterPicker/utils/configs';
import { useMyEncounters } from './useMyEncounters';

type EncounterChoiceProps = Pick<ConnectedChoiceFieldProps<Encounter, string>, 'modObject' | 'config' | 'options' | 'placeholder' | 'renderValue'>;

// The encounter choice the picker and the Advanced dialog share: Custom, the presets and My encounters. An entry
// reads as edited while the encounter differs from what the list holds for it, and an import that carries the
// name of an entry reads as imported.
export const useEncounterChoice = (encounter: Encounter, field: Parameters<typeof encounterChoiceConfig>[2]): EncounterChoiceProps => {
	const { entries: own } = useMyEncounters();
	const config = useMemo(() => encounterChoiceConfig(encounter, own, field), [encounter, own, field]);
	const options = useMemo(() => encounterChoices(encounter.sim.db.getAllPresetEncounters(), own), [encounter, own]);

	const current = useStoreSubscribe(subscribeEncounterChange(encounter), () => ({
		name: encounter.getName(),
		mine: encounter.isMyEncounter(),
		// The list holds its entries without the mark, so the comparison leaves it out too.
		json: JSON.stringify(EncounterProto.toJson({ ...encounter.toProto(), myEncounter: false })),
	}));
	const entry = own.find(entry => entry.name === current.name);
	const edited = current.mine && !!entry && entry.json !== current.json;
	const imported = !current.mine && !!entry;

	return {
		modObject: encounter,
		config,
		options,
		placeholder: imported ? i18n.t('settings_tab.encounter.encounter_preset.imported', { name: current.name }) : current.name,
		renderValue: edited ? () => i18n.t('settings_tab.encounter.encounter_preset.edited', { name: current.name }) : undefined,
	};
};
