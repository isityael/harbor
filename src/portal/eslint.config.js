const angular = require('angular-eslint');
const fs = require('node:fs');
const path = require('node:path');
const prettierRecommended = require('eslint-plugin-prettier/recommended');
const tseslint = require('typescript-eslint');

const copyrightHeader = fs
    .readFileSync(path.join(__dirname, 'copyright.tmpl.js'), 'utf8')
    .trimEnd();

const headerPlugin = {
    rules: {
        header: {
            meta: {
                type: 'layout',
                fixable: 'whitespace',
                schema: [],
            },
            create(context) {
                return {
                    Program(node) {
                        if (
                            !context.sourceCode.text.startsWith(copyrightHeader)
                        ) {
                            context.report({
                                node,
                                message: 'missing copyright header',
                                fix: fixer =>
                                    fixer.insertTextBefore(
                                        node,
                                        `${copyrightHeader}\n\n`
                                    ),
                            });
                        }
                    },
                };
            },
        },
    },
};

module.exports = tseslint.config(
    {
        ignores: ['projects/**/*'],
    },
    {
        files: ['**/*.ts'],
        extends: [...angular.configs.tsRecommended],
        languageOptions: {
            parserOptions: {
                project: [
                    'server/tsconfig.json',
                    'tsconfig.json',
                    'cypress/tsconfig.json',
                ],
                tsconfigRootDir: __dirname,
            },
        },
        processor: angular.processInlineTemplates,
        rules: {
            'no-console': ['error', { allow: ['warn', 'error'] }],
            '@angular-eslint/prefer-standalone': 'off',
            '@angular-eslint/prefer-on-push-component-change-detection': 'off',
            '@angular-eslint/no-output-native': 'off',
            '@angular-eslint/prefer-inject': 'off',
        },
    },
    {
        files: ['**/*.html'],
        extends: [...angular.configs.templateRecommended],
        rules: {
            '@angular-eslint/template/prefer-control-flow': 'off',
            '@angular-eslint/template/prefer-self-closing-tags': 'off',
        },
    },
    {
        files: ['**/*.ts', '**/*.html'],
        plugins: prettierRecommended.plugins,
        rules: prettierRecommended.rules,
    },
    {
        files: ['**/*.html'],
        ignores: ['**/*inline-template-*.component.html'],
        rules: {
            'prettier/prettier': ['error', { parser: 'angular' }],
        },
    },
    {
        files: ['src/**/*.ts'],
        plugins: {
            header: headerPlugin,
        },
        rules: {
            'header/header': 'error',
        },
    }
);
