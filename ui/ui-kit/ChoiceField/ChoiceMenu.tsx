import { Menu as BaseMenu } from '@base-ui/react/menu';
import { useFieldDescribedBy } from '@ui-kit/Field';
import { Menu } from '@ui-kit/Menu';
import { type ReactNode, useMemo } from 'react';

import { ChoiceMenuItems } from './ChoiceMenuItems';
import type { ChoiceEntry, ChoiceFieldProps, ChoiceOption } from './types';
import { flattenChoices } from './utils';

interface ChoiceMenuListProps<V> {
	entries: ReadonlyArray<ChoiceEntry<V>> | (() => ReadonlyArray<ChoiceEntry<V>>);
	value: V;
	equals: (a: V, b: V) => boolean;
	onSelect: (value: V) => void;
	tooltipId: string;
}

/** Mounts with each opening, so a tree builder runs once per opening and never while the menu is closed. */
const ChoiceMenuList = <V,>({ entries, value, equals, onSelect, tooltipId, onClose }: ChoiceMenuListProps<V> & { onClose: () => void }) => {
	const { tree, flat, indexOf } = useMemo(() => {
		const tree = typeof entries === 'function' ? entries() : entries;
		const flat = flattenChoices(tree);
		const positions = new Map<ChoiceOption<V>, number>(flat.map((option, index) => [option, index]));
		return { tree, flat, indexOf: (option: ChoiceOption<V>) => positions.get(option)! };
	}, [entries]);
	const select = (index: number) => onSelect(flat[index].value);

	return (
		<BaseMenu.RadioGroup render={<ul />} className="m-0 list-none p-0" value={flat.findIndex(option => equals(option.value, value))} onValueChange={select}>
			<ChoiceMenuItems entries={tree} indexOf={indexOf} tooltipId={tooltipId} onSelect={select} onClose={onClose} />
		</BaseMenu.RadioGroup>
	);
};

export interface ChoiceMenuProps<V> extends ChoiceMenuListProps<V>, Pick<ChoiceFieldProps<V>, 'side' | 'positionMethod'> {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	trigger: ReactNode;
	triggerProps: Record<string, unknown>;
}

/** The list as a menu, for options filed under submenus, which a select has no part for. */
export const ChoiceMenu = <V,>({
	entries,
	value,
	equals,
	onSelect,
	open,
	onOpenChange,
	trigger,
	triggerProps,
	tooltipId,
	side,
	positionMethod,
}: ChoiceMenuProps<V>) => {
	// A Select's trigger gets the field's description ids from Base UI's own Field wiring, added to its own; a Menu's
	// trigger is not one of its field-aware controls, so they are joined here.
	const describedBy = [triggerProps['aria-describedby'], useFieldDescribedBy()].filter(Boolean).join(' ') || undefined;
	return (
		<Menu
			surface="menu"
			open={open}
			onOpenChange={onOpenChange}
			trigger={trigger}
			triggerProps={{ ...triggerProps, 'aria-describedby': describedBy }}
			align="start"
			side={side}
			positionMethod={positionMethod}
			sideOffset={2}
			portalProps={{ className: 'contents' }}
			popupRender={<div />}>
			{open && (
				<ChoiceMenuList entries={entries} value={value} equals={equals} onSelect={onSelect} tooltipId={tooltipId} onClose={() => onOpenChange(false)} />
			)}
		</Menu>
	);
};
