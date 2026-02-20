import { type Subprocess, spawn } from 'bun';

const PORT = process.env.PORT || '8080';
const BACKEND_PORT = '3001';
const BACKEND_URL = `http://localhost:${BACKEND_PORT}`;

// Logging defaults: JSON in production (entrypoint only runs in Docker).
// Both the Go backend and SvelteKit frontend read LOG_JSON identically,
// so we normalize once here and propagate via env to both subprocesses.
const LOG_JSON = process.env.LOG_JSON ?? 'true';
const LOG_LEVEL = process.env.LOG_LEVEL;

type LogLevel = 'info' | 'warn' | 'error' | 'debug';

function log(level: LogLevel, message: string, fields?: Record<string, unknown>) {
	if (LOG_JSON === 'true' || LOG_JSON === '1') {
		const entry = {
			timestamp: new Date().toISOString(),
			level,
			target: 'motophoto::entrypoint',
			message,
			...fields
		};
		const out = level === 'error' ? process.stderr : process.stdout;
		out.write(JSON.stringify(entry) + '\n');
	} else {
		const prefix = level === 'error' ? 'ERROR: ' : '';
		const suffix = fields
			? ` ${Object.entries(fields)
					.map(([k, v]) => `${k}=${v}`)
					.join(' ')}`
			: '';
		const out = level === 'error' ? console.error : console.log;
		out(`${prefix}${message}${suffix}`);
	}
}

const sharedEnv: Record<string, string | undefined> = {
	...process.env,
	LOG_JSON
};
if (LOG_LEVEL) {
	sharedEnv.LOG_LEVEL = LOG_LEVEL;
}

log('info', 'Starting Go backend');
const goProc = spawn({
	cmd: ['/app/motophoto'],
	env: { ...sharedEnv, PORT: BACKEND_PORT },
	stdout: 'inherit',
	stderr: 'inherit'
});

// Wait for backend to be healthy (15s timeout)
const startTime = Date.now();
let healthy = false;
while (!healthy) {
	if (Date.now() - startTime > 15_000) {
		log('error', 'Go backend failed to become healthy within 15s');
		goProc.kill();
		process.exit(1);
	}

	try {
		const response = await fetch(`${BACKEND_URL}/api/health`);
		if (response.ok) {
			healthy = true;
		}
	} catch {
		// Backend not ready yet
	}

	if (!healthy) {
		await Bun.sleep(250);
	}
}
log('info', 'Go backend is healthy');

log('info', 'Starting SvelteKit SSR', { host: '0.0.0.0', port: PORT });
const bunProc = spawn({
	cmd: ['bun', 'build/index.js'],
	cwd: '/app/web',
	env: {
		...sharedEnv,
		PORT,
		HOST: '0.0.0.0',
		BACKEND_URL
	},
	stdout: 'inherit',
	stderr: 'inherit'
});

// Monitor both processes — exit if either dies
async function monitor(name: string, proc: Subprocess) {
	const exitCode = await proc.exited;
	log('error', `${name} exited`, { exit_code: exitCode });
	return { name, exitCode };
}

const result = await Promise.race([monitor('Go', goProc), monitor('SvelteKit', bunProc)]);

log('error', 'Shutting down', { trigger: result.name });
goProc.kill();
bunProc.kill();
process.exit(result.exitCode || 1);
