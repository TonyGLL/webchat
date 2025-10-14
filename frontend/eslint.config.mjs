import next from 'next/eslint';

/** @type {import('eslint').Linter.Config} */
const config = {
  ...next,
  rules: {
    ...next.rules,
    // Add custom rules here
  },
};

export default config;