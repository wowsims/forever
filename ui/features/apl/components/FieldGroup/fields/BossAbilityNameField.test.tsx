import { APLRotation } from '@generated/proto/apl';
import { DesignBossAbility, DesignBossAbility_Kind, EncounterDesign } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { Encounter } from '@sim/raid/encounter';
import { makePlayer } from '@testing/apl/player';
import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { BossAbilityNameField } from './BossAbilityNameField';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

let player: ReturnType<typeof makePlayer>;
let picked: string;

const ability = (name: string, kind = DesignBossAbility_Kind.Pulse) => DesignBossAbility.create({ name, kind });
const design = (...names: Array<string>) => EncounterDesign.create({ bossAbilities: names.map(name => ability(name)) });

const mount = () =>
	render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<BossAbilityNameField
				player={player as unknown as Player<any>}
				config={{
					id: 'boss-ability-field',
					getValue: () => picked,
					setValue: (_player, value) => {
						picked = value;
					},
				}}
			/>
		</SimHostProvider>,
	);

const trigger = () => document.getElementById('boss-ability-field')!;
const items = () => screen.queryAllByRole('option').map(item => item.textContent);

beforeEach(() => {
	document.body.innerHTML = '';
	picked = '';
	player = makePlayer(APLRotation.create());
	player.sim.encounter = new Encounter(player.sim as never);
});

describe('BossAbilityNameField', () => {
	it("lists the design's named boss abilities once each and writes the one picked", async () => {
		player.sim.encounter.setDesign(design('Lava Burst', '', 'Wrath of Ragnaros', 'Lava Burst'));
		mount();

		await userEvent.click(trigger());
		expect(items()).toEqual(['Lava Burst', 'Wrath of Ragnaros']);
		await userEvent.click(screen.getByRole('option', { name: 'Wrath of Ragnaros' }));
		expect(picked).toBe('Wrath of Ragnaros');
	});

	it('follows the design as its abilities change', async () => {
		player.sim.encounter.setDesign(design('Lava Burst'));
		mount();

		act(() => player.sim.encounter.setDesign(design('Lava Burst', 'Magma Blast')));
		await userEvent.click(trigger());
		expect(items()).toEqual(['Lava Burst', 'Magma Blast']);
	});

	it('takes a typed name when the design names no ability, since the sim finds an unnamed one by no name', () => {
		player.sim.encounter.setDesign(design(''));
		mount();

		expect(trigger().tagName).toBe('INPUT');
	});
});
