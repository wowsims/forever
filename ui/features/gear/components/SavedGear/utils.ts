import type { SpecPresets } from '@features/hooks/useSavedPanel';
import { EquipmentSpec, UnitStats } from '@generated/proto/common';
import { SavedGearSet } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import type { SavedDataPanelEntry } from '@ui-kit/SavedDataPanel';

export const gearSetData = (player: Player<any>): SavedGearSet =>
	SavedGearSet.create({ gear: player.getGear().asSpec(), bonusStatsStats: player.getBonusStats().toProto() });

export const serializeGearSet = (data: SavedGearSet): string => JSON.stringify(SavedGearSet.toJson(data));

export const gearSetPresets = (presets: SpecPresets, { player, sim }: IndividualSimHost<any>): Array<SavedDataPanelEntry<SavedGearSet>> =>
	presets.gear.map(presetGear => {
		const data = SavedGearSet.create({ gear: sim.db.lookupEquipmentSpec(presetGear.gear).asSpec(), bonusStatsStats: new Stats().toProto() });
		return {
			name: presetGear.name,
			data,
			json: serializeGearSet(data),
			tooltip: presetGear.tooltip,
			isPreset: true,
			disabled: !!presetGear.enableWhen && !presetGear.enableWhen(player),
			afterLoad: () => presetGear.onLoad?.(player),
		};
	});

export const loadGearSet = ({ data }: SavedDataPanelEntry<SavedGearSet>, { player, sim }: IndividualSimHost<any>) =>
	batch(() => {
		player.setGear(sim.db.lookupEquipmentSpec(data.gear || EquipmentSpec.create()));
		player.setBonusStats(Stats.fromProto(data.bonusStatsStats || UnitStats.create()));
	});
