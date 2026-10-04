import mainCss from '../../styles/main.css?raw';
import { parseCssRules } from './ui_classes';

export interface CustomUtility {
	name: string;
	text: string;
}

/** Every `@utility` block in `main.css`, in file order. */
export const UTILITIES: ReadonlyArray<CustomUtility> = parseCssRules(mainCss)
	.filter(rule => rule.selector.startsWith('@utility '))
	.map(rule => ({ name: rule.selector.slice('@utility '.length).trim(), text: rule.text }));
