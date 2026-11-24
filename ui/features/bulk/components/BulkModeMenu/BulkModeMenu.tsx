import type { BaseUIEvent } from '@base-ui/react/types';
import { BULK_TAB_ID, useActivateTab } from '@features/hooks/useActivateTab';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { BulkMode } from '@sim/state/sim_store';
import { Icon } from '@ui-kit/Icon';
import { Menu, MenuItem } from '@ui-kit/Menu';
import { type MouseEvent, type PointerEvent, type ReactElement, useRef, useState } from 'react';

import { useBulkMode } from '../../hooks/useBulkMode';
import { bulkMenuModes, setBulkMode } from '../../model/mode';

export interface BulkModeMenuViewProps {
	/** The batch tab, which opens the menu on mouseover, on a tap and from the keyboard; a click opens the tab in its mode. */
	tab: ReactElement;
	/** The modes the menu lists, in order. */
	modes: readonly BulkMode[];
	/** The mode the tab opens in, which the menu marks. */
	current: BulkMode;
	onPick: (mode: BulkMode) => void;
}

export const BulkModeMenuView = ({ tab, modes, current, onPick }: BulkModeMenuViewProps) => {
	// Read at `pointerdown` because a `mousedown` does not name its pointer type.
	const pointerType = useRef('');
	const tabRef = useRef<HTMLElement>(null);
	// A closing menu hands focus back to the element that held it before a hover opened it, and the tab strip opens
	// whichever tab takes focus. After a pick focus goes to this tab, so it stays open; any other close keeps the default.
	const [picked, setPicked] = useState(false);

	// `nativeButton`: the tab renders a native button, which the kit cannot tell from the element, and a non-native
	// trigger's role="button" would replace role="tab". Each item picks on click, so the current mode still opens the tab.
	// A press opens the tab in its mode like any other tab. A mouse press leaves the hover menu alone, since a toggle
	// would keep it open after the pointer leaves; a tap opens it too, having no hover to open it with.
	return (
		<Menu
			surface="menu"
			triggerRender={tab}
			onOpenChange={open => {
				if (open) setPicked(false);
			}}
			triggerProps={{
				ref: tabRef,
				openOnHover: true,
				delay: 0,
				nativeButton: true,
				className: 'data-popup-open:text-white',
				onPointerDown: (event: PointerEvent) => (pointerType.current = event.pointerType),
				onMouseDown: (event: BaseUIEvent<MouseEvent>) => {
					if (pointerType.current !== 'touch') event.preventBaseUIHandler();
				},
			}}
			popupProps={{ finalFocus: picked ? tabRef : undefined }}
			className="grid gap-1 py-1"
			align="start"
			sideOffset={-1}>
			{modes.map(mode => (
				<MenuItem
					key={mode}
					layout="block"
					className="grid gap-1 py-2"
					testId={`bulk-mode-${mode}`}
					aria-current={mode === current ? 'true' : undefined}
					data-checked={mode === current ? '' : undefined}
					onClick={() => {
						setPicked(true);
						onPick(mode);
					}}>
					<span className="flex items-center justify-between gap-4 text-sm font-medium">
						{i18n.t(`bulk_tab.modes.${mode}.label`)}
						{mode === current && <Icon name="check" className="text-primary" />}
					</span>
					<span className="text-xs text-muted">{i18n.t(`bulk_tab.modes.${mode}.description`)}</span>
				</MenuItem>
			))}
		</Menu>
	);
};

export const BulkModeMenu = ({ tab }: Pick<BulkModeMenuViewProps, 'tab'>) => {
	const player = usePlayer();
	const activate = useActivateTab();
	const { powerTools, runMode } = useBulkMode();

	return (
		<BulkModeMenuView
			tab={tab}
			modes={bulkMenuModes(powerTools)}
			current={runMode}
			onPick={mode => {
				setBulkMode(player, mode);
				activate(BULK_TAB_ID);
			}}
		/>
	);
};
