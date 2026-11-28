// The categories a PresetBuild can carry. Lives here rather than beside the picker so ui/i18n
// does not have to import a UI component for its translation map.
export enum PresetConfigurationCategory {
	EPWeights = 'epWeights',
	Gear = 'gear',
	Talents = 'talents',
	Rotation = 'rotation',
	RotationType = 'rotationType',
	Encounter = 'encounter',
	Settings = 'settings',
}
