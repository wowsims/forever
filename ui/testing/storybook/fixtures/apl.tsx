import { AplProvider, type AplProviderProps } from '@features/apl/context/AplContext';
import { APLRotation, type APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { IndividualSimHost } from '@sim/sim_host';
import { patchKeyed } from '@sim/state/sim_store';
import type { Decorator } from '@storybook/react-vite';
import { type AplTestPlayer, makePlayer } from '@testing/apl/player';
import { type ReactNode, useState } from 'react';

import { makeStubHost } from './models';
import { makeAplRotation } from './protos';

const touchRotation = (host: IndividualSimHost<any>) => patchKeyed(host.sim.store, 'players', host.player.storeKey, {}, ['rotation']);

/**
 * `makePlayer` from `@testing/apl/player`, on a store with a seeded player slice: a rotation write
 * bumps the slice's `rotation` version, so the pickers subscribed to it re-read and a story stays
 * editable.
 */
export const makeAplStoryPlayer = (rotation: APLRotation = makeAplRotation()): AplTestPlayer => {
	const host = makeStubHost();
	const player = makePlayer(rotation, () => touchRotation(host));
	player.storeKey = host.player.storeKey;
	player.sim.store = host.sim.store;
	return player;
};

const AplHost = ({ rotation, children }: { rotation: () => APLRotation; children: ReactNode }) => {
	const [host] = useState(() => {
		const player = makeAplStoryPlayer(rotation());
		const applyEmptyAplRotation = () =>
			player.modifyAplRotation(target => {
				target.prepullActions = [];
				target.priorityList = [];
			});
		return { player, rootElem: document.body, applyEmptyAplRotation } as unknown as IndividualSimHost<any>;
	});
	return <SimHostProvider host={host}>{children}</SimHostProvider>;
};

/** A host whose player holds `rotation()`, for a story of a whole APL pane: the panes read the player off the host. */
export const withAplHost =
	(rotation: () => APLRotation): Decorator =>
	Story => (
		<AplHost rotation={rotation}>
			<Story />
		</AplHost>
	);

/** Renders the story inside `<AplProvider>`, e.g. `{ isPrepull: true }` for a pre-pull row. */
export const withAplScope =
	(scope: Omit<AplProviderProps, 'children'>): Decorator =>
	Story => (
		<AplProvider {...scope}>
			<Story />
		</AplProvider>
	);

export interface RotationTypeHostOptions {
	type: APLRotationType;
	hasSimple: boolean;
}

const makeRotationTypeHost = ({ type, hasSimple }: RotationTypeHostOptions): IndividualSimHost<any> => {
	const rotation = APLRotation.create({ type });
	const host = makeStubHost({
		player: {
			hasSimpleRotationGenerator: () => hasSimple,
			getRotationType: () => rotation.type,
			modifyAplRotation: (change: (target: APLRotation) => void) => {
				change(rotation);
				touchRotation(host);
			},
		},
	});
	return host;
};

// A component, not a closure with hooks: react-hooks/rules-of-hooks is an error at the lint root.
const RotationTypeHost = ({ options, children }: { options: RotationTypeHostOptions; children: ReactNode }) => {
	const [host] = useState(() => makeRotationTypeHost(options));
	return <SimHostProvider host={host}>{children}</SimHostProvider>;
};

/** A host whose player holds a rotation of `type` and answers the reads and write `RotationTypePicker` makes. */
export const withRotationTypeHost =
	(options: RotationTypeHostOptions): Decorator =>
	Story => (
		<RotationTypeHost options={options}>
			<Story />
		</RotationTypeHost>
	);
