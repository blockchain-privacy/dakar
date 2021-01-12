// webpack is part of vue-cli, DO NOT add it as dependency. It is going to result in several errors.
// eslint-disable-next-line import/no-extraneous-dependencies
const webpack = require('webpack');
const gitCommitHash = require('child_process')
  .execSync('git rev-parse --short HEAD')
  .toString();

const gitBranch = require('child_process')
  .execSync('git branch --show-current')
  .toString();

module.exports = {
  devServer: {
    proxy: 'http://localhost:8081',
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST',
      'Access-Control-Allow-Headers': 'Origin, Content-Type, X-Auth-Token',
    },
  },

  transpileDependencies: [
    'vuetify',
  ],
  runtimeCompiler: true,
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        __COMMIT_HASH__: JSON.stringify(gitCommitHash.trim()),
      }),
      new webpack.DefinePlugin({
        __BRANCH__: JSON.stringify(gitBranch.trim()),
      }),
    ],
  },
};
