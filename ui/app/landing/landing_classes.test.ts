import { LaunchStatus } from '@sim/constants/other';
import { PlayerClasses } from '@sim/player/classes/index';
import { describe, expect, it } from 'vitest';

import specsCss from '../../styles/theme/specs.css?raw';
import { classLaunchStatus, LANDING_CLASS_ORDER, SPEC_THEME_CLASS } from './landing_classes';

const classes = () => LANDING_CLASS_ORDER.map(id => PlayerClasses.fromProto(id));

describe('the landing page class list', () => {
	it('covers every class exactly once', () => {
		expect(LANDING_CLASS_ORDER).toHaveLength(PlayerClasses.naturalOrder.length);
		expect(new Set(LANDING_CLASS_ORDER).size).toBe(LANDING_CLASS_ORDER.length);
	});

	it('sorts the classes by name', () => {
		const names = classes().map(playerClass => playerClass.friendlyName);
		expect(names).toEqual([...names].sort((a, b) => a.localeCompare(b)));
	});

	// The landing page is the only place all 18 spec links exist, so a class that lost its specs
	// would drop them silently.
	it('reaches all eighteen specs', () => {
		expect(classes().flatMap(playerClass => Object.values(playerClass.specs))).toHaveLength(18);
	});
});

describe('SPEC_THEME_CLASS', () => {
	it('names a theme class for every spec on the page, each with its own background in theme/specs.css', () => {
		const specs = classes().flatMap(playerClass => Object.values(playerClass.specs));
		specs.forEach(spec => {
			const themeClass = SPEC_THEME_CLASS[spec.specID as keyof typeof SPEC_THEME_CLASS];
			expect(themeClass, spec.friendlyName).toBeDefined();
			expect(specsCss, themeClass).toMatch(new RegExp(`\\.${themeClass} \\{[^}]*--theme-background-image: url\\(`));
		});
	});
});

describe('classLaunchStatus', () => {
	// A class whose specs sit at the given statuses; `classLaunchStatus` reads only `specs`.
	const classWith = (...statuses: LaunchStatus[]) =>
		({ specs: Object.fromEntries(statuses.map((status, i) => [i, { launch: { status } }])) }) as unknown as Parameters<typeof classLaunchStatus>[0];

	it('reports the furthest-along spec, so one unlaunched spec does not demote the class', () => {
		expect(classLaunchStatus(classWith(LaunchStatus.Unlaunched, LaunchStatus.Alpha))).toBe(LaunchStatus.Alpha);
		// A gear planner ranks above unlaunched and below alpha, so the class keeps its alpha label.
		expect(classLaunchStatus(classWith(LaunchStatus.GearPlanner, LaunchStatus.Alpha, LaunchStatus.Unlaunched))).toBe(LaunchStatus.Alpha);
		expect(classLaunchStatus(classWith(LaunchStatus.Unlaunched, LaunchStatus.Unlaunched))).toBe(LaunchStatus.Unlaunched);
	});
});
