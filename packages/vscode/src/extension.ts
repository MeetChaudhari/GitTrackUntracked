import * as path from 'node:path';
import { execFile } from 'node:child_process';
import * as vscode from 'vscode';

const output = vscode.window.createOutputChannel('GitTrackUntracked');

type CommandResult = { stdout: string; stderr: string };
type NodeKind = 'section' | 'action' | 'file' | 'message';

class SidebarNode extends vscode.TreeItem {
  constructor(
    public readonly kind: NodeKind,
    label: string,
    public readonly children: SidebarNode[] = [],
    options: { description?: string; tooltip?: string; command?: vscode.Command; icon?: string } = {},
  ) {
    super(label, children.length > 0 ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.None);
    this.description = options.description;
    this.tooltip = options.tooltip ?? label;
    this.command = options.command;
    if (options.icon) this.iconPath = new vscode.ThemeIcon(options.icon);
    if (kind === 'section') this.contextValue = 'section';
  }
}

function binary(): string {
  return vscode.workspace.getConfiguration('gittrackuntracked').get<string>('binaryPath', 'gitu');
}

function workspacePath(): string | undefined {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}

function execute(args: string[], cwd = workspacePath(), reveal = true): Promise<CommandResult> {
  if (reveal) output.show(true);
  output.appendLine(`$ ${binary()} ${args.join(' ')}`);
  return new Promise((resolve, reject) => {
    execFile(binary(), args, { cwd }, (error, stdout, stderr) => {
      if (stdout) output.append(stdout);
      if (stderr) output.append(stderr);
      if (error) {
        const detail = stderr.trim() || stdout.trim() || error.message;
        if (reveal) vscode.window.showErrorMessage(`GitTrackUntracked failed: ${detail}`);
        reject(error);
        return;
      }
      resolve({ stdout, stderr });
    });
  });
}

async function run(args: string[], cwd = workspacePath()): Promise<CommandResult | undefined> {
  try {
    return await execute(args, cwd);
  } catch {
    return undefined;
  }
}

function requireWorkspace(): string | undefined {
  const root = workspacePath();
  if (!root) vscode.window.showWarningMessage('Open a project folder first.');
  return root;
}

class GitTrackUntrackedProvider implements vscode.TreeDataProvider<SidebarNode> {
  private readonly changed = new vscode.EventEmitter<SidebarNode | undefined>();
  readonly onDidChangeTreeData = this.changed.event;
  private files: string[] = [];
  private status = 'Open a Git workspace to get started.';
  private state = 'unknown';

  refresh(): void {
    void this.load();
  }

  getTreeItem(node: SidebarNode): vscode.TreeItem {
    return node;
  }

  getChildren(node?: SidebarNode): SidebarNode[] {
    if (node) return node.children;
    const workspace = workspacePath();
    if (!workspace) {
      return [this.message('Open a project folder to manage local files.', 'folder-opened')];
    }

    const statusIcon = this.state === 'ready' ? 'pass-filled' : this.state === 'attention' ? 'warning' : 'circle-slash';
    const files = this.files.length > 0
      ? this.files.map(file => this.fileNode(workspace, file))
      : [this.message('No files registered yet.', 'file-add')];

    return [
      new SidebarNode('section', 'Vault & Project', [
        this.message(this.status, statusIcon),
      ], { description: this.state === 'ready' ? 'Ready' : 'Needs attention', icon: statusIcon }),
      new SidebarNode('section', 'Registered Local Files', files, { description: `${this.files.length}`, icon: 'files' }),
      new SidebarNode('section', 'Actions', [
        this.action('Initialize personal vault', 'gittrackuntracked.initializeVault', 'cloud-upload'),
        this.action('Initialize this project', 'gittrackuntracked.initializeProject', 'repo'),
        this.action('Add active file', 'gittrackuntracked.addActiveFile', 'add'),
        this.action('Choose files or folders', 'gittrackuntracked.addFiles', 'folder-opened'),
        this.action('Sync selected files', 'gittrackuntracked.sync', 'sync'),
        this.action('Restore from vault', 'gittrackuntracked.restore', 'cloud-download'),
        this.action('Refresh', 'gittrackuntracked.refresh', 'refresh'),
      ], { icon: 'play' }),
    ];
  }

  private async load(): Promise<void> {
    const workspace = workspacePath();
    if (!workspace) {
      this.files = [];
      this.status = 'Open a project folder to get started.';
      this.state = 'unknown';
      this.changed.fire(undefined);
      return;
    }
    try {
      const list = await execute(['list'], workspace, false);
      const status = await execute(['status'], workspace, false);
      this.files = list.stdout.trim() === '' || list.stdout.includes('No paths registered.')
        ? []
        : list.stdout.split(/\r?\n/).map(line => line.trim()).filter(Boolean);
      this.status = status.stdout.trim() || 'All registered paths are up to date.';
      this.state = this.status.includes('up to date') ? 'ready' : 'attention';
    } catch {
      this.files = [];
      this.status = 'Set up the private vault and initialize this project.';
      this.state = 'attention';
    }
    this.changed.fire(undefined);
  }

  private action(label: string, command: string, icon: string): SidebarNode {
    return new SidebarNode('action', label, [], {
      command: { command, title: label },
      icon,
    });
  }

  private message(label: string, icon?: string): SidebarNode {
    return new SidebarNode('message', label, [], { icon });
  }

  private fileNode(workspace: string, relative: string): SidebarNode {
    const uri = vscode.Uri.file(path.join(workspace, relative));
    return new SidebarNode('file', relative, [], {
      tooltip: uri.fsPath,
      command: { command: 'revealInExplorer', title: 'Reveal local path', arguments: [uri] },
      icon: 'file',
    });
  }
}

export function activate(context: vscode.ExtensionContext): void {
  context.subscriptions.push(output);
  const provider = new GitTrackUntrackedProvider();
  context.subscriptions.push(vscode.window.registerTreeDataProvider('gittrackuntracked.sidebar', provider));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.refresh', () => provider.refresh()));
  context.subscriptions.push(vscode.workspace.onDidChangeWorkspaceFolders(() => provider.refresh()));
  context.subscriptions.push(vscode.workspace.onDidSaveTextDocument(document => {
    if (workspacePath() && document.uri.fsPath.startsWith(workspacePath()!)) provider.refresh();
  }));

  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.initializeVault', async () => {
    const remote = await vscode.window.showInputBox({ prompt: 'Private vault Git remote URL', ignoreFocusOut: true });
    if (remote) await run(['vault', 'init', '--remote', remote], undefined);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.initializeProject', async () => {
    if (!requireWorkspace()) return;
    const projectID = await vscode.window.showInputBox({
      prompt: 'Optional stable project ID (required only when this repository has no origin)',
      placeHolder: 'Leave empty to use the Git origin',
      ignoreFocusOut: true,
    });
    if (projectID === undefined) return;
    await run(projectID.trim() ? ['init', '--project', projectID.trim()] : ['init']);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.addActiveFile', async () => {
    const root = requireWorkspace();
    const editor = vscode.window.activeTextEditor;
    if (!root || !editor) return;
    const relative = path.relative(root, editor.document.uri.fsPath);
    if (relative.startsWith('..') || path.isAbsolute(relative)) {
      vscode.window.showWarningMessage('The active file is outside the workspace.');
      return;
    }
    await run(['add', relative]);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.addFiles', async () => {
    const root = requireWorkspace();
    if (!root) return;
    const selected = await vscode.window.showOpenDialog({
      title: 'Select local files or folders to track privately',
      defaultUri: vscode.Uri.file(root),
      canSelectFiles: true,
      canSelectFolders: true,
      canSelectMany: true,
      openLabel: 'Add to private vault',
    });
    if (!selected || selected.length === 0) return;
    const paths = selected.map(uri => path.relative(root, uri.fsPath));
    if (paths.some(relative => relative.startsWith('..') || path.isAbsolute(relative))) {
      vscode.window.showWarningMessage('Every selected path must be inside the workspace.');
      return;
    }
    await run(['add', ...paths]);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.status', async () => {
    if (requireWorkspace()) await run(['status']);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.sync', async () => {
    if (!requireWorkspace()) return;
    const message = await vscode.window.showInputBox({
      prompt: 'Optional private-vault commit message',
      placeHolder: 'Leave empty to use the automatic message',
      ignoreFocusOut: true,
    });
    if (message === undefined) return;
    await run(message.trim() ? ['sync', '-m', message.trim()] : ['sync']);
    provider.refresh();
  }));
  context.subscriptions.push(vscode.commands.registerCommand('gittrackuntracked.restore', async () => {
    if (!requireWorkspace()) return;
    const choice = await vscode.window.showQuickPick([
      { label: 'Restore safely', description: 'Never overwrite existing files', force: false },
      { label: 'Restore and overwrite', description: 'Overwrite existing files from the vault', force: true },
    ], { placeHolder: 'Choose how to restore local files' });
    if (!choice) return;
    await run(choice.force ? ['restore', '--force'] : ['restore']);
    provider.refresh();
  }));
  provider.refresh();
}

export function deactivate(): void {}
