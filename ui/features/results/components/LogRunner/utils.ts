import i18n from '@i18n/config';
import type { CombatLog, Outcome } from '@sim/proto/combat_log';
import { formattedTimestamp, rawWithoutTimestamp } from '@sim/proto/combat_log';
import { OUTCOMES } from '@sim/proto/combat_log/types';

import type { SuggestionSource } from '../../model/log/search/indexes';
import { TYPE_SUGGESTIONS } from '../../model/log/search/indexes';
import type { ClauseField, SearchGroup } from '../../model/log/search/query';
import type { SimResultData } from '../../model/result_data';
import { outcomeLabel } from '../DamageResult';

/**
 * A `SearchGroup` with the identity React needs to keep a group's half-typed value across the removal
 * of an earlier one.
 */
export type IdentifiedSearchGroup = SearchGroup & { id: number };

export const DEBUG_MARKER = '[DEBUG]';

/** Fields whose values are typed into a box rather than picked, and the placeholder each shows. */
export const TYPED_FIELDS: Partial<Record<ClauseField, string>> = { time: '10-30', amount: '>5000' };

export const fieldLabel = (field: ClauseField): string => {
	switch (field) {
		case 'source':
			return i18n.t('results_tab.details.logs.search_fields.source');
		case 'target':
			return i18n.t('results_tab.details.logs.search_fields.target');
		case 'spell':
			return i18n.t('results_tab.details.logs.search_fields.spell');
		case 'type':
			return i18n.t('results_tab.details.logs.search_fields.type');
		case 'school':
			return i18n.t('results_tab.details.logs.search_fields.school');
		case 'outcome':
			return i18n.t('results_tab.details.logs.search_fields.outcome');
		case 'time':
			return i18n.t('results_tab.details.logs.search_fields.time');
		case 'amount':
			return i18n.t('results_tab.details.logs.search_fields.amount');
	}
};

const typeLabel = (token: string): string | undefined => {
	switch (token) {
		case 'damage':
			return i18n.t('results_tab.details.logs.search_types.damage');
		case 'heal':
			return i18n.t('results_tab.details.logs.search_types.heal');
		case 'shield':
			return i18n.t('results_tab.details.logs.search_types.shield');
		case 'resource':
			return i18n.t('results_tab.details.logs.search_types.resource');
		case 'aura':
			return i18n.t('results_tab.details.logs.search_types.aura');
		case 'buff':
			return i18n.t('results_tab.details.logs.search_types.buff');
		case 'cast':
			return i18n.t('results_tab.details.logs.search_types.cast');
		case 'delay':
			return i18n.t('results_tab.details.logs.search_types.delay');
		case 'major-cooldown':
			return i18n.t('results_tab.details.logs.search_types.major_cooldown');
		case 'stat-change':
			return i18n.t('results_tab.details.logs.search_types.stat_change');
		case 'debug':
			return i18n.t('results_tab.details.logs.search_types.debug');
	}
};

export const labelOf = (field: ClauseField, value: string): string => {
	switch (field) {
		case 'outcome':
			return outcomeLabel(value as Outcome) ?? value;
		case 'type':
			return typeLabel(value) ?? value;
		default:
			return value;
	}
};

/** Quoted phrases stay whole, everything else splits on whitespace. */
export const keywordsOf = (text: string): Array<string> => {
	const keywords: Array<string> = [];
	for (const match of text.matchAll(/"([^"]+)"|\S+/g)) keywords.push(match[1] ?? match[0]);
	return keywords;
};

/**
 * The timeline narrows its unit list with this filter; the log has no unit rows, so the
 * equivalent is to keep the lines naming that target. Null whenever it would be a no-op, which
 * includes every single-target encounter.
 */
export const selectedTargetNumber = (resultData: SimResultData): number | null => {
	if (resultData.result.getTargets().length < 2) return null;
	const selected = resultData.result.getTargets(resultData.filter);
	return selected.length === 1 ? selected[0].index + 1 : null;
};

/** The values a picked field offers. A typed field (see `TYPED_FIELDS`) offers none. */
export const valueCandidates = (field: ClauseField, suggestions: SuggestionSource): ReadonlyArray<string> => {
	switch (field) {
		case 'type':
			return TYPE_SUGGESTIONS;
		case 'outcome':
			return OUTCOMES;
		case 'school':
			return suggestions.schools;
		case 'spell':
			return suggestions.spells;
		case 'source':
		case 'target':
			return suggestions.units;
		default:
			return [];
	}
};

/** What the exporter dialog writes out: every line, unfiltered, timestamp first. */
export const combinedLogText = (logs: ReadonlyArray<CombatLog>): string =>
	logs.map(log => `${formattedTimestamp(log)};${rawWithoutTimestamp(log.raw)}`).join('\n');
