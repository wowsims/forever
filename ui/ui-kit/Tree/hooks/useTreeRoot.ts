import { useContext } from 'react';

import { TreeRootContext } from '../context/TreeContext';

export const useTreeRoot = () => {
	const value = useContext(TreeRootContext);
	if (value === null) throw new Error('A Tree part must be rendered inside a Tree.Root.');
	return value;
};
