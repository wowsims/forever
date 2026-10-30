import { Encounter as EncounterProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Encounter } from '@sim/raid/encounter';

// The name as given when no other entry has it, else the first of "name (2)", "name (3)", … that is free.
export const uniqueEncounterName = (name: string, taken: ReadonlyArray<string>): string => {
	if (!taken.includes(name)) return name;
	let number = 2;
	while (taken.includes(i18n.t('settings_tab.encounter.unique_name', { name, number }))) number++;
	return i18n.t('settings_tab.encounter.unique_name', { name, number });
};

export const newEncounterProto = (): EncounterProto =>
	EncounterProto.create({ ...Encounter.defaultEncounterProto(), name: i18n.t('settings_tab.encounter.new_encounter') });
