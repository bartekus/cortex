import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const SCHEMAS_DIR = path.join(__dirname, '../../spec/schemas');
const COMMON_SCHEMA = 'common.schema.json';

// Tools categorized by type
const SNAPSHOT_ONLY_TOOLS = [
    'snapshot.create',
    'snapshot.info',
    'snapshot.changes',
    'snapshot.export',
];

const HYBRID_TOOLS = [
    'snapshot.list',
    'snapshot.file',
    'snapshot.grep',
    'snapshot.diff',
    'workspace.apply_patch',
];

interface LintError {
    file: string;
    message: string;
}

const errors: LintError[] = [];

function lintCommonSchema() {
    const commonPath = path.join(SCHEMAS_DIR, COMMON_SCHEMA);
    if (!fs.existsSync(commonPath)) {
        errors.push({ file: COMMON_SCHEMA, message: 'File not found' });
        return;
    }

    const content = JSON.parse(fs.readFileSync(commonPath, 'utf-8'));
    const errorEnum = content?.$defs?.error?.properties?.error?.properties?.code?.enum;

    if (!Array.isArray(errorEnum)) {
        errors.push({ file: COMMON_SCHEMA, message: 'Could not find error.code enum definition' });
        return;
    }

    if (!errorEnum.includes('STALE_LEASE')) {
        errors.push({ file: COMMON_SCHEMA, message: 'Error enum must include STALE_LEASE' });
    }
}

function lintResponseSchemas() {
    const allTools = [...SNAPSHOT_ONLY_TOOLS, ...HYBRID_TOOLS];

    for (const tool of allTools) {
        const filename = `${tool}.response.schema.json`;
        const filePath = path.join(SCHEMAS_DIR, filename);

        if (!fs.existsSync(filePath)) {
            errors.push({ file: filename, message: 'Schema file must exist' });
            continue;
        }

        const schema = JSON.parse(fs.readFileSync(filePath, 'utf-8'));

        // Check if it's OneOf (Failure + Success(es))
        if (!schema.oneOf || !Array.isArray(schema.oneOf)) {
            errors.push({ file: filename, message: 'Root must be oneOf array (Success | Error)' });
            continue;
        }

        const successBranches = schema.oneOf.filter((b: any) => !b['$ref'] || !b['$ref'].includes('error'));

        if (SNAPSHOT_ONLY_TOOLS.includes(tool)) {
            // Snapshot only: Should have exactly 1 success branch
            if (successBranches.length !== 1) {
                errors.push({ file: filename, message: `Snapshot-only tool must have exactly 1 success branch, found ${successBranches.length}` });
            } else {
                const branch = successBranches[0];
                validateCacheHint(filename, branch, 'immutable');
            }

        } else if (HYBRID_TOOLS.includes(tool)) {
            // Hybrid: Should have exactly 2 success branches (worktree + snapshot)
            if (successBranches.length !== 2) {
                errors.push({ file: filename, message: `Hybrid tool must have exactly 2 success branches, found ${successBranches.length}` });
            } else {
                let foundImmutable = false;
                let foundWorktree = false;

                for (const branch of successBranches) {
                    const hint = getCacheHintConst(branch);
                    if (hint === 'immutable') {
                        foundImmutable = true;
                    } else if (hint === 'until_dirty') {
                        foundWorktree = true;
                        validateWorktreeRequirements(filename, branch);
                    } else {
                        errors.push({ file: filename, message: `Unknown or missing cache_hint const value: ${hint}` });
                    }
                }

                if (!foundImmutable) errors.push({ file: filename, message: 'Missing "immutable" branch' });
                if (!foundWorktree) errors.push({ file: filename, message: 'Missing "until_dirty" (worktree) branch' });
            }
        }
    }
}

function getCacheHintConst(branch: any): string | undefined {
    // Traverse properties.cache_hint.const
    // Or if it's nested in $defs (less likely for response root, but possible)
    // We assume direct structure for now
    return branch?.properties?.cache_hint?.const;
}

function validateCacheHint(file: string, branch: any, expected: string) {
    const actual = getCacheHintConst(branch);
    if (actual !== expected) {
        errors.push({ file, message: `Expected cache_hint "${expected}", found "${actual}"` });
    }
}

function validateWorktreeRequirements(file: string, branch: any) {
    const required = branch.required || [];
    const props = branch.properties || {};

    if (!required.includes('lease_id') || !props.lease_id) {
        errors.push({ file, message: 'Worktree branch must require lease_id' });
    }
    if (!required.includes('fingerprint') || !props.fingerprint) {
        errors.push({ file, message: 'Worktree branch must require fingerprint' });
    }
}

function main() {
    console.log("Linting Schemas...");
    lintCommonSchema();
    lintResponseSchemas();

    if (errors.length > 0) {
        console.error("\nSchema Lint Failed:");
        for (const err of errors) {
            console.error(`[${err.file}] ${err.message}`);
        }
        process.exit(1);
    } else {
        console.log("Schema Lint Passed.");
    }
}

main();
