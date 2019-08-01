/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

const React = require('react');

const CompLibrary = require('../../core/CompLibrary.js');

const MarkdownBlock = CompLibrary.MarkdownBlock;
const Container = CompLibrary.Container;
const GridBlock = CompLibrary.GridBlock;

class Index extends React.Component {
  render() {
    const {config: siteConfig, language = ''} = this.props;
    const {baseUrl} = siteConfig;

    const TitleBlock = () => (
      <div className="homeContainer">
        <div className="homeSplashFade">
          <div className="wrapper homeWrapper">
            <div className="inner">
              <h2 className="projectTitle">{siteConfig.title}</h2>
            </div>
          </div>
        </div>
      </div>
    );

    const Block = props => (
      <Container
        padding={['bottom', 'top']}
        id={props.id}
        background={props.background}>
        <GridBlock
          contents={props.children}
          layout={props.layout}
        />
      </Container>
    );

    const About = () => (
      <Block id="try">
        {[
          {
            title: 'What is riff?',
            content: `
riff is an Open Source platform for building and running Functions, Applications, and Containers on [Kubernetes](https://kubernetes.io/). To get started running your own functions on riff, see our [Docs](/docs).

This project is sponsored by [Pivotal](https://pivotala.io)  
_Transforming How The World Builds Software_
`,
            image: `${baseUrl}img/riff-logo.png`,
            imageAlign: 'left'
          },
        ]}
      </Block>
    );

    const Buildpacks= () => (
      <Block background="light">
        {[
          {
            title: 'Buildpacks and Invokers',
            content: `
[Cloud Native Buildpacks]() translate source code into container images.
This release comes with buildpacks for functions using the following invokers:

- [Java](https://github.com/projectriff/java-function-invoker)
- [JavaScript](https://github.com/projectriff/node-function-invoker)
- [Command](https://github.com/projectriff/command-function-invoker)
`,
            image: `${baseUrl}img/cnb.png`,
            imageAlign: 'right',
          },
        ]}
      </Block>
    );
    
    const Knative = () => (
      <Block id="try">
        {[
          {
            title: 'Knative Serving',
            content: `
riff runs containers using [Knative serving](https://github.com/knative/serving).  
This provides support for

- 0-N autoscaling
- Revisions
- HTTP routing using Istio ingress
`,
            image: `${baseUrl}img/knative.png`,
            imageAlign: 'left',
          },
        ]}
      </Block>
    );

    return (
      <div>
        <TitleBlock />
        <About />
        <Buildpacks />
        <Knative />
      </div>
    );
  }
}

module.exports = Index;
