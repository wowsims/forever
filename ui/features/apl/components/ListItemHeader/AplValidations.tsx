import type { APLValidation } from '@generated/proto/api';
import { LogLevel } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { subscribePlayerField } from '@sim/state/subscriptions';
import { NoticeIcon, type NoticeLevel } from '@ui-kit/Notice';
import { Fragment, useEffect, useRef, useState } from 'react';

export interface AplValidationsViewProps {
	/** The item's validations, every spell id already resolved to its name. Empty hides the button. */
	validations: ReadonlyArray<APLValidation>;
}

const LEVEL = new Map<LogLevel, NoticeLevel>([
	[LogLevel.Information, 'info'],
	[LogLevel.Warning, 'warning'],
	[LogLevel.Error, 'error'],
]);

const HEADER = new Map<LogLevel, string>([
	[LogLevel.Information, i18n.t('common.list_picker.additional_information')],
	[LogLevel.Warning, i18n.t('common.list_picker.action_has_warnings')],
	[LogLevel.Error, i18n.t('common.list_picker.action_has_errors')],
]);

/** Written out rather than derived: `LogLevel` is a numeric enum, so iterating it yields the reverse mappings too. */
const LEVEL_CLASS = new Map<LogLevel, string>([
	[LogLevel.Undefined, 'apl-validation-undefined'],
	[LogLevel.Information, 'apl-validation-information'],
	[LogLevel.Warning, 'apl-validation-warning'],
	[LogLevel.Error, 'apl-validation-error'],
]);

const groupByLevel = (validations: ReadonlyArray<APLValidation>) => {
	const grouped = new Map<LogLevel, Array<string>>();
	let maxLogLevel = LogLevel.Undefined;
	for (const entry of validations) {
		maxLogLevel = Math.max(entry.logLevel, maxLogLevel);
		const messages = grouped.get(entry.logLevel);
		if (messages) messages.push(entry.validation);
		else grouped.set(entry.logLevel, [entry.validation]);
	}
	return { maxLogLevel, groups: Array.from(grouped, ([logLevel, messages]) => ({ header: HEADER.get(logLevel), messages })) };
};

/**
 * The warning triangle on an APL list item's header, and the grouped list its tooltip shows.
 *
 * The button is present but `display: none` while there is nothing to say, which keeps the header's
 * element count stable. A message is text, never markup: the sim formats rotation-supplied group and
 * variable names straight into it, so it reaches the tooltip as a text child and not as HTML.
 */
export const AplValidationsView = ({ validations }: AplValidationsViewProps) => {
	const shown = validations.length > 0;
	const { maxLogLevel, groups } = groupByLevel(validations);
	const level = (shown && LEVEL.get(maxLogLevel)) || 'warning';

	return (
		<NoticeIcon
			level={level}
			label={i18n.t('common.list_picker.warnings')}
			className={shown ? LEVEL_CLASS.get(maxLogLevel) : undefined}
			testId="apl-validations"
			hidden={!shown}>
			{groups.map((entry, index) => (
				<Fragment key={index}>
					<p>{entry.header}</p>
					<ul>
						{entry.messages.map((message, messageIndex) => (
							<li key={messageIndex}>{message}</li>
						))}
					</ul>
				</Fragment>
			))}
		</NoticeIcon>
	);
};

export interface AplValidationsProps {
	/** The validations for this one list item, read fresh on every `currentStats` notification. */
	getValidations: (player: Player<any>) => Array<APLValidation>;
}

const NONE: ReadonlyArray<APLValidation> = [];

const sameValidations = (a: Array<APLValidation>, b: Array<APLValidation>) =>
	a.length === b.length && a.every((entry, index) => entry.logLevel === b[index].logLevel && entry.validation === b[index].validation);

const resolveNames = (validations: Array<APLValidation>): Promise<Array<APLValidation>> =>
	Promise.all(validations.map(async entry => ({ ...entry, validation: await ActionId.replaceAllInString(entry.validation) })));

/**
 * The item's validations, read off `currentStats` and handed to the view once every spell id in them
 * is resolved to its name. The view shows nothing until then, so an unresolved id never reaches it.
 */
export const AplValidations = ({ getValidations }: AplValidationsProps) => {
	const player = usePlayer();
	const read = useRef(getValidations);
	read.current = getValidations;

	const subscribe = subscribePlayerField(player, 'currentStats');
	const last = useRef<Array<APLValidation>>([]);
	const validations = useStoreSubscribe(subscribe, () => {
		const next = read.current(player);
		if (!sameValidations(last.current, next)) last.current = next;
		return last.current;
	});

	const [resolved, setResolved] = useState(NONE);
	useEffect(() => {
		if (!validations.length) {
			setResolved(NONE);
			return;
		}
		let live = true;
		resolveNames(validations).then(next => {
			if (live) setResolved(next);
		});
		return () => {
			live = false;
		};
	}, [validations]);

	return <AplValidationsView validations={resolved} />;
};
