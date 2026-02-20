/**
 * Shared formatting, color, and CLI argument parsing utilities.
 */

const isTTY = process.stdout.isTTY ?? false;

/** Whether stderr is a TTY (useful for progress spinners and status output) */
export const isStderrTTY = process.stderr.isTTY ?? false;

/**
 * ANSI color wrapper - automatically disables colors when not in TTY.
 */
export function c(code: string, text: string): string {
	return isTTY ? `\x1b[${code}m${text}\x1b[0m` : text;
}

/**
 * Format elapsed time since a start timestamp.
 */
export function elapsed(start: number): string {
	return ((Date.now() - start) / 1000).toFixed(1);
}

/**
 * Parse CLI flags from argument array with support for short/long flags.
 */
export function parseFlags<T extends Record<string, "bool" | "string">>(
	argv: string[],
	spec: T,
	shortMap: Record<string, keyof T>,
	defaults: { [K in keyof T]: T[K] extends "bool" ? boolean : string },
): { flags: typeof defaults; passthrough: string[] } {
	const flags = { ...defaults };
	const passthrough: string[] = [];
	let i = 0;

	while (i < argv.length) {
		const arg = argv[i];

		if (arg === "--") {
			passthrough.push(...argv.slice(i + 1));
			break;
		}

		if (arg.startsWith("--")) {
			const name = arg.slice(2);
			if (!(name in spec)) {
				console.error(`Unknown flag: ${arg}`);
				process.exit(1);
			}
			if (spec[name] === "string") {
				i++;
				if (i >= argv.length || argv[i].startsWith("-")) {
					console.error(`Flag ${arg} requires a value`);
					process.exit(1);
				}
				(flags as Record<string, unknown>)[name] = argv[i];
			} else {
				(flags as Record<string, unknown>)[name] = true;
			}
		} else if (arg.startsWith("-") && arg.length > 1) {
			const chars = arg.slice(1);
			for (let j = 0; j < chars.length; j++) {
				const ch = chars[j];
				const mapped = shortMap[ch];
				if (!mapped) {
					console.error(`Unknown flag: -${ch}`);
					process.exit(1);
				}
				if (spec[mapped as string] === "string") {
					i++;
					if (i >= argv.length || argv[i].startsWith("-")) {
						console.error(`Flag -${ch} requires a value`);
						process.exit(1);
					}
					(flags as Record<string, unknown>)[mapped as string] = argv[i];
				} else {
					(flags as Record<string, unknown>)[mapped as string] = true;
				}
			}
		} else {
			passthrough.push(arg);
		}

		i++;
	}

	return { flags, passthrough };
}
