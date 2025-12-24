
import * as fs from 'fs';
import * as path from 'path';
import { spawn } from 'child_process';

const ROOT_DIR = path.join(path.resolve(process.cwd()), '..');
const TESTS_ROOT = path.resolve(process.cwd());
const BIN_PATH = path.join(ROOT_DIR, 'rust/target/debug/cortex-mcp');
const SUITE_PATH = path.join(TESTS_ROOT, 'golden/v1/suite.json');
const REPOS_DIR = path.join(TESTS_ROOT, 'fixtures/repos');
const TOY_REPO_ROOT = path.join(REPOS_DIR, 'toy-repo-01');

// Ensure binary exists
if (!fs.existsSync(BIN_PATH)) {
    console.error(`Binary not found at ${BIN_PATH}. Build it first!`);
    process.exit(1);
}

// Load Suite
const suite = JSON.parse(fs.readFileSync(SUITE_PATH, 'utf-8'));

// Prepare Replacements
const replacements: Record<string, string> = {
    "{{REPO_ROOT}}": TOY_REPO_ROOT
};

// State for dynamic replacements
const dynamicReplacements: Record<string, string> = {};

import { execSync } from 'child_process';

async function run() {
    console.log("Cleaning toy repo...");
    try {
        execSync('git clean -fd && git checkout .', { cwd: TOY_REPO_ROOT, stdio: 'inherit' });
    } catch (e) {
        console.error("Failed to clean repo:", e);
        process.exit(1);
    }

    console.log("Starting MCP Server...");

    const env = { ...process.env, CORTEX_WORKSPACE_ROOTS: REPOS_DIR };
    const proc = spawn(BIN_PATH, [], { env, stdio: ['pipe', 'pipe', 'inherit'] });

    let buffer = Buffer.alloc(0);

    proc.stdout.on('data', (chunk) => {
        buffer = Buffer.concat([buffer, chunk]);
        processBuffer();
    });

    let queue: any[] = [];
    let processing = false;

    // Helper to request next test
    function nextTest() {
        if (queue.length > 0) {
            const { resolve, reject, test } = queue.shift();
            // Prepare Request
            let reqStr = JSON.stringify(test.request);

            // Apply Static Replacements
            for (const [k, v] of Object.entries(replacements)) {
                reqStr = reqStr.split(k).join(v);
            }
            // Apply Dynamic Replacements
            for (const [k, v] of Object.entries(dynamicReplacements)) {
                // k is like {{LEASE_ID}}
                reqStr = reqStr.split(k).join(v);
            }

            const reqBuf = Buffer.from(reqStr, 'utf-8');
            const header = `Content-Length: ${reqBuf.length}\r\n\r\n`;

            try {
                proc.stdin.write(header);
                proc.stdin.write(reqBuf);
            } catch (e) {
                reject(e);
            }

            // Wait for response logic is handled by processBuffer which resolves promise
            activeRequest = { resolve, reject, test };
        } else {
            // Finished?
        }
    }

    let activeRequest: any = null;

    function processBuffer() {
        while (true) {
            // Check for header
            // Look for \r\n\r\n
            const headerEnd = buffer.indexOf('\r\n\r\n');
            if (headerEnd === -1) return; // Need more data

            const headerPart = buffer.slice(0, headerEnd).toString('utf-8');
            const match = headerPart.match(/Content-Length: (\d+)/i);
            if (!match) {
                console.error("Invalid header:", headerPart);
                process.exit(1);
            }
            const len = parseInt(match[1], 10);

            const bodyStart = headerEnd + 4;
            if (buffer.length < bodyStart + len) return; // Need more data body

            const body = buffer.slice(bodyStart, bodyStart + len).toString('utf-8');
            console.log("Raw Response for test:", activeRequest?.test?.name, body); // ADDED LOGGING
            buffer = buffer.slice(bodyStart + len);

            handleResponse(JSON.parse(body));
        }
    }

    function handleResponse(res: any) {
        if (!activeRequest) {
            // Maybe logging or unprompted notification?
            // "notifications/initialized" usually gets no response if it's a notification?
            return;
        }

        // Notifications might be interleaved?
        if (res.id === activeRequest.test.request.id) {
            activeRequest.resolve(res);
            activeRequest = null;
        } else {
            // Ignore non-matching ID or notifications??
        }
    }

    // Execute Suite
    for (const test of suite) {
        // If request has no id, it's a notification. send and continue.
        if (test.request.id === undefined) {
            // Just send
            let reqStr = JSON.stringify(test.request);
            const reqBuf = Buffer.from(reqStr, 'utf-8');
            const header = `Content-Length: ${reqBuf.length}\r\n\r\n`;
            proc.stdin.write(header);
            proc.stdin.write(reqBuf);
            // Wait small delay?
            await new Promise(r => setTimeout(r, 100));
            continue;
        }

        console.log(`Running test: ${test.name}`);

        const response: any = await new Promise((resolve, reject) => {
            queue.push({ resolve, reject, test });
            if (!activeRequest) nextTest();
        });

        // Validation
        if (test.response) {
            validateResponse(test.response, response);
        }
    }

    proc.kill();
    console.log("Suite passed!");
}

function validateResponse(expected: any, actual: any) {
    recursiveCompare(expected, actual, "");
}

function recursiveCompare(exp: any, act: any, path: string) {
    if (exp === "__IGNORE__") return;

    if (typeof exp === 'string') {
        if (exp.startsWith("{{") && exp.endsWith("}}")) {
            // Capture?
            let expectedVal = exp;
            for (const [k, v] of Object.entries(replacements)) {
                expectedVal = expectedVal.split(k).join(v);
            }

            if (dynamicReplacements[exp]) {
                if (act !== dynamicReplacements[exp]) {
                    throw new Error(`Mismatch at ${path}: expected ${dynamicReplacements[exp]} (bound to ${exp}), got ${act}`);
                }
                return;
            }

            dynamicReplacements[exp] = act;
            return;
        }

        let expectedVal = exp;
        for (const [k, v] of Object.entries(replacements)) {
            expectedVal = expectedVal.split(k).join(v);
        }

        if (act !== expectedVal) {
            throw new Error(`Mismatch at ${path}: expected '${expectedVal}', got '${act}'`);
        }
        return;
    }

    if (Array.isArray(exp)) {
        if (!Array.isArray(act)) throw new Error(`Mismatch at ${path}: expected array, got ${typeof act}`);
        if (exp.length !== act.length) throw new Error(`Mismatch at ${path}: array length expected ${exp.length}, got ${act.length}`);
        for (let i = 0; i < exp.length; i++) {
            recursiveCompare(exp[i], act[i], `${path}[${i}]`);
        }
        return;
    }

    if (typeof exp === 'object' && exp !== null) {
        if (typeof act !== 'object' || act === null) throw new Error(`Mismatch at ${path}: expected object, got ${act === null ? 'null' : typeof act}`); // IMPROVED ERROR
        for (const k of Object.keys(exp)) {
            if (!(k in act)) throw new Error(`Mismatch at ${path}: missing key ${k}`);
            recursiveCompare(exp[k], act[k], `${path}.${k}`);
        }
        return;
    }

    if (exp !== act) {
        throw new Error(`Mismatch at ${path}: expected ${JSON.stringify(exp)}, got ${JSON.stringify(act)}`); // IMPROVED ERROR
    }
}

run().catch(e => {
    console.error(e);
    process.exit(1);
});
