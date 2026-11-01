import { loadRotation } from '@features/apl/components/SavedRotation/utils';
import { loadGearSet } from '@features/gear/components/SavedGear/utils';
import { APLRotation } from '@generated/proto/apl';
import { EquipmentSpec, ItemSlot, ItemSpec, ItemType, Spec } from '@generated/proto/common';
import { SavedGearSet, SavedRotation, SavedTalents, UIDatabase, UIItem } from '@generated/proto/ui';
import { fromJsonStringLenient } from '@sim/proto/lenient_json';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { SAVED_DATA_TYPE_RULES, specFromName, specFullName } from './saved_data_types';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const RING_ID = 1001;
const UNKNOWN_ID = 999_999_999;

const entryOf = <T>(data: T) => ({ name: 'Imported', data, json: '' });
const gearOf = (...ids: number[]) => SavedGearSet.create({ gear: EquipmentSpec.create({ items: ids.map(id => ItemSpec.create({ id })) }) });

const RING = UIItem.create({ id: RING_ID, name: 'Ring', type: ItemType.ItemTypeFinger, scalingOptions: { 0: { ilvl: 60 } } });

beforeAll(() => serveDatabase(UIDatabase.create({ items: [RING] })));

describe('an imported entry naming what this database does not have', () => {
	it('loads gear without the unknown item, as a stale saved set does', async () => {
		const { host, player } = await makeIndividualSim();
		const data = gearOf(RING_ID, UNKNOWN_ID);

		expect(() => SAVED_DATA_TYPE_RULES.gear.check!(data, host)).not.toThrow();
		loadGearSet(entryOf(data), host);

		expect(player.getGear().getEquippedItem(ItemSlot.ItemSlotFinger1)?.item.id).toBe(RING_ID);
		expect(
			player
				.getGear()
				.asSpec()
				.items.filter(item => item.id !== 0),
		).toHaveLength(1);
	});

	it('refuses gear with more items than their slots hold, which the database throws on', async () => {
		const { host } = await makeIndividualSim();

		expect(() => SAVED_DATA_TYPE_RULES.gear.check!(gearOf(RING_ID, RING_ID), host)).not.toThrow();
		expect(() => SAVED_DATA_TYPE_RULES.gear.check!(gearOf(RING_ID, RING_ID, RING_ID), host)).toThrow(/No slots left/);
	});

	it('loads a rotation that names an unknown spell, and drops an action name the protos do not list', async () => {
		const { host, player } = await makeIndividualSim();
		const json = JSON.stringify({
			rotation: {
				type: 'TypeAPL',
				priorityList: [
					{ action: { castSpell: { spellId: { spellId: UNKNOWN_ID } } } },
					{ action: { castSpell: { spellId: { otherId: 'OtherActionNotInTheProtos' } } } },
					{ action: { notAnAction: {} } },
				],
			},
		});
		const data = fromJsonStringLenient(SavedRotation, json);

		loadRotation(entryOf(data), host);

		const loaded = player.getResolvedAplRotation();
		expect(loaded.priorityList).toHaveLength(3);
		expect(() => APLRotation.toBinary(loaded)).not.toThrow();
	});
});

describe('the talents check', () => {
	const check = (talentsString: string) => SAVED_DATA_TYPE_RULES.talents.check!(SavedTalents.create({ talentsString }), {} as never);

	it.each(['', '2305', '-2305', '2500052300030150330125--053500031003001'])('takes %s', talentsString => {
		expect(() => check(talentsString)).not.toThrow();
	});

	it.each(['=HYPERLINK("x")', '+1', '@A1', '12a'])('refuses %s, which would reach the Batch CSV as it stands', talentsString => {
		expect(() => check(talentsString)).toThrow();
	});
});

describe('specFromName', () => {
	it('reads an enum member', () => {
		expect(specFromName('SpecMage')).toBe(Spec.SpecMage);
	});

	it.each(['constructor', '__proto__', 'toString', '1', 'SpecUnknown', 'SpecNope'])('takes %s for no spec', name => {
		expect(specFromName(name)).toBeUndefined();
	});
});

describe('specFullName', () => {
	it('leaves no space where a one-spec class has an empty spec name', () => {
		expect(specFullName(Spec.SpecMage)).not.toMatch(/^\s/);
	});
});
