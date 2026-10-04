import { useContext } from 'react';

import { TreeItemsContext } from '../context/TreeContext';

export const useTreeItems = () => {
	const value = useContext(TreeItemsContext);
	if (value === null) throw new Error('A Tree.Item must be rendered inside a Tree.Items.');
	return value;
};
