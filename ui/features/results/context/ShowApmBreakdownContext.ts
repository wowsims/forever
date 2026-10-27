import { createContext, useContext } from 'react';

export const ShowApmBreakdownContext = createContext(false);

export const useShowApmBreakdown = (): boolean => useContext(ShowApmBreakdownContext);
