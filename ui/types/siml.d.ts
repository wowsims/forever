// What tools/vite/siml.mts gives an import.

// The JSON of the message the file's kind loads into (sim/siml/SPEC.md §1): an APLRotation for x.apl.siml, an
// EquipmentSpec for x.gear.siml and an IndividualSimSettings for any other file. `any`, as the preset helpers take the
// message's JSON and build the message from it.
declare module '*.siml' {
	const json: any;
	export default json;
}
