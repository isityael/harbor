const fs = require('node:fs');
const path = require('node:path');

const servicesDirectory = path.join(
    __dirname,
    '..',
    'ng-swagger-gen',
    'services'
);

for (const entry of fs.readdirSync(servicesDirectory, {
    withFileTypes: true,
})) {
    if (!entry.isFile() || !entry.name.endsWith('.ts')) {
        continue;
    }

    const file = path.join(servicesDirectory, entry.name);
    const source = fs.readFileSync(file, 'utf8');
    const updated = source.replace(
        /^module (\w+Service) \{$/gm,
        'namespace $1 {'
    );

    if (updated !== source) {
        fs.writeFileSync(file, updated);
    }
}
