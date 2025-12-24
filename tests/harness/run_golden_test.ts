import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { exec, execSync, spawn } from 'child_process';
import { RepoSetup } from './mutation.ts';

// Feature: MCP_SNAPSHOT_WORKSPACE_SUBSTRATE
// Spec: spec/mcp/snapshot-workspace-v1.md

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Configuration
const UPDATE_GOLDEN = process.env.UPDATE_GOLDEN === '1';
const GOLDEN_DIR = path.resolve(__dirname, '../golden');
const FIXTURES_DIR = path.resolve(__dirname, '../fixtures/run');
const REPO_ROOT = path.join(FIXTURES_DIR, 'toy-repo-01');

// Setup environment
if (fs.existsSync(FIXTURES_DIR)) {
    fs.rmSync(FIXTURES_DIR, { recursive: true, force: true });
}
fs.mkdirSync(FIXTURES_DIR, { recursive: true });

// 1. Schema Lint
console.log("Running Schema Lint...");
const SCHEMA_LINT_PATH = path.join(__dirname, 'schema_lint_test.ts');
try {
    execSync(`npx -y ts-node "${SCHEMA_LINT_PATH}"`, { stdio: 'inherit' });
} catch (e) {
    console.error("Schema lint failed.");
    process.exit(1);
}

// 2. Setup Repo
console.log("Initializing deterministic repo...");
const repo = new RepoSetup(FIXTURES_DIR, 'toy-repo-01');
repo.init();
repo.writeFile('src/main.rs', 'fn main() {}');
repo.writeFile('README.md', '# Toy Repo');
repo.git('add .');
repo.commit('Initial commit', '2024-01-01T00:00:00Z');

// 3. Build Server
console.log("Building MCP server...");
execSync('cargo build -p cortex-mcp', {
    cwd: path.resolve(__dirname, '../../rust/mcp'),
    stdio: 'inherit'
});
const SERVER_BIN = path.resolve(__dirname, '../../rust/target/debug/cortex-mcp');

// 4. Run Tests
async function run() {
    // Collect tests
    const files = fs.readdirSync(GOLDEN_DIR).filter(f => f.endsWith('.json'));
    files.sort();

    for (const file of files) {
        console.log(`Running golden test: ${file}`);
        await runTestFile(file);
    }
}

async function runTestFile(file: string) {
    const filePath = path.join(GOLDEN_DIR, file);
    const testCase = JSON.parse(fs.readFileSync(filePath, 'utf8'));

    // Spawn server
    const proc = spawn(SERVER_BIN, [], {
        env: { ...process.env, RUST_LOG: 'debug' }, // Optional logging
    });

    let stdoutBuffer = Buffer.alloc(0);
    let pendingRequests: any[] = [];

    proc.stderr.on('data', (d) => {
        // console.error(`SERVER STDERR: ${d}`);
    });

    proc.stdout.on('data', (d) => {
        stdoutBuffer = Buffer.concat([stdoutBuffer, d]);
        processBuffer();
    });

    function processBuffer() {
        while (true) {
            // Find header delimiter \r\n\r\n
            const idx = stdoutBuffer.indexOf('\r\n\r\n');
            if (idx === -1) break;

            const headerPart = stdoutBuffer.subarray(0, idx).toString();
            // Parse content length
            const match = headerPart.match(/Content-Length: (\d+)/i);
            if (!match) {
                console.error("Invalid header: " + headerPart);
                process.exit(1);
            }
            const len = parseInt(match[1]);
            const bodyStart = idx + 4;

            if (stdoutBuffer.length < bodyStart + len) {
                // Wait for more data
                return;
            }

            const body = stdoutBuffer.subarray(bodyStart, bodyStart + len).toString();
            stdoutBuffer = stdoutBuffer.subarray(bodyStart + len);

            handleMessage(JSON.parse(body));
        }
    }

    let requestIdx = 0;

    // Send Input[0]
    sendNext();

    function sendNext() {
        if (requestIdx >= testCase.interactions.length) {
            proc.kill();
            return;
        }
        const interaction = testCase.interactions[requestIdx];
        const req = interaction.request;
        // Inject dynamic values if needed (REPO_ROOT)
        const reqStr = JSON.stringify(req).replace(/__REPO_ROOT__/g, REPO_ROOT);

        const payload = Buffer.from(reqStr);
        proc.stdin.write(`Content-Length: ${payload.length}\r\n\r\n`);
        proc.stdin.write(payload);
    }

    function handleMessage(msg: any) {
        const interaction = testCase.interactions[requestIdx];

        // Normalize response
        const normalized = normalize(msg);

        if (UPDATE_GOLDEN) {
            interaction.response = normalized;
        } else {
            // Validate
            if (JSON.stringify(normalized) !== JSON.stringify(interaction.response)) {
                console.error(`Mismatch in ${file} request ${requestIdx}`);
                console.error("Expected:", JSON.stringify(interaction.response, null, 2));
                console.error("Actual:", JSON.stringify(normalized, null, 2));
                process.exit(1);
            }
        }

        requestIdx++;
        if (requestIdx < testCase.interactions.length) {
            sendNext();
        } else {
            proc.kill();
            finish();
        }
    }

    function finish() {
        if (UPDATE_GOLDEN) {
            fs.writeFileSync(filePath, JSON.stringify(testCase, null, 4));
            console.log("Updated golden file.");
        }
    }

    // Wait for process exit?
    await new Promise<void>(resolve => {
        proc.on('close', resolve);
    });
}

function normalize(obj: any): any {
    // Recursive normalization
    // 1. UUIDs -> valid fake UUIDs? Or just canonicalize?
    // 2. Dates
    // 3. REPO_ROOT -> __REPO_ROOT__

    const str = JSON.stringify(obj);
    let replaced = str.replace(new RegExp(escapeRegExp(REPO_ROOT), 'g'), '__REPO_ROOT__');

    // Normalize UUIDs (simple regex for v4)
    // We can't consistently map random UUIDs without state.
    // But `lease_id` is random.
    // We should mask it: "lease_id": "UUID-STUB" if it looks like uuid?
    // Regex: [0-9a-f]{8}-[0-9a-f]{4}-...
    // Only verify format?

    // For golden tests, hard to match exact UUID.
    // We update the expected to match the pattern or use a stable replacement if we track it.
    // Simple approach: Replace any UUID with "<UUID>" in the string?
    // But we might have multiple.

    const uuidRegex = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/g;
    replaced = replaced.replace(uuidRegex, '<UUID>');

    // Also timestamps if any? Fingerprints have shas.
    // Git OIDs depend on commit time/author.
    // Our RepoSetup forces date. So SHAs should be STABLE!
    // That's the key "Determinism".

    return JSON.parse(replaced);
}

function escapeRegExp(string: string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

run().catch(e => {
    console.error(e);
    process.exit(1);
});
