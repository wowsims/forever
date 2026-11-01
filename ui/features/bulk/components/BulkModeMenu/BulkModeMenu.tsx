import type { BaseUIEvent } from '@base-ui/react/types';
import { BULK_TAB_ID, useActivateTab } from '@features/hooks/useActivateTab';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { BulkMode } from '@sim/state/sim_store';
import { Menu, MenuItem } from '@ui-kit/Menu';
import { type MouseEvent, type PointerEvent, type ReactElement, useRef, useState } from 'react';

import { useBulkMode } from '../../hooks/useBulkMode';
import { setBulkMode } from '../../model/mode';

export interface BulkModeMenuViewProps {
	/** The batch tab, which opens the menu on mouseover, on a tap and from the keyboard; a click opens the tab in its mode. */
	tab: ReactElement;
	onPick: (mode: BulkMode) => void;
}

export const BulkModeMenuView = ({ tab, onPick }: BulkModeMenuViewProps) => {
	const modes: ReadonlyArray<{ mode: BulkMode; label: string; description: string }> = [
		{ mode: 'gear', label: i18n.t('bulk_tab.modes.gear.label'), description: i18n.t('bulk_tab.modes.gear.description') },
		{ mode: 'builds', label: i18n.t('bulk_tab.modes.builds.label'), description: i18n.t('bulk_tab.modes.builds.description') },
	];

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
			align="start"
			sideOffset={-1}>
			{modes.map(({ mode, label, description }) => (
				<MenuItem
					key={mode}
					layout="block"
					testId={`bulk-mode-${mode}`}
					onClick={() => {
						setPicked(true);
						onPick(mode);
					}}>
					<span className="block text-sm font-medium">{label}</span>
					<span className="block text-xs text-muted">{description}</span>
				</MenuItem>
			))}
		</Menu>
	);
};

export const BulkModeMenu = ({ tab }: Pick<BulkModeMenuViewProps, 'tab'>) => {
	const player = usePlayer();
	const activate = useActivateTab();
	const { powerTools } = useBulkMode();
	if (!powerTools) return tab;

	return (
		<BulkModeMenuView
			tab={tab}
			onPick={mode => {
				setBulkMode(player, mode);
				activate(BULK_TAB_ID);
			}}
		/>
	);
};
