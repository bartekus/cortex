
import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';

const FIXTURES_DIR = path.resolve(process.cwd(), 'tests/fixtures/repos');
const REPO_NAME = 'toy-repo-01';
const REPO_PATH = path.join(FIXTURES_DIR, REPO_NAME);

// Ensure clean slate
if (fs.existsSync(REPO_PATH)) {
    fs.rmSync(REPO_PATH, { recursive: true, force: true });
}
fs.mkdirSync(REPO_PATH, { recursive: true });


function exec(cmd: string, cwd: string = REPO_PATH) {
    try {
        execSync(cmd, { cwd, stdio: 'pipe' });
    } catch (e: any) {
        console.error(`Command failed: ${cmd}`);
        if (e.stdout) console.error(`stdout: ${e.stdout.toString()}`);
        if (e.stderr) console.error(`stderr: ${e.stderr.toString()}`);
        throw e;
    }
}

console.log(`Creating deterministic repo at ${REPO_PATH}`);

// Initialize Git
exec('git init');
exec('git config user.name "Toy Robot"');
exec('git config user.email "toy@robot.com"');
exec('git config init.defaultBranch main'); // Force main

// Initial Commit
fs.writeFileSync(path.join(REPO_PATH, 'README.md'), '# Toy Repo 01\n');
exec('git add README.md');
exec('git commit -m "Initial commit"');

// Generate Deterministic Data
// ...

// Add some structure
fs.mkdirSync(path.join(REPO_PATH, 'src'));
fs.writeFileSync(path.join(REPO_PATH, 'src/main.rs'), 'fn main() {\n    println!("Hello");\n}\n');
fs.writeFileSync(path.join(REPO_PATH, 'src/utils.rs'), 'pub fn help() {}\n');
exec('git add src');
exec('git commit -m "Add source code"');

// Create a branch
exec('git checkout -b feature/a');
fs.writeFileSync(path.join(REPO_PATH, 'src/feature.rs'), '// Feature A\n');
exec('git add src/feature.rs');
exec('git commit -m "Add feature A"');

// Back to main
exec('git checkout main');

// Create a Tag
exec('git tag v0.1.0');

// Log the state
const head = execSync('git rev-parse HEAD', { cwd: REPO_PATH }).toString().trim();
console.log(`Repo created. HEAD: ${head}`);
