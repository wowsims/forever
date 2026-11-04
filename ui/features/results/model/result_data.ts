import { SimResult, SimResultFilter } from '@sim/proto/sim_result';

export interface SimResultData {
	result: SimResult;
	filter: SimResultFilter;
}

/**
 * A result a view has already drawn, as a string: two emits carrying the same run under the same filter draw the
 * same view, so rebuilding for the second is pure cost.
 */
export const resultKey = (resultData: SimResultData | null): string =>
	resultData ? [resultData.result.request.requestId, JSON.stringify(resultData.filter)].join('|') : '';
