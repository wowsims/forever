import { AplProvider } from '@features/apl/context/AplContext';
import { useAplReveal } from '@features/apl/hooks/useAplReveal';
import { actionKinds, type APLActionKind } from '@features/apl/model/action_kinds';
import { type AplMove, applyMove, canApplyMove, type MoveSource, type OffersKind, resolveMoveSource } from '@features/apl/model/apply_move';
import type { AplRevealRequest } from '@features/apl/model/reveal';
import type { APLAction, APLListItem, APLPrepullAction, APLRotation } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { Tree, type TreeDensity, type TreeLayout, type TreeSortable, useTreeItem } from '@ui-kit/Tree';
import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react';

export interface AplListRootProps {
	/** `APL_ROWS`, `APL_GROUPS` or `APL_VARIABLES`: which lists of the rotation the tree's moves land in. */
	group: string;
	/** Inside the action groups pane, the only place a variable placeholder may be used. */
	isGroup?: boolean;
	/** `inline` for a pane whose items are rows, as `Rows` lays out the rows under a section. */
	layout?: TreeLayout;
	/** `compact` (the default here, unlike the kit's) leaves the items unframed; `default` boxes each, as the rows under a section are. */
	density?: TreeDensity;
	testId?: string;
	children: ReactNode;
}

/**
 * One APL pane's tree. Its items are the pane's top-level entries; a drop anywhere under it, the
 * nested lists included, goes through `applyMove` as one rotation write.
 */
const Root = ({ group, isGroup, layout, density = 'compact', testId, children }: AplListRootProps) => {
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
			<Tree.Root group={group} sortable={sortable} density={density} layout={layout} testId={testId}>
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
	/** For a section or a group: whether an entry a cast's click opens sits under it, so the item opens. */
	revealOn?: (request: AplRevealRequest) => boolean;
	testId?: string;
	/** The item's `Tree.Item.Body` and/or `Tree.Item.Children`. */
	children?: ReactNode;
}

const RevealOpen = ({ on }: { on: (request: AplRevealRequest) => boolean }) => {
	const { setOpen } = useTreeItem();
	useAplReveal(request => {
		if (on(request)) setOpen(true);
	});
	return null;
};

/** One APL entry: the kit's handle, title, aside, trigger and menu around the body or children the caller gives. */
const Item = ({ id, label, title, titleAs, aside, movable = true, actions, disabled, muted, revealOn, testId, children }: AplListEntryProps) => (
	<Tree.Item id={id} disabled={disabled} muted={muted} testId={testId}>
		{revealOn && <RevealOpen on={revealOn} />}
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

const FLASH_MS = 1500;

export interface AplRevealTargetProps {
	/** The rotation's own row or step inside, read when a cast's click comes. */
	getItem: () => APLListItem | APLPrepullAction | APLAction | undefined;
	children: ReactNode;
}

/**
 * The body of a row or a sequence step, inside its `Tree.Item.Body`. When a cast's click opens an entry
 * through it, it opens its item; when the entry is this one, it scrolls into view and flashes. The
 * scroll waits a frame, for the tab, the sub-tab and the items above it to have opened.
 */
export const AplRevealTarget = ({ getItem, children }: AplRevealTargetProps) => {
	const { setOpen } = useTreeItem();
	const ref = useRef<HTMLDivElement>(null);
	const [flash, setFlash] = useState(false);
	const [reveals, setReveals] = useState(0);
	useAplReveal(({ item, steps }) => {
		const own = getItem();
		const step = steps.indexOf(own as APLAction);
		if (own !== item && step === -1) return;
		setOpen(true);
		if (own === item ? steps.length === 0 : step === steps.length - 1) setReveals(count => count + 1);
	});
	useEffect(() => {
		if (!reveals) return;
		const frame = requestAnimationFrame(() => {
			ref.current?.scrollIntoView({ block: 'center' });
			setFlash(true);
		});
		const timer = setTimeout(() => setFlash(false), FLASH_MS);
		return () => {
			cancelAnimationFrame(frame);
			clearTimeout(timer);
			setFlash(false);
		};
	}, [reveals]);
	return (
		<div
			ref={ref}
			data-flash={flash ? '' : undefined}
			className="outline-2 outline-transparent transition-[background-color,outline-color] duration-500 outline-solid data-flash:bg-primary-dampened data-flash:outline-primary">
			{children}
		</div>
	);
};

export const AplSectionTitle = ({ text, tooltip }: { text: string; tooltip: string }) => (
	<>
		{text}
		<TooltipButton tooltip={tooltip} />
	</>
);

export const AplList = { Root, Rows, Item };
