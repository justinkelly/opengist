import { defineConfig } from 'vite'
import tailwindcss from "@tailwindcss/vite";

/** Firefox rejects `-webkit-text-size-adjust: 100%` (github-markdown, Tailwind preflight). */
function stripTextSizeAdjustPlugin() {
    const strip = (code) => code
        .replace(/-webkit-text-size-adjust:\s*100%;?/g, '')
        .replace(/-ms-text-size-adjust:\s*100%;?/g, '');

    return {
        name: 'strip-text-size-adjust',
        transform(code, id) {
            if (id.endsWith('.css') || id.includes('type=css')) {
                return { code: strip(code), map: null };
            }
        },
        generateBundle(_, bundle) {
            for (const item of Object.values(bundle)) {
                if (item.type !== 'asset' || !item.fileName.endsWith('.css')) continue;
                const src = item.source;
                const text = typeof src === 'string' ? src : Buffer.from(src).toString('utf8');
                const next = strip(text);
                item.source = typeof src === 'string' ? next : Buffer.from(next);
            }
        },
    };
}

export default defineConfig({
    root: './public',
    plugins: [
        tailwindcss(),
        stripTextSizeAdjustPlugin(),
    ],
    server: {
        cors: {
            origin: 'http://localhost:6157',
        },
    },
    build: {
        // generate manifest.json in outDir
        outDir: '',
        assetsDir: 'assets',
        manifest: true,
        rollupOptions: {
            input: [
                './public/ts/admin.ts',
                './public/ts/auto.ts',
                './public/ts/dark.ts',
                './public/ts/editor.ts',
                './public/ts/embed.ts',
                './public/ts/gist.ts',
                './public/ts/light.ts',
                './public/ts/main.ts',
                './public/ts/style_preferences.ts',
                './public/ts/webauthn.ts',
            ]
        },
        assetsInlineLimit: 0,
    }
})
