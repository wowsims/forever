import { PrecisionMode as SavedPrecisionMode } from '@generated/proto/ui';

import type { PrecisionMode, StatWeightsSlice } from '../state/sim_store';

export type PrecisionSettings = Pick<StatWeightsSlice, 'precisionMode' | 'iterations' | 'targetCi' | 'maxIterationMultiplier'>;

/** Whether a run can start at `precision`: Target ± needs a target above 0. */
export const precisionRunnable = ({ precisionMode, targetCi }: PrecisionSettings) => precisionMode !== 'target' || targetCi > 0;

export const precisionModeFromProto = (saved: SavedPrecisionMode): PrecisionMode =>
	saved === SavedPrecisionMode.PrecisionModeTarget ? 'target' : 'iterations';

export const precisionModeToProto = (mode: PrecisionMode): SavedPrecisionMode =>
	mode === 'target' ? SavedPrecisionMode.PrecisionModeTarget : SavedPrecisionMode.PrecisionModeIterations;
