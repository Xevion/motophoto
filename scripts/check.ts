/**
 * Run all project checks in parallel.
 *
 * Usage: bun scripts/check.ts [--fix|-f] [--help|-h]
 */

import { c, elapsed, isStderrTTY, parseFlags } from './lib/fmt';
import { type CollectResult, raceInOrder, runPiped, spawnCollect } from './lib/proc';

const { flags } = parseFlags(
	process.argv.slice(2),
	{ fix: 'bool', help: 'bool' } as const,
	{ f: 'fix', h: 'help' },
	{ fix: false, help: false },
);

if (flags.help) {
	console.log(`Usage: bun scripts/check.ts [flags]

Runs all project checks in parallel.

Flags:
  -f, --fix     Format code first, then verify
  -h, --help    Show this help message and exit`);
	process.exit(0);
}

if (flags.fix) {
	console.log(c('1;36', '→ Fixing...'));
	runPiped(['gofmt', '-w', '.']);
	runPiped(['bun', 'run', '--cwd', 'web', 'lint:fix']);
	console.log(c('1;36', '→ Verifying...'));
}

interface Check {
	name: string;
	cmd: string[];
	cwd?: string;
	hint?: string;
	subsystem: 'frontend' | 'backend';
}

const checks: Check[] = [
	// Frontend checks
	{
		name: 'frontend-check',
		subsystem: 'frontend',
		cmd: ['bun', 'run', '--cwd', 'web', 'check']
	},
	{
		name: 'frontend-lint',
		subsystem: 'frontend',
		cmd: ['bun', 'run', '--cwd', 'web', 'lint']
	},
	// Backend checks
	{
		name: 'backend-vet',
		subsystem: 'backend',
		cmd: ['go', 'vet', './...']
	},
	{
		name: 'backend-build',
		subsystem: 'backend',
		cmd: ['go', 'build', '-o', '/dev/null', '.']
	},
	{
		name: 'backend-test',
		subsystem: 'backend',
		cmd: ['go', 'test', './...']
	},
];

const start = Date.now();
const remaining = new Set(checks.map((ch) => ch.name));

const promises = checks.map(async (check) => ({
	...check,
	...(await spawnCollect(check.cmd, start, { cwd: check.cwd }))
}));

const interval = isStderrTTY
	? setInterval(() => {
			const cols = process.stderr.columns || 80;
			const line = `${elapsed(start)}s [${Array.from(remaining).join(', ')}]`;
			process.stderr.write(`\r\x1b[K${line.length > cols ? line.slice(0, cols - 1) + '…' : line}`);
		}, 100)
	: null;

const results: Record<string, Check & CollectResult> = {};

await raceInOrder(promises, checks, (r) => {
	results[r.name] = r;
	remaining.delete(r.name);
	if (isStderrTTY) process.stderr.write('\r\x1b[K');

	const subsystemLabel = c('2', `[${r.subsystem}]`);
	if (r.exitCode !== 0) {
		process.stdout.write(c('31', `✗ ${r.name}`) + ` ${subsystemLabel} (${r.elapsed}s)\n`);
		if (r.hint) {
			process.stdout.write(c('2', `  ${r.hint}`) + '\n');
		} else {
			if (r.stdout) process.stdout.write(r.stdout);
			if (r.stderr) process.stderr.write(r.stderr);
		}
	} else {
		process.stdout.write(c('32', `✓ ${r.name}`) + ` ${subsystemLabel} (${r.elapsed}s)\n`);
	}
});

if (interval) clearInterval(interval);
if (isStderrTTY) process.stderr.write('\r\x1b[K');

const failed = Object.values(results).some((r) => r.exitCode !== 0);

process.exit(failed ? 1 : 0);
