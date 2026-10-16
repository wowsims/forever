import { StatCapConfig, StatCapType, StatWeightsResult, UIStat as UnitStatProto } from '@generated/proto/api';
import { Class, PseudoStat, Stat, UnitStats } from '@generated/proto/common';
import { translatePseudoStat, translateStat } from '@i18n/localization';

import * as Mechanics from '../constants/mechanics';
import { CURRENT_API_VERSION } from '../constants/other';
import { getEnumValues } from '../utils/collections';

const STATS_LEN = getEnumValues(Stat).length;
const PSEUDOSTATS_LEN = getEnumValues(PseudoStat).length;

export const SCHOOL_HIT_PERCENT_PSEUDO_STATS = [
	PseudoStat.PseudoStatSchoolHitPercentArcane,
	PseudoStat.PseudoStatSchoolHitPercentFire,
	PseudoStat.PseudoStatSchoolHitPercentFrost,
	PseudoStat.PseudoStatSchoolHitPercentHoly,
	PseudoStat.PseudoStatSchoolHitPercentNature,
	PseudoStat.PseudoStatSchoolHitPercentShadow,
];

export class UnitStat {
	private readonly stat: Stat | null;
	private readonly pseudoStat: PseudoStat | null;

	// Used to link a "child" PseudoStat like PhysicalHitPercent to a
	// "parent" Stat like HitRating, so that both values can be displayed
	// together in the character sheet.
	private readonly rootStat: Stat | null;

	private constructor(stat: Stat | null, pseudoStat: PseudoStat | null, rootStat: Stat | null) {
		this.stat = stat;
		this.pseudoStat = pseudoStat;
		this.rootStat = rootStat;
	}

	isStat(): boolean {
		return this.stat != null;
	}
	isPseudoStat(): boolean {
		return this.pseudoStat != null;
	}
	hasRootStat(): boolean {
		return this.rootStat != null;
	}

	getStat(): Stat {
		if (!this.isStat()) {
			throw new Error('Not a stat!');
		}
		return this.stat!;
	}
	getPseudoStat(): PseudoStat {
		if (!this.isPseudoStat()) {
			throw new Error('Not a pseudo stat!');
		}
		return this.pseudoStat!;
	}
	getRootStat(): Stat {
		if (!this.hasRootStat()) {
			throw new Error('No root stat for this PseudoStat');
		}
		return this.rootStat!;
	}

	equals(other: UnitStat): boolean {
		return this.stat == other.stat && this.pseudoStat == other.pseudoStat;
	}

	equalsStat(other: Stat): boolean {
		return this.isStat() && this.stat == other;
	}

	equalsPseudoStat(other: PseudoStat): boolean {
		return this.isPseudoStat() && this.pseudoStat == other;
	}

	linkedToStat(other: Stat): boolean {
		return this.stat == other || this.rootStat == other;
	}

	getFullName(playerClass: Class): string {
		const fullName = this.isStat() ? translateStat(this.getStat()) : translatePseudoStat(this.getPseudoStat());
		// A hunter reads its physical stats as ranged.
		if (playerClass == Class.ClassHunter) {
			return fullName.replace('Physical', 'Ranged');
		} else {
			return fullName.replace('Physical', 'Melee');
		}
	}

	getShortName(playerClass: Class): string {
		const fullName = this.getFullName(playerClass);
		return fullName.replace(' Rating', '').replace(' Percent', '');
	}

	getKey(): string {
		return this.isStat() ? Stat[this.stat!] : PseudoStat[this.getPseudoStat()];
	}

	// The rating a percentage point of this stat costs, or null where a rating representation does not
	// make sense for the stat in question (Strength for example).
	private ratingPerPercent(): number | null {
		if (this.linkedToStat(Stat.StatHitRating) || this.isSchoolHitPercent()) {
			return Mechanics.HIT_RATING_PER_HIT_PERCENT;
		} else if (this.linkedToStat(Stat.StatCritRating)) {
			return Mechanics.CRIT_RATING_PER_CRIT_PERCENT;
		} else if (this.linkedToStat(Stat.StatHasteRating)) {
			return Mechanics.HASTE_RATING_PER_HASTE_PERCENT;
		} else if (this.linkedToStat(Stat.StatExpertiseRating)) {
			return Mechanics.EXPERTISE_RATING_PER_EXPERTISE_PERCENT;
		} else if (this.linkedToStat(Stat.StatDefenseRating)) {
			return Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL;
		} else if (this.linkedToStat(Stat.StatDodgeRating)) {
			return Mechanics.DODGE_RATING_PER_DODGE_PERCENT;
		} else if (this.linkedToStat(Stat.StatParryRating)) {
			return Mechanics.PARRY_RATING_PER_PARRY_PERCENT;
		} else if (this.linkedToStat(Stat.StatBlockRating)) {
			return Mechanics.BLOCK_RATING_PER_BLOCK_PERCENT;
		} else {
			return null;
		}
	}

	// Convert a UnitStat value from its Rating representation to a percentage representation
	// (0-100). If a percentage representation does not make sense for the stat in question
	// (Strength for example), then null is returned.
	convertRatingToPercent(ratingValue: number): number | null {
		const perPercent = this.ratingPerPercent();
		return perPercent === null ? null : ratingValue / perPercent;
	}

	// Convert a UnitStat value from its percentage representation (0-100) to the equivalent amount of
	// Rating. If a Rating representation does not make sense for the stat in question (Block in Cata
	// for example), then null is returned.
	// For PseudoStatReducedCritTakenPercent, parentStat specifies the source (DefenseRating).
	convertPercentToRating(percentOrPointsValue: number, parentStat?: Stat): number | null {
		const perPercent = this.ratingPerPercent();
		if (perPercent !== null) {
			return percentOrPointsValue * perPercent;
		}
		if (this.equalsPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent) && parentStat === Stat.StatDefenseRating) {
			return percentOrPointsValue * (Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL / Mechanics.MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE);
		}
		return null;
	}

	private isSchoolHitPercent(): boolean {
		return this.isPseudoStat() && SCHOOL_HIT_PERCENT_PSEUDO_STATS.includes(this.pseudoStat!);
	}

	convertDefaultUnitsToPercent(value: number): number | null {
		if (this.isStat()) {
			// Proper stats are either in Rating units, or do not
			// have a percent representation.
			return this.convertRatingToPercent(value);
		} else {
			// PseudoStats are either already in percent units, or do
			// not have a percent representation.
			if (PseudoStat[this.getPseudoStat()].includes('Percent')) {
				return value;
			} else {
				return null;
			}
		}
	}

	convertRatingToDefaultUnits(ratingValue: number): number | null {
		if (this.isStat()) {
			return Stat[this.stat!].includes('Rating') ? ratingValue : null;
		} else {
			return this.convertRatingToPercent(ratingValue);
		}
	}

	convertPercentToDefaultUnits(percentValue: number): number | null {
		if (this.isStat()) {
			return this.convertPercentToRating(percentValue);
		} else {
			return PseudoStat[this.getPseudoStat()].includes('Percent') ? percentValue : null;
		}
	}

	convertEpToRatingScale(epValue: number): number {
		if (this.isPseudoStat() && PseudoStat[this.pseudoStat!].includes('Percent')) {
			return this.convertRatingToPercent(epValue) ?? epValue;
		} else {
			return epValue;
		}
	}

	getProtoValue(proto: UnitStats): number {
		if (this.isStat()) {
			return proto.stats[this.stat!];
		} else {
			return proto.pseudoStats[this.pseudoStat!];
		}
	}

	setProtoValue(proto: UnitStats, val: number) {
		if (this.isStat()) {
			proto.stats[this.stat!] = val;
		} else {
			proto.pseudoStats[this.pseudoStat!] = val;
		}
	}

	toJson(): object {
		return UnitStatProto.toJson(this.toProto()) as object;
	}

	static fromJson(obj: any): UnitStat {
		return UnitStat.fromProto(UnitStatProto.fromJson(obj));
	}

	toProto(): UnitStatProto {
		const protoMessage = UnitStatProto.create({
			apiVersion: CURRENT_API_VERSION,
		});

		if (this.isStat()) {
			protoMessage.unitStat = {
				oneofKind: 'stat',
				stat: this.stat!,
			};
		} else if (this.isPseudoStat()) {
			protoMessage.unitStat = {
				oneofKind: 'pseudoStat',
				pseudoStat: this.pseudoStat!,
			};
		} else {
			throw new Error('Neither a Stat nor a PseudoStat!');
		}

		return protoMessage;
	}

	static fromProto(protoMessage: UnitStatProto): UnitStat {
		if (protoMessage.unitStat.oneofKind == 'stat') {
			return UnitStat.fromStat(protoMessage.unitStat.stat);
		} else if (protoMessage.unitStat.oneofKind == 'pseudoStat') {
			return UnitStat.fromPseudoStat(protoMessage.unitStat.pseudoStat);
		} else {
			return new UnitStat(null, null, null);
		}
	}

	static fromStat(stat: Stat): UnitStat {
		return new UnitStat(stat, null, stat);
	}
	static fromPseudoStat(pseudoStat: PseudoStat): UnitStat {
		return new UnitStat(null, pseudoStat, UnitStat.getRootStat(pseudoStat));
	}

	static getAllStats(): Array<UnitStat> {
		const allStats = getEnumValues(Stat) as Array<Stat>;
		return allStats.map(stat => UnitStat.fromStat(stat));
	}
	static getAllPseudoStats(): Array<UnitStat> {
		const allPseudoStats = getEnumValues(PseudoStat) as Array<PseudoStat>;
		return allPseudoStats.map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat));
	}
	static getAll(): Array<UnitStat> {
		return [UnitStat.getAllStats(), UnitStat.getAllPseudoStats()].flat();
	}

	// Returns the "parent" Stat (such as HitRating) associated with a
	// "child" PseudoStat (such as PhysicalHitPercent), or null if there is
	// no such root Stat.
	static getRootStat(pseudoStat: PseudoStat): Stat | null {
		const pseudoStatName = PseudoStat[pseudoStat];

		if (pseudoStatName.endsWith('Rating')) {
			// The per-attack hit and crit ratings are already in rating units; the character sheet
			// pairs each with its percent row.
			return null;
		} else if (pseudoStatName.includes('Expertise')) {
			return Stat.StatExpertiseRating;
		} else if (pseudoStatName.includes('Dodge')) {
			return Stat.StatDodgeRating;
		} else if (pseudoStatName.includes('Parry')) {
			return Stat.StatParryRating;
		} else if (pseudoStatName.includes('BlockValue')) {
			return Stat.StatBlockValue;
		} else if (pseudoStatName.includes('Block')) {
			return Stat.StatBlockRating;
		} else if (pseudoStatName.includes('SchoolHit')) {
			return null;
		} else if (pseudoStatName.includes('Haste')) {
			return Stat.StatHasteRating;
		} else if (pseudoStatName.includes('Hit')) {
			return Stat.StatHitRating;
		} else if (pseudoStatName.includes('ReducedCritTaken')) {
			return null;
		} else if (pseudoStatName.includes('Crit')) {
			return Stat.StatCritRating;
		} else {
			return null;
		}
	}

	static createDisplayStatArray(statList: Stat[], pseudoStatList: PseudoStat[]): UnitStat[] {
		return displayStatOrder.filter(
			displayStat =>
				(displayStat.isStat() && statList.includes(displayStat.stat!)) ||
				(displayStat.isPseudoStat() && pseudoStatList.includes(displayStat.pseudoStat!)),
		);
	}
}

const RATING_WEIGHTED_PERCENT_PSEUDO_STATS = new Map<PseudoStat, { root: Stat; perPercent: number }>(
	(getEnumValues(PseudoStat) as PseudoStat[]).flatMap(pseudoStat => {
		const root = UnitStat.getRootStat(pseudoStat);
		if (root === null || root === Stat.StatBlockValue) {
			return [];
		}
		return [[pseudoStat, { root, perPercent: UnitStat.fromStat(root).convertPercentToRating(1)! }]];
	}),
);

// The melee, ranged and spell percents one hit, crit or haste rating raises together, at perPercent rating
// per percent point. Each percent's weight measures its own channel alone: melee attacks, ranged attacks or
// spells. RatingChannelTable in sim/core/unit.go is the back end's copy of the same channels; the two are
// kept in sync by hand.
const RATING_CHANNELS = new Map<Stat, { channels: [PseudoStat, PseudoStat, PseudoStat]; perPercent: number }>([
	[
		Stat.StatHitRating,
		{
			channels: [PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatRangedHitPercent, PseudoStat.PseudoStatSpellHitPercent],
			perPercent: Mechanics.HIT_RATING_PER_HIT_PERCENT,
		},
	],
	[
		Stat.StatCritRating,
		{
			channels: [PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatRangedCritPercent, PseudoStat.PseudoStatSpellCritPercent],
			perPercent: Mechanics.CRIT_RATING_PER_CRIT_PERCENT,
		},
	],
	[
		Stat.StatHasteRating,
		{
			channels: [PseudoStat.PseudoStatMeleeHastePercent, PseudoStat.PseudoStatRangedHastePercent, PseudoStat.PseudoStatSpellHastePercent],
			perPercent: Mechanics.HASTE_RATING_PER_HASTE_PERCENT,
		},
	],
]);
const CHANNEL_PSEUDO_STATS = new Set([...RATING_CHANNELS.values()].flatMap(({ channels }) => channels));

const displayStatOrder: Array<UnitStat> = [
	UnitStat.fromStat(Stat.StatHealth),
	UnitStat.fromStat(Stat.StatMana),
	UnitStat.fromStat(Stat.StatArmor),
	UnitStat.fromStat(Stat.StatBonusArmor),
	UnitStat.fromStat(Stat.StatArmorNoMods),
	UnitStat.fromStat(Stat.StatStamina),
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatIntellect),
	UnitStat.fromStat(Stat.StatSpirit),
	UnitStat.fromStat(Stat.StatHealingPower),
	UnitStat.fromStat(Stat.StatSpellDamage),
	UnitStat.fromStat(Stat.StatArcaneDamage),
	UnitStat.fromStat(Stat.StatFireDamage),
	UnitStat.fromStat(Stat.StatFrostDamage),
	UnitStat.fromStat(Stat.StatHolyDamage),
	UnitStat.fromStat(Stat.StatNatureDamage),
	UnitStat.fromStat(Stat.StatShadowDamage),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFrost),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentNature),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellCritPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent),
	UnitStat.fromStat(Stat.StatMP5),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromStat(Stat.StatRangedAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent),
	UnitStat.fromStat(Stat.StatArmorPenetration),
	UnitStat.fromStat(Stat.StatSpellPiercing),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent),
	UnitStat.fromStat(Stat.StatDefenseRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent),
	UnitStat.fromStat(Stat.StatBlockValue),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent),
	UnitStat.fromStat(Stat.StatArcaneResistance),
	UnitStat.fromStat(Stat.StatFireResistance),
	UnitStat.fromStat(Stat.StatFrostResistance),
	UnitStat.fromStat(Stat.StatNatureResistance),
	UnitStat.fromStat(Stat.StatShadowResistance),
];

export interface StatSets {
	stats: Set<Stat>;
	pseudoStats: Set<PseudoStat>;
}

export const statSetsOf = (unitStats: ReadonlyArray<UnitStat>): StatSets => ({
	stats: new Set(unitStats.filter(unitStat => unitStat.isStat()).map(unitStat => unitStat.getStat())),
	pseudoStats: new Set(unitStats.filter(unitStat => unitStat.isPseudoStat()).map(unitStat => unitStat.getPseudoStat())),
});

export const unitStatsOf = ({ stats, pseudoStats }: StatSets): UnitStat[] => [
	...[...stats].map(stat => UnitStat.fromStat(stat)),
	...[...pseudoStats].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat)),
];

/**
 * Represents values for all character stats (stam, agi, spell power, hit raiting, etc).
 *
 * This is an immutable type.
 */
export class Stats {
	private readonly stats: Array<number>;
	private readonly pseudoStats: Array<number>;

	constructor(stats?: Array<number>, pseudoStats?: Array<number>) {
		this.stats = Stats.initStatsArray(STATS_LEN, stats);
		this.pseudoStats = Stats.initStatsArray(PSEUDOSTATS_LEN, pseudoStats);
	}

	private static initStatsArray(expectedLen: number, newStats?: Array<number>): Array<number> {
		let stats = newStats?.slice(0, expectedLen) || [];

		if (stats.length < expectedLen) {
			stats = stats.concat(new Array(expectedLen - (newStats?.length || 0)).fill(0));
		}

		for (let i = 0; i < expectedLen; i++) {
			if (stats[i] == null) stats[i] = 0;
		}
		return stats;
	}

	equals(other: Stats): boolean {
		return (
			this.stats.every((newStat, statIdx) => newStat == other.getStat(statIdx)) &&
			this.pseudoStats.every((newStat, statIdx) => newStat == other.getPseudoStat(statIdx))
		);
	}

	getStat(stat: Stat): number {
		return this.stats[stat];
	}
	getPseudoStat(stat: PseudoStat): number {
		return this.pseudoStats[stat];
	}
	getUnitStat(stat: UnitStat): number {
		if (stat.isStat()) {
			return this.stats[stat.getStat()];
		} else {
			return this.pseudoStats[stat.getPseudoStat()];
		}
	}

	withStat(stat: Stat, value: number): Stats {
		const newStats = this.stats.slice();
		newStats[stat] = value;
		return new Stats(newStats, this.pseudoStats);
	}
	withPseudoStat(stat: PseudoStat, value: number): Stats {
		const newStats = this.pseudoStats.slice();
		newStats[stat] = value;
		return new Stats(this.stats, newStats);
	}
	withUnitStat(stat: UnitStat, value: number): Stats {
		if (stat.isStat()) {
			return this.withStat(stat.getStat(), value);
		} else {
			return this.withPseudoStat(stat.getPseudoStat(), value);
		}
	}

	addStat(stat: Stat, value: number): Stats {
		return this.withStat(stat, this.getStat(stat) + value);
	}
	addPseudoStat(pseudoStat: PseudoStat, value: number): Stats {
		return this.withPseudoStat(pseudoStat, this.getPseudoStat(pseudoStat) + value);
	}

	add(other: Stats): Stats {
		return new Stats(
			this.stats.map((value, stat) => value + other.stats[stat]),
			this.pseudoStats.map((value, stat) => value + other.pseudoStats[stat]),
		);
	}

	subtract(other: Stats): Stats {
		return new Stats(
			this.stats.map((value, stat) => value - other.stats[stat]),
			this.pseudoStats.map((value, stat) => value - other.pseudoStats[stat]),
		);
	}

	scale(scalar: number): Stats {
		return new Stats(
			this.stats.map((value, _stat) => value * scalar),
			this.pseudoStats.map((value, _stat) => value * scalar),
		);
	}

	// The part of a rating's melee, ranged and spell percents every channel holds is worth the rating's
	// weight, or, where the weights leave the rating at 0, the sum of the channel weights they give.
	// What a channel holds beyond it is worth that percent's own weight. A channel the weights leave at
	// 0 falls back to what the rating is worth past the channels they weigh, and any other percent to
	// its rating's weight. The fallbacks of one rating count once, as the largest, since the one rating
	// raises them all.
	computeEP(epWeights: Stats): number {
		let total = 0;
		this.stats.forEach((stat, idx) => {
			total += stat * epWeights.stats[idx];
		});

		let fallbacks: Map<Stat, number> | undefined;
		const addPercent = (rating: Stat, pseudoStat: PseudoStat, value: number, fallbackPerPercent: number) => {
			if (value === 0) {
				return;
			}
			const weight = epWeights.pseudoStats[pseudoStat];
			if (weight !== 0) {
				total += value * weight;
				return;
			}
			fallbacks ??= new Map<Stat, number>();
			const ep = value * fallbackPerPercent;
			const counted = fallbacks.get(rating);
			fallbacks.set(rating, counted === undefined ? ep : Math.max(counted, ep));
		};

		RATING_CHANNELS.forEach(({ channels, perPercent: ratingPerPercent }, rating) => {
			const values = channels.map(pseudoStat => this.pseudoStats[pseudoStat]);
			const perPercent = ratingPerPercent * epWeights.stats[rating];
			const weighted = channels.map(pseudoStat => epWeights.pseudoStats[pseudoStat]).reduce((sum, weight) => sum + weight);
			const common = Math.max(0, Math.min(...values));
			total += common * (perPercent || weighted);
			const residual = Math.max(0, perPercent - weighted);
			channels.forEach((pseudoStat, i) => addPercent(rating, pseudoStat, values[i] - common, residual));
		});

		this.pseudoStats.forEach((value, idx) => {
			if (value === 0 || CHANNEL_PSEUDO_STATS.has(idx)) {
				return;
			}
			const rated = RATING_WEIGHTED_PERCENT_PSEUDO_STATS.get(idx);
			if (rated) {
				addPercent(rated.root, idx, value, rated.perPercent * epWeights.stats[rated.root]);
			} else {
				total += value * epWeights.pseudoStats[idx];
			}
		});

		fallbacks?.forEach(ep => {
			total += ep;
		});
		return total;
	}

	computeGapToCap(unitStat: UnitStat, cap: number): number {
		let statDelta = cap - this.getUnitStat(unitStat);

		if (unitStat.equalsPseudoStat(PseudoStat.PseudoStatMeleeHastePercent)) {
			statDelta /= this.getPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier);
		} else if (unitStat.equalsPseudoStat(PseudoStat.PseudoStatRangedHastePercent)) {
			statDelta /= this.getPseudoStat(PseudoStat.PseudoStatRangedSpeedMultiplier);
		} else if (unitStat.equalsPseudoStat(PseudoStat.PseudoStatSpellHastePercent)) {
			statDelta /= this.getPseudoStat(PseudoStat.PseudoStatCastSpeedMultiplier);
		}

		return statDelta == 0 ? 1e-12 : statDelta;
	}

	computeStatCapsDelta(statCaps: Stats): Stats {
		return new Stats(
			statCaps.stats.map((value, key) => {
				return value > 0 ? this.computeGapToCap(UnitStat.fromStat(key), value) : 0;
			}),
			statCaps.pseudoStats.map((value, key) => {
				return value > 0 ? this.computeGapToCap(UnitStat.fromPseudoStat(key), value) : 0;
			}),
		);
	}

	getBuffedStats(): Map<Stat, number> {
		const buffedStats = new Map<Stat, number>();

		for (const [statIdx, statValue] of this.stats.entries()) {
			if (statValue > 0) {
				buffedStats.set(statIdx, statValue);
			}
		}

		return buffedStats;
	}

	asProtoArray(): Array<number> {
		return this.stats.slice();
	}

	asUnitStatArray(): [UnitStat, number][] {
		const statValues = this.stats.map((value, key) => [UnitStat.fromStat(key), value] as [UnitStat, number]);
		const pseudoStatValues = this.pseudoStats.map((value, key) => [UnitStat.fromPseudoStat(key), value] as [UnitStat, number]);
		return statValues.concat(pseudoStatValues);
	}

	toJson(): object {
		return UnitStats.toJson(this.toProto()) as object;
	}

	toProto(): UnitStats {
		return UnitStats.create({
			stats: this.stats.slice(),
			pseudoStats: this.pseudoStats.slice(),
			apiVersion: CURRENT_API_VERSION,
		});
	}

	static fromJson(obj: any): Stats {
		return Stats.fromProto(UnitStats.fromJson(obj));
	}

	static fromMap(statsMap: Partial<Record<Stat, number>>, pseudoStatsMap?: Partial<Record<PseudoStat, number>>): Stats {
		const statsArr = new Array(STATS_LEN).fill(0);
		Object.entries(statsMap).forEach(entry => {
			const [statStr, value] = entry;
			statsArr[Number(statStr)] = value;
		});

		const pseudoStatsArr = new Array(PSEUDOSTATS_LEN).fill(0);
		if (pseudoStatsMap) {
			Object.entries(pseudoStatsMap).forEach(entry => {
				const [pseudoStatstr, value] = entry;
				pseudoStatsArr[Number(pseudoStatstr)] = value;
			});
		}

		return new Stats(statsArr, pseudoStatsArr);
	}

	static fromProto(unitStats?: UnitStats): Stats {
		if (unitStats) {
			return new Stats(unitStats.stats, unitStats.pseudoStats);
		} else {
			return new Stats();
		}
	}
}

// Folds a stat weights result down to a single EP number, weighted by the
// player's EP ratios (dps, hps, tps, dtps, tmi, pDeath — in that order).
export const scaledEpValue = (stat: UnitStat, epRatios: number[], result: StatWeightsResult | null): number => {
	if (!result) return 0;

	return (
		(result.dps?.epValues ? epRatios[0] * stat.getProtoValue(result.dps.epValues) : 0) +
		(result.hps?.epValues ? epRatios[1] * stat.getProtoValue(result.hps.epValues) : 0) +
		(result.tps?.epValues ? epRatios[2] * stat.getProtoValue(result.tps.epValues) : 0) +
		(result.dtps?.epValues ? epRatios[3] * stat.getProtoValue(result.dtps.epValues) : 0) +
		(result.tmi?.epValues ? epRatios[4] * stat.getProtoValue(result.tmi.epValues) : 0) +
		(result.pDeath?.epValues ? epRatios[5] * stat.getProtoValue(result.pDeath.epValues) : 0)
	);
};

// Used for spec specific stat presets to be used
// as a easy to access reference
export interface UnitStatPresets {
	unitStat: UnitStat;
	// Name of the preset and the value in percentage
	presets: Map<string, number>;
}

export interface BreakpointConfig {
	breakpoints: number[];
	capType: StatCapType;
	postCapEPs: number[];
}

// Represents a StatCapConfig proto message as a proper class for UI convenience
export class StatCap {
	readonly unitStat: UnitStat;
	readonly capType: StatCapType;
	breakpoints: number[] = [];
	postCapEPs: number[] = [];

	constructor(unitStat: UnitStat, breakpoints: number[], capType: StatCapType, postCapEPs: number[]) {
		// Check for valid inputs
		if (capType == StatCapType.TypeSoftCap && breakpoints.length != postCapEPs.length) {
			throw new Error('Breakpoint and EP counts do not match!');
		}
		if (capType != StatCapType.TypeSoftCap && capType != StatCapType.TypeThreshold) {
			throw new Error('Only SoftCap and Threshold cap types are supported currently!');
		}

		this.unitStat = unitStat;
		this.capType = capType;
		this.breakpoints = breakpoints;
		this.postCapEPs = postCapEPs;
	}

	static fromStat(stat: Stat, config: BreakpointConfig): StatCap {
		return new StatCap(UnitStat.fromStat(stat), config.breakpoints, config.capType, config.postCapEPs);
	}

	static fromPseudoStat(pseudoStat: PseudoStat, config: BreakpointConfig): StatCap {
		return new StatCap(UnitStat.fromPseudoStat(pseudoStat), config.breakpoints, config.capType, config.postCapEPs);
	}

	static fromProto(message: StatCapConfig[]): StatCap[] {
		const statCapObjects: StatCap[] = [];

		message
			.filter(config => config.unitStat)
			.forEach(config => {
				statCapObjects.push(new StatCap(UnitStat.fromProto(config.unitStat!), config.breakpoints, config.capType, config.postCapEPs));
			});

		return statCapObjects;
	}

	static cloneSoftCaps(softCaps: StatCap[]): StatCap[] {
		const clonedSoftCaps: StatCap[] = [];

		softCaps.forEach(config => {
			clonedSoftCaps.push(new StatCap(config.unitStat, config.breakpoints.slice(), config.capType, config.postCapEPs.slice()));
		});

		return clonedSoftCaps;
	}
}

// ---------------------------------------------------------------------------
// Stat attribution (moved out of components/character_stats.tsx so the math is
// UI-free and reusable).

export type StatMods = {
	base?: Stats;
	gear?: Stats;
	talents?: Stats;
	buffs?: Stats;
	consumes?: Stats;
	debuffs?: Stats;
	final?: Stats;
	stats?: Array<Stat>;
};
export type StatWrites = {
	base: Stats;
	gear: Stats;
	talents: Stats;
	buffs: Stats;
	consumes: Stats;
	debuffs: Stats;
	final: Stats;
	stats: Array<Stat>;
};

export interface StatAttribution {
	base: Stats;
	gear: Stats;
	talents: Stats;
	buffs: Stats;
	consumes: Stats;
	debuffs: Stats;
	final: Stats;
}

// Derives the per-source stat deltas shown in the character sheet from the
// cumulative server-computed stages (base ⊂ gear ⊂ talents ⊂ buffs ⊂ consumes).
// Debuffs are not one of those stages — the caller computes them from the raid's
// debuff settings and passes them in.
export function computeStatAttribution(
	playerStats: {
		baseStats?: UnitStats;
		gearStats?: UnitStats;
		talentsStats?: UnitStats;
		buffsStats?: UnitStats;
		consumesStats?: UnitStats;
		finalStats?: UnitStats;
	},
	bonusStats: Stats,
	debuffStats: Stats,
	statMods: StatMods,
	statOverwrites?: StatWrites,
): StatAttribution {
	const baseStats = Stats.fromProto(playerStats.baseStats);
	const gearStats = Stats.fromProto(playerStats.gearStats);
	const talentsStats = Stats.fromProto(playerStats.talentsStats);
	const buffsStats = Stats.fromProto(playerStats.buffsStats);
	const consumesStats = Stats.fromProto(playerStats.consumesStats);

	let finalStats = Stats.fromProto(playerStats.finalStats)
		.add(statMods.base || new Stats())
		.add(statMods.gear || new Stats())
		.add(statMods.talents || new Stats())
		.add(statMods.buffs || new Stats())
		.add(statMods.consumes || new Stats())
		.add(statMods.debuffs || new Stats())
		.add(statMods.final || new Stats())
		.add(debuffStats);

	let baseDelta = baseStats.add(statMods.base || new Stats());
	let gearDelta = gearStats
		.subtract(baseStats)
		.subtract(bonusStats)
		.add(statMods.gear || new Stats());
	let talentsDelta = talentsStats.subtract(gearStats).add(statMods.talents || new Stats());
	let buffsDelta = buffsStats.subtract(talentsStats).add(statMods.buffs || new Stats());
	let consumesDelta = consumesStats.subtract(buffsStats).add(statMods.consumes || new Stats());
	let debuffsDelta = debuffStats.add(statMods.debuffs || new Stats());

	if (statOverwrites?.stats) {
		statOverwrites.stats.forEach(stat => {
			baseDelta = baseDelta.withStat(stat, statOverwrites.base.getStat(stat));
			gearDelta = gearDelta.withStat(stat, statOverwrites.gear.getStat(stat));
			talentsDelta = talentsDelta.withStat(stat, statOverwrites.talents.getStat(stat));
			buffsDelta = buffsDelta.withStat(stat, statOverwrites.buffs.getStat(stat));
			consumesDelta = consumesDelta.withStat(stat, statOverwrites.consumes.getStat(stat));
			debuffsDelta = debuffsDelta.withStat(stat, statOverwrites.debuffs.getStat(stat));
			finalStats = finalStats.withStat(stat, statOverwrites.final.getStat(stat));
		});
	}

	return { base: baseDelta, gear: gearDelta, talents: talentsDelta, buffs: buffsDelta, consumes: consumesDelta, debuffs: debuffsDelta, final: finalStats };
}
