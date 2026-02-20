/**
 * Run all project checks in parallel.
 *
 * Usage: bun scripts/check.ts [--fix|-f] [--help|-h]
 */

import { c, elapsed, isStderrTTY, parseFlags } from './lib/fmt';
import { type CollectResult, hasTool, raceInOrder, runPiped, spawnCollect, warnMissingTool } from './lib/proc';

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

function runStep(name: string, subsystem: 'frontend' | 'backend', cmd: string[], opts?: { cwd?: string }) {
	const t0 = Date.now();
	const result = runPiped(cmd, { cwd: opts?.cwd });
	const dt = ((Date.now() - t0) / 1000).toFixed(1);
	const subsystemLabel = c('2', `[${subsystem}]`);
	if (result.exitCode !== 0) {
		process.stdout.write(c('31', `✗ ${name}`) + ` ${subsystemLabel} (${dt}s)\n`);
		if (result.stdout) process.stdout.write(result.stdout);
		if (result.stderr) process.stderr.write(result.stderr);
		process.exit(1);
	}
	process.stdout.write(c('32', `✓ ${name}`) + ` ${subsystemLabel} (${dt}s)\n`);
}

// Detect available tools
const hasGo = hasTool("go");
const hasTygo = hasTool("tygo");
const hasLinter = hasTool("golangci-lint");
const hasGoimports = hasTool("goimports");

if (!hasGo) warnMissingTool("go", "skipping backend checks");
if (!hasTygo) warnMissingTool("tygo", "skipping binding generation");
if (!hasLinter) warnMissingTool("golangci-lint", "skipping backend lint");
if (!hasGoimports) warnMissingTool("goimports", "skipping Go import formatting");

if (flags.fix) {
	if (hasGoimports) runStep('fix-goimports', 'backend', ['goimports', '-w', '.']);
	runStep('fix-eslint', 'frontend', ['bun', 'run', '--cwd', 'web', 'lint:fix']);
	runStep('fix-biome', 'frontend', ['bun', 'run', '--cwd', 'web', 'format']);
}

// Generate TypeScript bindings from Go types before running checks
if (hasTygo) {
	runStep('generate-bindings', 'backend', ['tygo', 'generate']);
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
	{
		name: 'frontend-format',
		subsystem: 'frontend',
		cmd: ['bun', 'run', '--cwd', 'web', 'format:check']
	},
	// Backend checks (conditional on tool availability)
	...(hasGoimports ? [{
		name: 'backend-format',
		subsystem: 'backend' as const,
		cmd: ['bash', '-c', 'test -z "$(goimports -l .)"'],
		hint: 'Run `just format` or `goimports -w .` to fix formatting'
	}] : []),
	...(hasLinter ? [{
		name: 'backend-lint',
		subsystem: 'backend' as const,
		cmd: ['golangci-lint', 'run', '--timeout=5m']
	}] : []),
	...(hasGo ? [
		{
			name: 'backend-build',
			subsystem: 'backend' as const,
			cmd: ['go', 'build', '-o', '/dev/null', '.']
		},
		{
			name: 'backend-test',
			subsystem: 'backend' as const,
			cmd: ['go', 'test', './...']
		},
	] : []),
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
