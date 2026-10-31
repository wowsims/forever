import { AplProvider } from '@features/apl/context/AplContext';
import { APLRotation, APLValue } from '@generated/proto/apl';
import { ActionID, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { ActionId } from '@sim/proto/action_id';
import { makePlayer } from '@testing/apl/player';
import { source } from '@testing/apl/source';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ValuePicker } from '../../ValuePicker';
import { ActionIdField } from './ActionIdField';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

// Named, as `fill()` returns them: an unfilled id would send the menu icon off to fetch one.
const named = (spellId: number, name: string) => Object.assign(Object.create(ActionId.prototype), ActionId.fromSpellId(spellId), { name }) as ActionId;

const spell = (id: ActionId, flags: { prepullOnly?: boolean; encounterOnly?: boolean } = {}) => ({
	id,
	data: { isCastable: true, isPotion: false, isConjured: false, isMajorCooldown: false, prepullOnly: false, encounterOnly: false, ...flags },
});

let player: ReturnType<typeof makePlayer>;

const setup = () => {
	source.listeners.clear();
	player = makePlayer(APLRotation.create(), () => source.notify());
	const spells = [
		spell(named(1, 'Plain Spell')),
		spell(named(2, 'Prepull Spell'), { prepullOnly: true }),
		spell(named(3, 'Encounter Spell'), { encounterOnly: true }),
	];
	player.sim.getUnitMetadata = () => ({ getSpells: () => spells, getAuras: () => [], getName: () => '', getMenuGroupOrder: () => [] });
};

const mount = (isPrepull: boolean) =>
	render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<AplProvider isPrepull={isPrepull}>
				<ActionIdField
					player={player as never}
					config={{ id: 'spell', storeSubscribe: () => source.subscribe, getValue: () => ActionID.create(), setValue: () => {} } as never}
					actionIdSet="castable_spells"
					unitRefField={undefined}
					defaultUnitRef="self"
					getParentValue={() => ({})}
				/>
			</AplProvider>
		</SimHostProvider>,
	);

/** The Spell submenu, which is where `castable_spells` files every spell the sim gives no category. */
const openSpellSubmenu = async () => {
	await act(async () => {});
	await act(async () => {
		fireEvent.click(document.getElementById('spell')!);
	});
	await act(async () => {
		fireEvent.click(screen.getAllByRole('menuitem')[0]);
	});
	return within(screen.getByTestId('choice-field-submenu'))
		.queryAllByRole('menuitemradio')
		.map(item => item.textContent);
};

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('ActionIdField — the list a spell may be used in', () => {
	it.each([
		['no pre-pull-only spell outside', false, ['Plain Spell', 'Encounter Spell']],
		['no encounter-only spell inside', true, ['Plain Spell', 'Prepull Spell']],
	])('offers %s the pre-pull list', async (_case, isPrepull, offered) => {
		setup();
		mount(isPrepull);

		expect(await openSpellSubmenu()).toEqual(offered);
	});
});

describe('ActionIdField — the trigger element', () => {
	it('keeps the same trigger node once the async spell list resolves and gains its submenu', async () => {
		setup();
		mount(false);

		const before = document.getElementById('spell');
		expect(before).not.toBeNull();

		await act(async () => {});

		expect(document.getElementById('spell')).toBe(before);
	});
});

describe('ActionIdField — the unit named by unitRefField', () => {
	const TARGET_REF = UnitReference.create({ type: UnitType.Target, index: 3 });
	const PLAYER_REF = UnitReference.create({ type: UnitType.Player, index: 1 });

	it('reads the unit named by unitRefField on mount', async () => {
		setup();
		const reads: Array<string> = [];
		const metadata = { getSpells: () => [], getAuras: () => [], getName: () => '', getMenuGroupOrder: () => [] };
		player.sim.getUnitMetadata = (ref: UnitReference | undefined) => {
			reads.push(ref ? `${UnitType[ref.type]}:${ref.index}` : 'none');
			return metadata;
		};
		const parentValue = { sourceUnit: TARGET_REF, otherUnit: PLAYER_REF, auraId: ActionID.create() };
		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<ActionIdField
					player={player as never}
					config={{ id: 'aura', storeSubscribe: () => source.subscribe, getValue: () => parentValue.auraId, setValue: () => {} } as never}
					actionIdSet="auras"
					unitRefField="sourceUnit"
					defaultUnitRef="self"
					getParentValue={() => parentValue}
				/>
			</SimHostProvider>,
		);
		await act(async () => {});

		expect(new Set(reads)).toEqual(new Set(['Target:3']));
	});

	it('replaces the whole field group when the kind changes, so the mounted ActionIdField is unmounted before the second spec exists', async () => {
		setup();
		player.aplRotation = APLRotation.create({
			valueVariables: [{ name: 'root', value: APLValue.create({ value: { oneofKind: 'auraIsKnown', auraIsKnown: {} }, uuid: { value: 'u' } }) }],
		});
		const valueConfig = {
			getValue: () => player.aplRotation.valueVariables[0].value,
			setValue: (_subject: unknown, next: APLValue | undefined) => {
				player.aplRotation.valueVariables[0].value = next;
				source.notify();
			},
		};
		const builderRoots = () => Array.from(document.querySelectorAll('[data-testid="apl-picker-builder-root"]'));
		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<ValuePicker player={player as never} config={valueConfig as never} />
			</SimHostProvider>,
		);
		await act(async () => {});
		const before = builderRoots()[0];
		expect(before).toBeTruthy();

		await act(async () => {
			player.aplRotation.valueVariables[0].value!.value = { oneofKind: 'dotIsActive', dotIsActive: {} } as never;
			source.notify();
		});

		expect(builderRoots()[0]).toBeTruthy();
		expect(builderRoots()[0]).not.toBe(before);
		expect(before.isConnected).toBe(false);
	});
});
