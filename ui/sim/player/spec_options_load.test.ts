import { Player as PlayerProto } from '@generated/proto/api';
import { Spec } from '@generated/proto/common';
import { ProtectionWarrior_Options, WarriorOptions, WarriorStance } from '@generated/proto/warrior';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

const OWN = WarriorOptions.create({ useBattleShout: true, defaultStance: WarriorStance.WarriorStanceDefensive, queueDelay: 250 });

const settingsWith = (options: ProtectionWarrior_Options) => PlayerProto.create({ spec: { oneofKind: 'protectionWarrior', protectionWarrior: { options } } });

describe('Player.fromProto spec options', () => {
	const setup = async () => {
		const { player } = await makeIndividualSim({ spec: Spec.SpecProtectionWarrior });
		player.setSpecOptions(ProtectionWarrior_Options.create({ classOptions: OWN }));
		return player;
	};

	it('keeps the class options when the settings carry none', async () => {
		const player = await setup();
		player.fromProto(settingsWith(ProtectionWarrior_Options.create()), [SimSettingCategories.Miscellaneous]);
		expect(player.getClassOptions()).toEqual(OWN);
	});

	it('takes the class options the settings carry', async () => {
		const player = await setup();
		const loaded = WarriorOptions.create({ useBattleShout: false, defaultStance: WarriorStance.WarriorStanceBattle });
		player.fromProto(settingsWith(ProtectionWarrior_Options.create({ classOptions: loaded })), [SimSettingCategories.Miscellaneous]);
		expect(player.getClassOptions()).toEqual(loaded);
	});
});
