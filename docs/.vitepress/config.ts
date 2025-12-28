import { defineConfig } from "vitepress";

export default defineConfig({
  lang: "en-US",
  title: "sparktype",
  description: "Generate static types from OpenAPI specifications",
  base: "/sparktype/",

  head: [
    [
      "link",
      { rel: "icon", type: "image/svg+xml", href: "/sparktype/logo.svg" },
    ],
    [
      "script",
      {
        defer: "true",
        src: "https://assets.onedollarstats.com/stonks.js",
      },
    ],
  ],

  lastUpdated: false,
  ignoreDeadLinks: true,
  cleanUrls: true,

  themeConfig: {
    logo: "/logo.svg",

    nav: [
      { text: "Guide", link: "/getting-started" },
      { text: "Config", link: "/configuration/" },
      { text: "CLI", link: "/cli/" },
      { text: "Formats", link: "/formats/" },
    ],

    sidebar: {
      "/": [
        {
          items: [
            { text: "Introduction", link: "/" },
            { text: "Getting Started", link: "/getting-started" },
          ],
        },
        {
          text: "Configuration",
          items: [
            { text: "Overview", link: "/configuration/" },
            { text: "Specs", link: "/configuration/specs" },
            { text: "Outputs", link: "/configuration/outputs" },
            { text: "Contents", link: "/configuration/contents" },
            { text: "Options", link: "/configuration/options" },
          ],
        },
        {
          text: "CLI Reference",
          items: [
            { text: "Overview", link: "/cli/" },
            { text: "generate", link: "/cli/generate" },
            { text: "check", link: "/cli/check" },
            { text: "validate", link: "/cli/validate" },
            { text: "init", link: "/cli/init" },
          ],
        },
        {
          text: "Output Formats",
          items: [
            { text: "Overview", link: "/formats/" },
            { text: "TypeScript", link: "/formats/typescript" },
            { text: "Zod", link: "/formats/zod" },
            { text: "Python", link: "/formats/python" },
            { text: "Go", link: "/formats/go" },
          ],
        },
        {
          text: "Guides",
          items: [
            { text: "CI/CD Integration", link: "/guides/ci-cd" },
            { text: "Multiple Specs", link: "/guides/multi-spec" },
            { text: "Remote Specs", link: "/guides/remote-specs" },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: "github", link: "https://github.com/hntrl/sparktype" },
    ],

    footer: {
      message:
        "Released under the <a href='https://github.com/hntrl/sparktype/blob/main/LICENSE'>MIT License</a>.",
      copyright: "Copyright © 2025-present",
    },

    search: {
      provider: "local",
    },

    editLink: {
      pattern: "https://github.com/hntrl/sparktype/edit/main/docs/:path",
      text: "Edit this page on GitHub",
    },
  },
});
