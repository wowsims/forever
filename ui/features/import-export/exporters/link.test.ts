import { Player } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { ConsumesSpec, EquipmentSpec, Race } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { tryParseUrlLocation } from '@sim/state/sim_links';
import { fakeHost } from '@testing/sim/fake_host';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { senderStatFilters } from '@testing/sim/stat_filters';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { createLink } from './link';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

// The Fury warrior's Pre-BiS preset, written out so that a preset edit cannot move the link below.
const FURY_SETTINGS = IndividualSimSettings.create({
	apiVersion: 3,
	player: Player.create({
		name: 'Player',
		race: Race.RaceOrc,
		talentsString: '30305213002-15053105102210501',
		equipment: EquipmentSpec.fromJson({
			items: [
				{ id: 12640, enchant: 2543 },
				{ id: 19491 },
				{ id: 10263, randomSuffix: 12727 },
				{ id: 18511, enchant: 7667 },
				{ id: 252483, enchant: 1891 },
				{ id: 10147, randomSuffix: 12727, enchant: 7656 },
				{ id: 21998, enchant: 8206 },
				{ id: 20216 },
				{ id: 22385, enchant: 2543 },
				{ id: 21995, enchant: 1887 },
				{ id: 272408 },
				{ id: 272409 },
				{ id: 249469 },
				{ id: 272437 },
				{ id: 22384, enchant: 1900 },
				{ id: 12798, enchant: 1900 },
				{ id: 22347 },
			],
		}),
		consumables: ConsumesSpec.create({ elixirIds: [9187, 13452], mhImbueId: 12643, ohImbueId: 12404 }),
		rotation: APLRotation.fromJson({
			type: 'TypeAPL',
			priorityList: [{ action: { castSpell: { spellId: { spellId: 23894 } } } }, { action: { castSpell: { spellId: { spellId: 1680 } } } }],
		}),
	}),
});

// FURY_SETTINGS as pako, the deflate library the site shipped before, wrote it: the form every older shared link has.
const OLDER_FURY_HASH =
	'#eJyrYJa6wziLmS0gJ7EytYiDTaqZkYuN40GSxHthLhaOxTNAvOkBAtuTudg5zk9glPhsDWQc3sIv8ZiPi5NjsT9QRuIFSOzdakaJPgegnh9zGYHcj+sYQUawc7wGit/nA4rPuCAAJGeCyb9z+IHkViCbneMDUOUbPqA1/5JBNAvH6XWMSrwBj5MiviSsYmR57N6TuYtJ1tjA2MDUyNDYwMBI19DUwNTY0MDU0MDICEgZGD5ikpRgFuKS4pBiA+q/totRiFOKXYqVi5ljAi8AMTE8EA==';

describe('createLink', () => {
	it('writes a link that reads back as the settings it was made from', async () => {
		const url = new URL(await createLink(fakeHost({ toProto: () => FURY_SETTINGS })));
		const parsed = (await tryParseUrlLocation(url))!;

		expect(IndividualSimSettings.equals(parsed.settings, FURY_SETTINGS)).toBe(true);
		expect(url.searchParams.has('i')).toBe(false);
	});

	it('leaves the stat filters out of a link that carries the other filters', async () => {
		const { sim, host } = await makeIndividualSim();
		sim.setFilters({ ...sim.getFilters(), minIlvl: 50, statFilters: senderStatFilters });

		const url = new URL(await createLink(host, [SimSettingCategories.Gear, SimSettingCategories.UISettings]));
		const filters = (await tryParseUrlLocation(url))!.settings.settings!.filters!;

		expect(filters.statFilters).toBeUndefined();
		expect(filters.minIlvl).toBe(50);
		expect(sim.getFilters().statFilters).toBeDefined();
	});
});

describe('tryParseUrlLocation', () => {
	it('reads a link the earlier zlib library wrote', async () => {
		const parsed = (await tryParseUrlLocation({ hash: OLDER_FURY_HASH, search: '' }))!;

		expect(IndividualSimSettings.equals(parsed.settings, FURY_SETTINGS)).toBe(true);
		expect(parsed.settings.player!.race).toBe(Race.RaceOrc);
		expect(parsed.settings.player!.equipment!.items).toHaveLength(17);
	});
});
