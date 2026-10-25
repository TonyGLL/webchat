import nextVitals from 'eslint-config-next/core-web-vitals';
import nextTs from 'eslint-config-next/typescript';

const config = [
  ...nextVitals,
  ...nextTs,
  {
    rules: {
      // Fetching data inside effects is the pattern used by the hooks in this app.
      'react-hooks/set-state-in-effect': 'warn',
      // No raw HTML sinks (see src/security.test.ts); the constant theme script is the one exception.
      'react/no-danger': 'error',
      'react/jsx-no-target-blank': 'error',
    },
  },
  { ignores: ['.next/**', 'node_modules/**', 'next-env.d.ts'] },
];

export default config;
