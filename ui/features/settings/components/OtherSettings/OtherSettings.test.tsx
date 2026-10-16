import { OpenSelectorModalContext } from '@features/gear/hooks/useSelectorModal';
import { Class, ConsumableType, ConsumesSpec, HandType, ItemSlot, Race, Spec, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { DatabaseFilters, UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { ItemSwapGear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { createSimStore, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { fakeHost } from '@testing/sim/fake_host';
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const CONSUMABLES = [Consumable.create({ id: 12404, name: 'Dense Sharpening Stone', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 })];

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', name: '', href: '', ready: true }) }));
vi.mock('@sim/proto/database', () => ({ Database: { getSync: () => ({ getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined }) } }));
vi.mock('@features/inputs/components/ConnectedBooleanPicker', () => ({ ConnectedBooleanPicker: () => <div /> }));
vi.mock('@features/inputs/components/ConnectedIconEnumPicker', async importActual => ({
	...(await importActual<typeof import('@features/inputs/components/ConnectedIconEnumPicker')>()),
	ConnectedIconEnumPicker: () => <span data-testid="swap-consume-picker" />,
}));

const { OtherSettings } = await import('./OtherSettings');

const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
const weights = new Stats();

const setup = (enabled: boolean) => {
	const store = createSimStore();
	const storeKey = 0;
	seedKeyed(store, 'players', storeKey, { itemSwapEnabled: enabled, itemSwapGear: new ItemSwapGear({}), v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = {
		storeKey,
		sim: {
			store,
			getMinConsumableLvl: () => 0,
			getFilters: () => DatabaseFilters.create(),
			encounter: { getTarget: () => undefined, getZoneId: () => 0 },
		},
		getClass: () => Class.ClassWarrior,
		getSpec: () => Spec.SpecDpsWarrior,
		getRace: () => Race.RaceOrc,
		hasProfession: () => false,
		getEquippedItem: () => null,
		itemSwapSettings: {
			getEnableItemSwap: () => enabled,
			getItem: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotMainHand ? { item: sword } : null),
		},
		getSpecConfig: () => ({ epStats: [Stat.StatStrength, Stat.StatAttackPower], displayStats: [] }),
		getParty: () => ({ getBuffs: () => ({}) }),
		getConsumes: () => ConsumesSpec.create(),
		getEpWeights: () => weights,
	} as unknown as Player<any>;
	render(
		<SimHostProvider host={fakeHost({ player })}>
			<OpenSelectorModalContext value={vi.fn()}>
				<OtherSettings inputs={[]} itemSlots={[ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand]} />
			</OpenSelectorModalContext>
		</SimHostProvider>,
	);
	return screen.getByTestId('item-swap-box');
};

describe('OtherSettings', () => {
	it("puts the swap weapons' stones and oils in the item swap box under the swap gear, plainly captioned, and hides a row no swap weapon takes", () => {
		const box = setup(true);

		const row = within(box).getByRole('group', { name: i18n.t('settings_tab.consumables.imbue.title') });
		expect(row.contains(screen.getByTestId('consumes-swap-imbue'))).toBe(true);
		expect(box.children[0]).toBe(screen.getByTestId('input-item-swap-container'));
		expect(within(box).queryByRole('group', { name: i18n.t('settings_tab.consumables.weaponImbue.title') })).toBeNull();
	});

	it('keeps the swap consumable pickers mounted in the hidden box while item swap is off', () => {
		const box = setup(false);

		expect(box.hasAttribute('hidden')).toBe(true);
		expect(within(box).getAllByTestId('swap-consume-picker')).toHaveLength(4);
	});
});
