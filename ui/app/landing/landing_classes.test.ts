import { LaunchStatus } from '@sim/constants/other';
import { PlayerClasses } from '@sim/player/classes/index';
import { describe, expect, it } from 'vitest';

import { classLaunchStatus, LANDING_CLASS_ORDER } from './landing_classes';

const classes = () => LANDING_CLASS_ORDER.map(id => PlayerClasses.fromProto(id));

describe('the landing page class list', () => {
	it('covers every class exactly once', () => {
		expect(LANDING_CLASS_ORDER).toHaveLength(PlayerClasses.naturalOrder.length);
		expect(new Set(LANDING_CLASS_ORDER).size).toBe(LANDING_CLASS_ORDER.length);
	});

	// The landing page is the only place all 18 spec links exist, so a class that lost its specs
	// would drop them silently.
	it('reaches all eighteen specs', () => {
		expect(classes().flatMap(playerClass => Object.values(playerClass.specs))).toHaveLength(18);
	});
});

describe('classLaunchStatus', () => {
	it('reports the furthest-along spec, so one unlaunched spec does not demote the class', () => {
		expect(classLaunchStatus(PlayerClasses.Druid)).toBe(LaunchStatus.Alpha);
		expect(classLaunchStatus(PlayerClasses.Paladin)).toBe(LaunchStatus.Alpha);
		// A gear planner ranks above unlaunched and below alpha, so the class keeps its alpha label.
		expect(classLaunchStatus(PlayerClasses.Priest)).toBe(LaunchStatus.Alpha);
	});
});
