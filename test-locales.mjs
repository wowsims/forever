import Ajv from 'ajv';
import path from 'node:path';
import fs from 'node:fs/promises';

const localesPath = 'assets/locales';

const ajv = new Ajv();

const __dirname = import.meta.dirname;

const loadSchemaFiles = async () => {
	const schemas = {};
	const schemaFiles = await fs.readdir(path.join(__dirname, 'schemas'));
	for (const file of schemaFiles) {
		const data = await fs.readFile(path.join(__dirname, 'schemas', file), 'utf-8');
		const schemaName = file.split('.')[0];
		schemas[schemaName] = JSON.parse(data);
	}
	return schemas;
};

const validateSchemas = async () => {
	const schemas = await loadSchemaFiles();
	let hasError = false;

	for (const [name, schema] of Object.entries(schemas)) {
		const validate = ajv.compile(schema);
		const filePaths = await Array.fromAsync(fs.glob(path.join(__dirname, localesPath, `**/${name}.json`)));

		for (const filePath of filePaths) {
			const relativePath = filePath.split(localesPath)[1].replace('/', '');
			const data = await fs.readFile(filePath, 'utf-8');

			const valid = validate(JSON.parse(data));

			if (valid) {
				console.log(`✅ ${relativePath} is valid`);
			} else {
				if (!hasError) hasError = true;
				console.log(
					`❌ ${relativePath} is invalid:`,
					ajv.errorsText(validate.errors, {
						dataVar: 'schema',
					}),
				);
			}
		}
	}
	if (hasError) process.exit(1);
};

validateSchemas();
