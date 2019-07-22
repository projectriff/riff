// See https://docusaurus.io/docs/site-config 

const siteConfig = {
  title: 'riff is for functions',
  tagline: '',

  url: 'https://jldec.projectriff.io', // 'https://projectriff.io'
  baseUrl: '/',

  // for docusaurus publish
  // https://docusaurus.io/docs/en/publishing#deploying-to-github-pages
  projectName: 'riff',
  organizationName: 'jldec', // 'projectriff',
  cname: 'jldec.projectriff.io', // projectriff.io

  headerLinks: [
    {doc: 'getting-started-on-gke', label: 'Docs'},
    {doc: 'riff', label: 'CLI'},
    {blog: true, label: 'Blog'},
  ],

  headerIcon: 'img/riff-white.svg',
  footerIcon: 'img/riff-white.svg',
  favicon: 'img/favicon.ico',

  colors: {
    primaryColor: '#52adc8',
    secondaryColor: '#111111',
  },

  // theme for syntax highlighting
  highlight: {
    theme: 'default',
  },

  // on-page navigation
  onPageNav: 'separate',

  // no .html extensions
  cleanUrl: true,

  // open Graph and twitter card images
  ogImage: 'img/riff.svg',
  twitterImage: 'img/riff.svg',
  
  // show all blog posts in sidebar
  blogSidebarCount: 'ALL',

  // other keys
  // repoUrl: 'https://github.com/projectriff/riff',
};

module.exports = siteConfig;
