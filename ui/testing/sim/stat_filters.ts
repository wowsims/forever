import { Player as PlayerProto } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { type IndividualSettingsHost, loadIndividualSettings } from '@sim/state/persistence';

export const STORAGE_KEY = 'test__currentSettings__';

export const selectionOf = (...stats: Stat[]) => StatSelection.create({ stats: stats.map(stat => UnitStat.fromStat(stat).toProto()) });

export const viewerStatFilters = StatFilters.create({ characterStats: selectionOf(Stat.StatIntellect), consumables: selectionOf(Stat.StatStamina) });
export const senderStatFilters = StatFilters.create({ ep: selectionOf(Stat.StatStrength), buffs: selectionOf() });

// What a viewer has set on their filters besides the stat filters: a hand filter off and a favourite.
export const viewerFiltersWith = (statFilters: StatFilters | undefined): Partial<DatabaseFilters> => ({
	twoHandedWeapons: false,
	favoriteItems: [19019],
	statFilters,
});

export const envelope = (settings?: SimSettings) => IndividualSimSettings.create({ apiVersion: 3, player: PlayerProto.create(), settings });

// Loads the settings as the page does at start: the autosave under STORAGE_KEY, then the link the env's location carries.
export const loadSettings = ({ host, player }: { host: IndividualSettingsHost; player: Player<any> }) =>
	loadIndividualSettings(host, {
		storageKey: STORAGE_KEY,
		player,
		autosaveSubscribe: () => () => {},
		statWeightSettings: { load: () => {} } as unknown as StatWeightActionSettings,
		statSweepSettings: { load: () => {} } as unknown as StatSweepSettings,
	});
