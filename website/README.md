This website was created with [Docusaurus v1.12](https://docusaurus.io/).

### Get Started

1. Install:

```sh
# Install docusaurus with yarn
$ yarn

# or with npm
$ npm install
```

2. Run local dev server:

```sh
# Start the site with yarn
$ yarn start

# or with npm
$ npm start
```

### Directory Structure

```
├── docs
│   ├── ... next (unreleased) docs ...
└── website
    ├── README.md
    ├── blog
    │   ├── ...blog posts...
    ├── core
    │   └── Footer.js (react)
    ├── i18n
    │   └── en.json
    ├── package.json
    ├── pages
    │   └── en
    │       ├── index.js (react)
    │       └── versions.js
    ├── sidebars.json
    ├── siteConfig.js
    ├── static
    │   ├── css
    │   │   └── custom.css
    │   └── img
    └──     ├── ... images ...
```

### docs pages

Docs markdown should include at least the following frontmatter:

```markdown
---
id: page-needs-edit
title: This Doc Needs To Be Edited
---

Content...
```

Refer to a doc page ID in `website/sidebar.json`:

### blog posts

Create the blog post with the format `YYYY-MM-DD-my-blog-post-title.md` in `website/blog`.
Note lowercase, slugified filename (which will be used for the page url).

Add a `title` to the markdown frontmatter.

```markdown
---
title: New Blog Post
---

Lorem Ipsum...
```

### versioning

See https://docusaurus.io/docs/en/versioning