/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Theme colors (use CSS variables for light/dark switching)
        'bg-primary': 'rgb(var(--color-bg-primary) / <alpha-value>)',
        'bg-secondary': 'rgb(var(--color-bg-secondary) / <alpha-value>)',
        'bg-tertiary': 'rgb(var(--color-bg-tertiary) / <alpha-value>)',
        'border-custom': 'rgb(var(--color-border-custom) / <alpha-value>)',
        // Accent colors
        'accent-primary': 'rgb(var(--color-accent-primary) / <alpha-value>)',
        'accent-success': 'rgb(var(--color-accent-success) / <alpha-value>)',
        'accent-warning': 'rgb(var(--color-accent-warning) / <alpha-value>)',
        'accent-error': 'rgb(var(--color-accent-error) / <alpha-value>)',
        // Message colors
        'msg-sent': 'rgb(var(--color-msg-sent) / <alpha-value>)',
        'msg-received': 'rgb(var(--color-msg-received) / <alpha-value>)',
        // Text colors
        'text-primary': 'rgb(var(--color-text-primary) / <alpha-value>)',
        'text-secondary': 'rgb(var(--color-text-secondary) / <alpha-value>)',
        'text-muted': 'rgb(var(--color-text-muted) / <alpha-value>)',
      },
      fontFamily: {
        'mono': ['JetBrains Mono', 'Fira Code', 'Consolas', 'monospace'],
      },
      borderColor: {
        'subtle': 'rgb(var(--color-border-subtle) / <alpha-value>)',
        'muted': 'rgb(var(--color-border-muted) / <alpha-value>)',
      },
    },
  },
  plugins: [],
}
