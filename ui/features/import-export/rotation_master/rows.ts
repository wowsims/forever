// The report as rows a dialog can list: one per converted item, the ones that need a look first.
import type { FromAplItem } from './from_apl';
import { describeReasons, type Fidelity } from './report';
import type { ToAplItem } from './to_apl';

export interface ReportRow {
	key: string;
	/** Where the item is in the original ("rotation[3].onTrue[1]", "priorityList[2]"). */
	path: string;
	/** What it is ("spell #133", "item #19019"). */
	title: string;
	/** The original, or what it became, as short text. */
	detail: string;
	fidelity: Fidelity;
	outcome: string;
	/** What did not carry over exactly, as sentences. Empty for an exact item. */
	sentences: string[];
}

const RANK: Record<Fidelity, number> = { unsupported: 0, lossy: 1, approximate: 2, exact: 3 };
const byAttention = (a: ReportRow, b: ReportRow): number => RANK[a.fidelity] - RANK[b.fidelity];

/** Rows for an import (Rotation Master -> APL). Branches and Stops are structure, not items, and have no row. */
export function importRows(items: ReadonlyArray<ToAplItem>): ReportRow[] {
	return items
		.filter(item => item.type !== 'branch' && item.type !== 'stop')
		.map((item, i) => ({
			key: `${i}:${item.path}`,
			path: item.path,
			title: [item.type, item.label, item.name].filter(Boolean).join(' '),
			detail: item.section === 'cooldowns' ? 'Cooldowns list' : '',
			fidelity: item.fidelity,
			outcome: item.outcome,
			sentences: describeReasons(item.reasons),
		}))
		.sort(byAttention);
}

/** Rows for an export (APL -> Rotation Master). */
export function exportRows(items: ReadonlyArray<FromAplItem>): ReportRow[] {
	return items
		.filter(item => !(item.hidden && item.outcome === 'skipped'))
		.map((item, i) => ({
			key: `${i}:${item.path}`,
			path: item.path,
			title: item.became.length
				? item.became.map(b => `${b.type} ${Array.isArray(b.action) ? b.action.join('+') : '#' + b.action}`).join(', ')
				: item.path,
			detail: item.original,
			fidelity: item.fidelity,
			outcome: item.outcome,
			sentences: describeReasons(item.reasons),
		}))
		.sort(byAttention);
}
