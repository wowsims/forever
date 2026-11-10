import { patchSlice } from '@sim/state/sim_store';

import { makeStubHost } from './models';

export const makePowerToolsHost = () => {
	const host = makeStubHost();
	patchSlice(host.sim.store, 'ui', { showPowerTools: true });
	return host;
};
