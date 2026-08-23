import { execFile } from 'node:child_process';

/** A small Promise-based adapter for a separately installed `gitu` binary. */
export class GituClient {
  constructor({ binary = process.env.GITU_BINARY || 'gitu', cwd = process.cwd() } = {}) {
    this.binary = binary;
    this.cwd = cwd;
  }

  run(args) {
    return new Promise((resolve, reject) => {
      execFile(this.binary, args, { cwd: this.cwd }, (error, stdout, stderr) => {
        if (error) {
          error.stdout = stdout;
          error.stderr = stderr;
          reject(error);
          return;
        }
        resolve({ stdout, stderr });
      });
    });
  }

  init(project) { return this.run(project ? ['init', '--project', project] : ['init']); }
  add(...paths) { return this.run(['add', ...paths]); }
  status() { return this.run(['status']); }
  sync(message) { return this.run(message ? ['sync', '-m', message] : ['sync']); }
  restore(paths = [], force = false) { return this.run(['restore', ...(force ? ['--force'] : []), ...paths]); }
}
