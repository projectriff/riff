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

    const Buildpacks= () => (
      <Block background="light">
        {[
          {
            content:
              'Buildpacks combine functions with invokers producing runnable containers.',
            image: `${baseUrl}img/build.png`,
            imageAlign: 'right',
            title: 'Cloud Native Buildpacks',
          },
        ]}
      </Block>
    );
    
    const Knative = () => (
      <Block id="try">
        {[
          {
            content:
              '[Knative serving](https://github.com/knative/serving) runs container workloads with multiple revisions and 0-to-N autoscaling.',
            image: `${baseUrl}img/knative.png`,
            imageAlign: 'left',
            title: 'Knative',
          },
        ]}
      </Block>
    );

    return (
      <div>
        <TitleBlock />
        <Buildpacks />
        <Knative />
      </div>
    );
  }
}

module.exports = Index;
