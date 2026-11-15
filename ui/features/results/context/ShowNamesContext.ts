import { createContext, useContext } from 'react';

export interface ShowNames {
	/** Whether the metrics tables name each action next to its icon. */
	shown: boolean;
	/** Folds the names out or back in. Absent from `lg` up, where the names always show. */
	toggle?: () => void;
}

export const NAMES_SHOWN: ShowNames = { shown: true };

export const ShowNamesContext = createContext<ShowNames>(NAMES_SHOWN);

export const useShowNames = (): ShowNames => useContext(ShowNamesContext);
