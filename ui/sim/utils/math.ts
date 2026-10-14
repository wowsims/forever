// Arithmetic and statistics: sums, standard deviation and confidence intervals
// for sim results, and the combinatorics the bulk sim enumerates with.

import { BULK_SIM_Z95 } from '../wasm/bulk_sim/constants_auto_gen';
import { STAT_WEIGHT_Z90 } from '../wasm/stat_weights_constants_auto_gen';
import { distinct } from './collections';

export function sum(arr: Array<number>): number {
	return arr.reduce((total, cur) => total + cur, 0);
}

export const percentOf = (part: number, whole: number) => (part / whole) * 100;

// What is left of `total` at the rate `completed` took `elapsed`, in `elapsed`'s unit; undefined before anything completes.
export const timeRemainingAtRate = (elapsed: number, completed: number, total: number): number | undefined =>
	completed > 0 ? (elapsed / completed) * Math.max(0, total - completed) : undefined;

export interface Aggregate {
	avg: number;
	n: number;
	sumSq: number;
}

// Two samples as one sample over both; n must be above 0.
export const mergeAggregates = (a: Aggregate, b: Aggregate): Aggregate & { stdev: number } => {
	const n = a.n + b.n;
	const avg = (a.avg * a.n + b.avg * b.n) / n;
	const sumSq = a.sumSq + b.sumSq;
	return { avg, n, sumSq, stdev: Math.sqrt(Math.max(0, sumSq / n - avg * avg)) };
};

export const Z_95 = BULK_SIM_Z95;
// The two-sided 90% interval, the ± shown behind each EP and the one Target ± sizes runs by.
export const Z_90 = STAT_WEIGHT_Z90;

// The sim's standard two-sample significance test; every "are these two results actually
// different" decision should go through this so thresholds never diverge.
export function zTest(
	n1: number,
	avg1: number,
	stdev1: number,
	n2: number,
	avg2: number,
	stdev2: number,
	preNormalized = false,
): { z: number; isDiff: boolean } {
	const delta = avg1 - avg2;
	const err1 = preNormalized ? stdev1 : stdev1 / Math.sqrt(n1);
	const err2 = preNormalized ? stdev2 : stdev2 / Math.sqrt(n2);
	const denom = Math.sqrt(Math.pow(err1, 2) + Math.pow(err2, 2));
	const z = Math.abs(delta / denom);
	return { z, isDiff: z > Z_95 };
}

export function stDevToConf90(stDev: number, N: number) {
	return (Z_90 * stDev) / Math.sqrt(N);
}

export function stDevToConf95(stDev: number, N: number) {
	return (Z_95 * stDev) / Math.sqrt(N);
}

export function permutations<T>(arr: Array<T>, k: number): Array<Array<T>> {
	if (k == 0) {
		return [];
	} else if (k == 1) {
		return arr.map(v => [v]);
	} else {
		return arr
			.map((v, i) => {
				const withoutThisElem = arr.slice();
				withoutThisElem.splice(i, 1);
				const permutationsWithoutThisElem = permutations(withoutThisElem, k - 1);
				return permutationsWithoutThisElem.map(perm => [v].concat(perm));
			})
			.flat();
	}
}

// Returns all N choose K combinations of the elements in arr of size N.
export function combinations<T>(arr: Array<T>, k: number, comparator?: (_a: T, _b: T) => number): Array<Array<T>> {
	const perms = permutations(arr, k);
	const sorted = perms.map(permutation => permutation.sort(comparator));

	const equals: (_a: T, _b: T) => boolean = comparator ? (a, b) => comparator(a, b) == 0 : (a, b) => a == b;
	return distinct(sorted, (permutationA, permutationB) => permutationA.every((elem, i) => equals(elem, permutationB[i])));
}

// Returns all N pick K permutations of the elements in arr of size N, allowing duplicates.
export function permutationsWithDups<T>(arr: Array<T>, k: number): Array<Array<T>> {
	if (k == 0) {
		return [];
	} else if (k == 1) {
		return arr.map(v => [v]);
	} else {
		const smaller = permutationsWithDups(arr, k - 1);
		return arr
			.map(v => {
				return smaller.map(permutation => {
					const newPerm = permutation.slice();
					newPerm.push(v);
					return newPerm;
				});
			})
			.flat();
	}
}

export const mod = (n: number, m: number): number => {
	return ((n % m) + m) % m;
};
