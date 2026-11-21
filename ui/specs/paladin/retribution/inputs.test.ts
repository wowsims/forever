// The Unsealed Wrath twist picker and the Echo stack dropdowns it shows, on a player with the Unsealed Wrath talents.
import { Spec } from '@generated/proto/common';
import { PaladinSealTwist } from '@generated/proto/paladin';
import type { Player } from '@sim/player/player';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { RotationIconInputs, RotationInputs } from './inputs';
import { UnsealedWrathTalents } from './presets';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => {
	serveDatabase();
});

type RetPlayer = Player<Spec.SpecRetributionPaladin>;
type Input = {
	id: string;
	showWhen?: (player: RetPlayer) => boolean;
	getValue: (player: RetPlayer) => number;
	setValue: (player: RetPlayer, value: number) => void;
};

const input = (id: string) => [...RotationIconInputs, ...RotationInputs.inputs].find(config => config.id === id) as unknown as Input;

const unsealedWrathPlayer = async () => {
	const player = (await makeIndividualSim({ spec: Spec.SpecRetributionPaladin })).player as unknown as RetPlayer;
	player.setTalentsString(UnsealedWrathTalents.data.talentsString);
	const rotation = player.getSimpleRotation();
	rotation.uwTwist = PaladinSealTwist.SealTwistUnknown;
	rotation.uwSocStacks = 0;
	rotation.uwSofStacks = 0;
	rotation.uwSorStacks = 0;
	player.setSimpleRotation(rotation);
	return player;
};

const shown = (player: RetPlayer) => ['uwSocStacks', 'uwSofStacks', 'uwSorStacks'].filter(id => input(id).showWhen!(player));

describe('the Unsealed Wrath twist picker', () => {
	it("twists Seal of Command until Seal of Fury is picked, and shows that twist's Echo stacks", async () => {
		const player = await unsealedWrathPlayer();
		expect(input('uwTwist').showWhen!(player)).toBe(true);
		expect(input('uwTwist').getValue(player)).toBe(PaladinSealTwist.SealTwistCommand);
		expect(shown(player)).toEqual(['uwSocStacks', 'uwSorStacks']);
		expect(input('uwSocStacks').getValue(player)).toBe(3);
		expect(input('uwSorStacks').getValue(player)).toBe(1);

		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);

		expect(input('uwTwist').getValue(player)).toBe(PaladinSealTwist.SealTwistFury);
		expect(shown(player)).toEqual(['uwSofStacks', 'uwSorStacks']);
	});

	it("resets the Echo stacks to the picked twist's defaults", async () => {
		const player = await unsealedWrathPlayer();
		input('uwSorStacks').setValue(player, 2);

		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);
		expect(player.getSimpleRotation()).toMatchObject({ uwSofStacks: 1, uwSorStacks: 3 });

		input('uwSofStacks').setValue(player, 2);
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistCommand);
		expect(player.getSimpleRotation()).toMatchObject({ uwSocStacks: 3, uwSorStacks: 1, uwSofStacks: 2 });
	});

	it('keeps the Echo stacks when the twist already running is picked again', async () => {
		const player = await unsealedWrathPlayer();
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);
		input('uwSofStacks').setValue(player, 3);
		input('uwSorStacks').setValue(player, 2);

		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);

		expect(player.getSimpleRotation()).toMatchObject({ uwSofStacks: 3, uwSorStacks: 2 });
	});
});
