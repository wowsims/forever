import { createContext, useContext } from 'react';

/** Whether a metrics table names each action next to its icon: always from `lg` up, below it once the viewer folds the names out. */
export const ShowNamesContext = createContext(true);

export const useShowNames = (): boolean => useContext(ShowNamesContext);
