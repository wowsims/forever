import { createLink } from '@features/import-export/exporters/link';
import { ConsumesPicker } from '@features/settings/components/ConsumesPicker';
import { StatOptionIcons } from '@features/settings/components/StatOptionIcons';
import { BUFFS_CONFIG, PARTY_BUFFS_CONFIG } from '@features/settings/model/buffs_debuffs';
import { relevantStatOptions, statOptionScope } from '@features/settings/model/stat_options';
import { shownStats } from '@features/stat-filters/model/stat_filters';
import { ConsumableType, ConsumesSpec, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { IndividualSimSettings, StatFilters, UIDatabase } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { ActionId } from '@sim/proto/action_id';
import { getEnumValues } from '@sim/utils/collections';
import { renderWithHost } from '@testing/react';
import { filled } from '@testing/sim/hooks';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { memoryEnv } from '@testing/sim/memory_env';
import { loadSettings, selectionOf, STORAGE_KEY } from '@testing/sim/stat_filters';
import { waitFor } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import warrior from './warrior/dps/spec';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// The icon of the Arcane Brilliance input: its first rank, Arcane Intellect.
const ARCANE_BRILLIANCE_ICON = 10157;
const ELIXIR_OF_GREATER_AGILITY = Consumable.create({
	id: 9187,
	name: 'Elixir of Greater Agility',
	type: ConsumableType.ConsumableTypeElixir,
	lvl: 38,
	stats: Object.assign(new Array<number>(42).fill(0), { [Stat.StatAgility]: 25 }),
	groupKey: '29/1',
	groupName: 'Agility',
});

// Strength only, so the Intellect raid buff and the Agility elixir are both hidden.
const viewerStatFilters = StatFilters.create({ buffs: selectionOf(Stat.StatStrength), consumables: selectionOf(Stat.StatStrength) });

const warriorSim = (env = memoryEnv()) => makeIndividualSim({ spec: warrior.spec, config: warrior, env });

beforeAll(() => serveDatabase(UIDatabase.create({ consumables: [ELIXIR_OF_GREATER_AGILITY] })));

describe('a shared build', () => {
	it("arrives picked where the viewer's stat filters hide it, offered as kept, and runs", async () => {
		const sender = await warriorSim();
		sender.sim.raid.setBuffs({ ...sender.sim.raid.getBuffs(), arcaneBrilliance: 1 });
		sender.player.setConsumes(ConsumesSpec.create({ elixirIds: [ELIXIR_OF_GREATER_AGILITY.id] }));
		sender.sim.setFilters({ ...sender.sim.getFilters(), statFilters: StatFilters.create({ buffs: selectionOf(Stat.StatIntellect) }) });
		const link = new URL(createLink(sender.host, getEnumValues(SimSettingCategories) as Array<SimSettingCategories>));

		const autosave = await warriorSim();
		autosave.sim.setFilters({ ...autosave.sim.getFilters(), statFilters: viewerStatFilters });
		const env = memoryEnv({ hash: link.hash, search: link.search });
		env.storage.setItem(STORAGE_KEY, IndividualSimSettings.toJsonString(autosave.host.toProto()));
		const viewer = await warriorSim(env);
		loadSettings(viewer);

		expect(StatFilters.equals(viewer.sim.getFilters().statFilters!, viewerStatFilters)).toBe(true);
		const buffStats = shownStats(viewer.sim.getFilters(), 'buffs', statOptionScope([...BUFFS_CONFIG, ...PARTY_BUFFS_CONFIG], warrior));
		const shown = new Set(buffStats.map(stat => stat.getStat()));
		expect(shown.has(Stat.StatIntellect)).toBe(false);

		vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
			return Promise.resolve(filled(this, String(this.spellId || this.itemId), 'icon.jpg'));
		});
		const { container } = renderWithHost(
			<>
				<StatOptionIcons options={relevantStatOptions(BUFFS_CONFIG, warrior, shown)} />
				<ConsumesPicker />
			</>,
			viewer.host as unknown as Record<string, unknown>,
		);
		const swatch = (href: string) =>
			waitFor(() => {
				const anchor = container.querySelector(`a[href="${href}"]`);
				expect(anchor).not.toBeNull();
				return anchor!;
			});

		const buff = await swatch(ActionId.makeSpellUrl(ARCANE_BRILLIANCE_ICON));
		expect(buff.hasAttribute('data-active')).toBe(true);
		expect(buff.parentElement!.hasAttribute('data-kept')).toBe(true);
		const elixir = await swatch(ActionId.makeItemUrl(ELIXIR_OF_GREATER_AGILITY.id));
		expect(elixir.parentElement!.hasAttribute('data-kept')).toBe(true);

		const request = viewer.sim.makeRaidSimRequest();
		expect(request.raid!.buffs!.arcaneBrilliance).toBe(1);
		expect(request.raid!.parties[0].players[0].consumables!.elixirIds).toEqual([ELIXIR_OF_GREATER_AGILITY.id]);
	});
});
