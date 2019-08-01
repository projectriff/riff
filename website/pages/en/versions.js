const React = require('react');

const CompLibrary = require('../../core/CompLibrary');

const Container = CompLibrary.Container;

function Versions(props) {
  const {config: siteConfig} = props;
  const repoUrl = `https://github.com/${siteConfig.organizationName}/${
    siteConfig.projectName
  }`;
  return (
    <div className="docMainWrapper wrapper">
      <Container className="mainContainer versionsContainer">
        <div className="post">
          <header className="postHeader">
            <h1>Versions</h1>
          </header>
          <ul>
            {siteConfig.versions.map(
              (version, index) => (
                <li key={index}>
                  <a href={`${siteConfig.baseUrl}${version.url}`}>
                    {version.name}
                  </a>
                </li>
              ),
            )}
          </ul>
          <p>
            See <a href={repoUrl}>GitHub</a> for earlier versions.
          </p>
        </div>
      </Container>
    </div>
  );
}

module.exports = Versions;