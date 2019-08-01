/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

const React = require('react');

class Footer extends React.Component {
  docUrl(doc, language) {
    const baseUrl = this.props.config.baseUrl;
    const docsUrl = this.props.config.docsUrl;
    const docsPart = `${docsUrl ? `${docsUrl}/` : ''}`;
    const langPart = `${language ? `${language}/` : ''}`;
    return `${baseUrl}${docsPart}${langPart}${doc}`;
  }

  pageUrl(doc, language) {
    const baseUrl = this.props.config.baseUrl;
    return baseUrl + (language ? `${language}/` : '') + doc;
  }

  render() {
    return (
      <footer className="nav-footer" id="footer">
        <section className="sitemap">
          <a href={this.props.config.baseUrl} className="nav-home">
            {this.props.config.footerIcon && (
              <img
                src={this.props.config.baseUrl + this.props.config.footerIcon}
                alt={this.props.config.title}
                width="66"
                height="58"
              />
            )}
          </a>
          <div>
            <h5>Docs</h5>
            <a href={this.docUrl('getting-started', this.props.language)}>
              Getting Started
            </a>
            <a href={this.docUrl('riff', this.props.language)}>
              CLI
            </a>
          </div>
          <div>
            <h5>Community</h5>
            <a href="https://github.com/projectriff/">GitHub</a>
            <a
              href="https://twitter.com/projectriff"
              target="_blank"
              rel="noreferrer noopener">
              Twitter
            </a>
            <a href={`${this.props.config.baseUrl}blog`}>Blog</a>
          </div>
          <div>
            <h5>More</h5>
            <a href="" target="blank">Pivotal Function Service</a>
            <a href="https://pivotal.io/privacy-policy" target="blank">Privacy Policy</a>
            <a href="https://pivotal.io/legal" target="blank">Terms of Use</a>
          </div>
        </section>
        <section className="copyright">
          Copyright © {new Date().getFullYear()} Pivotal Software, Inc. All Rights Reserved.
        </section>
      </footer>
    );
  }
}

module.exports = Footer;
