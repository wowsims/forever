import i18n from '@i18n/config';
import type { Entity } from '@sim/proto/combat_log';

export interface EntityLabelProps {
	entity: Entity;
}

export const EntityLabel = ({ entity }: EntityLabelProps) => {
	if (entity.isTarget) {
		return <span className="text-danger">[{i18n.t('results_tab.details.target_number', { number: entity.index + 1 })}]</span>;
	}

	if (entity.isPet) {
		return (
			<>
				<span className="text-primary">{`[${entity.ownerName} ${entity.index + 1}]`}</span>
				{` - `}
				{entity.name}
			</>
		);
	}

	return <span className="text-primary">{`[${entity.name} ${entity.index + 1}]`}</span>;
};
