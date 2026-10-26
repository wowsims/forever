import { createContext, useContext } from 'react';

export const HidePassivesContext = createContext(false);

export const useHidePassives = (): boolean => useContext(HidePassivesContext);
