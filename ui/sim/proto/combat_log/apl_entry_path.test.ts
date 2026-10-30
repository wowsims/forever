import { RaidSimResult } from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

// Resolving an ActionId goes to the item/spell database; every assertion here is about the text
// around the id, so the log string stands in for the resolved action.
vi.mock('../action_id', () => {
	class ActionId {
		constructor(readonly logString: string) {}
		static fromLogString(logString: string) {
			return new ActionId(logString);
		}
		static async replaceAllInString(str: string) {
			return str;
		}
		async fill() {
			return this;
		}
		toString() {
			return this.logString;
		}
	}
	return { ActionId };
});

import { parseAplEntryPath } from './apl_entry_path';
import { buildCastLogs } from './derive';
import { parseAll } from './parse';
import type { CastBeganLog } from './types';

const CASTING = '[0.00] [Player (#1)] Casting {SpellID: 5308} (Cost = 15.000, Cast Time = 0s, GCD = 1.5s, Effective Time = 1.5s)';

const parseLine = async (line: string) => (await parseAll(RaidSimResult.create({ logs: line })))[0] as CastBeganLog;

describe('parseAplEntryPath', () => {
	it('reads 1-based log positions as 0-based segments', () => {
		expect(parseAplEntryPath('priority 6 > group 2.2 > step 3')).toEqual([
			{ kind: 'priority', index: 5 },
			{ kind: 'group', group: 1, index: 1 },
			{ kind: 'step', index: 2 },
		]);
		expect(parseAplEntryPath('prepull 1')).toEqual([{ kind: 'prepull', index: 0 }]);
	});

	it.each(['priority x', 'priority 0', 'group 2', 'group 1.x', 'group 1.2.3', 'bogus 1', 'priority 1 > ', 'priority 1 2'])('gives null for %j', text => {
		expect(parseAplEntryPath(text)).toBeNull();
	});
});

describe('Casting lines with an APL suffix', () => {
	it('parses the cost, times and entry path', async () => {
		const log = await parseLine(`${CASTING} APL: priority 6 > group 2.2 > step 3`);
		expect(log.kind).toBe('cast-began');
		expect(log.manaCost).toBe(15);
		expect(log.castTime).toBe(0);
		expect(log.gcd).toBe(1.5);
		expect(log.effectiveTime).toBe(1.5);
		expect(log.aplEntryPath).toEqual([
			{ kind: 'priority', index: 5 },
			{ kind: 'group', group: 1, index: 1 },
			{ kind: 'step', index: 2 },
		]);
	});

	it('parses a line without the suffix as before', async () => {
		const log = await parseLine(CASTING);
		expect(log.manaCost).toBe(15);
		expect(log.effectiveTime).toBe(1.5);
		expect(log.aplEntryPath).toBeNull();
	});

	it('gives a null entry path for a malformed suffix', async () => {
		expect((await parseLine(`${CASTING} APL: priority x`)).aplEntryPath).toBeNull();
	});

	it('carries the entry path onto the cast', async () => {
		const logs = await parseAll(RaidSimResult.create({ logs: `${CASTING} APL: priority 2\n[0.00] [Player (#1)] Completed cast {SpellID: 5308}` }));
		expect(buildCastLogs(logs)[0].aplEntryPath).toEqual([{ kind: 'priority', index: 1 }]);
	});
});
