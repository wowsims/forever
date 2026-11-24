// The seal picker, the Unsealed Wrath twist picker and the Echo stack dropdowns it shows, on a player with the
// Unsealed Wrath talents.
import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedInput } from '@features/inputs/components/ConnectedInput';
import { Spec } from '@generated/proto/common';
import { PaladinAura, PaladinSeal, PaladinSealTwist } from '@generated/proto/paladin';
import type { Player } from '@sim/player/player';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { act, screen } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { RotationIconInputs, RotationInputs, SEAL_SPELLS } from './inputs';
import { HolyShockTalents, UnsealedWrathTalents } from './presets';

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

// A new player's simple rotation holds no twist and no Echo stacks, as settings saved without those fields do.
const unsealedWrathPlayer = async () => {
	const made = await makeIndividualSim({ spec: Spec.SpecRetributionPaladin });
	const player = made.player as unknown as RetPlayer;
	player.setTalentsString(UnsealedWrathTalents.data.talentsString);
	return { made, player };
};

// Unsealed Wrath is the last talent of the last tree.
const withoutUnsealedWrath = UnsealedWrathTalents.data.talentsString.slice(0, -1);

// The Unsealed Wrath talents with the points of Seal of Command, and of Vengeance behind it, in Deflection.
const withoutSealOfCommand = '050250002-501-45233010201301051';

// A rotation change, which is when a hidden icon picker writes its zero value unless it keeps its value.
const changeAura = (player: RetPlayer) => {
	const rotation = player.getSimpleRotation();
	rotation.aura = PaladinAura.DevotionAura;
	player.setSimpleRotation(rotation);
};

// The link of the spell the mounted picker shows as picked, empty when it shows none.
const shownSpell = () => screen.getByTestId('icon-enum-picker-root').getAttribute('href') ?? '';
const spellOf = (seal: PaladinSeal) => `spell=${SEAL_SPELLS[seal]!.spellId}`;

const shown = (player: RetPlayer) => ['uwSocStacks', 'uwSofStacks', 'uwSorStacks'].filter(id => input(id).showWhen!(player));

describe('the Unsealed Wrath twist picker', () => {
	it("twists Seal of Command until Seal of Fury is picked, and shows that twist's Echo stacks", async () => {
		const { player } = await unsealedWrathPlayer();
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
		const { player } = await unsealedWrathPlayer();
		input('uwSorStacks').setValue(player, 2);

		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);
		expect(player.getSimpleRotation()).toMatchObject({ uwSofStacks: 1, uwSorStacks: 3 });

		input('uwSofStacks').setValue(player, 2);
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistCommand);
		expect(player.getSimpleRotation()).toMatchObject({ uwSocStacks: 3, uwSorStacks: 1, uwSofStacks: 2 });
	});

	it('keeps the Echo stacks when the twist already running is picked again', async () => {
		const { player } = await unsealedWrathPlayer();
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);
		input('uwSofStacks').setValue(player, 3);
		input('uwSorStacks').setValue(player, 2);

		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);

		expect(player.getSimpleRotation()).toMatchObject({ uwSofStacks: 3, uwSorStacks: 2 });
	});

	it('starts the twist that follows no twist from its own Echo stacks', async () => {
		const { player } = await unsealedWrathPlayer();
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury);
		player.setTalentsString(withoutSealOfCommand);
		expect(player.getTalents()).toMatchObject({ unsealedWrath: true, sealOfCommand: false });
		input('uwTwist').setValue(player, PaladinSealTwist.SealTwistUnknown);

		player.setTalentsString(UnsealedWrathTalents.data.talentsString);

		expect(input('uwTwist').getValue(player)).toBe(PaladinSealTwist.SealTwistCommand);
		expect(input('uwSorStacks').getValue(player)).toBe(1);
	});

	it('shows the Seal of Command twist once Seal of Command is talented', async () => {
		const { made, player } = await unsealedWrathPlayer();
		act(() => player.setTalentsString(withoutSealOfCommand));
		renderWithHost(<ConnectedIconEnumPicker modObject={player} config={input('uwTwist') as never} />, made);
		await act(async () => {});
		expect(shownSpell()).not.toContain('spell=');

		await act(async () => player.setTalentsString(UnsealedWrathTalents.data.talentsString));

		expect(shownSpell()).toContain(spellOf(PaladinSeal.SealOfCommand));
	});

	it('hides the Echo stack dropdowns when Unsealed Wrath is untalented and shows them when it is back', async () => {
		const { made, player } = await unsealedWrathPlayer();
		renderWithHost(
			<>
				{['uwSocStacks', 'uwSofStacks', 'uwSorStacks'].map(id => (
					<ConnectedInput key={id} config={input(id) as never} inline />
				))}
			</>,
			made,
		);
		const mounted = () => ['uwSocStacks', 'uwSofStacks', 'uwSorStacks'].filter(id => document.getElementById(id));
		expect(mounted()).toEqual(['uwSocStacks', 'uwSorStacks']);

		act(() => player.setTalentsString(withoutUnsealedWrath));
		expect(mounted()).toEqual([]);

		act(() => player.setTalentsString(UnsealedWrathTalents.data.talentsString));
		expect(mounted()).toEqual(['uwSocStacks', 'uwSorStacks']);
	});

	it('keeps the picked twist and its Echo stacks while Unsealed Wrath is untalented', async () => {
		const { made, player } = await unsealedWrathPlayer();
		expect(player.getTalents().unsealedWrath).toBe(true);
		renderWithHost(<ConnectedIconEnumPicker modObject={player} config={input('uwTwist') as never} />, made);
		act(() => input('uwTwist').setValue(player, PaladinSealTwist.SealTwistFury));
		act(() => input('uwSofStacks').setValue(player, 2));

		act(() => player.setTalentsString(withoutUnsealedWrath));
		expect(player.getTalents().unsealedWrath).toBe(false);
		act(() => changeAura(player));
		act(() => player.setTalentsString(UnsealedWrathTalents.data.talentsString));

		expect(player.getSimpleRotation()).toMatchObject({ uwTwist: PaladinSealTwist.SealTwistFury, uwSofStacks: 2, uwSorStacks: 3 });
	});
});

describe('the seal picker', () => {
	// The picker reads a stored zero as the talents' seal, here Seal of Command.
	it('keeps the picked seal while Unsealed Wrath is talented', async () => {
		const { made, player } = await unsealedWrathPlayer();
		act(() => player.setTalentsString(withoutUnsealedWrath));
		expect(input('seal').showWhen!(player)).toBe(true);
		renderWithHost(<ConnectedIconEnumPicker modObject={player} config={input('seal') as never} />, made);
		act(() => input('seal').setValue(player, PaladinSeal.SealOfRighteousness));

		act(() => player.setTalentsString(UnsealedWrathTalents.data.talentsString));
		expect(input('seal').showWhen!(player)).toBe(false);
		act(() => changeAura(player));
		act(() => player.setTalentsString(withoutUnsealedWrath));

		expect(input('seal').getValue(player)).toBe(PaladinSeal.SealOfRighteousness);
	});

	// The picker shows only with Seal of Command talented, so the talents' seal without it is read from the input.
	it('reads the seal the new talents decide when none is picked', async () => {
		const { player } = await unsealedWrathPlayer();
		act(() => player.setTalentsString(withoutSealOfCommand));
		expect(input('seal').getValue(player)).toBe(PaladinSeal.SealOfRighteousness);

		act(() => player.setTalentsString(HolyShockTalents.data.talentsString));

		expect(input('seal').getValue(player)).toBe(PaladinSeal.SealOfCommand);
	});
});
