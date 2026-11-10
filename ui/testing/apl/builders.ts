import { APLAction, APLValue, APLValueCompare_ComparisonOperator as Op } from '@generated/proto/apl';

// No vitest import here: Storybook loads this module in the browser through the stories.

export const constant = (val: string) => APLValue.create({ value: { oneofKind: 'const', const: { val } } });

export const after = (val: string) =>
	APLValue.create({ value: { oneofKind: 'cmp', cmp: { op: Op.OpGt, lhs: { value: { oneofKind: 'currentTime', currentTime: {} } }, rhs: constant(val) } } });

export const waitFor = (val: string, condition?: APLValue) => APLAction.create({ condition, action: { oneofKind: 'wait', wait: { duration: constant(val) } } });

export const sequence = (actions: Array<APLAction>) => APLAction.create({ action: { oneofKind: 'sequence', sequence: { name: 'opener', actions } } });
