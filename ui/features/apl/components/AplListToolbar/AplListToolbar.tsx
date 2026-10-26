import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStickyBottom } from '@ui-kit/hooks/useStickyBottom';
import { Icon } from '@ui-kit/Icon';
import { Toolbar, ToolbarButton } from '@ui-kit/Toolbar';
import clsx from 'clsx';
import { useState } from 'react';

export interface AplListToolbarViewProps {
	/** Names the thing the button creates, e.g. "Action" or "Variable". */
	itemName: string;
	/**
	 * When given, the new button asks for a name first and hands it to `onCreate`. The dialog's
	 * title is the button's own label — both callers built the same string twice.
	 */
	nameDialog?: {
		inputLabel: string;
		existingNames: () => Array<string>;
	};
	onCreate: (name?: string) => void;
	/** When given, a paste button beside the new button appends the item the clipboard holds. */
	onPaste?: () => void;
	onReset: () => void;
	className?: string;
}

/**
 * The sticky "new / paste / reset" bar under an APL list; the paste button shows only when `onPaste` is given.
 */
export const AplListToolbarView = ({ itemName, nameDialog, onCreate, onPaste, onReset, className }: AplListToolbarViewProps) => {
	const newLabel = i18n.t('rotation_tab.apl.floatingActionBar.new', { itemName });
	const { ref: rootRef, stuck } = useStickyBottom<HTMLDivElement>();
	const [naming, setNaming] = useState(false);

	return (
		<Toolbar
			ref={rootRef}
			className={clsx(
				'sticky bottom-0 mt-2 flex items-center border-0 border-border transition-[padding,border-width,background-color] duration-150 ease-in-out data-stuck:border data-stuck:bg-background data-stuck:p-2',
				className,
			)}
			testId="apl-floating-action-bar-root"
			data-stuck={stuck ? '' : undefined}>
			<ToolbarButton variant="primary" onClick={() => (nameDialog ? setNaming(true) : onCreate())}>
				<Icon name="plus" />
				{newLabel}
			</ToolbarButton>
			{onPaste && (
				<ToolbarButton variant="outline-primary" className="ml-2" testId="apl-paste" onClick={onPaste}>
					<Icon name="paste" />
					{i18n.t('rotation_tab.apl.floatingActionBar.paste', { itemName })}
				</ToolbarButton>
			)}
			<ToolbarButton variant="link-danger" size="sm" className="ml-auto" onClick={onReset}>
				<Icon name="xmark" />
				{i18n.t('rotation_tab.apl.floatingActionBar.reset')}
			</ToolbarButton>
			{nameDialog && (
				<AplNameDialog
					open={naming}
					title={newLabel}
					inputLabel={nameDialog.inputLabel}
					existingNames={naming ? nameDialog.existingNames() : []}
					onSubmit={onCreate}
					onClose={() => setNaming(false)}
				/>
			)}
		</Toolbar>
	);
};

export type AplListToolbarProps = Omit<AplListToolbarViewProps, 'onReset'>;

export const AplListToolbar = (props: AplListToolbarProps) => {
	const host = useSimHost();
	return <AplListToolbarView {...props} onReset={() => host.applyEmptyAplRotation()} />;
};
