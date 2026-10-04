import { PlayerStats } from '@generated/proto/api';
import { IndividualBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { Gear, ItemSwapGear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import type { Decorator } from '@storybook/react-vite';
import { fakeHost, type FakeHostParts } from '@testing/sim/fake_host';
import { type ReactNode, useState } from 'react';

export type StubHostParts = FakeHostParts & { playerSlice?: Partial<PlayerSlice> };

const STUB_PLAYER_KEY = 0;

// The values a freshly constructed Player seeds, so `usePlayerStore` reads a real value for every field.
const defaultPlayerSlice = (): PlayerSlice => ({
	name: '',
	race: 0,
	profession1: 0,
	profession2: 0,
	prof1Rank: 0,
	prof2Rank: 0,
	buffs: IndividualBuffs.create(),
	consumables: ConsumesSpec.create(),
	bonusStats: new Stats(),
	gear: new Gear({}),
	talentsString: '',
	specOptions: {},
	reactionTime: 0,
	channelClipDelay: 0,
	inFrontOfTarget: false,
	distanceFromTarget: 0,
	weaponSkills: [],
	healingModel: HealingModel.create(),
	epWeights: new Stats(),
	epRatios: new Array<number>(6).fill(0),
	currentStats: PlayerStats.create(),
	itemSwapEnabled: false,
	itemSwapGear: new ItemSwapGear({}),
	itemSwapBonusStats: new Stats(),
	dpsRefStat: undefined,
	healRefStat: undefined,
	tankRefStat: undefined,
	v: zeroVersions(PLAYER_FIELDS),
});

/**
 * A host that answers the reads every stateful container makes on mount: `useSimReady` (a resolved
 * `waitForInit`), the store hooks (one `sim.store`, reached from the player as `player.sim`, with the
 * player's slice seeded under `player.storeKey` from the defaults and `playerSlice`, which
 * `player.getGear()` reads back for an item's tooltip) and `useSpecPresets` (the four preset lists a
 * spec always declares, empty). The `player` and `sim` parts are spread over those defaults, so they
 * must be plain objects; any other part replaces its default whole.
 */
export const makeStubHost = ({ player, sim, playerSlice, ...parts }: StubHostParts = {}): IndividualSimHost<any> => {
	const store = createSimStore();
	seedKeyed(store, 'players', STUB_PLAYER_KEY, { ...defaultPlayerSlice(), ...playerSlice });
	const stubSim = { store, waitForInit: () => Promise.resolve(), ...(sim as object) };
	const getGear = () => store.getState().players[STUB_PLAYER_KEY].gear;
	return fakeHost({
		player: { storeKey: STUB_PLAYER_KEY, sim: stubSim, getGear, ...(player as object) },
		sim: stubSim,
		individualConfig: { presets: { epWeights: [], gear: [], talents: [], rotations: [] } },
		...parts,
	});
};

// A component, not a closure with hooks: react-hooks/rules-of-hooks is an error at the lint root.
const StubHost = ({ parts, children }: { parts: StubHostParts; children: ReactNode }) => {
	const [host] = useState(() => makeStubHost(parts));
	return <SimHostProvider host={host}>{children}</SimHostProvider>;
};

/** Wraps a story in `<SimHostProvider>` with `makeStubHost(parts)`, built once per mount. */
export const withSimHost =
	(parts: StubHostParts = {}): Decorator =>
	Story => (
		<StubHost parts={parts}>
			<Story />
		</StubHost>
	);
