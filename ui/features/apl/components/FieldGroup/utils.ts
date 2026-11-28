import type { AplFieldSpec } from '@features/apl/model/field_specs';
import type { InputConfig } from '@features/inputs/input';
import type { Player } from '@sim/player/player';
import type { StoreSubscribe } from '@sim/state/subscriptions';

const INLINE_KINDS: ReadonlySet<AplFieldSpec['kind']> = new Set(['boolean', 'number', 'string']);

/**
 * The binding one field of a kind's impl message gets.
 *
 * `getValue` is a mutation on a read path and it stays one, because the whole tree is built on
 * reading the live proto rather than a copy of it.
 */
export const fieldInputConfig = (
	spec: AplFieldSpec,
	id: string,
	getParentValue: () => any,
	changeSource: (player: Player<any>) => StoreSubscribe,
): InputConfig<Player<any>, any> & { id: string } => ({
	id,
	label: spec.label,
	labelTooltip: spec.labelTooltip,
	// A wide field stacks, so its input takes the field's own width instead of the inline one.
	layout: INLINE_KINDS.has(spec.kind) && !spec.wide ? 'inline' : undefined,
	extraClassNames: spec.wide ? ['w-64', 'max-w-full'] : undefined,
	storeSubscribe: changeSource,
	getValue: () => {
		const source = getParentValue();
		if (!source[spec.field]) source[spec.field] = spec.newValue();
		return source[spec.field];
	},
	setValue: (player: Player<any>, newValue: any) => {
		getParentValue()[spec.field] = newValue;
		player.touchRotation();
	},
});
