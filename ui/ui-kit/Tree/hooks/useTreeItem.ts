import { useContext } from 'react';

import { TreeItemContext } from '../context/TreeContext';

export const useTreeItem = () => {
	const value = useContext(TreeItemContext);
	if (value === null) throw new Error('A Tree.Item part must be rendered inside a Tree.Item.');
	return value;
};
