/**
 * Shared process spawning utilities.
 */

import { elapsed } from "./fmt";

const baseEnv = { ...process.env, CI: "1" };

export interface CollectResult {
	stdout: string;
	stderr: string;
	exitCode: number;
	elapsed: string;
}

/**
 * Spawn a command synchronously with inherited stdio.
 * Exits the parent process if the command fails.
 */
export function run(cmd: string[], options?: { cwd?: string; env?: Record<string, string> }): void {
	const proc = Bun.spawnSync(cmd, {
		stdio: ["ignore", "inherit", "inherit"],
		env: options?.env ? { ...baseEnv, ...options.env } : baseEnv,
		cwd: options?.cwd,
	});
	if (proc.exitCode !== 0) process.exit(proc.exitCode);
}

/**
 * Spawn a command synchronously with captured output.
 */
export function runPiped(cmd: string[], options?: { cwd?: string; env?: Record<string, string> }): {
	exitCode: number;
	stdout: string;
	stderr: string;
} {
	const proc = Bun.spawnSync(cmd, {
		stdout: "pipe",
		stderr: "pipe",
		env: options?.env ? { ...baseEnv, ...options.env } : baseEnv,
		cwd: options?.cwd,
	});
	return {
		exitCode: proc.exitCode,
		stdout: proc.stdout?.toString() ?? "",
		stderr: proc.stderr?.toString() ?? "",
	};
}

/**
 * Spawn a command asynchronously and collect output.
 */
export async function spawnCollect(
	cmd: string[],
	startTime: number,
	options?: { cwd?: string },
): Promise<CollectResult> {
	try {
		const proc = Bun.spawn(cmd, {
			env: { ...baseEnv, FORCE_COLOR: "1" },
			stdout: "pipe",
			stderr: "pipe",
			cwd: options?.cwd,
		});
		const [stdout, stderr] = await Promise.all([
			new Response(proc.stdout).text(),
			new Response(proc.stderr).text(),
		]);
		await proc.exited;
		return {
			stdout,
			stderr,
			exitCode: proc.exitCode ?? 1,
			elapsed: elapsed(startTime),
		};
	} catch (err) {
		return {
			stdout: "",
			stderr: String(err),
			exitCode: 1,
			elapsed: elapsed(startTime),
		};
	}
}

/**
 * Execute promises in parallel, yielding results in completion order.
 */
export async function raceInOrder<T extends { name: string }>(
	promises: Promise<T & CollectResult>[],
	fallbacks: T[],
	onResult: (r: T & CollectResult) => void,
): Promise<void> {
	const tagged = promises.map((p, i) =>
		p
			.then((r) => ({ i, r }))
			.catch((err) => ({
				i,
				r: {
					...fallbacks[i],
					exitCode: 1,
					stdout: "",
					stderr: String(err),
					elapsed: "?",
				} as T & CollectResult,
			})),
	);

	for (let n = 0; n < promises.length; n++) {
		const { i, r } = await Promise.race(tagged);
		tagged[i] = new Promise(() => {});
		onResult(r);
	}
}

/**
 * Managed process group with coordinated lifecycle and cleanup.
 */
export class ProcessGroup {
	private procs: ReturnType<typeof Bun.spawn>[] = [];
	private signalHandlers: { signal: NodeJS.Signals; handler: () => void }[] = [];
	private cleanupFns: (() => void)[] = [];

	constructor() {
		const cleanup = () => {
			for (const p of this.procs) {
				try {
					p.kill("SIGTERM");
				} catch {}
			}
			for (const fn of this.cleanupFns) {
				try {
					fn();
				} catch {}
			}
			this.removeSignalHandlers();
			ProcessGroup.resetTerminal();
			process.exit(130);
		};
		for (const sig of ["SIGINT", "SIGTERM"] as const) {
			process.on(sig, cleanup);
			this.signalHandlers.push({ signal: sig, handler: cleanup });
		}
	}

	onCleanup(fn: () => void): void {
		this.cleanupFns.push(fn);
	}

	private removeSignalHandlers(): void {
		for (const { signal, handler } of this.signalHandlers) {
			process.off(signal, handler);
		}
		this.signalHandlers = [];
	}

	static resetTerminal(): void {
		try {
			process.stdout.write("\x1b[0m\x1b[?25h\x1b[?1049l");
		} catch {}
		try {
			Bun.spawnSync(["stty", "sane"], { stdio: ["inherit", "ignore", "ignore"] });
		} catch {}
	}

	spawn(
		cmd: string[],
		options?: { env?: Record<string, string>; cwd?: string; inheritStdin?: boolean },
	): ReturnType<typeof Bun.spawn> {
		const proc = Bun.spawn(cmd, {
			stdio: [options?.inheritStdin ? "inherit" : "ignore", "inherit", "inherit"],
			env: { ...baseEnv, ...options?.env },
			cwd: options?.cwd,
		});
		this.procs.push(proc);
		return proc;
	}

	async killAll(): Promise<void> {
		for (const fn of this.cleanupFns) {
			try {
				fn();
			} catch {}
		}
		for (const p of this.procs) {
			try {
				p.kill("SIGTERM");
			} catch {}
		}
		const timeout = 5000;
		const exitPromises = this.procs.map((p) => p.exited);
		const timeoutPromise = new Promise<void>((resolve) => setTimeout(resolve, timeout));
		await Promise.race([Promise.all(exitPromises), timeoutPromise]);
		for (const p of this.procs) {
			try {
				p.kill("SIGKILL");
			} catch {}
		}
		this.removeSignalHandlers();
		ProcessGroup.resetTerminal();
	}

	async waitForFirst(): Promise<number> {
		const results = this.procs.map((p, i) => p.exited.then((code) => ({ i, code })));
		const first = await Promise.race(results);
		await this.killAll();
		return first.code;
	}

	async waitForAll(): Promise<number> {
		const codes = await Promise.all(this.procs.map((p) => p.exited));
		this.removeSignalHandlers();
		return Math.max(0, ...codes);
	}
}
