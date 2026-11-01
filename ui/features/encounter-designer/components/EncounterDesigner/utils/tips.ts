import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, EncounterDesign, Target } from '@generated/proto/common';
import { DesignBossAbility_Kind } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateMobType, translateSpellSchool } from '@i18n/localization';

import type { DesignEventRef } from '../../../model/lanes';
import { eventLabelKey, eventWindow, firesAtInstants, getEvent, healthSegments, moveSeconds } from '../../../model/lanes';
import type { TimeWindow } from '../../../model/time';
import { fractionToSeconds } from '../../../model/time';
import { ABILITY_KINDS, AMP_TYPES, designerText } from './configs';

// A health percent as read, to a tenth: a hold across a downtime can start at a third of the way down a line.
export const formatPercent = (percent: number): number => Number(percent.toFixed(1));

export const formatSeconds = (seconds: number): string => designerText('seconds', { value: Number(seconds.toFixed(1)) });

export const eventName = (design: EncounterDesign, ref: DesignEventRef): string => {
	const named = getEvent(design, ref)?.name;
	return named || designerText('event_name', { event: i18n.t(eventLabelKey(ref.kind)), number: ref.index + 1 });
};

// The first occurrence keeps the bare name; the repeats after it count from 2.
export const occurrenceName = (name: string, occurrence: number): string =>
	occurrence > 1 ? designerText('occurrence_name', { name, number: occurrence }) : name;

// Reads the times from `window` when one is given, so a bar can describe where a drag has it.
export const eventValueText = (design: EncounterDesign, ref: DesignEventRef, baseSeconds: number, window?: TimeWindow): string => {
	const name = eventName(design, ref);
	const stored = eventWindow(design, ref);
	if (!stored) return name;
	if (ref.kind === 'bossHealth') {
		const segment = healthSegments(design).find(found => found.index === ref.index);
		const start = formatSeconds(fractionToSeconds(window?.start ?? segment?.start ?? 0, baseSeconds));
		const end = formatSeconds(fractionToSeconds(window?.end ?? segment?.end ?? 0, baseSeconds));
		return designerText('health_value', {
			name,
			from: formatPercent(segment?.fromPercent ?? 100),
			to: formatPercent(segment?.toPercent ?? 100),
			start,
			end,
		});
	}
	const shown = window ?? ('at' in stored ? { start: stored.at, end: stored.at } : stored);
	const at = formatSeconds(fractionToSeconds(shown.start, baseSeconds));
	if (!('at' in stored) || (ref.kind === 'move' && window))
		return designerText('span_value', { name, start: at, end: formatSeconds(fractionToSeconds(shown.end, baseSeconds)) });
	return designerText('point_value', { name, at });
};

// A hover tooltip: the title on the first line, then one line per rule.
interface EventTip {
	title: string;
	lines: ReadonlyArray<string>;
}

// A tooltip anchor carries its text in one attribute, a line each; `parseTip` reads it back for the Tooltip to render.
export const serializeTip = (tip: EventTip): string => [tip.title, ...tip.lines].join('\n');
export const parseTip = (text: string | null | undefined): EventTip | undefined => {
	if (!text) return undefined;
	const [title, ...lines] = text.split('\n');
	return { title, lines };
};

// "Adds #2 — Add wave 'Adds'": a repeat of an event leads its title with the occurrence.
export const occurrenceTip = (tip: EventTip, label: string): EventTip => ({ ...tip, title: designerText('tip.occurrence_title', { label, title: tip.title }) });

// "3 adds (Lvl 60 Beast ×2, Lvl 61 Demon)": the adds grouped by level and mob type, in the order they first appear.
const addsSummary = (adds: ReadonlyArray<Target>): string => {
	const groups = new Map<string, { level: number; type: string; count: number }>();
	for (const add of adds) {
		const key = `${add.level}:${add.mobType}`;
		const group = groups.get(key) ?? { level: add.level, type: translateMobType(add.mobType), count: 0 };
		group.count++;
		groups.set(key, group);
	}
	const listed = [...groups.values()].map(group =>
		designerText(group.count > 1 ? 'tip.add_group_count' : 'tip.add_group', { level: group.level, type: group.type, count: group.count }),
	);
	return designerText('tip.adds', { count: adds.length, groups: listed.join(', ') });
};

export const eventTooltip = (
	design: EncounterDesign,
	ref: DesignEventRef,
	baseSeconds: number,
	targetNames: ReadonlyArray<string>,
	window?: TimeWindow,
): EventTip => {
	const event = getEvent(design, ref);
	if (!event) return { title: '', lines: [] };
	const tip = (key: string, options?: Record<string, string | number>) => designerText(`tip.${key}`, options);
	const sec = (fraction: number) => formatSeconds(fractionToSeconds(fraction, baseSeconds));
	const found = eventWindow(design, ref);
	const stored = !found ? { start: 0, end: 0 } : 'at' in found ? { start: found.at, end: found.at } : found;
	const shown = window ?? stored;
	const range = tip('range', { start: sec(shown.start), end: sec(shown.end) });
	const repeat = 'repeat' in event && event.repeat && event.repeat.interval > 0 ? sec(event.repeat.interval) : undefined;
	const every = repeat ? [tip('every', { interval: repeat })] : [];
	const built = (title: string, lines: ReadonlyArray<string | undefined>): EventTip => ({ title, lines: lines.filter((line): line is string => !!line) });
	switch (ref.kind) {
		case 'bossHealth': {
			const segment = healthSegments(design).find(found => found.index === ref.index);
			const covered = window ?? { start: segment?.start ?? 0, end: segment?.end ?? 0 };
			const [from, to] = [segment?.fromPercent ?? 100, segment?.toPercent ?? 100];
			const curve =
				from === to ? tip('health_hold', { percent: formatPercent(from) }) : tip('health', { from: formatPercent(from), to: formatPercent(to) });
			return built(eventName(design, ref), [curve, tip('range', { start: sec(covered.start), end: sec(covered.end) })]);
		}
		case 'downtime': {
			const downtime = event as DesignDowntime;
			const target = targetNames[downtime.targetIndex] ?? designerText('target_name', { number: downtime.targetIndex + 1 });
			return built(tip('downtime'), [target, range, downtime.clearDebuffs ? tip('clears_debuffs') : undefined]);
		}
		case 'ability': {
			const ability = event as DesignBossAbility;
			const kind = designerText(`ability_kinds.${ABILITY_KINDS.find(([value]) => value === ability.kind)?.[1]}`);
			const details: Array<string> = [];
			switch (ability.kind) {
				case DesignBossAbility_Kind.SpellHit:
					details.push(
						tip('school_damage', { school: translateSpellSchool(ability.school), min: ability.minDamage, max: ability.maxDamage }),
						tip('cast', { time: formatSeconds(ability.castTime) }),
					);
					break;
				case DesignBossAbility_Kind.Strike:
					details.push(tip('weapon', { percent: ability.weaponPercent }));
					break;
				case DesignBossAbility_Kind.Thrash:
					details.push(tip('proc', { chance: ability.procChance }));
					if (ability.extraAttacks > 0) details.push(tip('extra_attacks', { count: ability.extraAttacks }));
					if (ability.icd > 0) details.push(tip('icd', { time: formatSeconds(ability.icd) }));
					break;
				default:
					details.push(tip('swing_speed', { value: ability.swingSpeedMultiplier }));
			}
			const when = firesAtInstants(ability)
				? [repeat ? tip('every_from', { interval: repeat, start: sec(shown.start), end: sec(shown.end) }) : tip('at', { at: sec(shown.start) })]
				: [range, ...every];
			return built(kind, [...details, ...when]);
		}
		case 'wave': {
			const wave = event as AddWave;
			const jitter = wave.spawn?.jitter ?? 0;
			const spawn = jitter > 0 ? tip('spawns_jitter', { at: sec(shown.start), jitter: sec(jitter) }) : tip('spawns', { at: sec(shown.start) });
			return built(tip('wave', { name: eventName(design, ref) }), [
				addsSummary(wave.adds),
				spawn,
				tip('lives', { time: sec(shown.end - shown.start) }),
				...every,
				tip('kill_share', { percent: Number((wave.killShare * 100).toFixed(1)) }),
			]);
		}
		case 'move': {
			const move = event as DesignMove;
			const time = formatSeconds(Number(moveSeconds(move).toFixed(2)));
			return built(tip('move'), [tip('at', { at: sec(shown.start) }), move.yards > 0 ? tip('move_yards', { yards: move.yards, time }) : time, ...every]);
		}
		default: {
			const amp = event as DesignDamageAmp;
			const type = designerText(`amp_types.${AMP_TYPES.find(([value]) => value === amp.type)?.[1]}`);
			return built(tip('amp'), [tip('amp_value', { type, percent: amp.percent }), range, ...every]);
		}
	}
};
