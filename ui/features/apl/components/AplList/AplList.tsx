import { AplProvider } from '@features/apl/context/AplContext';
import { actionKinds, type APLActionKind } from '@features/apl/model/action_kinds';
import { type AplMove, applyMove, canApplyMove, type MoveSource, type OffersKind, resolveMoveSource } from '@features/apl/model/apply_move';
import type { APLRotation } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { Tree, type TreeLayout, type TreeSortable } from '@ui-kit/Tree';
import { type ReactNode, useMemo } from 'react';

export interface AplListRootProps {
	/** `APL_ROWS`, `APL_GROUPS` or `APL_VARIABLES`: which lists of the rotation the tree's moves land in. */
	group: string;
	/** Inside the action groups pane, the only place a variable placeholder may be used. */
	isGroup?: boolean;
	/** `inline` for a pane whose items are rows, as `Rows` lays out the rows under a section. */
	layout?: TreeLayout;
	testId?: string;
	children: ReactNode;
}

/**
 * One APL pane's tree. Its items are the pane's top-level entries; a drop anywhere under it, the
 * nested lists included, goes through `applyMove` as one rotation write.
 */
const Root = ({ group, isGroup, layout, testId, children }: AplListRootProps) => {
	const player = usePlayer();
	const sortable = useMemo<TreeSortable>(() => {
		const offersKind: OffersKind = (kind, isPrepull) => actionKinds[kind as NonNullable<APLActionKind>]?.includeIf?.(player, isPrepull) ?? true;
		// The dragged item's list is resolved once per drag, so a dragover looks up the target alone.
		let cached: { rotation: APLRotation; key: string; source: MoveSource | undefined } | null = null;
		const sourceOf = (from: AplMove['from']) => {
			const rotation = player.aplRotation;
			const key = `${from.group}\n${from.parentId}\n${from.index}\n${from.id}`;
			if (cached?.rotation !== rotation || cached.key !== key) cached = { rotation, key, source: resolveMoveSource(rotation, from) };
			return cached.source;
		};
		return {
			onMove: move => player.modifyAplRotation(rotation => applyMove(rotation, move, offersKind)),
			canDrop: (from, to) => canApplyMove(player.aplRotation, { from, to }, offersKind, sourceOf(from)),
		};
	}, [player]);

	return (
		<AplProvider isGroup={isGroup}>
			<Tree.Root group={group} sortable={sortable} density="compact" layout={layout} testId={testId}>
				<Tree.Items>{children}</Tree.Items>
			</Tree.Root>
		</AplProvider>
	);
};

/**
 * The rows under a section or a group, inside its `Tree.Item.Children`: each row's body sits on its
 * header line, and the rows stack once the tree is narrow. The root inherits the enclosing group,
 * and its items take the enclosing item as their parent, as the children list's own would.
 */
const Rows = ({ children }: { children: ReactNode }) => (
	<Tree.Root layout="inline" density="default">
		<Tree.Items>{children}</Tree.Items>
	</Tree.Root>
);

export interface AplListEntryProps {
	/** `aplItemId(item, index)`, or a section's fixed id. */
	id: string;
	/** The translated singular, e.g. "Action": names the item in the trigger and menu labels. */
	label: string;
	/** A name or a heading; an action row has none, so its body starts right after the handle. */
	title?: ReactNode;
	titleAs?: 'span' | 'h6';
	/** Status and toggles beside the trigger: the validations badge, the hide eye. */
	aside?: ReactNode;
	/** A drag handle; a section stays where it is. */
	movable?: boolean;
	/** Tree.Item.Action entries; without them the item has no menu. */
	actions?: ReactNode;
	disabled?: boolean;
	/** A row the eye has switched off: its handle and title fade with its body. */
	muted?: boolean;
	testId?: string;
	/** The item's `Tree.Item.Body` and/or `Tree.Item.Children`. */
	children?: ReactNode;
}

/** One APL entry: the kit's handle, title, aside, trigger and menu around the body or children the caller gives. */
const Item = ({ id, label, title, titleAs, aside, movable = true, actions, disabled, muted, testId, children }: AplListEntryProps) => (
	<Tree.Item id={id} disabled={disabled} muted={muted} testId={testId}>
		<Tree.Item.Header>
			{movable && <Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} />}
			{title !== undefined && <Tree.Item.Title as={titleAs}>{title}</Tree.Item.Title>}
			{aside !== undefined && <Tree.Item.Aside>{aside}</Tree.Item.Aside>}
			<Tree.Item.Trigger label={i18n.t('common.list_picker.toggle_item', { itemLabel: label })} />
			{actions && <Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel: label })}>{actions}</Tree.Item.Actions>}
		</Tree.Item.Header>
		{children}
	</Tree.Item>
);

export interface AplItemActionsProps {
	/** The translated singular the copy and delete labels name. */
	label: string;
	onCopy?: () => void;
	/** Copies the item as text that a paste in any rotation reads back. */
	onCopyToClipboard?: () => void;
	onExtract?: () => void;
	/** Read when the menu opens: the entry shows only for an item that holds a value to extract. */
	canExtract?: () => boolean;
	onDelete?: () => void;
	/** The entries' testids are `${testIdPrefix}-copy`, `-clipboard`, `-extract` and `-delete`. */
	testIdPrefix?: string;
}

/** The menu entries an APL item offers, in a fixed order: copy, copy to the clipboard, extract to a variable, delete. */
export const AplItemActions = ({ label, onCopy, onCopyToClipboard, onExtract, canExtract, onDelete, testIdPrefix = 'apl-item' }: AplItemActionsProps) => (
	<>
		{onCopy && (
			<Tree.Item.Action
				label={i18n.t('common.list_picker.copy_to_new', { itemLabel: label })}
				icon="copy"
				testId={`${testIdPrefix}-copy`}
				onClick={onCopy}
			/>
		)}
		{onCopyToClipboard && (
			<Tree.Item.Action
				label={i18n.t('common.list_picker.copy_to_clipboard', { itemLabel: label })}
				icon="clipboard"
				testId={`${testIdPrefix}-clipboard`}
				onClick={onCopyToClipboard}
			/>
		)}
		{onExtract && (canExtract?.() ?? true) && (
			<Tree.Item.Action
				label={i18n.t('rotation_tab.apl.variables.extractToVariable')}
				icon="arrow-right-from-bracket"
				testId={`${testIdPrefix}-extract`}
				onClick={onExtract}
			/>
		)}
		{onDelete && (
			<Tree.Item.Action
				label={i18n.t('common.list_picker.delete_item', { itemLabel: label })}
				icon="xmark"
				tone="danger"
				testId={`${testIdPrefix}-delete`}
				onClick={onDelete}
			/>
		)}
	</>
);

export const AplSectionTitle = ({ text, tooltip }: { text: string; tooltip: string }) => (
	<>
		{text}
		<TooltipButton tooltip={tooltip} />
	</>
);

export const AplList = { Root, Rows, Item };
