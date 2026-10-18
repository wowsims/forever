import { createContext, useContext } from 'react';

export const GroupRanksContext = createContext(true);

export const useGroupRanks = (): boolean => useContext(GroupRanksContext);
