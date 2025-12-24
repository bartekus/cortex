
import * as fs from 'fs';
import * as path from 'path';

const ROOT_DIR = path.join(path.resolve(process.cwd()), '..');
const SCHEMAS_DIR = path.resolve(ROOT_DIR, 'spec/schemas');

interface Schema {
    $id?: string;
    oneOf?: any[];
    [key: string]: any;
}

function loadSchemas(): string[] {
    return fs.readdirSync(SCHEMAS_DIR)
        .filter(f => f.endsWith('.json') && f !== 'common.schema.json');
}

function lintSchema(filename: string) {
    const content = fs.readFileSync(path.join(SCHEMAS_DIR, filename), 'utf-8');
    const schema: Schema = JSON.parse(content);

    console.log(`Linting ${filename}...`);

    // Rule 1: Response schemas must use oneOf
    if (filename.includes('.response.')) {
        if (!schema.oneOf || !Array.isArray(schema.oneOf)) {
            throw new Error(`${filename}: Response schemas MUST use top-level 'oneOf'.`);
        }

        // Classify tool type
        const isHybrid = [
            'snapshot.list',
            'snapshot.file',
            'snapshot.grep',
            'snapshot.diff',
            'snapshot.export',
            'snapshot.changes',
            'workspace.apply_patch'
        ].some(tool => filename.includes(tool));

        const successBranches = schema.oneOf.filter(b => !b['$ref']);

        if (isHybrid) {
            // Hybrid tools: Expect Worktree and Snapshot success branches (or combined oneOf if implicit, but our spec says explicit oneOf branches preferred usually,
            // actually for hybrid we might have 2 success definitions in the oneOf array, plus error)
            // Our spec says: oneOf [WorktreeSuccess, SnapshotSuccess, Error]

            let worktreeBranch = successBranches.find(b =>
                b.properties?.mode?.const === 'worktree' ||
                (b.title && b.title.toLowerCase().includes(' worktree '))
            );
            let snapshotBranch = successBranches.find(b =>
                b.properties?.mode?.const === 'snapshot' ||
                (b.title && b.title.toLowerCase().includes(' snapshot '))
            );

            if (!worktreeBranch) throw new Error(`${filename}: Hybrid tool missing Worktree success branch.`);
            if (!snapshotBranch) throw new Error(`${filename}: Hybrid tool missing Snapshot success branch.`);

            // Validate Worktree Branch
            if (worktreeBranch.properties.cache_hint.const !== 'until_dirty') {
                throw new Error(`${filename}: Worktree branch must have cache_hint: "until_dirty".`);
            }
            if (!worktreeBranch.required.includes('lease_id')) {
                throw new Error(`${filename}: Worktree branch must require 'lease_id'.`);
            }
            if (!worktreeBranch.required.includes('fingerprint')) {
                throw new Error(`${filename}: Worktree branch must require 'fingerprint'.`);
            }

            // Validate Snapshot Branch
            if (snapshotBranch.properties.cache_hint.const !== 'immutable') {
                throw new Error(`${filename}: Snapshot branch must have cache_hint: "immutable".`);
            }

        } else {
            // Snapshot-only or simple tools
            // Should have 1 success branch
            if (successBranches.length !== 1) {
                throw new Error(`${filename}: Non-hybrid tool should have exactly 1 success branch (found ${successBranches.length}).`);
            }
            const success = successBranches[0];

            // workspace.write_file/delete don't have cache_hint enforced by spec in the same way, but let's check snapshot ones
            if (filename.includes('snapshot.')) {
                if (success.properties?.cache_hint?.const !== 'immutable') {
                    throw new Error(`${filename}: Snapshot-only tool must have cache_hint: "immutable".`);
                }
            }
        }

        // Check strictness on all success branches
        successBranches.forEach(branch => {
            if (branch.additionalProperties !== false) {
                throw new Error(`${filename}: Success branch must set additionalProperties: false.`);
            }
            if (filename.includes('snapshot.') && !branch.required.includes('cache_key')) {
                throw new Error(`${filename}: Success branch must require 'cache_key'.`);
            }
        });

    }
}

function main() {
    const schemas = loadSchemas();
    let errors = 0;

    for (const f of schemas) {
        try {
            lintSchema(f);
        } catch (e: any) {
            console.error(`❌ ${e.message}`);
            errors++;
        }
    }

    if (errors > 0) {
        console.error(`\nFound ${errors} schema lint errors.`);
        process.exit(1);
    } else {
        console.log("\n✅ All schemas passed linting.");
    }
}

main();
