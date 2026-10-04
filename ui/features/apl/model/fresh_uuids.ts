import { randomUUID } from '@sim/utils/misc';

/**
 * Gives every value inside `message` a new uuid and returns it. A copy made with `clone` keeps its
 * original's uuids, and a uuid is a value's identity: its validations, its tree item and the group
 * of the operand list it owns are all keyed on it.
 */
export const freshUuids = <T>(message: T): T => {
	const visit = (node: unknown) => {
		if (!node || typeof node !== 'object') return;
		for (const [key, child] of Object.entries(node)) {
			if (key === 'uuid' && (child as { value?: string } | undefined)?.value) (node as { uuid: object }).uuid = { value: randomUUID() };
			else visit(child);
		}
	};
	visit(message);
	return message;
};
