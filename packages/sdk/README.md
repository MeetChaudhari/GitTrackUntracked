# `gittrackuntracked-sdk` (preview)

The SDK is a thin Promise-based wrapper around a separately installed `gitu`
binary. It does not bypass the CLI's safeguards or manage credentials.

```js
import { GituClient } from 'gittrackuntracked-sdk';

const gitu = new GituClient({ cwd: '/path/to/project' });
await gitu.status();
await gitu.sync('Save local notes');
```

It is source-only while the project remains experimental.
