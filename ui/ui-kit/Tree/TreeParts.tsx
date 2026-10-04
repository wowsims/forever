import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useCallback, useContext, useId, useLayoutEffect, useState } from 'react';

import { TreeActionsContext, TreeBodyContext } from './context/TreeContext';
import { useTreeItem } from './hooks/useTreeItem';
import { useTreeRoot } from './hooks/useTreeRoot';
import { Root } from './Tree';
import type {
	TreeAddButtonProps,
	TreeItemActionProps,
	TreeItemActionsProps,
	TreeItemAsideProps,
	TreeItemBodyProps,
	TreeItemChildrenProps,
	TreeItemHandleProps,
	TreeItemHeaderProps,
	TreeItemTitleProps,
	TreeItemTriggerProps,
} from './types';

export const Header = ({ className, testId, children }: TreeItemHeaderProps) => {
	useTreeItem();
	return (
		<div className={clsx('ui-tree-item-header', className)} data-testid={testId ?? 'tree-item-header'}>
			{children}
		</div>
	);
};

export const Handle = ({ label, className, testId }: TreeItemHandleProps) => {
	const { disabled } = useTreeItem();
	return (
		<Button iconOnly className={clsx('ui-tree-item-handle', className)} aria-label={label} disabled={disabled} data-testid={testId}>
			<Icon name="grip-lines" size="sm" />
		</Button>
	);
};

export const Title = ({ as: Tag = 'span', className, testId, children }: TreeItemTitleProps) => {
	useTreeItem();
	return (
		<Tag className={clsx('ui-tree-item-title', className)} data-testid={testId}>
			{children}
		</Tag>
	);
};

export const Trigger = ({ label, className, testId }: TreeItemTriggerProps) => {
	const { open, setOpen, disabled, controls } = useTreeItem();
	return (
		<Button
			iconOnly
			className={clsx('ui-tree-item-trigger', className)}
			aria-label={label}
			aria-expanded={open}
			aria-controls={controls}
			disabled={disabled}
			data-testid={testId}
			onClick={() => setOpen(!open)}>
			<Icon name={open ? 'chevron-down' : 'chevron-right'} fixedWidth />
		</Button>
	);
};

export const Actions = ({ label, className, testId, children }: TreeItemActionsProps) => {
	const { disabled } = useTreeItem();
	const [open, setOpen] = useState(false);
	const close = useCallback(() => setOpen(false), []);
	const triggerTestId = testId ?? 'tree-item-actions';
	return (
		<Popover
			trigger={<Icon name="ellipsis" fixedWidth />}
			triggerRender={<Button iconOnly aria-label={label} />}
			triggerClassName={clsx('ui-tree-item-actions', className)}
			triggerTestId={triggerTestId}
			triggerProps={{ disabled }}
			open={open}
			onOpenChange={setOpen}
			side="bottom"
			align="end"
			className="ui-tree-item-actions-popup"
			testId={`${triggerTestId}-popup`}>
			<TreeActionsContext.Provider value={close}>{children}</TreeActionsContext.Provider>
		</Popover>
	);
};

export const Action = ({ label, icon, tone = 'default', disabled, onClick, className, testId }: TreeItemActionProps) => {
	const closeMenu = useContext(TreeActionsContext);
	return (
		<Button
			variant="unstyled"
			className={clsx('ui-tree-item-action', className)}
			data-tone={tone === 'danger' ? 'danger' : undefined}
			data-testid={testId ?? 'tree-item-action'}
			disabled={disabled}
			onClick={() => {
				closeMenu?.();
				onClick?.();
			}}>
			{icon && <Icon name={icon} fixedWidth />}
			{label}
		</Button>
	);
};

export const Aside = ({ className, testId, children }: TreeItemAsideProps) => {
	useTreeItem();
	return (
		<span className={clsx('ui-tree-item-aside', className)} data-testid={testId}>
			{children}
		</span>
	);
};

export const Body = ({ className, testId, children }: TreeItemBodyProps) => {
	const { open, bodyId, mountPart } = useTreeItem();
	useLayoutEffect(() => mountPart('body'), [mountPart]);
	return (
		<div id={bodyId} className={clsx('ui-tree-item-body', className)} hidden={!open} data-testid={testId}>
			<TreeBodyContext.Provider value>{children}</TreeBodyContext.Provider>
		</div>
	);
};

export const Children = ({ className, testId, children }: TreeItemChildrenProps) => {
	const { open, childrenProps, childrenId, mountPart } = useTreeItem();
	const { depth, group } = useTreeRoot();
	useLayoutEffect(() => mountPart('children'), [mountPart]);
	return (
		<div id={childrenId} className={clsx('ui-tree-item-children', className)} hidden={!open} data-testid={testId} {...childrenProps}>
			<Root depth={depth + 1} group={group}>
				{children}
			</Root>
		</div>
	);
};

export const AddButton = ({ label, icon, tone = 'default', disabled, onClick, className, testId }: TreeAddButtonProps) => {
	const root = useTreeRoot();
	const tooltipId = useId();
	const shared = {
		className: clsx('ui-tree-add', className),
		disabled: root.disabled || disabled,
		'data-tone': tone === 'success' ? 'success' : undefined,
		'data-testid': testId ?? 'tree-add',
		onClick,
	};
	if (!icon)
		return (
			<Button {...shared} className={clsx(shared.className, 'inline-flex items-baseline gap-2')}>
				<Icon name="plus" />
				{label}
			</Button>
		);
	return (
		<>
			<Button iconOnly aria-label={label} {...shared} {...tooltipAnchorProps(tooltipId, label)}>
				<Icon name={icon} fixedWidth />
			</Button>
			<Tooltip id={tooltipId} />
		</>
	);
};
