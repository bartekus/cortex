import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';

export class RepoSetup {
    repoPath: string;

    constructor(baseDir: string, name: string) {
        this.repoPath = path.join(baseDir, name);
    }

    // Initialize a deterministic repo
    init() {
        if (fs.existsSync(this.repoPath)) {
            fs.rmSync(this.repoPath, { recursive: true, force: true });
        }
        fs.mkdirSync(this.repoPath, { recursive: true });

        this.git('init');
        // Configure user
        this.git('config user.name "Test User"');
        this.git('config user.email "test@example.com"');
        this.git('config commit.gpgsign false');
        // Deterministic dates?
        // We can force dates in commit env vars
    }

    git(args: string) {
        execSync(`git ${args}`, { cwd: this.repoPath, stdio: 'pipe' }); // Pipe or ignore to reduce noise
    }

    commit(msg: string, date: string = "2024-01-01T00:00:00Z") {
        // GIT_AUTHOR_DATE, GIT_COMMITTER_DATE
        const env = { ...process.env, GIT_AUTHOR_DATE: date, GIT_COMMITTER_DATE: date };
        execSync(`git commit -m "${msg}"`, { cwd: this.repoPath, env, stdio: 'pipe' });
    }

    writeFile(relPath: string, content: string) {
        const p = path.join(this.repoPath, relPath);
        fs.mkdirSync(path.dirname(p), { recursive: true });
        fs.writeFileSync(p, content);
    }
}
