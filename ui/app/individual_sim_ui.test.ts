// What the host's `fromProto` hands the settings loader: a gear planner never loads the encounter, and an import
// that held nothing else loads nothing, where the empty list it would leave reads as every category.
import { IndividualSimSettings } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { getEnumValues } from '@sim/utils/collections';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SimHostObject } from './individual_sim_ui';

const apply = vi.hoisted(() => vi.fn());
vi.mock('@sim/state/serialization', async importOriginal => ({
	...(await importOriginal<typeof import('@sim/state/serialization')>()),
	applyIndividualSimSettings: apply,
}));

const context = { player: {}, sim: {} };
// The host as its constructor leaves it, for the two fields `fromProto` reads.
const hostFor = (simDisabled: boolean) =>
	Object.assign(Object.create(SimHostObject.prototype), { simDisabled, serializationContext: () => context }) as SimHostObject<any>;
const settings = IndividualSimSettings.create();
const ALL = getEnumValues<SimSettingCategories>(SimSettingCategories);
const ALL_BUT_ENCOUNTER = ALL.filter(category => category !== SimSettingCategories.Encounter);

beforeEach(() => apply.mockClear());

describe('a gear planner', () => {
	const planner = hostFor(true);

	it('loads nothing from an import of the encounter alone', () => {
		planner.fromProto(settings, [SimSettingCategories.Encounter]);
		expect(apply).not.toHaveBeenCalled();
	});

	it('loads the rest of an import that also holds the encounter', () => {
		planner.fromProto(settings, [SimSettingCategories.Encounter, SimSettingCategories.Gear]);
		expect(apply).toHaveBeenCalledWith(context, settings, [SimSettingCategories.Gear]);
	});

	it.each([undefined, []])('loads every category but the encounter when given %j', categories => {
		planner.fromProto(settings, categories);
		expect(apply).toHaveBeenCalledWith(context, settings, ALL_BUT_ENCOUNTER);
	});
});

describe('a page that sims', () => {
	const page = hostFor(false);

	it.each([undefined, [], [SimSettingCategories.Encounter]])('hands the categories %j on as they are', categories => {
		page.fromProto(settings, categories);
		expect(apply).toHaveBeenCalledWith(context, settings, categories);
	});
});
